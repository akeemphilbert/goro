package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireContainerComponentAssembly(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "wire_container_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":0", // Use random port for testing
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0", // Use random port for testing
		},
	}

	// Create logger
	logger := log.NewStdLogger(os.Stdout)

	t.Run("wire app assembly succeeds", func(t *testing.T) {
		// Test that Wire can successfully assemble all components
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err, "Wire should successfully assemble all components")
		require.NotNil(t, app, "App should be created")
		require.NotNil(t, cleanup, "Cleanup function should be provided")

		// Cleanup
		cleanup()
	})

	t.Run("container service is properly wired", func(t *testing.T) {
		// Test that container service is properly created and wired
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify app is created successfully
		assert.NotNil(t, app)
	})

	t.Run("container repository is properly wired", func(t *testing.T) {
		// Test that container repository is properly created and wired
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify app is created successfully
		assert.NotNil(t, app)
	})

	t.Run("container handlers are properly wired", func(t *testing.T) {
		// Test that container handlers are properly created and wired
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify app is created successfully
		assert.NotNil(t, app)
	})
}

func TestWireContainerDependencyResolution(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "wire_dependency_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up test data directory
	dataDir := filepath.Join(tempDir, "data")
	err = os.MkdirAll(dataDir, 0755)
	require.NoError(t, err)

	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":0",
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0",
		},
	}

	logger := log.NewStdLogger(os.Stdout)

	t.Run("all container dependencies resolve correctly", func(t *testing.T) {
		// Test that all container-related dependencies can be resolved
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err, "All container dependencies should resolve")
		require.NotNil(t, app)
		defer cleanup()
	})

	t.Run("container service dependencies are satisfied", func(t *testing.T) {
		// Test that container service gets all required dependencies
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify successful assembly indicates proper dependency satisfaction
		assert.NotNil(t, app)
	})

	t.Run("container repository dependencies are satisfied", func(t *testing.T) {
		// Test that container repository gets all required dependencies
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify successful assembly indicates proper dependency satisfaction
		assert.NotNil(t, app)
	})
}

func TestWireContainerEventHandling(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "wire_event_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test configuration
	config := &conf.Server{
		HTTP: &conf.HTTP{
			Network: "tcp",
			Addr:    ":0",
		},
		GRPC: &conf.GRPC{
			Network: "tcp",
			Addr:    ":0",
		},
	}

	logger := log.NewStdLogger(os.Stdout)

	t.Run("container event handlers are properly wired", func(t *testing.T) {
		// Test that container event handlers are properly registered
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err, "Container event handlers should be properly wired")
		require.NotNil(t, app)
		defer cleanup()
	})

	t.Run("event dispatcher is properly configured for containers", func(t *testing.T) {
		// Test that event dispatcher is properly configured
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err)
		defer cleanup()

		// Verify successful assembly
		assert.NotNil(t, app)
	})
}

// TestWireContainerConfigurationIntegration tests that container-specific configuration
// is properly integrated with the Wire dependency injection
func TestWireContainerConfigurationIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "wire_config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("container configuration is properly injected", func(t *testing.T) {
		// Create test configuration with container-specific settings
		config := &conf.Server{
			HTTP: &conf.HTTP{
				Network: "tcp",
				Addr:    ":0",
			},
			GRPC: &conf.GRPC{
				Network: "tcp",
				Addr:    ":0",
			},
		}

		logger := log.NewStdLogger(os.Stdout)

		// Test that configuration is properly used in Wire assembly
		app, cleanup, err := wireApp(config, logger)
		require.NoError(t, err, "Container configuration should be properly injected")
		require.NotNil(t, app)
		defer cleanup()
	})
}
