package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/logger"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

func main() {

	err := logger.Init()
	if err != nil {
		slog.Error("Failed to setup logger",
			slog.Any("error", err))
		panic(err)
	}

	cfg, err := config.Load("config.json")

	if err != nil {
		slog.Error("Failed to load config",
			slog.Any("error", err))
		panic(err)
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

	col := collector.New(cfg.MaxHistorySize, targets, cfg.DiskPaths, cfg.DiskCollectInterval.Duration)
	col.Start(cfg.NetworkCollectInterval.Duration)

	srv := server.New(col)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Server staring",
			slog.String("port", cfg.HTTPPort))

		if err := srv.Start(cfg.HTTPPort); err != nil {

			slog.Error("Server statup failed",
				slog.Any("error", err))
		}
	}()

	<-sigChan
	slog.Info("Shutting down gracefully...")

	col.Stop()

	time.Sleep(1 * time.Second)
	slog.Info("Shutdown complete")
}
