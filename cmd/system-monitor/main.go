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
		slog.Error("failed to load config",
			slog.Any("error", err))
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	application, err := app.New(*cfg)
	if err != nil {
		slog.Error("failed to setup application",
			slog.Any("error", err))
		panic(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	if err := application.Start(ctx); err != nil {
		slog.Error("application failed to start",
			slog.Any("error", err))
		panic(err)
	}

	slog.Info("application started", slog.Any("version", info.Version))

	if err := application.Wait(ctx); err != nil {
		slog.Error("application failed",
			slog.Any("error", err))
		panic(err)
	}

	slog.Info("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := application.Shutdown(shutdownCtx); err != nil {

		slog.Error("application shutdown failed", slog.Any("error", err))
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
