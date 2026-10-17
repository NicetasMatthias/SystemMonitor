package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		createFile bool
		wantErr    bool
	}{
		{
			name:       "valid JSON",
			content:    `{}`,
			createFile: true,
			wantErr:    false,
		},
		{
			name:       "malformed JSON",
			content:    `{"server":`,
			createFile: true,
			wantErr:    true,
		},
		{
			name:       "empty file",
			content:    "",
			createFile: true,
			wantErr:    true,
		},
		{
			name:       "missing file",
			createFile: false,
			wantErr:    true,
		},
		{
			name:       "JSON array instead of object",
			content:    `[]`,
			createFile: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "json")
			if tt.createFile {
				if err := os.WriteFile(path, []byte(tt.content), 0666); err != nil {
					t.Fatalf(
						"os.WriteFile() error = \"%v\"",
						err,
					)
				}
			}

			cfg, err := Load(path)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Load() error = \"%v\", wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if cfg == nil && err == nil {
				t.Fatal("Load() returned nil config without an error")
			}
		})
	}
}

func TestLoad_InvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "json")

	content := `{
    "server": {
        "port": "70000"
    },
    "collector": {
        "cpu": {
            "interval": 5,
            "history_size": 100
        },
        "disk": {
            "interval": 10,
            "history_size": 100
        },
        "memory": {
            "interval": 5,
            "history_size": 100
        },
        "network": {
            "targets": []
        },
        "system": {
            "interval": 10
        }
    }
	}`

	if err := os.WriteFile(path, []byte(content), 0666); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() succeeded with an invalid configuration")
	}
}
