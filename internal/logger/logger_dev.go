//go:build dev

package logger

import (
	"log/slog"
	"os"
)

func getHandler() (slog.Handler, error) {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	return slog.NewTextHandler(os.Stdout, opts), nil
}
