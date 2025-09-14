package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainerServiceProvider(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "container_service_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up test dependencies
	basePath := filepath.Join(tempDir, "data")
	err = os.MkdirAll(basePath, 0755)
	require.NoError(t, err)

	// Create membership indexer
	indexer, err := infrastructure.MembershipIndexerProvider(basePath)
	require.NoError(t, err)

	// Create container repository
	containerRepo, err := infrastructure.NewFileSystemContainerRepository(basePath, indexer)
	require.NoError(t, err)

	// Create event store and dispatcher
	db, err := infrastructure.DatabaseProvider()
	require.NoError(t, err)

	eventStore, err := infrastructure.EventStoreProvider(db)
	require.NoError(t, err)

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	require.NoError(t, err)

	// Create unit of work factory
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()

	t.Run("creates container service successfully", func(t *testing.T) {
		service, err := NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			eventDispatcher,
			rdfConverter,
		)

		require.NoError(t, err, "Container service provider should create service successfully")
		assert.NotNil(t, service, "Container service should not be nil")
		assert.IsType(t, &ContainerService{}, service, "Should return ContainerService type")
	})

	t.Run("registers event handlers", func(t *testing.T) {
		service, err := NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			eventDispatcher,
			rdfConverter,
		)

		require.NoError(t, err, "Container service provider should register event handlers")
		assert.NotNil(t, service, "Container service should not be nil")

		// Verify that event handlers are registered by checking the service is created
		// (actual event handler registration is tested in the event handler tests)
	})

	t.Run("handles nil dependencies gracefully", func(t *testing.T) {
		// Test with nil container repository
		_, err := NewContainerServiceProvider(
			nil,
			unitOfWorkFactory,
			eventDispatcher,
			rdfConverter,
		)
		assert.Error(t, err, "Should return error with nil container repository")

		// Test with nil unit of work factory
		_, err = NewContainerServiceProvider(
			containerRepo,
			nil,
			eventDispatcher,
			rdfConverter,
		)
		assert.Error(t, err, "Should return error with nil unit of work factory")

		// Test with nil event dispatcher
		_, err = NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			nil,
			rdfConverter,
		)
		assert.Error(t, err, "Should return error with nil event dispatcher")

		// Test with nil RDF converter
		_, err = NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			eventDispatcher,
			nil,
		)
		assert.Error(t, err, "Should return error with nil RDF converter")
	})
}

func TestNewEventHandlerRegistrarProvider(t *testing.T) {
	t.Run("creates event handler registrar successfully", func(t *testing.T) {
		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		registrar := NewEventHandlerRegistrarProvider(eventDispatcher)

		assert.NotNil(t, registrar, "Event handler registrar should not be nil")
		assert.IsType(t, &EventHandlerRegistrar{}, registrar, "Should return EventHandlerRegistrar type")
	})

	t.Run("handles nil event dispatcher", func(t *testing.T) {
		registrar := NewEventHandlerRegistrarProvider(nil)

		// Should still create registrar but it won't be functional
		assert.NotNil(t, registrar, "Event handler registrar should be created even with nil dispatcher")
	})
}

func TestContainerServiceProviderIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "container_provider_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up complete test environment
	basePath := filepath.Join(tempDir, "data")
	err = os.MkdirAll(basePath, 0755)
	require.NoError(t, err)

	t.Run("full provider chain works correctly", func(t *testing.T) {
		// Create all dependencies as they would be created by Wire
		indexer, err := infrastructure.MembershipIndexerProvider(basePath)
		require.NoError(t, err)

		containerRepo, err := infrastructure.NewFileSystemContainerRepository(basePath, indexer)
		require.NoError(t, err)

		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := infrastructure.EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
			return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
		}

		rdfConverter := infrastructure.NewContainerRDFConverter()

		// Create container service using provider
		service, err := NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			eventDispatcher,
			rdfConverter,
		)

		require.NoError(t, err, "Full provider chain should work correctly")
		assert.NotNil(t, service, "Container service should be created")

		// Verify service is functional by checking its type and non-nil state
		assert.IsType(t, &ContainerService{}, service)
	})

	t.Run("provider creates service with proper dependencies", func(t *testing.T) {
		// Test that the provider creates a service with all required dependencies
		indexer, err := infrastructure.MembershipIndexerProvider(basePath)
		require.NoError(t, err)

		containerRepo, err := infrastructure.NewFileSystemContainerRepository(basePath, indexer)
		require.NoError(t, err)

		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := infrastructure.EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
			return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
		}

		rdfConverter := infrastructure.NewContainerRDFConverter()

		service, err := NewContainerServiceProvider(
			containerRepo,
			unitOfWorkFactory,
			eventDispatcher,
			rdfConverter,
		)

		require.NoError(t, err)
		assert.NotNil(t, service)

		// Verify the service has the expected structure
		// (internal fields are not directly accessible, but we can verify the service was created)
		assert.IsType(t, &ContainerService{}, service)
	})
}
