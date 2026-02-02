package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yuanweize/RouteLens/pkg/geoip"
	"github.com/yuanweize/RouteLens/pkg/logging"
	"github.com/yuanweize/RouteLens/pkg/prober"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

type Service struct {
	db              *storage.DB
	targets         []storage.Target
	targetsMu       sync.RWMutex // Protects targets slice
	pingTicker      *time.Ticker
	speedTicker     *time.Ticker
	refreshTicker   *time.Ticker
	heartbeatTicker *time.Ticker
	stopChan        chan struct{}
	geoProvider     *geoip.Provider
}

func NewService(db *storage.DB) *Service {
	geoProvider := initGeoProvider()
	s := &Service{
		db:          db,
		stopChan:    make(chan struct{}),
		geoProvider: geoProvider,
	}
	s.refreshTargets() // Initial load
	return s
}

func (s *Service) refreshTargets() {
	targets, err := s.db.GetTargets(true)
	if err != nil {
		log.Printf("Failed to refresh targets: %v", err)
		return
	}
	s.targetsMu.Lock()
	s.targets = targets
	s.targetsMu.Unlock()
}

func (s *Service) Start() {
	s.pingTicker = time.NewTicker(30 * time.Second)
	s.speedTicker = time.NewTicker(5 * time.Minute) // Speed tests every 5 minutes
	s.refreshTicker = time.NewTicker(1 * time.Minute)
	s.heartbeatTicker = time.NewTicker(60 * time.Second) // Heartbeat every 60s

	// Run initial cycles immediately on startup
	go func() {
		time.Sleep(2 * time.Second) // Wait for service to fully initialize
		logging.Info("monitor", "Running initial probe cycle on startup...")
		s.runPingTraceCycle()
		s.runSpeedCycle()
	}()

	go s.runLoop()
}

func (s *Service) Stop() {
	if s.geoProvider != nil {
		s.geoProvider.Close()
	}
	close(s.stopChan)
}

func (s *Service) runLoop() {
	log.Println("Monitor Service Started")
	logging.Info("monitor", "Monitor Service Started")
	for {
		select {
		case <-s.pingTicker.C:
			s.runPingTraceCycle()
		case <-s.speedTicker.C:
			s.runSpeedCycle()
		case <-s.refreshTicker.C:
			s.refreshTargets()
		case <-s.heartbeatTicker.C:
			s.targetsMu.RLock()
			targetCount := len(s.targets)
			s.targetsMu.RUnlock()
			logging.Info("monitor", "Heartbeat: System healthy, monitoring %d targets", targetCount)
		case <-s.stopChan:
			log.Println("Monitor Service Stopped")
			logging.Info("monitor", "Monitor Service Stopped")
			return
		}
	}
}

func (s *Service) runPingTraceCycle() {
	// Ping/Trace is common for almost all modes except maybe pure HTTP?
	// User said "Mode A: ICMP/MTR Only (默认)": 仅监控延迟和丢包。
	// So we keep ICMP/Trace as a baseline.
	s.targetsMu.RLock()
	targetsCopy := make([]storage.Target, len(s.targets))
	copy(targetsCopy, s.targets)
	s.targetsMu.RUnlock()

	enabledTargets := 0
	for _, target := range targetsCopy {
		if !target.Enabled {
			continue // Skip disabled targets
		}
		enabledTargets++
		go s.runPingTraceForTarget(target)
	}
	logging.Info("monitor", "Starting ping/trace cycle for %d targets (total: %d)", enabledTargets, len(targetsCopy))
}

func (s *Service) runSpeedCycle() {
	s.targetsMu.RLock()
	targetsCopy := make([]storage.Target, len(s.targets))
	copy(targetsCopy, s.targets)
	s.targetsMu.RUnlock()

	speedTargets := []storage.Target{}
	for _, target := range targetsCopy {
		if !target.Enabled {
			continue // Skip disabled targets
		}
		if target.ProbeType == storage.ProbeModeICMP || target.ProbeType == "" {
			continue // No speed test for ICMP only mode
		}
		speedTargets = append(speedTargets, target)
	}

	if len(speedTargets) == 0 {
		logging.Debug("speedtest", "No speed test targets configured (all targets are ICMP-only)")
		return
	}

	logging.Info("speedtest", "=== Starting speed test cycle for %d targets ===", len(speedTargets))
	for _, target := range speedTargets {
		logging.Info("speedtest", "Queuing speed test: %s (%s) [%s]", target.Name, target.Address, target.ProbeType)
		go s.runSpeedForTarget(target)
	}
}

