package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

type multiHandler struct {
	handlers []slog.Handler
	level    slog.Level
}

func Setup(dev bool) {

	once.Do(func() {
		var handlers []slog.Handler

		if dev {
			handlers = append(handlers, slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
			}))
		} else {
			handlers = append(handlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: false,
			}))
		}

		handlers = append(handlers, newStorageHandler())

		defaultLogger = slog.New(&multiHandler{
			handlers: handlers,
		})
		slog.SetDefault(defaultLogger)
	})
}

func Default() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	} else {
		return defaultLogger
	}
}

func (h *multiHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))

	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}

	return &multiHandler{
		handlers: handlers,
		level:    h.level,
	}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))

	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}

	return &multiHandler{
		handlers: handlers,
		level:    h.level,
	}
}

// func (h *multiHandler) Enabled(ctx context.Context, l slog.Level) bool {
// 	return l > h.level
// }

// func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
// 	var errs []error
// 	for _, handler := range h.handlers {
// 		if err := handler.Handle(ctx, r); err != nil {
// 			errs = append(errs, err)
// 		}
// 	}

// 	return errors.Join(errs...)
// }

// func GetDefaultLogger() *slog.Logger {
// 	return defaultLogger
// }

// func Init() error {
// 	var err error
// 	once.Do(func() {
// 		err = initLogger()
// 	})
// 	return err
// }

// func initLogger() error {
// 	handler, err := getHandler()

// 	if err != nil {
// 		return err
// 	}

// 	defaultLogger = slog.New(handler)
// 	slog.SetDefault(defaultLogger)

// 	return nil
// }
