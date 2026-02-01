package monitor

import (
	"log"
	"strconv"
	"strings"
	"time"

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
}

func NewService(db *storage.DB) *Service {
	s := &Service{
		db:       db,
		stopChan: make(chan struct{}),
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
		go func(t storage.Target) {
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

			// Serialize Trace (Mock for now or minimal JSON)
			_ = traceRes
			traceBytes := []byte("[]")

			rec := &storage.MonitorRecord{
				Target:     t.Address,
				CreatedAt:  time.Now(),
				LatencyMs:  float64(res.AvgRtt.Milliseconds()),
				PacketLoss: res.LossRate,
				TraceJson:  traceBytes,
			}
			s.db.SaveRecord(rec)
		}(target)
	}
}

func (s *Service) runSpeedCycle() {
	for _, target := range s.targets {
		if target.ProbeMode == "ICMP" || target.ProbeMode == "" {
			continue // No speed test for ICMP only mode
		}

		go func(t storage.Target) {
			var speedRes *prober.SpeedResult
			var err error

			switch t.ProbeMode {
			case "SSH":
				// Parse config (expects user:keypath)
				parts := strings.Split(t.ProbeConfig, ":")
				if len(parts) < 2 {
					log.Printf("Invalid SSH config for %s", t.Name)
					return
				}
				sshCfg := prober.SSHConfig{
					User:      parts[0],
					KeyPath:   parts[1],
					Host:      t.Address,
					Port:      22,
					TestBytes: 20 * 1024 * 1024,
				}
				runner := prober.NewSSHSpeedTester(sshCfg)
				speedRes, err = runner.Run()

			case "HTTP":
				// ProbeConfig is the URL
				runner := prober.NewHTTPSpeedTester(t.ProbeConfig)
				speedRes, err = runner.Run()

			case "IPERF3":
				port := 5201
				if t.ProbeConfig != "" {
					if p, err := strconv.Atoi(t.ProbeConfig); err == nil {
						port = p
					}
				}
				runner := prober.NewIperfProber(t.Address, port)
				speedRes, err = runner.Run()
			}

			if err != nil {
				log.Printf("Speed test failed for %s (%s): %v", t.Name, t.ProbeMode, err)
				return
			}

			if speedRes != nil {
				rec := &storage.MonitorRecord{
					Target:    t.Address,
					CreatedAt: time.Now(),
					SpeedUp:   speedRes.UploadSpeed,
					SpeedDown: speedRes.DownloadSpeed,
				}
				s.db.SaveRecord(rec)
			}
		}(target)
	}
}
