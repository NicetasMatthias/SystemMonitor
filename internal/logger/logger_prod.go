//go:build !dev

package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// TODO: move to system dirs
// TODO: log rotation
const logFile string = "./log/1.log"

func getHandler() (slog.Handler, error) {
	var logLevel slog.Level
	if _, exists := os.LookupEnv("DEBUG"); exists {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	var writers []io.Writer

	file, err := getFileWriter()
	if err != nil {
		return nil, err
	}
	writers = append(writers, file)

	// TODO: Add web logger

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false,
	}

	multiWriter := io.MultiWriter(writers...)

	return slog.NewJSONHandler(multiWriter, opts), nil
}

func getFileWriter() (io.Writer, error) {
	logDir := filepath.Dir(logFile)
	err := os.MkdirAll(logDir, 0775)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

		return nil, err
	}

	return file, nil
}
