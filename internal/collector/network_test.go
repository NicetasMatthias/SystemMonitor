package collector

import "testing"

func TestNewNetworkCheck(t *testing.T) {
	tests := []struct {
		name    string
		target  NetworkTarget
		wantErr bool
	}{
		{
			name: "DNS",
			target: NetworkTarget{
				Type: "dns",
			},
			wantErr: false,
		},
		{
			name: "HTTP",
			target: NetworkTarget{
				Type: "http",
			},
			wantErr: false,
		},
		{
			name: "HTTPS",
			target: NetworkTarget{
				Type: "https",
			},
			wantErr: false,
		},
		{
			name: "unsupported type",
			target: NetworkTarget{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check, err := newNetworkCheck(tt.target)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"newNetworkCheck() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}

			if !tt.wantErr && check == nil {
				t.Fatal("newNetworkCheck() returned nil check without an error")
			}
		})
	}
}
