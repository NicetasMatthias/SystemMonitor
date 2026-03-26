package logger

import (
	"log/slog"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

func Init() error {
	var err error
	once.Do(func() {
		err = initLogger()
	})
	return err
}

func initLogger() error {
	handler, err := getHandler()

	if err != nil {
		return err
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	return nil
}

func GetDefaultLogger() *slog.Logger {
	return defaultLogger
}
