package conf

import (
	"errors"
	"net"
	"strings"
	"time"
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
	Network         string        `yaml:"network"`
	Addr            string        `yaml:"addr"`
	Timeout         time.Duration `yaml:"timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes"`
	TLS             TLS           `yaml:"tls"`
}

// TLS holds the TLS configuration for HTTPS
type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// GRPC holds the gRPC server configuration
type GRPC struct {
	Network string        `yaml:"network"`
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
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
		h.Timeout = 30 * time.Second
	}
	if h.ReadTimeout == 0 {
		h.ReadTimeout = 30 * time.Second
	}
	if h.WriteTimeout == 0 {
		h.WriteTimeout = 30 * time.Second
	}
	if h.ShutdownTimeout == 0 {
		h.ShutdownTimeout = 10 * time.Second
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
