package app

import (
	"context"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/config"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

type Application struct {
	collector *collector.Collector
	server    *server.Server
}

func New(cfg config.Config) (*Application, error) {

	//=== TODO: убрать это внутрь реализации collector.New, чтобы он принимал только конфиг
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
	//=== TODO: Тут должен быть контекст и возврат ошибки
	app.collector.Start(ctx)

	// if err != nil {
	// 	//=== TODO: log
	// 	return nil, err
	// }

	//=== TODO: Тут должен быть контекст
	err := app.server.Start("")

	if err != nil {
		//=== TODO: log
		return err
	}

	return nil

}
