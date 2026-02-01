package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yuanweize/RouteLens/pkg/prober"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

func main() {
	mode := flag.String("mode", "ping", "Mode: ping, trace, speed")
	target := flag.String("target", "", "Target IP or Hostname")

	// SSH Flags
	sshPort := flag.Int("port", 22, "SSH Port")
	sshUser := flag.String("user", "root", "SSH User")
	sshPass := flag.String("pass", "", "SSH Password")
	sshKey := flag.String("key", "", "SSH Key Path")

	// Database Test Flag
	dbPath := flag.String("db", "test.db", "Database path for db-test mode")

	flag.Parse()

	if *mode == "db-test" {
		runDBTest(*dbPath)
		return
	}

	if *target == "" {
		fmt.Println("Please provide -target")
		os.Exit(1)
	}

	switch *mode {
	case "ping":
		runPing(*target)
	case "trace":
		runTrace(*target)
	case "speed":
		runSpeed(*target, *sshPort, *sshUser, *sshPass, *sshKey)
	default:
		fmt.Println("Unknown mode. Use ping, trace, speed, or db-test")
	}
}

func runDBTest(path string) {
	fmt.Printf("Initializing DB at %s...\n", path)
	db, err := storage.NewDB(path)
	if err != nil {
		log.Fatalf("DB Init failed: %v", err)
	}

	fmt.Println("Querying history...")
	recs, err := db.GetHistory("8.8.8.8", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Found %d records.\n", len(recs))
	for _, r := range recs {
		fmt.Printf("- ID: %d, Time: %s, Latency: %.2fms, TraceJson: %v (Should be nil)\n",
			r.ID, r.CreatedAt.Format(time.RFC3339), r.LatencyMs, r.TraceJson)
	}

	if len(recs) > 0 {
		fmt.Println("Fetching detail for first record...")
		full, err := db.GetRecordDetail(recs[0].ID)
		if err != nil {
			log.Fatalf("Detail fetch failed: %v", err)
		}
		fmt.Printf("Detail trace data: %s\n", string(full.TraceJson))
	}

	// Test pruning
	fmt.Println("Testing pruner (no-op since data is new)...")
	if err := db.PruneOldData(30); err != nil {
		log.Fatalf("Prune failed: %v", err)
	}
	fmt.Println("DB Test Complete.")
}

func runPing(target string) {
	fmt.Printf("Pinging %s...\n", target)
	pinger := prober.NewICMPPinger(target, 4)
	res, err := pinger.Run()
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	fmt.Printf("\n--- %s ping statistics ---\n", target)
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n",
		res.PacketsSent, res.PacketsRecv, res.LossRate)
	fmt.Printf("rtt min/avg/max = %v / %v / %v\n",
		res.MinRtt, res.AvgRtt, res.MaxRtt)
}

func runTrace(target string) {
	fmt.Printf("Tracing route to %s over a maximum of 30 hops...\n", target)

	runner := prober.NewTracerouteRunner(target)
	res, err := runner.Run()
	if err != nil {
		log.Fatalf("Trace failed: %v", err)
	}

	for _, hop := range res.Hops {
		latencyStr := fmt.Sprintf("%v", hop.Latency)
		if hop.IP == "*" {
			latencyStr = "*"
		}
		fmt.Printf("%2d  %s  %s\n", hop.Hop, hop.IP, latencyStr)
	}
}

func runSpeed(host string, port int, user, pass, key string) {
	fmt.Printf("Running SSH Speed Test to %s:%d (User: %s)...\n", host, port, user)

	cfg := prober.SSHConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		KeyPath:  key,
		Timeout:  10 * time.Second,
	}

	tester := prober.NewSSHSpeedTester(cfg)
	res, err := tester.Run()
	if err != nil {
		log.Fatalf("Speed test failed: %v", err)
	}

	fmt.Printf("\n--- Speed Test Results ---\n")
	fmt.Printf("Download: %.2f Mbps\n", res.DownloadSpeed)
	fmt.Printf("Upload:   %.2f Mbps\n", res.UploadSpeed)
}
