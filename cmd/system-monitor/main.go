package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

func main() {
	cfg, err := config.Load("config.json")

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var targets []collector.NetworkTarget

	for _, t := range cfg.NetworkTargets {
		timeout := t.Timeout.Duration
		if timeout == 0 {
			timeout = 5 * time.Second
		}

		targets = append(targets, collector.NetworkTarget{
			Name:     t.Name,
			Address:  t.Address,
			Interval: t.Interval.Duration,
			Timeout:  timeout,
		})
	}

	col := collector.New(cfg.MaxHistorySize, targets)
	col.Start(cfg.CollectInterval.Duration)

	srv := server.New(col)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", cfg.HTTPPort)
		if err := srv.Start(cfg.HTTPPort); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")

	col.Stop()

	time.Sleep(1 * time.Second)
	log.Println("Shutdown complete")
}
