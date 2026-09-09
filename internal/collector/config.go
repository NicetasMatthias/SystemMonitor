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
	Name      string `json:"name"`
	Address   string `json:"addr"`
	Type      string `json:"type"`
	Transport string `json:"transport,omitempty"`
	Interval  int    `json:"interval"`
	Timeout   int    `json:"timeout,omitempty"`
}

var allowedTypes = map[string]struct{}{
	"dns":   {},
	"http":  {},
	"https": {},
}

var allowedTransports = map[string]struct{}{
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

		if err := target.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid network target at %d: %v", index, err))
		}
	}

	return errors.Join(errs...)
}

func (t *NetworkTarget) Validate() error {
	var errs []error

	if len(t.Name) == 0 {
		errs = append(errs, fmt.Errorf("network target name must not be empty"))
	}
	if len(t.Address) == 0 {
		errs = append(errs, fmt.Errorf("network target address must not be empty"))
	}
	if _, ok := allowedTypes[strings.ToLower(t.Type)]; !ok {
		errs = append(errs, fmt.Errorf("network target has unsupported type %q", t.Type))
	}
	if _, ok := allowedTransports[strings.ToLower(t.Transport)]; !ok {
		errs = append(errs, fmt.Errorf("network target has unsupported transport %q", t.Type))
	}
	if t.Interval <= 0 {
		errs = append(errs, fmt.Errorf("network target interval must be positive, got = %d", t.Interval))
	}
	if t.Timeout <= 0 {
		errs = append(errs, fmt.Errorf("network target timeout must be positive, got = %d", t.Timeout))
	}

	switch t.Type {
	case "dns":
		// DNS supports both UDP and TCP.

	case "http", "https":
		if t.Transport != "tcp" {
			errs = append(errs, fmt.Errorf("network target of type %q requires tcp transport, got = %q", t.Type, t.Transport))
		}
	}

	return errors.Join(errs...)
}

func (t *NetworkTarget) UnmarshalJSON(data []byte) error {
	type alias NetworkTarget

	var raw alias

	raw.Timeout = 5

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	raw.Transport = strings.ToLower(raw.Transport)
	raw.Type = strings.ToLower(raw.Type)

	if raw.Transport == "" {
		switch raw.Type {
		case "dns":
			raw.Transport = "udp"
		case "http", "https":
			raw.Transport = "tcp"
		}
	}

	*t = NetworkTarget(raw)
	return nil
}
