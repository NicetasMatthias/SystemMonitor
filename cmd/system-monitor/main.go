package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/app"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/info"
	"github.com/NicetasMatthias/SystemMonitor/internal/logger"
)

func main() {
	if len(os.Args) < 2 {
		run()
	} else if os.Args[1] == "version" || os.Args[1] == "--version" {
		printVersion()
	} else {
		printHelp()
	}

}

func run() {

	err := logger.Init()
	if err != nil {
		slog.Error("Failed to setup logger",
			slog.Any("error", err))
		panic(err)
	}

	cfg, err := config.Load("config.json") //=== TODO: set proper config path

	if err != nil {
		slog.Error("Failed to load config",
			slog.Any("error", err))
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	application, err := app.New(*cfg)
	if err != nil {
		// === TODO: log
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	if err := application.Start(ctx); err != nil {
		slog.Error("system-monitor failed",
			slog.Any("error", err))
		panic(err)
	}

	slog.Info("system-monitor started", slog.Any("version", info.Version))

	<-sigChan

	slog.Info("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := application.Shutdown(shutdownCtx); err != nil {

		slog.Error("Server shutdown failed", slog.Any("error", err))
		panic(err)
	} else {
		slog.Info("Shutdown complete")
	}

}

func printVersion() {
	fmt.Println(info.Printable())
}

func printHelp() {
	fmt.Println("wrong args") //=== TODO: write help
}