func (s *Service) runPingTraceForTarget(t storage.Target) {
	logging.Debug("probe", "[MTR] Starting probe for %s (%s)", t.Name, t.Address)

	// 1. Ping (fallback latency)
	pinger := prober.NewICMPPinger(t.Address, 5)
	pingRes, err := pinger.Run()
	if err != nil {
		log.Printf("Ping failed for %s: %v", t.Name, err)
		logging.Error("probe", "[ICMP] Ping failed for %s (%s): %v", t.Name, t.Address, err)
		return
	}
	logging.Info("probe", "[ICMP] Ping OK for %s: latency=%.1fms, loss=%.1f%%", t.Name, float64(pingRes.AvgRtt.Milliseconds()), pingRes.LossRate)

	// 2. MTR (preferred) or Traceroute
	var traceBytes []byte
	latencyMs := float64(pingRes.AvgRtt.Milliseconds())
	packetLoss := pingRes.LossRate

	if mtrRes, mtrErr := prober.NewMTRRunner(t.Address).Run(); mtrErr == nil && mtrRes != nil && len(mtrRes.Hops) > 0 {
		selectedLatency, truncated := selectTargetLatency(mtrRes, latencyMs)
		traceBytes = s.serializeTraceFromMTR(mtrRes, truncated)
		latencyMs = selectedLatency
		packetLoss = selectTargetLoss(mtrRes, packetLoss)
		logging.Info("probe", "[MTR] Trace complete for %s: %d hops, latency=%.1fms", t.Name, len(mtrRes.Hops), latencyMs)
	} else {
		if mtrErr != nil {
			log.Printf("MTR unavailable for %s: %v", t.Name, mtrErr)
			logging.Warn("probe", "[MTR] Fallback to traceroute for %s: %v", t.Name, mtrErr)
		}
		traceRunner := prober.NewTracerouteRunner(t.Address)
		traceRes, _ := traceRunner.Run()
		traceBytes = s.serializeTraceFromTraceroute(traceRes)
	}

	rec := &storage.MonitorRecord{
		Target:     t.Address,
		CreatedAt:  time.Now(),
		LatencyMs:  latencyMs,
		PacketLoss: packetLoss,
		TraceJson:  traceBytes,
		SpeedUp:    0,
		SpeedDown:  0,
	}
	if err := s.db.SaveRecord(rec); err != nil {
		log.Printf("Failed to save record for %s: %v", t.Name, err)
	}
}

