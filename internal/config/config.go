package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
	"github.com/NicetasMatthias/SystemMonitor/internal/server"
)

type Config struct {
	Server    server.Config    `json:"server"`
	Collector collector.Config `json:"collector"`
}

func DefaultConfig() *Config {
	return &Config{
		Server:    server.DefaultConfig(),
		Collector: collector.DefaultConfig(),
	}
}

func (c *Config) Validate() error {
	var errs []error

	if err := c.Server.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid server config: %v", err))
	}

	if err := c.Collector.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid collector config: %v", err))
	}

	return errors.Join(errs...)
}

func Load(path string) (*Config, error) {

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not exist: %v", err)
		} else {
			return nil, fmt.Errorf("cannot open config file: %v", err)
		}
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Failed to close config file", slog.Any("error", err))
		}
	}()

	cfg := DefaultConfig()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config file: %v", err)
	}

	return cfg, nil
}
