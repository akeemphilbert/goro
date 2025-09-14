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
	HTTP      *HTTP      `yaml:"http"`
	GRPC      *GRPC      `yaml:"grpc"`
	Container *Container `yaml:"container"`
}

// HTTP holds the HTTP server configuration
type HTTP struct {
	Network         string   `json:"network"`
	Addr            string   `json:"addr"`
	Timeout         Duration `json:"timeout"`
	ReadTimeout     Duration `json:"read_timeout"`
	WriteTimeout    Duration `json:"write_timeout"`
	ShutdownTimeout Duration `json:"shutdown_timeout"`
	MaxHeaderBytes  int      `json:"max_header_bytes"`
	TLS             TLS      `json:"tls"`
}

// TLS holds the TLS configuration for HTTPS
type TLS struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
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
	Network string   `json:"network"`
	Addr    string   `json:"addr"`
	Timeout Duration `json:"timeout"`
}

// Container holds the container-specific configuration
type Container struct {
	StoragePath     string `json:"storage_path"`
	IndexPath       string `json:"index_path"`
	MaxDepth        int    `json:"max_depth"`
	PageSize        int    `json:"page_size"`
	CacheEnabled    bool   `json:"cache_enabled"`
	CacheSize       int    `json:"cache_size"`
	IndexingEnabled bool   `json:"indexing_enabled"`
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

// SetDefaults sets default values for Container configuration
func (c *Container) SetDefaults() {
	if c.StoragePath == "" {
		c.StoragePath = "./data/pod-storage"
	}
	if c.IndexPath == "" {
		c.IndexPath = "./data/pod-storage/index"
	}
	if c.MaxDepth == 0 {
		c.MaxDepth = 100 // Maximum container nesting depth
	}
	if c.PageSize == 0 {
		c.PageSize = 50 // Default page size for container listings
	}
	if c.CacheSize == 0 {
		c.CacheSize = 1000 // Default cache size for containers
	}
	// CacheEnabled and IndexingEnabled default to false (zero value)
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

// Validate validates the Container configuration
func (c *Container) Validate() error {
	// Validate storage path
	if c.StoragePath == "" {
		return errors.New("storage path cannot be empty")
	}

	// Validate index path
	if c.IndexPath == "" {
		return errors.New("index path cannot be empty")
	}

	// Validate max depth
	if c.MaxDepth < 1 {
		return errors.New("max depth must be at least 1")
	}
	if c.MaxDepth > 1000 {
		return errors.New("max depth cannot exceed 1000")
	}

	// Validate page size
	if c.PageSize < 1 {
		return errors.New("page size must be at least 1")
	}
	if c.PageSize > 10000 {
		return errors.New("page size cannot exceed 10000")
	}

	// Validate cache size
	if c.CacheSize < 0 {
		return errors.New("cache size cannot be negative")
	}

	return nil
}