func (s *Service) runSpeedForTarget(t storage.Target) {
	var speedRes *prober.SpeedResult
	var err error
	var configErr error

	logging.Info("speedtest", "[%s] >>> Starting speed test for %s (%s)", t.ProbeType, t.Name, t.Address)

	switch t.ProbeType {
	case storage.ProbeModeSSH:
		logging.Info("speedtest", "[SSH] Parsing SSH config for %s...", t.Name)
		sshCfg, cfgErr := parseSSHConfig(t.ProbeConfig)
		if cfgErr != nil {
			configErr = cfgErr
			log.Printf("Invalid SSH config for %s: %v", t.Name, cfgErr)
			logging.Error("speedtest", "[SSH] Invalid config for %s: %v", t.Name, cfgErr)
			s.db.UpdateTargetError(t.Address, fmt.Sprintf("Config error: %v", cfgErr))
			return
		}
		sshCfg.Host = t.Address
		logging.Info("speedtest", "[SSH] Connecting to %s@%s:%d...", sshCfg.User, sshCfg.Host, sshCfg.Port)
		runner := prober.NewSSHSpeedTester(sshCfg)
		speedRes, err = runner.Run()

	case storage.ProbeModeHTTP:
		url, cfgErr := parseHTTPConfig(t.ProbeConfig)
		if cfgErr != nil {
			configErr = cfgErr
			log.Printf("Invalid HTTP config for %s: %v", t.Name, cfgErr)
			s.db.UpdateTargetError(t.Address, fmt.Sprintf("Config error: %v", cfgErr))
			return
		}
		runner := prober.NewHTTPSpeedTester(url)
		speedRes, err = runner.Run()

	case storage.ProbeModeIPERF:
		port := 5201
		cfgPort, cfgErr := parseIperfConfig(t.ProbeConfig)
		if cfgErr != nil {
			configErr = cfgErr
			log.Printf("Invalid IPERF config for %s: %v", t.Name, cfgErr)
			s.db.UpdateTargetError(t.Address, fmt.Sprintf("Config error: %v", cfgErr))
			return
		}
		if cfgPort != 0 {
			port = cfgPort
		}
		runner := prober.NewIperfProber(t.Address, port)
		speedRes, err = runner.Run()
	}

	// Handle probe errors - store them for UI display
	if err != nil {
		errMsg := err.Error()
		// Categorize common SSH errors for better UX
		if strings.Contains(errMsg, "ssh") || strings.Contains(errMsg, "SSH") {
			if strings.Contains(errMsg, "handshake") || strings.Contains(errMsg, "key") {
				errMsg = "SSH: Authentication failed - check credentials/key"
			} else if strings.Contains(errMsg, "connection refused") {
				errMsg = "SSH: Connection refused - check host/port"
			} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "i/o timeout") {
				errMsg = "SSH: Connection timeout - host unreachable"
			}
		}
		log.Printf("Speed test failed for %s (%s): %v", t.Name, t.ProbeType, err)
		logging.Error("speedtest", "Speed test failed for %s (%s): %v", t.Name, t.ProbeType, err)
		s.db.UpdateTargetError(t.Address, errMsg)
		return
	}

	// Clear error on success and log
	if configErr == nil && err == nil {
		s.db.ClearTargetError(t.Address)
		if speedRes != nil {
			logging.Info("speedtest", "Speed test completed for %s: Down=%.1f Mbps, Up=%.1f Mbps", t.Name, speedRes.DownloadSpeed, speedRes.UploadSpeed)
		}
	}

	if speedRes != nil {
		rec := &storage.MonitorRecord{
			Target:     t.Address,
			CreatedAt:  time.Now(),
			LatencyMs:  0,
			PacketLoss: 0,
			SpeedUp:    speedRes.UploadSpeed,
			SpeedDown:  speedRes.DownloadSpeed,
		}
		if err := s.db.SaveRecord(rec); err != nil {
			log.Printf("Failed to save speed record for %s: %v", t.Name, err)
		}
	}
}

func (s *Service) TriggerProbe(target string) {
	s.targetsMu.RLock()
	targetsCopy := make([]storage.Target, len(s.targets))
	copy(targetsCopy, s.targets)
	s.targetsMu.RUnlock()

	if target == "" {
		for _, t := range targetsCopy {
			go s.runPingTraceForTarget(t)
			if t.ProbeType != storage.ProbeModeICMP && t.ProbeType != "" {
				go s.runSpeedForTarget(t)
			}
		}
		return
	}

	for _, t := range targetsCopy {
		if t.Address == target {
			go s.runPingTraceForTarget(t)
			if t.ProbeType != storage.ProbeModeICMP && t.ProbeType != "" {
				go s.runSpeedForTarget(t)
			}
			return
		}
	}
}

type sshProbeConfig struct {
	User      string `json:"user"`
	Password  string `json:"password"`
	KeyPath   string `json:"key_path"`
	KeyText   string `json:"key_text"`
	Port      int    `json:"port"`
	TestBytes int64  `json:"test_bytes"`
}

type httpProbeConfig struct {
	URL string `json:"url"`
}

type iperfProbeConfig struct {
	Port int `json:"port"`
}

func parseSSHConfig(raw string) (prober.SSHConfig, error) {
	if raw == "" {
		return prober.SSHConfig{}, fmt.Errorf("ssh config is required")
	}
	var cfg sshProbeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return prober.SSHConfig{}, err
	}
	sshCfg := prober.SSHConfig{
		User:      cfg.User,
		Password:  cfg.Password,
		KeyPath:   cfg.KeyPath,
		KeyText:   cfg.KeyText,
		Port:      cfg.Port,
		TestBytes: cfg.TestBytes,
	}
	if sshCfg.Port == 0 {
		sshCfg.Port = 22
	}
	if sshCfg.TestBytes == 0 {
		sshCfg.TestBytes = 20 * 1024 * 1024
	}
	return sshCfg, nil
}

