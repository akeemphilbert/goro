package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Server represents the HTTP server with routing and middleware capabilities
type Server struct {
	httpServer *http.Server
	router     Router
	middleware []Middleware
	config     *Config
	logger     log.Logger
}

// Config holds HTTP server configuration
type Config struct {
	Port            int           `yaml:"port" default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" default:"30s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" default:"30s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" default:"10s"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" default:"1048576"`
	TLS             *TLSConfig    `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" default:"false"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// NewServer creates a new HTTP server instance
func NewServer(config *Config, router Router, logger log.Logger) *Server {
	return &Server{
		config: config,
		router: router,
		logger: logger,
	}
}

// Start initializes and starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	// Implementation will be added in task 2.1
	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	// Implementation will be added in task 2.2
	return nil
}

// Use adds middleware to the server
func (s *Server) Use(middleware ...Middleware) {
	s.middleware = append(s.middleware, middleware...)
}
