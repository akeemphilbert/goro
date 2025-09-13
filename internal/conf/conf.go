package conf

import "time"

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
	Network      string        `yaml:"network"`
	Addr         string        `yaml:"addr"`
	Timeout      time.Duration `yaml:"timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// GRPC holds the gRPC server configuration
type GRPC struct {
	Network string        `yaml:"network"`
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}
