package main

import (
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/akeemphilbert/goro/internal/conf"
)

func TestWireAppCreation(t *testing.T) {
	// Create test configuration
	serverConf := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":8080",
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":9000",
		},
	}

	// Create test logger
	logger := log.NewStdLogger(nil)

	// Test that wireApp can create the application
	app, cleanup, err := wireApp(serverConf, logger)
	if err != nil {
		t.Fatalf("wireApp failed: %v", err)
	}
	defer cleanup()

	if app == nil {
		t.Fatal("wireApp returned nil app")
	}

	// Verify app has expected properties
	if app.Name() != "goro-server" {
		t.Errorf("Expected app name 'goro-server', got '%s'", app.Name())
	}
}

func TestHTTPServerCreation(t *testing.T) {
	// Create test configuration
	httpConf := &conf.HTTP{
		Network: "tcp",
		Addr:    ":8080",
	}

	// Create test logger
	logger := log.NewStdLogger(nil)

	// Test that NewHTTPServer can create HTTP server
	server := NewHTTPServer(httpConf, logger)
	if server == nil {
		t.Fatal("NewHTTPServer returned nil")
	}

	// Verify server is not nil (type checking is done at compile time)
	if server == nil {
		t.Error("NewHTTPServer returned nil server")
	}
}

func TestGRPCServerCreation(t *testing.T) {
	// Create test configuration
	grpcConf := &conf.GRPC{
		Network: "tcp",
		Addr:    ":9000",
	}

	// Create test logger
	logger := log.NewStdLogger(nil)

	// Test that NewGRPCServer can create gRPC server
	server := NewGRPCServer(grpcConf, logger)
	if server == nil {
		t.Fatal("NewGRPCServer returned nil")
	}

	// Verify server is not nil (type checking is done at compile time)
	if server == nil {
		t.Error("NewGRPCServer returned nil server")
	}
}

func TestProviderSetIntegration(t *testing.T) {
	// This test verifies that all providers can be wired together
	// without circular dependencies or missing providers

	serverConf := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":8080",
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":9000",
		},
	}

	logger := log.NewStdLogger(nil)

	// This should not panic or fail
	app, cleanup, err := wireApp(serverConf, logger)
	if err != nil {
		t.Fatalf("Provider integration failed: %v", err)
	}
	defer cleanup()

	if app == nil {
		t.Fatal("Provider integration returned nil app")
	}
}
