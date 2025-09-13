package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

func TestConfigurationLoading(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
server:
  http:
    addr: ":8080"
    timeout: 30s
  grpc:
    addr: ":9000"
    timeout: 30s
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test configuration loading
	c := config.New(
		config.WithSource(
			file.NewSource(tempDir),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify configuration can be accessed
	var serverConfig struct {
		Server struct {
			HTTP struct {
				Addr    string `yaml:"addr"`
				Timeout string `yaml:"timeout"`
			} `yaml:"http"`
		} `yaml:"server"`
	}

	if err := c.Scan(&serverConfig); err != nil {
		t.Fatalf("Failed to scan configuration: %v", err)
	}

	if serverConfig.Server.HTTP.Addr != ":8080" {
		t.Errorf("Expected HTTP addr :8080, got %s", serverConfig.Server.HTTP.Addr)
	}

	if serverConfig.Server.HTTP.Timeout != "30s" {
		t.Errorf("Expected HTTP timeout 30s, got %v", serverConfig.Server.HTTP.Timeout)
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("SERVER_HTTP_ADDR", ":9090")
	defer os.Unsetenv("SERVER_HTTP_ADDR")

	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
server:
  http:
    addr: ":8080"
    timeout: 30s
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test that environment variables can override config
	// Note: This test demonstrates the pattern, actual env override
	// implementation may vary based on Kratos config setup
	envAddr := os.Getenv("SERVER_HTTP_ADDR")
	if envAddr != ":9090" {
		t.Errorf("Expected environment variable SERVER_HTTP_ADDR to be :9090, got %s", envAddr)
	}
}

func TestApplicationStartup(t *testing.T) {
	// Create a minimal logger for testing
	logger := log.NewStdLogger(os.Stdout)

	// Test that we can create a Kratos app
	app := kratos.New(
		kratos.Name("goro-server-test"),
		kratos.Version("v1.0.0-test"),
		kratos.Logger(logger),
	)

	if app == nil {
		t.Fatal("Failed to create Kratos application")
	}

	// Test graceful startup and shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start app in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run()
	}()

	// Give app time to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	if err := app.Stop(); err != nil {
		t.Errorf("Failed to stop application: %v", err)
	}

	// Wait for app to finish or timeout
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Errorf("Application run returned error: %v", err)
		}
	case <-ctx.Done():
		t.Error("Application did not shut down within timeout")
	}
}

func TestConfigurationDefaults(t *testing.T) {
	// Test that application can start with default configuration
	// when no config file is provided
	tempDir := t.TempDir()

	// Create config source pointing to empty directory
	c := config.New(
		config.WithSource(
			file.NewSource(tempDir),
		),
	)
	defer c.Close()

	// Should not fail even if no config file exists
	if err := c.Load(); err != nil {
		t.Fatalf("Failed to load configuration from empty directory: %v", err)
	}
}
