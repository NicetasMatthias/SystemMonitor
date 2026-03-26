package config

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

type NetworkTarget struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Protocol string   `json:"protocol,omitempty"`
	Interval Duration `json:"interval"`
	Timeout  Duration `json:"timeout,omitempty"`
}

type Config struct {
	NetworkCollectInterval Duration        `json:"network_collect_interval"`
	DiskCollectInterval    Duration        `json:"disk_collect_interval"`
	HTTPPort               string          `json:"http_port"`
	MaxHistorySize         int             `json:"max_history_size"`
	NetworkTargets         []NetworkTarget `json:"network_targets"`
	DiskPaths              []string        `json:"sidk_paths"`
}

func (c *Config) setDefaults() {
	c.NetworkCollectInterval = Duration{5 * time.Second}
	c.HTTPPort = "8080"
	c.MaxHistorySize = 100
}

func Load(path string) (*Config, error) {
	config := &Config{}

	config.setDefaults()
	defer config.fallbackPath()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("cannot open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	return config, nil
}

func (c *Config) fallbackPath() {
	if len(c.DiskPaths) == 0 {
		if runtime.GOOS == "windows" {
			c.DiskPaths = append(c.DiskPaths, "C:\\\\")
		} else {
			c.DiskPaths = append(c.DiskPaths, "/")
		}
	}
}
