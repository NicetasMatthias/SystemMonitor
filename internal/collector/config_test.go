package collector_test

import (
	"encoding/json"
	"testing"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  collector.Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  collector.DefaultConfig(),
			wantErr: false,
		},
		{
			name: "zero cpu interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Cpu.Interval = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative cpu interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Cpu.Interval = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "zero cpu history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Cpu.MaxHistorySize = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative cpu history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Cpu.MaxHistorySize = -1
				return cfg
			}(),
			wantErr: true,
		},

		{
			name: "zero disk interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Disk.Interval = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative disk interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Disk.Interval = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "zero disk history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Disk.MaxHistorySize = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative disk history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Disk.MaxHistorySize = -1
				return cfg
			}(),
			wantErr: true,
		},

		{
			name: "zero memory interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Memory.Interval = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative memory interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Memory.Interval = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "zero memory history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Memory.MaxHistorySize = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative memory history size",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.Memory.MaxHistorySize = -1
				return cfg
			}(),
			wantErr: true,
		},

		{
			name: "zero system interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.System.Interval = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "negative system interval",
			config: func() collector.Config {
				cfg := collector.DefaultConfig()
				cfg.System.Interval = -1
				return cfg
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Config.Validate() error = \"%v\", wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestConfig_Validate_NetworkTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []collector.NetworkTarget
		wantErr bool
	}{
		{
			name:    "no network targets",
			targets: nil,
			wantErr: false,
		},
		{
			name: "valid http target",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "tcp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: false,
		},
		{
			name: "valid DNS target",
			targets: []collector.NetworkTarget{
				{
					Name:      "dns",
					Address:   "example.com:53",
					Type:      "dns",
					Transport: "udp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: false,
		},
		{
			name: "empty target name",
			targets: []collector.NetworkTarget{
				{
					Name:      "",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "tcp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "empty target address",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "",
					Type:      "http",
					Transport: "tcp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid target type",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "invalid",
					Transport: "tcp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid transport type",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "invalid",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "http with udp transport",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "udp",
					Interval:  10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "negative interval",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "tcp",
					Interval:  -10,
					Timeout:   5,
				},
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			targets: []collector.NetworkTarget{
				{
					Name:      "website",
					Address:   "https://example.com",
					Type:      "http",
					Transport: "tcp",
					Interval:  10,
					Timeout:   -2,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := collector.DefaultConfig()
			cfg.Network.Targets = tt.targets
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"NetworkTarget.Validate() error = \"%v\", wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestNetworkTarget_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid HTTP target",
			input: `{
				"name": "website",
				"address": "https://example.com",
				"type": "http",
				"transport": "tcp"
			}`,
			wantErr: false,
		},
		{
			name: "valid DNS target",
			input: `{
				"name": "dns",
				"address": "example.com:53",
				"type": "dns",
				"transport": "udp"
			}`,
			wantErr: false,
		},
		{
			name:    "empty JSON object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"name":`,
			wantErr: true,
		},
		{
			name:    "JSON array instead of object",
			input:   `[]`,
			wantErr: true,
		},
		{
			name:    "JSON string instead of object",
			input:   `"target"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target collector.NetworkTarget

			err := json.Unmarshal([]byte(tt.input), &target)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"NetworkTarget json.Unmarshal() error = \"%v\", wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestNetworkTarget_UnmarshalJSON_EmptyFields(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		transport string
		timeout   int
	}{
		{
			name: "detect transport udp based on dns",
			input: `{
				"name": "dns",
				"address": "example.com:53",
				"type": "dns",
				"interval": 10,
				"timeout": 5
			}`,
			transport: "udp",
			timeout:   5,
		},
		{
			name: "detect transport tcp based on http",
			input: `{
				"name": "dns",
				"address": "http://example.com",
				"type": "http",
				"interval": 10,
				"timeout": 5
			}`,
			transport: "tcp",
			timeout:   5,
		},
		{
			name: "detect transport tcp based on https",
			input: `{
				"name": "dns",
				"address": "http://example.com",
				"type": "https",
				"interval": 10,
				"timeout": 5
			}`,
			transport: "tcp",
			timeout:   5,
		},
		{
			name: "default timeout",
			input: `{
				"name": "dns",
				"address": "http://example.com",
				"type": "https",
				"transport": "tcp",
				"interval": 10
			}`,
			transport: "tcp",
			timeout:   5,
		},
		{
			name: "no timeout overwrite",
			input: `{
				"name": "dns",
				"address": "http://example.com",
				"type": "https",
				"transport": "tcp",
				"interval": 10,
				"timeout": 3
			}`,
			transport: "tcp",
			timeout:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target collector.NetworkTarget

			err := json.Unmarshal([]byte(tt.input), &target)

			if err != nil {
				t.Fatalf(
					"NetworkTarget json.Unmarshal() error = \"%v\", input:\n%v",
					err,
					tt.input,
				)
			}

			if target.Transport != tt.transport {
				t.Errorf(
					"NetworkTarget transport got: \"%v\"; expected : \"%v\"",
					target.Transport,
					tt.transport,
				)
			}

			if target.Timeout != tt.timeout {
				t.Errorf(
					"NetworkTarget timeout got: \"%v\"; expected : \"%v\"",
					target.Timeout,
					tt.timeout,
				)
			}
		})
	}
}