func parseHTTPConfig(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("http url is required")
	}
	var cfg httpProbeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return "", err
	}
	return cfg.URL, nil
}

func parseIperfConfig(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	var cfg iperfProbeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return 0, err
	}
	return cfg.Port, nil
}

type traceHop struct {
	Hop            int     `json:"hop"`
	Host           string  `json:"host,omitempty"`
	IP             string  `json:"ip"`
	LatencyLastMs  float64 `json:"latency_last_ms,omitempty"`
	LatencyAvgMs   float64 `json:"latency_avg_ms,omitempty"`
	LatencyBestMs  float64 `json:"latency_best_ms,omitempty"`
	LatencyWorstMs float64 `json:"latency_worst_ms,omitempty"`
	Loss           float64 `json:"loss"`
	ASN            string  `json:"asn,omitempty"`
	// Location fields - primary (zh-CN with fallback to en)
	City    string `json:"city,omitempty"`
	Subdiv  string `json:"subdiv,omitempty"`
	Country string `json:"country,omitempty"`
	// Location fields - English (for i18n support)
	CityEN    string  `json:"city_en,omitempty"`
	SubdivEN  string  `json:"subdiv_en,omitempty"`
	CountryEN string  `json:"country_en,omitempty"`
	ISP       string  `json:"isp,omitempty"`
	Latitude  float64 `json:"lat,omitempty"`
	Longitude float64 `json:"lon,omitempty"`
	// Precision indicates the accuracy of the geo data
	GeoPrecision string `json:"geo_precision,omitempty"`
}

type tracePayload struct {
	Target    string     `json:"target"`
	Hops      []traceHop `json:"hops"`
	Truncated bool       `json:"truncated,omitempty"`
}

func (s *Service) serializeTraceFromTraceroute(res *prober.TraceResult) []byte {
	if res == nil {
		return []byte("[]")
	}

	hops := make([]traceHop, 0, len(res.Hops))
	for _, h := range res.Hops {
		th := traceHop{
			Hop:           h.Hop,
			IP:            h.IP,
			LatencyLastMs: float64(h.Latency.Milliseconds()),
			Loss:          h.Loss,
		}
		s.enrichHopGeo(&th)
		hops = append(hops, th)
	}

	payload := tracePayload{Target: res.Target, Hops: hops}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return []byte("[]")
	}
	return bytes
}

func (s *Service) serializeTraceFromMTR(res *prober.MTRResult, truncated bool) []byte {
	if res == nil {
		return []byte("[]")
	}

	hops := make([]traceHop, 0, len(res.Hops))
	for _, h := range res.Hops {
		ip := resolveIP(h.Host)
		th := traceHop{
			Hop:            h.Hop,
			Host:           h.Host,
			IP:             ip,
			LatencyLastMs:  h.Last,
			LatencyAvgMs:   h.Avg,
			LatencyBestMs:  h.Best,
			LatencyWorstMs: h.Worst,
			Loss:           h.Loss,
			ASN:            h.ASN,
		}
		s.enrichHopGeo(&th)
		hops = append(hops, th)
	}

	payload := tracePayload{Target: res.Target, Hops: hops, Truncated: truncated}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return []byte("[]")
	}
	return bytes
}

func (s *Service) enrichHopGeo(th *traceHop) {
	if s.geoProvider == nil {
		return
	}
	if th.IP == "" || th.IP == "*" {
		return
	}
	if loc, err := s.geoProvider.Lookup(th.IP); err == nil {
		// Primary fields (zh-CN with fallback)
		th.City = loc.City
		th.Subdiv = loc.Subdiv
		th.Country = loc.Country
		// English fields for i18n
		th.CityEN = loc.CityEN
		th.SubdivEN = loc.SubdivEN
		th.CountryEN = loc.CountryEN
		// Other fields
		th.ISP = loc.ISP
		th.Latitude = loc.Latitude
		th.Longitude = loc.Longitude
		th.GeoPrecision = loc.Precision
	}
}

