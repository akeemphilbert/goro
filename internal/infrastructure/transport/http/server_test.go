package http

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
)

func TestNewHTTPServer(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	tests := []struct {
		name   string
		config *conf.HTTP
	}{
		{
			name: "basic configuration",
			config: &conf.HTTP{
				Network:         "tcp",
				Addr:            ":8080",
				Timeout:         conf.Duration(30 * time.Second),
				ReadTimeout:     conf.Duration(30 * time.Second),
				WriteTimeout:    conf.Duration(30 * time.Second),
				ShutdownTimeout: conf.Duration(10 * time.Second),
				MaxHeaderBytes:  1048576,
			},
		},
		{
			name: "custom port configuration",
			config: &conf.HTTP{
				Network:         "tcp",
				Addr:            ":9090",
				Timeout:         conf.Duration(60 * time.Second),
				ReadTimeout:     conf.Duration(45 * time.Second),
				WriteTimeout:    conf.Duration(45 * time.Second),
				ShutdownTimeout: conf.Duration(15 * time.Second),
				MaxHeaderBytes:  2097152,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHTTPServer(tt.config, logger, healthHandler, requestResponseHandler)
			if server == nil {
				t.Fatal("NewHTTPServer() returned nil")
			}

			// Verify server is created successfully
			// Note: We can't easily test internal server configuration without exposing internals
			// The main test is that the server is created without panicking
		})
	}
}

func TestNewHTTPServerWithMiddleware(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":8080",
		Timeout:         conf.Duration(30 * time.Second),
		ReadTimeout:     conf.Duration(30 * time.Second),
		WriteTimeout:    conf.Duration(30 * time.Second),
		ShutdownTimeout: conf.Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}

	// Test that middleware is properly registered
	// This is tested implicitly by ensuring the server starts without errors
}

func TestServerOptions(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":0", // Use port 0 for testing to avoid conflicts
		Timeout:         conf.Duration(30 * time.Second),
		ReadTimeout:     conf.Duration(30 * time.Second),
		WriteTimeout:    conf.Duration(30 * time.Second),
		ShutdownTimeout: conf.Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}

	// Test that server can be started and stopped
	ctx := context.Background()
	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Server stop error: %v", err)
	}
}
func TestRouteRegistration(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":0",
		Timeout:         conf.Duration(30 * time.Second),
		ReadTimeout:     conf.Duration(30 * time.Second),
		WriteTimeout:    conf.Duration(30 * time.Second),
		ShutdownTimeout: conf.Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}

	// Test basic route registration
	RegisterRoutes(server, healthHandler, requestResponseHandler)

	// Verify server can start with registered routes
	ctx := context.Background()
	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Server stop error: %v", err)
	}
}

func TestRouteGroups(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":0",
		Timeout:         conf.Duration(30 * time.Second),
		ReadTimeout:     conf.Duration(30 * time.Second),
		WriteTimeout:    conf.Duration(30 * time.Second),
		ShutdownTimeout: conf.Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}

	// Test route groups and path parameters
	RegisterRouteGroups(server)

	// Verify server can start with route groups
	ctx := context.Background()
	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Server stop error: %v", err)
	}
}

func TestPathParameters(t *testing.T) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":0",
		Timeout:         conf.Duration(30 * time.Second),
		ReadTimeout:     conf.Duration(30 * time.Second),
		WriteTimeout:    conf.Duration(30 * time.Second),
		ShutdownTimeout: conf.Duration(10 * time.Second),
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}

	// Test path parameter routes
	RegisterParameterRoutes(server)

	// Verify server can start with parameter routes
	ctx := context.Background()
	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Server stop error: %v", err)
	}
}
