package smtp

import (
	"testing"
	"strings"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  &Config{Host: "smtp.test.com", Port: 587},
			wantErr: false,
		},
		{
			name:    "missing host",
			config:  &Config{Port: 587},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name:    "missing port",
			config:  &Config{Host: "smtp.test.com"},
			wantErr: true,
			errMsg:  "port is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigWithAuth(t *testing.T) {
	config := &Config{
		Host:     "smtp.maleon.run",
		Port:     5870,
		Username: "elus54",
		Password: "testpass",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Host: "smtp.maleon.run",
		Port: 5870,
	}

	if config.Host != "smtp.maleon.run" {
		t.Errorf("expected host smtp.maleon.run, got %s", config.Host)
	}
	if config.Port != 5870 {
		t.Errorf("expected port 5870, got %d", config.Port)
	}
}

func BenchmarkConfigValidate(b *testing.B) {
	config := &Config{
		Host: "smtp.test.com",
		Port: 587,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}
