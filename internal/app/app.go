package app

import (
	"context"
	"errors"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

type Application struct {
	collector *collector.Collector
	server    *server.Server
}

func New(cfg config.Config) (*Application, error) {

	coll, err := collector.New(cfg)

	if err != nil {
		//=== TODO: log
		return nil, err
	}

	//=== TODO: передавать конфиг тут
	srv, err := server.New(coll, cfg.HTTPPort)

	if err != nil {
		//=== TODO: log
		return nil, err
	}

	return &Application{
		collector: coll,
		server:    srv,
	}, nil
}

func (app *Application) Start(ctx context.Context) error {

	if err := app.collector.Start(ctx); err != nil {
		//=== TODO: log
		return err
	}

	//=== TODO: Тут должен быть контекст

	if err := app.server.Start(); err != nil {
		//=== TODO: log
		return err
	}

	return nil
}

func (app *Application) Shutdown(ctx context.Context) error {

	srvErr := app.server.Shutdown(ctx)
	if srvErr != nil {
		//==== TODO: log
	}
	collErr := app.collector.Stop(ctx)
	if collErr != nil {
		//==== TODO: log
	}

	return errors.Join(srvErr, collErr)
}
