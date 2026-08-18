package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "8080",
				MaxHistorySize:         100,
				NetworkTargets: []NetworkTarget{
					{
						Name:     "test",
						Address:  "127.0.0.1:8080",
						Protocol: "tcp",
						Interval: Duration{5 * time.Second},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid http port - string",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "invalid",
				MaxHistorySize:         100,
			},
			wantErr: true,
		},
		{
			name: "invalid http port - too low",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "0",
				MaxHistorySize:         100,
			},
			wantErr: true,
		},
		{
			name: "invalid http port - too high",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "70000",
				MaxHistorySize:         100,
			},
			wantErr: true,
		},
		{
			name: "invalid network interval - zero",
			config: Config{
				NetworkCollectInterval: Duration{0},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "8080",
				MaxHistorySize:         100,
			},
			wantErr: true,
		},
		{
			name: "invalid disk interval - zero",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{0},
				HTTPPort:               "8080",
				MaxHistorySize:         100,
			},
			wantErr: true,
		},
		{
			name: "invalid max history size - zero",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "8080",
				MaxHistorySize:         0,
			},
			wantErr: true,
		},
		{
			name: "invalid network target",
			config: Config{
				NetworkCollectInterval: Duration{5 * time.Second},
				DiskCollectInterval:    Duration{10 * time.Second},
				HTTPPort:               "8080",
				MaxHistorySize:         100,
				NetworkTargets: []NetworkTarget{
					{
						Name:     "",
						Address:  "127.0.0.1:8080",
						Protocol: "tcp",
						Interval: Duration{5 * time.Second},
					},
				},
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "valid duration - seconds",
			input:    []byte(`"5s"`),
			expected: 5 * time.Second,
			wantErr:  false,
		},
		{
			name:     "valid duration - minutes",
			input:    []byte(`"1m"`),
			expected: 1 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "valid duration - milliseconds",
			input:    []byte(`"100ms"`),
			expected: 100 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "invalid duration format",
			input:    []byte(`"invalid"`),
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    []byte(`""`),
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid json",
			input:    []byte(`"`),
			expected: 0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var d Duration
			if err := d.UnmarshalJSON(tt.input); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if d.Duration != tt.expected {
				t.Errorf("UnmarshalJSON() expected %v, received %v", tt.expected, d.Duration)
			}

		})
	}
}

func TestNetworkTarget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		target  NetworkTarget
		wantErr bool
	}{
		{
			name: "valid target",
			target: NetworkTarget{
				Name:     "test",
				Address:  "127.0.0.1:8080",
				Protocol: "tcp",
				Interval: Duration{5 * time.Second},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			target: NetworkTarget{
				Name:     "",
				Address:  "127.0.0.1:8080",
				Protocol: "tcp",
				Interval: Duration{5 * time.Second},
			},
			wantErr: true,
		},
		{
			name: "empty address",
			target: NetworkTarget{
				Name:     "test",
				Address:  "",
				Protocol: "tcp",
				Interval: Duration{5 * time.Second},
			},
			wantErr: true,
		},
		{
			name: "unsupported protocol",
			target: NetworkTarget{
				Name:     "test",
				Address:  "127.0.0.1:8080",
				Protocol: "http",
				Interval: Duration{5 * time.Second},
			},
			wantErr: true,
		},
		{
			name: "zero interval",
			target: NetworkTarget{
				Name:     "test",
				Address:  "127.0.0.1:8080",
				Protocol: "tcp",
				Interval: Duration{0},
			},
			wantErr: true,
		},
		{
			name: "negative interval",
			target: NetworkTarget{
				Name:     "test",
				Address:  "127.0.0.1:8080",
				Protocol: "tcp",
				Interval: Duration{-5 * time.Second},
			},
			wantErr: true,
		},
		{
			name: "protocol case insensitivity",
			target: NetworkTarget{
				Name:     "test",
				Address:  "127.0.0.1:8080",
				Protocol: "TCP",
				Interval: Duration{5 * time.Second},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.target.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("NetworkTarget.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()

	write := func(filepath, content string) error {
		return os.WriteFile(filepath, []byte(content), 0644)
	}
	//=== TODO: improve tests
	tests := []struct {
		name       string
		path       string
		configText string
		configNil  bool
		wantErr    bool
		prepare    func(string, string) error
	}{
		{
			name: "valid config file",
			path: filepath.Join(tempDir, "config.json"),
			configText: `{
    "network_collect_interval": "5s",
    "disk_collect_interval": "20s",
    "http_port": "32245",
    "max_history_size": 100,
    "network_targets": [
        {
            "name": "Google DNS",
            "address": "8.8.8.8:53",
            "protocol": "udp",
            "interval": "10s",
            "timeout": "5s"
        },
        {
            "name": "Local Router",
            "address": "192.168.0.1:80",
            "interval": "5s"
        }
    ],
    "diskPaths": [
        "/"
    ]
}`,
			configNil: false,
			wantErr:   false,
			prepare:   write,
		},

		{
			name:       "non-existent config file",
			path:       filepath.Join(tempDir, "non_config.json"),
			configText: "",
			configNil:  true,
			wantErr:    true,
			prepare:    func(_, _ string) error { return nil },
		},

		{
			name:       "invalid json",
			path:       filepath.Join(tempDir, "config.json"),
			configText: "{{\"}}",
			configNil:  true,
			wantErr:    true,
			prepare:    write,
		},

		//=== TODO: add test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.prepare(tt.path, tt.configText); err != nil {
				t.Errorf("Load() internal test error during preparation %v", err)
				return
			}
			_, err := Load(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
