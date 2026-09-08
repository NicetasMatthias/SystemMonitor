package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/app"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/info"
	"github.com/NicetasMatthias/SystemMonitor/internal/logger"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	version := flag.Bool("version", false, "print version")

	flag.Parse()

	if *version {
		printVersion()
		return
	}

	run(*configPath)

}

func run(cfgPath string) {

	err := logger.Init()
	if err != nil {
		slog.Error("Failed to setup logger",
			slog.Any("error", err))
		panic(err)
	}

	cfg, err := config.Load(cfgPath)

	if err != nil {
		slog.Error("failed to load config",
			slog.Any("error", err))
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer func() {
		shutdownCancel()
	}()
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
