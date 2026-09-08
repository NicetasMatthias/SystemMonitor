package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	Cpu     CpuConfig     `json:"cpu"`
	Disk    DiskConfig    `json:"disk"`
	Memory  MemoryConfig  `json:"memory"`
	Network NetworkConfig `json:"network"`
	System  SystemConfig  `json:"system"`
}

type CpuConfig struct {
	Interval       int `json:"interval"`
	MaxHistorySize int `json:"history_size"`
}

type DiskConfig struct {
	Interval       int `json:"interval"`
	MaxHistorySize int `json:"history_size"`
}

type MemoryConfig struct {
	Interval       int `json:"interval"`
	MaxHistorySize int `json:"history_size"`
}

type NetworkConfig struct {
	Targets []NetworkTarget `json:"targets"`
}

type SystemConfig struct {
	Interval int `json:"interval"`
}

type NetworkTarget struct {
	Name     string `json:"name"`
	Address  string `json:"addr"`
	Protocol string `json:"protocol,omitempty"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout,omitempty"`
}

var allowedProtocols = map[string]struct{}{
	"tcp": {},
	"udp": {},
}

func DefaultConfig() Config {
	return Config{
		Cpu: CpuConfig{
			Interval:       5,
			MaxHistorySize: 100,
		},
		Disk: DiskConfig{
			Interval:       10,
			MaxHistorySize: 100,
		},
		Memory: MemoryConfig{
			Interval:       5,
			MaxHistorySize: 100,
		},
		System: SystemConfig{
			Interval: 10,
		},
	}
}

func (c *Config) Validate() error {
	var errs []error

	if c.Cpu.Interval <= 0 {
		errs = append(errs, fmt.Errorf("cpu update interval must be positive, got = %d", c.Cpu.Interval))
	}
	if c.Cpu.MaxHistorySize <= 0 {
		errs = append(errs, fmt.Errorf("cpu max history size must be positive, got = %d", c.Cpu.MaxHistorySize))
	}

	if c.Disk.Interval <= 0 {
		errs = append(errs, fmt.Errorf("disk update interval must be positive, got = %d", c.Disk.Interval))
	}
	if c.Disk.MaxHistorySize <= 0 {
		errs = append(errs, fmt.Errorf("disk max history size must be positive, got = %d", c.Disk.MaxHistorySize))
	}

	if c.Memory.Interval <= 0 {
		errs = append(errs, fmt.Errorf("memory update interval must be positive, got = %d", c.Memory.Interval))
	}
	if c.Memory.MaxHistorySize <= 0 {
		errs = append(errs, fmt.Errorf("memory max history size must be positive, got = %d", c.Memory.MaxHistorySize))
	}

	if c.System.Interval <= 0 {
		errs = append(errs, fmt.Errorf("system update interval must be positive, got = %d", c.System.Interval))
	}

	for index, target := range c.Network.Targets {
		if len(target.Name) == 0 {
			errs = append(errs, fmt.Errorf("network target at %d name must not be empty", index))
		}
		if len(target.Address) == 0 {
			errs = append(errs, fmt.Errorf("network target at %d address must not be empty", index))
		}
		if _, ok := allowedProtocols[strings.ToLower(target.Protocol)]; !ok {
			errs = append(errs, fmt.Errorf("network target at %d address must be valid protocol , got = %s", index, target.Protocol))
		}
		if target.Interval <= 0 {
			errs = append(errs, fmt.Errorf("network target at %d interval must be positive, got = %d", index, target.Interval))
		}
		if target.Timeout <= 0 {
			errs = append(errs, fmt.Errorf("network target at %d timeout must be positive, got = %d", index, target.Timeout))
		}
	}

	return errors.Join(errs...)
}

func (t *NetworkTarget) UnmarshalJSON(data []byte) error {
	type alias NetworkTarget

	*t = NetworkTarget{
		Protocol: "tcp",
		Timeout:  5,
	}

	return json.Unmarshal(data, (*alias)(t))
}
