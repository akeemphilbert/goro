package conf

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestHTTPConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected HTTP
	}{
		{
			name: "default configuration",
			yaml: `
network: tcp
addr: ":8080"
timeout: 30s
read_timeout: 30s
write_timeout: 30s
`,
			expected: HTTP{
				Network:      "tcp",
				Addr:         ":8080",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
			},
		},
		{
			name: "configuration with TLS enabled",
			yaml: `
network: tcp
addr: ":8443"
timeout: 30s
read_timeout: 30s
write_timeout: 30s
shutdown_timeout: 10s
max_header_bytes: 1048576
tls:
  enabled: true
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
`,
			expected: HTTP{
				Network:         "tcp",
				Addr:            ":8443",
				Timeout:         Duration(30 * time.Second),
				ReadTimeout:     Duration(30 * time.Second),
				WriteTimeout:    Duration(30 * time.Second),
				ShutdownTimeout: Duration(10 * time.Second),
				MaxHeaderBytes:  1048576,
				TLS: TLS{
					Enabled:  true,
					CertFile: "/path/to/cert.pem",
					KeyFile:  "/path/to/key.pem",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config HTTP
			err := yaml.Unmarshal([]byte(tt.yaml), &config)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			if config.Network != tt.expected.Network {
				t.Errorf("Network = %v, want %v", config.Network, tt.expected.Network)
			}
			if config.Addr != tt.expected.Addr {
				t.Errorf("Addr = %v, want %v", config.Addr, tt.expected.Addr)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
			}
			if config.ReadTimeout != tt.expected.ReadTimeout {
				t.Errorf("ReadTimeout = %v, want %v", config.ReadTimeout, tt.expected.ReadTimeout)
			}
			if config.WriteTimeout != tt.expected.WriteTimeout {
				t.Errorf("WriteTimeout = %v, want %v", config.WriteTimeout, tt.expected.WriteTimeout)
			}
			if config.ShutdownTimeout != tt.expected.ShutdownTimeout {
				t.Errorf("ShutdownTimeout = %v, want %v", config.ShutdownTimeout, tt.expected.ShutdownTimeout)
			}
			if config.MaxHeaderBytes != tt.expected.MaxHeaderBytes {
				t.Errorf("MaxHeaderBytes = %v, want %v", config.MaxHeaderBytes, tt.expected.MaxHeaderBytes)
			}
			if config.TLS.Enabled != tt.expected.TLS.Enabled {
				t.Errorf("TLS.Enabled = %v, want %v", config.TLS.Enabled, tt.expected.TLS.Enabled)
			}
			if config.TLS.CertFile != tt.expected.TLS.CertFile {
				t.Errorf("TLS.CertFile = %v, want %v", config.TLS.CertFile, tt.expected.TLS.CertFile)
			}
			if config.TLS.KeyFile != tt.expected.TLS.KeyFile {
				t.Errorf("TLS.KeyFile = %v, want %v", config.TLS.KeyFile, tt.expected.TLS.KeyFile)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  HTTP
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: HTTP{
				Network:      "tcp",
				Addr:         ":8080",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
			},
			wantErr: false,
		},
		{
			name: "invalid network",
			config: HTTP{
				Network:      "invalid",
				Addr:         ":8080",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "invalid address",
			config: HTTP{
				Network:      "tcp",
				Addr:         "invalid-address",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
			},
			wantErr: true,
		},
		{
			name: "TLS enabled but missing cert file",
			config: HTTP{
				Network:      "tcp",
				Addr:         ":8443",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
				TLS: TLS{
					Enabled: true,
					KeyFile: "/path/to/key.pem",
				},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled but missing key file",
			config: HTTP{
				Network:      "tcp",
				Addr:         ":8443",
				Timeout:      Duration(30 * time.Second),
				ReadTimeout:  Duration(30 * time.Second),
				WriteTimeout: Duration(30 * time.Second),
				TLS: TLS{
					Enabled:  true,
					CertFile: "/path/to/cert.pem",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTP.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	config := &HTTP{}
	config.SetDefaults()

	expectedDefaults := HTTP{
		Network:         "tcp",
		Addr:            ":8080",
		Timeout:         Duration(30 * time.Second),
		ReadTimeout:     Duration(30 * time.Second),
		WriteTimeout:    Duration(30 * time.Second),
		ShutdownTimeout: Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
		TLS: TLS{
			Enabled: false,
		},
	}

	if config.Network != expectedDefaults.Network {
		t.Errorf("Default Network = %v, want %v", config.Network, expectedDefaults.Network)
	}
	if config.Addr != expectedDefaults.Addr {
		t.Errorf("Default Addr = %v, want %v", config.Addr, expectedDefaults.Addr)
	}
	if config.Timeout != expectedDefaults.Timeout {
		t.Errorf("Default Timeout = %v, want %v", config.Timeout, expectedDefaults.Timeout)
	}
	if config.ReadTimeout != expectedDefaults.ReadTimeout {
		t.Errorf("Default ReadTimeout = %v, want %v", config.ReadTimeout, expectedDefaults.ReadTimeout)
	}
	if config.WriteTimeout != expectedDefaults.WriteTimeout {
		t.Errorf("Default WriteTimeout = %v, want %v", config.WriteTimeout, expectedDefaults.WriteTimeout)
	}
	if config.ShutdownTimeout != expectedDefaults.ShutdownTimeout {
		t.Errorf("Default ShutdownTimeout = %v, want %v", config.ShutdownTimeout, expectedDefaults.ShutdownTimeout)
	}
	if config.MaxHeaderBytes != expectedDefaults.MaxHeaderBytes {
		t.Errorf("Default MaxHeaderBytes = %v, want %v", config.MaxHeaderBytes, expectedDefaults.MaxHeaderBytes)
	}
	if config.TLS.Enabled != expectedDefaults.TLS.Enabled {
		t.Errorf("Default TLS.Enabled = %v, want %v", config.TLS.Enabled, expectedDefaults.TLS.Enabled)
	}
}
