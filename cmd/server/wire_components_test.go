package main

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireComponents_InfrastructureLayer(t *testing.T) {
	t.Run("Database provider", func(t *testing.T) {
		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)
		assert.NotNil(t, db)
	})

	t.Run("Event store provider", func(t *testing.T) {
		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := infrastructure.EventStoreProvider(db)
		require.NoError(t, err)
		assert.NotNil(t, eventStore)
	})

	t.Run("Event dispatcher provider", func(t *testing.T) {
		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)
		assert.NotNil(t, eventDispatcher)
	})

	t.Run("Repository provider", func(t *testing.T) {
		repo, err := infrastructure.NewOptimizedFileSystemRepositoryProvider()
		require.NoError(t, err)
		assert.NotNil(t, repo)
	})

	t.Run("RDF converter provider", func(t *testing.T) {
		converter := infrastructure.NewRDFConverter()
		assert.NotNil(t, converter)
	})

	t.Run("Unit of work factory provider", func(t *testing.T) {
		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := infrastructure.EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		factory := infrastructure.NewUnitOfWorkFactory(eventStore, eventDispatcher)
		assert.NotNil(t, factory)

		// Test that factory creates unit of work
		uow := factory()
		assert.NotNil(t, uow)
	})
}

func TestWireComponents_ApplicationLayer(t *testing.T) {
	t.Run("Storage service provider", func(t *testing.T) {
		// Create dependencies
		repo, err := infrastructure.NewOptimizedFileSystemRepositoryProvider()
		require.NoError(t, err)

		converter := infrastructure.NewRDFConverter()

		db, err := infrastructure.DatabaseProvider()
		require.NoError(t, err)

		eventStore, err := infrastructure.EventStoreProvider(db)
		require.NoError(t, err)

		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		factory := infrastructure.NewUnitOfWorkFactory(eventStore, eventDispatcher)

		// Create storage service
		service, err := application.NewStorageServiceProvider(repo, converter, factory, eventDispatcher)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("Event handler registrar provider", func(t *testing.T) {
		eventDispatcher, err := infrastructure.NewEventDispatcher()
		require.NoError(t, err)

		registrar := application.NewEventHandlerRegistrarProvider(eventDispatcher)
		assert.NotNil(t, registrar)
	})
}

func TestWireComponents_InterfaceBindings(t *testing.T) {
	t.Run("Repository implements StreamingResourceRepository", func(t *testing.T) {
		repo, err := infrastructure.NewOptimizedFileSystemRepositoryProvider()
		require.NoError(t, err)

		// Verify it's not nil and is the expected type
		assert.NotNil(t, repo, "Repository should be created")

		// Since the provider returns the interface type, we can verify it's working
		// by checking that we can call interface methods (they exist)
		exists, err := repo.Exists(nil, "test-id")
		// We expect no error for a simple exists check, even if the resource doesn't exist
		assert.NoError(t, err)
		assert.False(t, exists, "Non-existent resource should return false")
	})

	t.Run("RDFConverter implements FormatConverter", func(t *testing.T) {
		converter := infrastructure.NewRDFConverter()

		// Verify it implements the interface by checking methods exist
		assert.NotNil(t, converter, "RDFConverter should be created")

		// Test that the methods exist by calling them with test data
		result, err := converter.Convert([]byte("test"), "turtle", "jsonld")
		// We expect an error since this is not valid RDF, but the method should exist
		assert.Error(t, err)
		assert.Nil(t, result)

		// Test ValidateFormat method
		isValid := converter.ValidateFormat("turtle")
		assert.True(t, isValid, "turtle should be a valid format")
	})
}
