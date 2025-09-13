package conf

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Bootstrap is the configuration structure for the application
type Bootstrap struct {
	Server *Server `yaml:"server"`
}

// Server holds the server configuration
type Server struct {
	HTTP *HTTP `yaml:"http"`
	GRPC *GRPC `yaml:"grpc"`
}

// HTTP holds the HTTP server configuration
type HTTP struct {
	Network         string   `yaml:"network"`
	Addr            string   `yaml:"addr"`
	Timeout         Duration `yaml:"timeout"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes  int      `yaml:"max_header_bytes"`
	TLS             TLS      `yaml:"tls"`
}

// TLS holds the TLS configuration for HTTPS
type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// Duration is a custom type that can unmarshal from YAML strings
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for Duration
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(duration)
	return nil
}

// MarshalYAML implements yaml.Marshaler for Duration
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// UnmarshalJSON implements json.Unmarshaler for Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(duration)
	return nil
}

// MarshalJSON implements json.Marshaler for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// GRPC holds the gRPC server configuration
type GRPC struct {
	Network string   `yaml:"network"`
	Addr    string   `yaml:"addr"`
	Timeout Duration `yaml:"timeout"`
}

// SetDefaults sets default values for HTTP configuration
func (h *HTTP) SetDefaults() {
	if h.Network == "" {
		h.Network = "tcp"
	}
	if h.Addr == "" {
		h.Addr = ":8080"
	}
	if h.Timeout == 0 {
		h.Timeout = Duration(30 * time.Second)
	}
	if h.ReadTimeout == 0 {
		h.ReadTimeout = Duration(30 * time.Second)
	}
	if h.WriteTimeout == 0 {
		h.WriteTimeout = Duration(30 * time.Second)
	}
	if h.ShutdownTimeout == 0 {
		h.ShutdownTimeout = Duration(10 * time.Second)
	}
	if h.MaxHeaderBytes == 0 {
		h.MaxHeaderBytes = 1048576 // 1MB
	}
}

// Validate validates the HTTP configuration
func (h *HTTP) Validate() error {
	// Validate network
	if h.Network != "tcp" && h.Network != "tcp4" && h.Network != "tcp6" {
		return errors.New("invalid network: must be tcp, tcp4, or tcp6")
	}

	// Validate address
	if h.Addr == "" {
		return errors.New("address cannot be empty")
	}

	// Check if address is valid
	if !strings.HasPrefix(h.Addr, ":") {
		// If it doesn't start with :, it should be a valid host:port
		if _, _, err := net.SplitHostPort(h.Addr); err != nil {
			return errors.New("invalid address format")
		}
	}

	// Validate TLS configuration
	if h.TLS.Enabled {
		if h.TLS.CertFile == "" {
			return errors.New("TLS cert file is required when TLS is enabled")
		}
		if h.TLS.KeyFile == "" {
			return errors.New("TLS key file is required when TLS is enabled")
		}
	}

	return nil
}
