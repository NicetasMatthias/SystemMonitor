package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

var allowedProtocols = map[string]struct{}{
	"tcp": {},
	"udp": {},
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
	DiskPaths              []string        `json:"disk_paths"`
}

func (c Config) Validate() error {

	var errs []error

	if port, err := strconv.Atoi(c.HTTPPort); err != nil {
		errs = append(errs, fmt.Errorf("http_port: failed to get from string. %w", err))
	} else if port < 1 || port > 65535 {
		errs = append(errs, fmt.Errorf("http_port: invalid value: %d, must be in 1..65535", port))
	}

	if c.NetworkCollectInterval.Duration <= 0 {
		errs = append(errs, errors.New("network_collect_interval: invalid value, must be positive"))
	}

	if c.DiskCollectInterval.Duration <= 0 {
		errs = append(errs, errors.New("disk_collect_interval: invalid value, must be positive"))
	}

	if c.MaxHistorySize <= 0 {
		errs = append(errs, errors.New("max_history_size: invalid value, must be positive"))
	}

	for i, netTarget := range c.NetworkTargets {
		if err := netTarget.Validate(); err != nil {

			errs = append(errs, fmt.Errorf("network_targets[%d]:\n%w", i, err))
		}
	}

	//=== FIXME: validate disk paths

	combinedErrors := errors.Join(errs...)

	if combinedErrors == nil {
		return nil
	} else {
		return fmt.Errorf("config validation failed:\n%w", combinedErrors)
	}
}
func (nt NetworkTarget) Validate() error {
	var errs []error

	if len(nt.Name) == 0 {
		errs = append(errs, errors.New("name: invalid value, must not be empty"))
	}

	if len(nt.Address) == 0 {
		errs = append(errs, errors.New("address: invalid value, must not be empty"))
	}

	if _, ok := allowedProtocols[strings.ToLower(nt.Protocol)]; !ok && len(nt.Protocol) > 0 {
		errs = append(errs, fmt.Errorf("protocol: unsupported protocol: %s", nt.Protocol))
	}

	if nt.Interval.Duration <= 0 {
		errs = append(errs, errors.New("interval: invalid value, must be positive"))
	}

	return errors.Join(errs...)
}

func Load(path string) (*Config, error) {

	r := &Config{}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not exist: %w", err)
		} else {
			return nil, fmt.Errorf("cannot open config file: %w", err)
		}
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Failed to close config file", slog.Any("error", err))
		}
	}()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(r); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	return r, nil
}
