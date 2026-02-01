package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yuanweize/RouteLens/pkg/geoip"
	"github.com/yuanweize/RouteLens/pkg/prober"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

type Service struct {
	db            *storage.DB
	targets       []storage.Target
	pingTicker    *time.Ticker
	speedTicker   *time.Ticker
	refreshTicker *time.Ticker
	stopChan      chan struct{}
	geoProvider   *geoip.Provider
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
	s.targets = targets
}

func (s *Service) Start() {
	s.pingTicker = time.NewTicker(30 * time.Second)
	s.speedTicker = time.NewTicker(30 * time.Minute) // More frequent speed tests
	s.refreshTicker = time.NewTicker(1 * time.Minute)

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
	for {
		select {
		case <-s.pingTicker.C:
			s.runPingTraceCycle()
		case <-s.speedTicker.C:
			s.runSpeedCycle()
		case <-s.refreshTicker.C:
			s.refreshTargets()
		case <-s.stopChan:
			log.Println("Monitor Service Stopped")
			return
		}
	}
}

func (s *Service) runPingTraceCycle() {
	// Ping/Trace is common for almost all modes except maybe pure HTTP?
	// User said "Mode A: ICMP/MTR Only (默认)": 仅监控延迟和丢包。
	// So we keep ICMP/Trace as a baseline.
	for _, target := range s.targets {
		go s.runPingTraceForTarget(target)
	}
}

func (s *Service) runSpeedCycle() {
	for _, target := range s.targets {
		if target.ProbeType == storage.ProbeModeICMP || target.ProbeType == "" {
			continue // No speed test for ICMP only mode
		}

		go s.runSpeedForTarget(target)
	}
}

func (s *Service) runPingTraceForTarget(t storage.Target) {
	// 1. Ping
	pinger := prober.NewICMPPinger(t.Address, 5)
	res, err := pinger.Run()
	if err != nil {
		log.Printf("Ping failed for %s: %v", t.Name, err)
		return
	}

	// 2. Trace
	traceRunner := prober.NewTracerouteRunner(t.Address)
	traceRes, _ := traceRunner.Run()
	traceBytes := s.serializeTrace(traceRes)

	rec := &storage.MonitorRecord{
		Target:     t.Address,
		CreatedAt:  time.Now(),
		LatencyMs:  float64(res.AvgRtt.Milliseconds()),
		PacketLoss: res.LossRate,
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

	switch t.ProbeType {
	case storage.ProbeModeSSH:
		sshCfg, cfgErr := parseSSHConfig(t.ProbeConfig)
		if cfgErr != nil {
			log.Printf("Invalid SSH config for %s: %v", t.Name, cfgErr)
			return
		}
		sshCfg.Host = t.Address
		runner := prober.NewSSHSpeedTester(sshCfg)
		speedRes, err = runner.Run()

	case storage.ProbeModeHTTP:
		url, cfgErr := parseHTTPConfig(t.ProbeConfig)
		if cfgErr != nil {
			log.Printf("Invalid HTTP config for %s: %v", t.Name, cfgErr)
			return
		}
		runner := prober.NewHTTPSpeedTester(url)
		speedRes, err = runner.Run()

	case storage.ProbeModeIPERF:
		port := 5201
		cfgPort, cfgErr := parseIperfConfig(t.ProbeConfig)
		if cfgErr != nil {
			log.Printf("Invalid IPERF config for %s: %v", t.Name, cfgErr)
			return
		}
		if cfgPort != 0 {
			port = cfgPort
		}
		runner := prober.NewIperfProber(t.Address, port)
		speedRes, err = runner.Run()
	}

	if err != nil {
		log.Printf("Speed test failed for %s (%s): %v", t.Name, t.ProbeType, err)
		return
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
	if target == "" {
		for _, t := range s.targets {
			go s.runPingTraceForTarget(t)
			if t.ProbeType != storage.ProbeModeICMP && t.ProbeType != "" {
				go s.runSpeedForTarget(t)
			}
		}
		return
	}

	for _, t := range s.targets {
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
	Hop       int     `json:"hop"`
	IP        string  `json:"ip"`
	LatencyMs float64 `json:"latency_ms"`
	Loss      float64 `json:"loss"`
	City      string  `json:"city,omitempty"`
	Country   string  `json:"country,omitempty"`
	ISP       string  `json:"isp,omitempty"`
	Latitude  float64 `json:"lat,omitempty"`
	Longitude float64 `json:"lon,omitempty"`
}

type tracePayload struct {
	Target string     `json:"target"`
	Hops   []traceHop `json:"hops"`
}

func (s *Service) serializeTrace(res *prober.TraceResult) []byte {
	if res == nil {
		return []byte("[]")
	}

	hops := make([]traceHop, 0, len(res.Hops))
	for _, h := range res.Hops {
		th := traceHop{
			Hop:       h.Hop,
			IP:        h.IP,
			LatencyMs: float64(h.Latency.Milliseconds()),
			Loss:      h.Loss,
		}
		if s.geoProvider != nil && h.IP != "*" {
			if loc, err := s.geoProvider.Lookup(h.IP); err == nil {
				th.City = loc.City
				th.Country = loc.Country
				th.ISP = loc.ISP
				th.Latitude = loc.Latitude
				th.Longitude = loc.Longitude
			}
		}
		hops = append(hops, th)
	}

	payload := tracePayload{Target: res.Target, Hops: hops}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return []byte("[]")
	}
	return bytes
}

func initGeoProvider() *geoip.Provider {
	cityDB := os.Getenv("RS_GEOIP_CITY_DB")
	ispDB := os.Getenv("RS_GEOIP_ISP_DB")
	geoPath := os.Getenv("RS_GEOIP_PATH")
	if geoPath != "" {
		if cityDB == "" {
			cityDB = filepath.Join(geoPath, "GeoLite2-City.mmdb")
		}
		if ispDB == "" {
			ispDB = filepath.Join(geoPath, "GeoLite2-ISP.mmdb")
		}
	}
	if cityDB == "" && ispDB == "" {
		return nil
	}
	provider, err := geoip.NewProvider(cityDB, ispDB)
	if err != nil {
		log.Printf("GeoIP disabled: %v", err)
		return nil
	}
	return provider
}
