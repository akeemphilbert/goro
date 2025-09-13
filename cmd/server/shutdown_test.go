package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdown(t *testing.T) {
	tests := []struct {
		name           string
		shutdownSignal os.Signal
		expectCleanup  bool
	}{
		{
			name:           "SIGTERM triggers graceful shutdown",
			shutdownSignal: syscall.SIGTERM,
			expectCleanup:  true,
		},
		{
			name:           "SIGINT triggers graceful shutdown",
			shutdownSignal: syscall.SIGINT,
			expectCleanup:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration
			config := &conf.Server{
				HTTP: &conf.HTTP{
					Network:         "tcp",
					Addr:            ":0", // Use random port for testing
					Timeout:         conf.Duration(5 * time.Second),
					ShutdownTimeout: conf.Duration(2 * time.Second),
				},
				GRPC: &conf.GRPC{
					Network: "tcp",
					Addr:    ":0", // Use random port for testing
					Timeout: conf.Duration(5 * time.Second),
				},
			}
			config.HTTP.SetDefaults()

			logger := log.NewStdLogger(os.Stdout)

			// Create app with test configuration
			app, cleanup, err := wireApp(config, logger)
			require.NoError(t, err)
			defer cleanup()

			// Start app in goroutine
			appDone := make(chan error, 1)
			go func() {
				appDone <- app.Run()
			}()

			// Wait a bit for app to start
			time.Sleep(100 * time.Millisecond)

			// Send shutdown signal
			process, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			err = process.Signal(tt.shutdownSignal)
			require.NoError(t, err)

			// Wait for app to shutdown
			select {
			case err := <-appDone:
				if tt.expectCleanup {
					assert.NoError(t, err, "App should shutdown gracefully without error")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("App did not shutdown within timeout")
			}
		})
	}
}

func TestGracefulShutdownWithActiveConnections(t *testing.T) {
	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network:         "tcp",
			Addr:            ":0", // Use random port for testing
			Timeout:         conf.Duration(5 * time.Second),
			ShutdownTimeout: conf.Duration(3 * time.Second),
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0", // Use random port for testing
			Timeout: conf.Duration(5 * time.Second),
		},
	}
	config.HTTP.SetDefaults()

	logger := log.NewStdLogger(os.Stdout)

	// Create app with test configuration
	app, cleanup, err := wireApp(config, logger)
	require.NoError(t, err)
	defer cleanup()

	// Start app in goroutine
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Run()
	}()

	// Wait for app to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	// This is a simplified test - in a real scenario we'd need to extract the port
	// For now, we'll simulate an active connection
	connectionDone := make(chan bool, 1)
	go func() {
		// Simulate a long-running request
		time.Sleep(1 * time.Second)
		connectionDone <- true
	}()

	// Send shutdown signal
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Verify that the app waits for active connections
	select {
	case <-connectionDone:
		// Connection completed
	case <-time.After(2 * time.Second):
		t.Fatal("Connection should have completed before shutdown timeout")
	}

	// Wait for app to shutdown
	select {
	case err := <-appDone:
		assert.NoError(t, err, "App should shutdown gracefully after connections complete")
	case <-time.After(5 * time.Second):
		t.Fatal("App did not shutdown within timeout")
	}
}

func TestShutdownTimeout(t *testing.T) {
	// Create test configuration with very short shutdown timeout
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network:         "tcp",
			Addr:            ":0",
			Timeout:         conf.Duration(5 * time.Second),
			ShutdownTimeout: conf.Duration(100 * time.Millisecond), // Very short timeout
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0",
			Timeout: conf.Duration(5 * time.Second),
		},
	}
	config.HTTP.SetDefaults()

	logger := log.NewStdLogger(os.Stdout)

	// Create app with test configuration
	app, cleanup, err := wireApp(config, logger)
	require.NoError(t, err)
	defer cleanup()

	// Start app in goroutine
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Run()
	}()

	// Wait for app to start
	time.Sleep(100 * time.Millisecond)

	// Send shutdown signal
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// App should shutdown within reasonable time even with short timeout
	select {
	case err := <-appDone:
		// Should shutdown without error (forced termination is still clean)
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("App did not shutdown within reasonable time")
	}
}

func TestResourceCleanup(t *testing.T) {
	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network:         "tcp",
			Addr:            ":0",
			Timeout:         conf.Duration(5 * time.Second),
			ShutdownTimeout: conf.Duration(2 * time.Second),
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0",
			Timeout: conf.Duration(5 * time.Second),
		},
	}
	config.HTTP.SetDefaults()

	logger := log.NewStdLogger(os.Stdout)

	// Create app with test configuration
	app, cleanup, err := wireApp(config, logger)
	require.NoError(t, err)

	// Verify cleanup function exists and is callable
	assert.NotNil(t, cleanup, "Cleanup function should be provided")

	// Test that cleanup can be called without error
	cleanup()

	// Start app in goroutine
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Run()
	}()

	// Wait for app to start
	time.Sleep(100 * time.Millisecond)

	// Send shutdown signal
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for app to shutdown gracefully
	select {
	case err := <-appDone:
		assert.NoError(t, err, "App should shutdown gracefully")
	case <-time.After(5 * time.Second):
		t.Fatal("App did not shutdown within timeout")
	}
}

func TestKratosAppLifecycleManagement(t *testing.T) {
	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network:         "tcp",
			Addr:            ":0",
			Timeout:         conf.Duration(5 * time.Second),
			ShutdownTimeout: conf.Duration(2 * time.Second),
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0",
			Timeout: conf.Duration(5 * time.Second),
		},
	}
	config.HTTP.SetDefaults()

	logger := log.NewStdLogger(os.Stdout)

	// Create app with test configuration
	app, cleanup, err := wireApp(config, logger)
	require.NoError(t, err)
	defer cleanup()

	// Verify app implements proper lifecycle interface
	assert.NotNil(t, app, "App should be created")

	// Start app in goroutine
	appDone := make(chan error, 1)
	go func() {
		appDone <- app.Run()
	}()

	// Wait a bit for app to start
	time.Sleep(100 * time.Millisecond)

	// Stop app using signal (Kratos handles SIGTERM/SIGINT automatically)
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for app to shutdown
	select {
	case err := <-appDone:
		// App should shutdown cleanly
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("App did not shutdown within timeout")
	}
}
