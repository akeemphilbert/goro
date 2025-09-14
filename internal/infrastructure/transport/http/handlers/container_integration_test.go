package handlers

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerHandlerIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "container_handler_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up dependencies - use existing working infrastructure
	repo, err := infrastructure.NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	converter := infrastructure.NewRDFConverter()

	// Create database and event infrastructure
	db, err := infrastructure.DatabaseProvider()
	require.NoError(t, err)

	eventStore, err := infrastructure.EventStoreProvider(db)
	require.NoError(t, err)

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	require.NoError(t, err)

	// Create unit of work factory using existing infrastructure
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create membership indexer
	membershipIndexer, err := infrastructure.NewSQLiteMembershipIndexer(tempDir + "/membership.db")
	require.NoError(t, err)
	defer membershipIndexer.Close()

	// Create container repository
	containerRepo, err := infrastructure.NewFileSystemContainerRepository(tempDir, membershipIndexer)
	require.NoError(t, err)

	// Create services
	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)
	containerService := application.NewContainerService(containerRepo, unitOfWorkFactory)
	logger := log.NewStdLogger(io.Discard)

	// Create container handler
	handler := NewContainerHandler(containerService, storageService, logger)

	// Test container creation and retrieval flow
	t.Run("container lifecycle integration", func(t *testing.T) {
		ctx := context.Background()

		// Create a container
		container, err := containerService.CreateContainer(ctx, "test-container", "", domain.BasicContainer)
		require.NoError(t, err)
		assert.Equal(t, "test-container", container.ID())
		assert.Equal(t, domain.BasicContainer, container.ContainerType)

		// Retrieve the container
		retrievedContainer, err := containerService.GetContainer(ctx, "test-container")
		require.NoError(t, err)
		assert.Equal(t, "test-container", retrievedContainer.ID())

		// List container members (should be empty initially)
		listing, err := containerService.ListContainerMembers(ctx, "test-container", domain.GetDefaultPagination())
		require.NoError(t, err)
		assert.Equal(t, "test-container", listing.ContainerID)
		assert.Empty(t, listing.Members)

		// Test container response building
		response := handler.buildContainerResponse(retrievedContainer, listing, "application/ld+json")
		assert.Equal(t, "test-container", response["@id"])
		assert.Contains(t, response["@type"], "ldp:BasicContainer")
		assert.Equal(t, 0, response["ldp:memberCount"])

		// Update container metadata
		retrievedContainer.SetTitle("Test Container")
		retrievedContainer.SetDescription("A test container for integration testing")
		err = containerService.UpdateContainer(ctx, retrievedContainer)
		require.NoError(t, err)

		// Verify metadata update
		updatedContainer, err := containerService.GetContainer(ctx, "test-container")
		require.NoError(t, err)
		assert.Equal(t, "Test Container", updatedContainer.GetTitle())
		assert.Equal(t, "A test container for integration testing", updatedContainer.GetDescription())

		// Test response with metadata
		updatedListing, err := containerService.ListContainerMembers(ctx, "test-container", domain.GetDefaultPagination())
		require.NoError(t, err)

		responseWithMetadata := handler.buildContainerResponse(updatedContainer, updatedListing, "application/ld+json")
		assert.Equal(t, "Test Container", responseWithMetadata["dcterms:title"])
		assert.Equal(t, "A test container for integration testing", responseWithMetadata["dcterms:description"])

		// Delete the container (should work since it's empty)
		err = containerService.DeleteContainer(ctx, "test-container")
		require.NoError(t, err)

		// Verify container is deleted
		_, err = containerService.GetContainer(ctx, "test-container")
		assert.Error(t, err)
		assert.True(t, domain.IsResourceNotFound(err))
	})

	t.Run("container with resources integration", func(t *testing.T) {
		ctx := context.Background()

		// Create a container
		container, err := containerService.CreateContainer(ctx, "container-with-resources", "", domain.BasicContainer)
		require.NoError(t, err)

		// Create a resource
		_, err = storageService.StoreResource(ctx, "test-resource", []byte(`{"test": "data"}`), "application/json")
		require.NoError(t, err)

		// Add resource to container
		err = containerService.AddResource(ctx, "container-with-resources", "test-resource")
		require.NoError(t, err)

		// List container members
		listing, err := containerService.ListContainerMembers(ctx, "container-with-resources", domain.GetDefaultPagination())
		require.NoError(t, err)
		assert.Equal(t, 1, len(listing.Members))
		assert.Contains(t, listing.Members, "test-resource")

		// Test container response with members
		response := handler.buildContainerResponse(container, listing, "application/ld+json")
		assert.Equal(t, 1, response["ldp:memberCount"])
		assert.Contains(t, response["ldp:contains"], "test-resource")

		// Try to delete container (should fail because it's not empty)
		err = containerService.DeleteContainer(ctx, "container-with-resources")
		assert.Error(t, err)
		assert.True(t, domain.IsContainerNotEmpty(err))

		// Remove resource from container
		err = containerService.RemoveResource(ctx, "container-with-resources", "test-resource")
		require.NoError(t, err)

		// Now delete should work
		err = containerService.DeleteContainer(ctx, "container-with-resources")
		require.NoError(t, err)

		// Clean up resource
		err = storageService.DeleteResource(ctx, "test-resource")
		require.NoError(t, err)
	})

	t.Run("content negotiation integration", func(t *testing.T) {
		tests := []struct {
			name         string
			acceptHeader string
			expected     string
		}{
			{
				name:         "JSON-LD format",
				acceptHeader: "application/ld+json",
				expected:     "application/ld+json",
			},
			{
				name:         "Turtle format",
				acceptHeader: "text/turtle",
				expected:     "text/turtle",
			},
			{
				name:         "RDF/XML format",
				acceptHeader: "application/rdf+xml",
				expected:     "application/rdf+xml",
			},
			{
				name:         "Default format",
				acceptHeader: "",
				expected:     "application/ld+json",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := handler.negotiateContentType(tt.acceptHeader)
				assert.Equal(t, tt.expected, result)

				contentType := handler.getResponseContentType(result)
				assert.Equal(t, tt.expected, contentType)
			})
		}
	})
}
