package main

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireApp_Integration(t *testing.T) {
	// Arrange
	logger := log.NewStdLogger(nil)
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":8080",
			Timeout: 30,
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":9090",
			Timeout: 30,
		},
	}

	// Act
	app, cleanup, err := wireApp(config, logger)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, cleanup)

	// Cleanup
	if cleanup != nil {
		cleanup()
	}
}

func TestWireApp_AllDependenciesResolved(t *testing.T) {
	// Arrange
	logger := log.NewStdLogger(nil)
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":8081",
			Timeout: 30,
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":9091",
			Timeout: 30,
		},
	}

	// Act
	app, cleanup, err := wireApp(config, logger)

	// Assert
	require.NoError(t, err, "Wire should successfully resolve all dependencies")
	assert.NotNil(t, app, "App should be created")
	assert.NotNil(t, cleanup, "Cleanup function should be provided")

	// Verify app can be started and stopped
	// Note: We don't actually start the app to avoid port conflicts in tests

	// Cleanup
	if cleanup != nil {
		cleanup()
	}
}

func TestWireApp_ErrorHandling(t *testing.T) {
	// Test with invalid config to verify error handling
	logger := log.NewStdLogger(nil)
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "invalid",
			Addr:    "invalid-address",
			Timeout: -1, // Invalid timeout
		},
		GRPC: &conf.GRPC{
			Network: "invalid",
			Addr:    "invalid-address",
			Timeout: -1, // Invalid timeout
		},
	}

	// Act
	app, cleanup, err := wireApp(config, logger)

	// Assert - Wire should still succeed even with invalid config
	// The actual validation happens when starting the servers
	require.NoError(t, err, "Wire should resolve dependencies even with invalid config")
	assert.NotNil(t, app)
	assert.NotNil(t, cleanup)

	// Cleanup
	if cleanup != nil {
		cleanup()
	}
}
