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

func TestResourceHandlerBusinessLogicIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "resource_handler_test")
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

	// Create storage service
	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)
	logger := log.NewStdLogger(io.Discard)

	// Create resource handler
	handler := NewResourceHandler(storageService, logger)

	t.Run("Content Negotiation Logic", func(t *testing.T) {
		tests := []struct {
			name           string
			acceptHeader   string
			expectedFormat string
		}{
			{
				name:           "JSON-LD exact match",
				acceptHeader:   "application/ld+json",
				expectedFormat: "application/ld+json",
			},
			{
				name:           "Turtle exact match",
				acceptHeader:   "text/turtle",
				expectedFormat: "text/turtle",
			},
			{
				name:           "JSON alias to JSON-LD",
				acceptHeader:   "application/json",
				expectedFormat: "application/ld+json",
			},
			{
				name:           "Quality preference",
				acceptHeader:   "text/turtle;q=0.8, application/ld+json;q=0.9",
				expectedFormat: "application/ld+json",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := handler.negotiateContentType(tt.acceptHeader)
				assert.Equal(t, tt.expectedFormat, result)
			})
		}
	})

	t.Run("Storage Service Integration", func(t *testing.T) {
		ctx := context.Background()

		// Test storing a resource
		jsonLDData := []byte(`{
			"@context": "https://www.w3.org/ns/activitystreams",
			"type": "Note",
			"content": "Integration test note"
		}`)

		resource, err := storageService.StoreResource(ctx, "integration-test", jsonLDData, "application/ld+json")
		require.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, "integration-test", resource.ID())
		assert.Equal(t, "application/ld+json", resource.GetContentType())

		// Test retrieving the resource
		retrieved, err := storageService.RetrieveResource(ctx, "integration-test", "application/ld+json")
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "integration-test", retrieved.ID())
		assert.Equal(t, "application/ld+json", retrieved.GetContentType())
		assert.Contains(t, string(retrieved.GetData()), "Integration test note")

		// Test resource exists
		exists, err := storageService.ResourceExists(ctx, "integration-test")
		require.NoError(t, err)
		assert.True(t, exists)

		// Test deleting the resource
		err = storageService.DeleteResource(ctx, "integration-test")
		require.NoError(t, err)

		// Verify resource is deleted
		exists, err = storageService.ResourceExists(ctx, "integration-test")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Error Handling Integration", func(t *testing.T) {
		ctx := context.Background()

		// Test retrieving non-existent resource
		_, err := storageService.RetrieveResource(ctx, "nonexistent", "application/ld+json")
		assert.Error(t, err)
		assert.True(t, domain.IsResourceNotFound(err))

		// Test deleting non-existent resource
		err = storageService.DeleteResource(ctx, "nonexistent")
		assert.Error(t, err)
		assert.True(t, domain.IsResourceNotFound(err))

		// Test storing with invalid ID
		_, err = storageService.StoreResource(ctx, "", []byte("test"), "application/ld+json")
		assert.Error(t, err)
	})

	t.Run("ETag Generation", func(t *testing.T) {
		// Create test resources
		resource1 := domain.NewResource("test-1", "application/ld+json", []byte(`{"test": "data1"}`))
		resource2 := domain.NewResource("test-2", "application/ld+json", []byte(`{"test": "data2"}`))
		resource3 := domain.NewResource("test-1", "application/ld+json", []byte(`{"test": "data1"}`)) // Same as resource1

		// Generate ETags
		etag1 := handler.generateETag(resource1)
		etag2 := handler.generateETag(resource2)
		etag3 := handler.generateETag(resource3)

		// Assert
		assert.NotEmpty(t, etag1)
		assert.NotEmpty(t, etag2)
		assert.NotEmpty(t, etag3)

		// Different resources should have different ETags
		assert.NotEqual(t, etag1, etag2)

		// Same resource should have same ETag
		assert.Equal(t, etag1, etag3)
	})

	t.Run("Binary File Support", func(t *testing.T) {
		ctx := context.Background()

		// Test storing binary data
		binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header

		resource, err := storageService.StoreResource(ctx, "binary-test", binaryData, "image/png")
		require.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, "binary-test", resource.ID())
		assert.Equal(t, "image/png", resource.GetContentType())

		// Test retrieving binary data
		retrieved, err := storageService.RetrieveResource(ctx, "binary-test", "")
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, binaryData, retrieved.GetData())
	})

	t.Run("Multiple Format Support", func(t *testing.T) {
		ctx := context.Background()

		// Test different RDF formats
		formats := map[string][]byte{
			"application/ld+json": []byte(`{"@context": "test", "name": "JSON-LD"}`),
			"text/turtle":         []byte(`@prefix ex: <http://example.org/> . ex:resource ex:name "Turtle" .`),
			"application/rdf+xml": []byte(`<?xml version="1.0"?><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="test"><name>RDF/XML</name></rdf:Description></rdf:RDF>`),
		}

		for contentType, data := range formats {
			t.Run("Format_"+contentType, func(t *testing.T) {
				resourceID := "format-test-" + contentType

				// Store resource
				resource, err := storageService.StoreResource(ctx, resourceID, data, contentType)
				require.NoError(t, err)
				assert.Equal(t, contentType, resource.GetContentType())

				// Retrieve resource
				retrieved, err := storageService.RetrieveResource(ctx, resourceID, contentType)
				require.NoError(t, err)
				assert.Equal(t, contentType, retrieved.GetContentType())
				assert.Equal(t, data, retrieved.GetData())
			})
		}
	})
}
