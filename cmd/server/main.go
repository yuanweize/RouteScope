package main

import (
	"log"

	"github.com/yuanweize/RouteLens/internal/api"
	"github.com/yuanweize/RouteLens/internal/cli"
	"github.com/yuanweize/RouteLens/internal/monitor"
	"github.com/yuanweize/RouteLens/pkg/storage"
	"github.com/yuanweize/RouteLens/web"
)

func main() {
	rootCmd := cli.NewRootCmd(runServer)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServer(port, dbPath string) {
	// Gin requires port like ":8080"
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	// 2. Storage
	log.Printf("Connecting to DB at %s...", dbPath)
	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// Seed default targets if none exist
	seedTargets(db)

	// 3. Monitor Service
	mon := monitor.NewService(db)
	mon.Start()
	defer mon.Stop()

	// 4. API Server
	server := api.NewServer(db, mon, web.DistFS)

	log.Printf("Starting API Server on %s...", port)
	if err := server.Run(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func seedTargets(db *storage.DB) {
	existing, _ := db.GetTargets(false)
	if len(existing) == 0 {
		log.Println("Seeding default targets...")
		db.SaveTarget(&storage.Target{Name: "Home NAS", Address: "nas.yuanweize.win", Desc: "NAS node in China", ProbeType: storage.ProbeModeICMP})
		db.SaveTarget(&storage.Target{Name: "Europe VPS", Address: "nue.eurun.top", Desc: "NUE node in Europe", ProbeType: storage.ProbeModeICMP})
	}
}
