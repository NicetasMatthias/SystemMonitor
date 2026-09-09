package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/logger"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

type Application struct {
	collector *collector.Collector
	server    *server.Server
}

func New(cfg config.Config) (*Application, error) {

	coll, err := collector.New(cfg.Collector)

	if err != nil {
		slog.Error("failed to setup collector", slog.Any("error", err))
		return nil, err
	}

	srv, err := server.New(coll, logger.Logs(), cfg.Server)

	if err != nil {
		slog.Error("failed to setup server", slog.Any("error", err))
		return nil, err
	}

	return &Application{
		collector: coll,
		server:    srv,
	}, nil
}

func (app *Application) Start(ctx context.Context) error {

	if err := app.collector.Start(ctx); err != nil {
		slog.Error("failed to start collector", slog.Any("error", err))
		return err
	}

	if err := app.server.Start(); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if stopErr := app.collector.Stop(shutdownCtx); stopErr != nil {
			slog.Error(
				"failed to rollback collector after server start failure",
				slog.Any("error", stopErr),
			)
		}

		return err
	}

	return nil
}

func (app *Application) Wait(ctx context.Context) error {
	select {
	case err := <-app.server.Errors():
		return err
	case <-ctx.Done():
		return nil
	}
}

func (app *Application) Shutdown(ctx context.Context) error {
	srvErr := app.server.Shutdown(ctx)
	if srvErr != nil {
		slog.Error("failed to shutdown collector", slog.Any("error", srvErr))
	}
	collErr := app.collector.Stop(ctx)
	if collErr != nil {
		slog.Error("failed to shutdown server", slog.Any("error", collErr))
	}

	return errors.Join(srvErr, collErr)
}
