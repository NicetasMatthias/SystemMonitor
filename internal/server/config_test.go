package server

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{
			name:    "valid port",
			port:    "8080",
			wantErr: false,
		},
		{
			name:    "minimum valid port",
			port:    "1",
			wantErr: false,
		},
		{
			name:    "maximum valid port",
			port:    "65535",
			wantErr: false,
		},
		{
			name:    "zero port",
			port:    "0",
			wantErr: true,
		},
		{
			name:    "negative port",
			port:    "-1",
			wantErr: true,
		},
		{
			name:    "port above maximum",
			port:    "65536",
			wantErr: true,
		},
		{
			name:    "non-numeric port",
			port:    "abc",
			wantErr: true,
		},
		{
			name:    "fractional port",
			port:    "8080.5",
			wantErr: true,
		},
		{
			name:    "empty port",
			port:    "",
			wantErr: true,
		},
		{
			name:    "port with whitespace",
			port:    " 8080 ",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Port: tt.port,
			}
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"Config.Validate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