func resolveIP(host string) string {
	if host == "" || host == "*" {
		return host
	}
	if net.ParseIP(host) != nil {
		return host
	}
	if ips, err := net.LookupIP(host); err == nil {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				return v4.String()
			}
		}
		if len(ips) > 0 {
			return ips[0].String()
		}
	}
	return host
}

// resolveGeoIPPaths determines cityDB and ispDB paths from environment variables
func resolveGeoIPPaths() (cityDB, ispDB string) {
	cityDB = os.Getenv("RS_GEOIP_CITY_DB")
	ispDB = os.Getenv("RS_GEOIP_ISP_DB")
	geoPath := os.Getenv("RS_GEOIP_PATH")

	if geoPath == "" && cityDB == "" && ispDB == "" {
		geoPath = filepath.Join("data", "geoip")
	}

	if geoPath == "" {
		return cityDB, ispDB
	}

	// Check if geoPath is a direct .mmdb file or a directory
	if strings.HasSuffix(strings.ToLower(geoPath), ".mmdb") {
		if cityDB == "" {
			cityDB = geoPath
		}
	} else {
		if cityDB == "" {
			cityDB = filepath.Join(geoPath, "GeoLite2-City.mmdb")
		}
		if ispDB == "" {
			ispDB = filepath.Join(geoPath, "GeoLite2-ISP.mmdb")
		}
	}
	return cityDB, ispDB
}

// ensureGeoIPFiles ensures the GeoIP database files exist, downloading if necessary
func ensureGeoIPFiles(cityDB, ispDB string) {
	if cityDB != "" {
		if dlErr := ensureGeoIPDatabase(cityDB); dlErr != nil {
			log.Printf("GeoIP download failed: %v", dlErr)
		}
	}
	if ispDB != "" {
		if _, err := os.Stat(ispDB); err != nil {
			log.Printf("GeoIP ISP DB not found at %s: %v", ispDB, err)
		}
	}
}

func initGeoProvider() *geoip.Provider {
	cityDB, ispDB := resolveGeoIPPaths()

	if cityDB == "" && ispDB == "" {
		log.Println("GeoIP disabled: RS_GEOIP_PATH/RS_GEOIP_CITY_DB not set. Map lines may be empty.")
		log.Println("GeoIP tip: place GeoLite2-City.mmdb under /opt/routelens/geoip and set RS_GEOIP_PATH=/opt/routelens/geoip")
		return nil
	}

	ensureGeoIPFiles(cityDB, ispDB)

	provider, err := geoip.NewProvider(cityDB, ispDB)
	if err != nil {
		log.Printf("GeoIP disabled: %v", err)
		return nil
	}
	log.Printf("GeoIP enabled: city=%s isp=%s", cityDB, ispDB)
	return provider
}

func downloadGeoIP(path string, url string) error {
	if url == "" {
		return fmt.Errorf("geoip download url missing")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("geoip download failed: %s", resp.Status)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func ensureGeoIPDatabase(path string) error {
	const minSizeBytes = 10 * 1024 * 1024
	if info, err := os.Stat(path); err == nil {
		if info.Size() > minSizeBytes {
			return nil
		}
	}
	log.Printf("[GeoIP] Downloading database from P3TERX mirror...")
	return downloadGeoIP(path, "https://raw.githubusercontent.com/P3TERX/GeoLite.mmdb/download/GeoLite2-City.mmdb")
}

func selectTargetLatency(res *prober.MTRResult, fallback float64) (float64, bool) {
	if res == nil || len(res.Hops) == 0 {
		return fallback, false
	}
	last := res.Hops[len(res.Hops)-1]
	if last.Loss < 100 {
		if last.Avg > 0 {
			return last.Avg, false
		}
		if last.Last > 0 {
			return last.Last, false
		}
	}
	for i := len(res.Hops) - 2; i >= 0; i-- {
		h := res.Hops[i]
		if h.Avg > 0 {
			return h.Avg, true
		}
		if h.Last > 0 {
			return h.Last, true
		}
	}
	return fallback, true
}

func selectTargetLoss(res *prober.MTRResult, fallback float64) float64 {
	if res == nil || len(res.Hops) == 0 {
		return fallback
	}
	return res.Hops[len(res.Hops)-1].Loss
}
