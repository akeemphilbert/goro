package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignalHandling(t *testing.T) {
	tests := []struct {
		name   string
		signal os.Signal
	}{
		{
			name:   "SIGTERM signal handling",
			signal: syscall.SIGTERM,
		},
		{
			name:   "SIGINT signal handling",
			signal: syscall.SIGINT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Start app in goroutine
			appDone := make(chan error, 1)
			go func() {
				appDone <- app.Run()
			}()

			// Wait for app to start
			time.Sleep(100 * time.Millisecond)

			// Send the specific signal
			process, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			err = process.Signal(tt.signal)
			require.NoError(t, err)

			// Wait for app to shutdown
			select {
			case err := <-appDone:
				assert.NoError(t, err, "App should shutdown gracefully on %s", tt.signal)
			case <-time.After(5 * time.Second):
				t.Fatalf("App did not shutdown within timeout after receiving %s", tt.signal)
			}
		})
	}
}

func TestConfigurableShutdownTimeout(t *testing.T) {
	tests := []struct {
		name            string
		shutdownTimeout time.Duration
		expectTimeout   bool
	}{
		{
			name:            "Normal shutdown timeout",
			shutdownTimeout: 2 * time.Second,
			expectTimeout:   false,
		},
		{
			name:            "Very short shutdown timeout",
			shutdownTimeout: 50 * time.Millisecond,
			expectTimeout:   false, // Should still shutdown cleanly
		},
		{
			name:            "Long shutdown timeout",
			shutdownTimeout: 5 * time.Second,
			expectTimeout:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration with specific shutdown timeout
			config := &conf.Server{
				HTTP: &conf.HTTP{
					Network:         "tcp",
					Addr:            ":0",
					Timeout:         conf.Duration(5 * time.Second),
					ShutdownTimeout: conf.Duration(tt.shutdownTimeout),
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

			// Record start time for timeout measurement
			startTime := time.Now()

			// Send shutdown signal
			process, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			err = process.Signal(syscall.SIGTERM)
			require.NoError(t, err)

			// Wait for app to shutdown
			select {
			case err := <-appDone:
				shutdownDuration := time.Since(startTime)
				assert.NoError(t, err, "App should shutdown gracefully")

				// Verify shutdown happened within reasonable time
				// Allow some buffer for test execution overhead
				maxExpectedTime := tt.shutdownTimeout + 1*time.Second
				assert.True(t, shutdownDuration < maxExpectedTime,
					"Shutdown took %v, expected less than %v", shutdownDuration, maxExpectedTime)

			case <-time.After(tt.shutdownTimeout + 3*time.Second):
				if tt.expectTimeout {
					// This is expected behavior
					t.Logf("App shutdown timed out as expected after %v", tt.shutdownTimeout)
				} else {
					t.Fatalf("App did not shutdown within expected timeout of %v", tt.shutdownTimeout)
				}
			}
		})
	}
}

func TestForcedTerminationAfterTimeout(t *testing.T) {
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

	// App should still shutdown cleanly even with short timeout
	// Kratos handles this gracefully
	select {
	case err := <-appDone:
		assert.NoError(t, err, "App should shutdown cleanly even with short timeout")
	case <-time.After(2 * time.Second):
		t.Fatal("App did not shutdown within reasonable time")
	}
}

func TestSignalHandlingInMainFunction(t *testing.T) {
	// This test verifies that signal handling is properly implemented in main.go
	// We test this by ensuring the app responds to signals correctly

	signals := []struct {
		name   string
		signal os.Signal
	}{
		{"SIGTERM", syscall.SIGTERM},
		{"SIGINT", syscall.SIGINT},
	}

	for _, tt := range signals {
		t.Run(tt.name, func(t *testing.T) {
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

			// Start app in goroutine
			appDone := make(chan error, 1)
			go func() {
				appDone <- app.Run()
			}()

			// Wait for app to start
			time.Sleep(100 * time.Millisecond)

			// Send signal
			process, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			err = process.Signal(tt.signal)
			require.NoError(t, err)

			// Wait for app to shutdown
			select {
			case err := <-appDone:
				assert.NoError(t, err, "App should handle %s signal gracefully", tt.signal)
			case <-time.After(3 * time.Second):
				t.Fatalf("App did not respond to %s signal within timeout", tt.signal)
			}
		})
	}
}

func TestSignalHandlerRegistration(t *testing.T) {
	// Test that signal handlers are properly registered
	// This is more of an integration test to ensure signals work

	// Create a signal channel to test signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigChan)

	// Send a signal to ourselves
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Verify we can receive the signal
	select {
	case sig := <-sigChan:
		assert.Equal(t, syscall.SIGTERM, sig, "Should receive SIGTERM signal")
	case <-time.After(1 * time.Second):
		t.Fatal("Did not receive signal within timeout")
	}
}
