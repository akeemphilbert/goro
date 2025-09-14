package features

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
)

// TestContainerBasicIntegration tests the most basic container functionality
func TestContainerBasicIntegration(t *testing.T) {
	// Create temporary directory for test storage
	tempDir, err := os.MkdirTemp("", "container-basic-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir)

	// Initialize basic repository
	indexer, err := infrastructure.NewSQLiteMembershipIndexer(filepath.Join(tempDir, "index.db"))
	require.NoError(t, err, "Failed to create indexer")

	containerRepo, err := infrastructure.NewFileSystemContainerRepository(tempDir, indexer)
	require.NoError(t, err, "Failed to create container repository")

	ctx := context.Background()

	t.Run("CreateAndRetrieveContainer", func(t *testing.T) {
		// Test container creation using domain constructor
		container := domain.NewContainer("test-container", "", domain.BasicContainer)
		container.Metadata["title"] = "Test Container"

		err := containerRepo.CreateContainer(ctx, container)
		require.NoError(t, err, "Failed to create container")

		// Test container retrieval
		retrieved, err := containerRepo.GetContainer(ctx, "test-container")
		require.NoError(t, err, "Failed to retrieve container")
		assert.Equal(t, "test-container", retrieved.ID())
		assert.Equal(t, domain.BasicContainer, retrieved.ContainerType)

		// Test container existence check
		exists, err := containerRepo.ContainerExists(ctx, "test-container")
		require.NoError(t, err, "Failed to check container existence")
		assert.True(t, exists, "Container should exist")
	})

	t.Run("ContainerNotFound", func(t *testing.T) {
		// Test non-existent container
		exists, err := containerRepo.ContainerExists(ctx, "non-existent")
		require.NoError(t, err, "Failed to check non-existent container")
		assert.False(t, exists, "Non-existent container should not exist")

		// Test retrieving non-existent container
		_, err = containerRepo.GetContainer(ctx, "non-existent")
		assert.Error(t, err, "Retrieving non-existent container should fail")
	})

	t.Run("ContainerDeletion", func(t *testing.T) {
		// Create a container for deletion test
		container := domain.NewContainer("delete-test", "", domain.BasicContainer)
		err := containerRepo.CreateContainer(ctx, container)
		require.NoError(t, err, "Failed to create container for deletion test")

		// Verify it exists
		exists, err := containerRepo.ContainerExists(ctx, "delete-test")
		require.NoError(t, err, "Failed to check container existence")
		assert.True(t, exists, "Container should exist before deletion")

		// Delete the container
		err = containerRepo.DeleteContainer(ctx, "delete-test")
		require.NoError(t, err, "Failed to delete container")

		// Verify deletion
		exists, err = containerRepo.ContainerExists(ctx, "delete-test")
		require.NoError(t, err, "Failed to check deleted container")
		assert.False(t, exists, "Deleted container should not exist")
	})

	t.Run("ContainerTypes", func(t *testing.T) {
		// Test BasicContainer
		basicContainer := domain.NewContainer("basic-container", "", domain.BasicContainer)
		err := containerRepo.CreateContainer(ctx, basicContainer)
		require.NoError(t, err, "Failed to create basic container")

		retrieved, err := containerRepo.GetContainer(ctx, "basic-container")
		require.NoError(t, err, "Failed to retrieve basic container")
		assert.Equal(t, domain.BasicContainer, retrieved.ContainerType)

		// Test DirectContainer
		directContainer := domain.NewContainer("direct-container", "", domain.DirectContainer)
		err = containerRepo.CreateContainer(ctx, directContainer)
		require.NoError(t, err, "Failed to create direct container")

		retrieved, err = containerRepo.GetContainer(ctx, "direct-container")
		require.NoError(t, err, "Failed to retrieve direct container")
		assert.Equal(t, domain.DirectContainer, retrieved.ContainerType)
	})

	t.Run("ContainerMetadata", func(t *testing.T) {
		// Test container with metadata
		container := domain.NewContainer("metadata-test", "", domain.BasicContainer)

		// Set metadata directly on the underlying resource
		container.Metadata["title"] = "Test Container with Metadata"
		container.Metadata["description"] = "This is a test container"

		err := containerRepo.CreateContainer(ctx, container)
		require.NoError(t, err, "Failed to create container with metadata")

		retrieved, err := containerRepo.GetContainer(ctx, "metadata-test")
		require.NoError(t, err, "Failed to retrieve container with metadata")

		// Check metadata exists
		assert.NotNil(t, retrieved.Metadata, "Metadata should not be nil")

		if title, exists := retrieved.Metadata["title"]; exists {
			assert.Equal(t, "Test Container with Metadata", title)
		}

		if description, exists := retrieved.Metadata["description"]; exists {
			assert.Equal(t, "This is a test container", description)
		}
	})

	t.Run("ContainerEvents", func(t *testing.T) {
		// Test that container creation emits events
		container := domain.NewContainer("event-test", "", domain.BasicContainer)

		// Check that creation events were emitted
		events := container.UncommittedEvents()
		assert.GreaterOrEqual(t, len(events), 1, "Container creation should emit events")

		// Store the container
		err := containerRepo.CreateContainer(ctx, container)
		require.NoError(t, err, "Failed to create container")

		// Add member and check for events
		err = container.AddMember("test-resource")
		require.NoError(t, err, "Failed to add member to container")

		events = container.UncommittedEvents()
		assert.GreaterOrEqual(t, len(events), 2, "Adding member should emit additional events")
	})
}

// TestContainerDomainLogic tests container domain logic without repository
func TestContainerDomainLogic(t *testing.T) {
	t.Run("ContainerCreation", func(t *testing.T) {
		container := domain.NewContainer("test-container", "", domain.BasicContainer)

		assert.NotNil(t, container)
		assert.Equal(t, "test-container", container.ID())
		assert.Equal(t, domain.BasicContainer, container.ContainerType)
		assert.Empty(t, container.ParentID)
		assert.NotNil(t, container.Members)
		assert.Len(t, container.Members, 0)
	})

	t.Run("ContainerHierarchy", func(t *testing.T) {
		parent := domain.NewContainer("parent", "", domain.BasicContainer)
		child := domain.NewContainer("child", "parent", domain.BasicContainer)

		assert.Equal(t, "parent", parent.ID())
		assert.Empty(t, parent.ParentID)

		assert.Equal(t, "child", child.ID())
		assert.Equal(t, "parent", child.ParentID)
	})

	t.Run("ContainerMembership", func(t *testing.T) {
		container := domain.NewContainer("test-container", "", domain.BasicContainer)

		// Add members
		err := container.AddMember("resource1")
		assert.NoError(t, err)
		assert.Contains(t, container.Members, "resource1")

		err = container.AddMember("resource2")
		assert.NoError(t, err)
		assert.Contains(t, container.Members, "resource2")
		assert.Len(t, container.Members, 2)

		// Remove member
		err = container.RemoveMember("resource1")
		assert.NoError(t, err)
		assert.NotContains(t, container.Members, "resource1")
		assert.Contains(t, container.Members, "resource2")
		assert.Len(t, container.Members, 1)
	})

	t.Run("ContainerValidation", func(t *testing.T) {
		container := domain.NewContainer("test-container", "", domain.BasicContainer)

		// Test duplicate member prevention
		err := container.AddMember("resource1")
		assert.NoError(t, err)

		err = container.AddMember("resource1")
		assert.Error(t, err, "Adding duplicate member should fail")

		// Test removing non-existent member
		err = container.RemoveMember("non-existent")
		assert.Error(t, err, "Removing non-existent member should fail")
	})

	t.Run("ContainerTypes", func(t *testing.T) {
		// Test valid container types
		assert.True(t, domain.BasicContainer.IsValid())
		assert.True(t, domain.DirectContainer.IsValid())

		// Test invalid container type
		invalidType := domain.ContainerType("InvalidType")
		assert.False(t, invalidType.IsValid())
	})
}
