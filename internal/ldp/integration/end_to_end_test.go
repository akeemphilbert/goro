package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndResourceWorkflows tests complete resource storage and retrieval workflows
// Requirements: 1.1, 1.2, 1.3, 3.4, 4.1
func TestEndToEndResourceWorkflows(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	service := createTestStorageService(t, tempDir)
	ctx := context.Background()

	t.Run("Complete_RDF_Resource_Lifecycle", func(t *testing.T) {
		testCompleteRDFResourceLifecycle(t, ctx, service)
	})

	t.Run("Complete_Binary_Resource_Lifecycle", func(t *testing.T) {
		testCompleteBinaryResourceLifecycle(t, ctx, service)
	})

	t.Run("Multiple_Resources_Workflow", func(t *testing.T) {
		testMultipleResourcesWorkflow(t, ctx, service)
	})

	t.Run("Resource_Update_Workflow", func(t *testing.T) {
		testResourceUpdateWorkflow(t, ctx, service)
	})
}

// TestFormatConversionWithSemanticPreservation tests format conversion with semantic preservation
// Requirements: 1.1, 1.2, 1.3
func TestFormatConversionWithSemanticPreservation(t *testing.T) {
	tempDir := t.TempDir()
	service := createTestStorageService(t, tempDir)
	ctx := context.Background()

	t.Run("JSON_LD_To_Turtle_Conversion", func(t *testing.T) {
		testJSONLDToTurtleConversion(t, ctx, service)
	})

	t.Run("Turtle_To_RDF_XML_Conversion", func(t *testing.T) {
		testTurtleToRDFXMLConversion(t, ctx, service)
	})

	t.Run("RDF_XML_To_JSON_LD_Conversion", func(t *testing.T) {
		testRDFXMLToJSONLDConversion(t, ctx, service)
	})

	t.Run("Round_Trip_Format_Conversion", func(t *testing.T) {
		testRoundTripFormatConversion(t, ctx, service)
	})

	t.Run("Semantic_Preservation_Validation", func(t *testing.T) {
		testSemanticPreservationValidation(t, ctx, service)
	})
}

// TestConcurrentAccessScenarios tests concurrent access scenarios
// Requirements: 3.4
func TestConcurrentAccessScenarios(t *testing.T) {
	tempDir := t.TempDir()
	service := createTestStorageService(t, tempDir)
	ctx := context.Background()

	t.Run("Concurrent_Resource_Creation", func(t *testing.T) {
		testConcurrentResourceCreation(t, ctx, service)
	})

	t.Run("Concurrent_Read_Write_Operations", func(t *testing.T) {
		testConcurrentReadWriteOperations(t, ctx, service)
	})

	t.Run("Concurrent_Format_Conversions", func(t *testing.T) {
		testConcurrentFormatConversions(t, ctx, service)
	})

	t.Run("Concurrent_Resource_Updates", func(t *testing.T) {
		testConcurrentResourceUpdates(t, ctx, service)
	})
}

// TestDataIntegrityAcrossOperations tests data integrity across operations
// Requirements: 4.1
func TestDataIntegrityAcrossOperations(t *testing.T) {
	tempDir := t.TempDir()
	service := createTestStorageService(t, tempDir)
	ctx := context.Background()

	t.Run("Checksum_Validation_Workflow", func(t *testing.T) {
		testChecksumValidationWorkflow(t, ctx, service, tempDir)
	})

	t.Run("Data_Corruption_Detection", func(t *testing.T) {
		testDataCorruptionDetection(t, ctx, service, tempDir)
	})

	t.Run("Transaction_Consistency", func(t *testing.T) {
		testTransactionConsistency(t, ctx, service)
	})

	t.Run("Event_Sourcing_Integrity", func(t *testing.T) {
		testEventSourcingIntegrity(t, ctx, service)
	})
}

// Helper function to create test storage service
func createTestStorageService(t *testing.T, tempDir string) *application.StorageService {
	// Create infrastructure dependencies
	db, err := infrastructure.DatabaseProvider()
	require.NoError(t, err)

	eventStore, err := infrastructure.EventStoreProvider(db)
	require.NoError(t, err)

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	require.NoError(t, err)

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create repository
	repo, err := infrastructure.NewFileSystemRepositoryWithPath(tempDir)
	require.NoError(t, err)

	// Create converter
	converter := infrastructure.NewRDFConverter()

	// Create storage service
	service := application.NewStorageService(repo, converter, unitOfWorkFactory)

	// Register event handlers
	registrar := application.NewEventHandlerRegistrar(eventDispatcher)
	err = registrar.RegisterAllHandlers(repo)
	require.NoError(t, err)

	return service
}

// Test complete RDF resource lifecycle
func testCompleteRDFResourceLifecycle(t *testing.T, ctx context.Context, service *application.StorageService) {
	// Test data - JSON-LD person resource
	jsonLDData := []byte(`{
		"@context": {
			"name": "http://schema.org/name",
			"email": "http://schema.org/email",
			"Person": "http://schema.org/Person"
		},
		"@type": "Person",
		"name": "John Doe",
		"email": "john.doe@example.com"
	}`)

	resourceID := "person-john-doe"

	// Step 1: Store the resource
	storedResource, err := service.StoreResource(ctx, resourceID, jsonLDData, "application/ld+json")
	require.NoError(t, err)
	assert.NotNil(t, storedResource)
	assert.Equal(t, resourceID, storedResource.ID())
	assert.Equal(t, "application/ld+json", storedResource.GetContentType())
	assert.Equal(t, jsonLDData, storedResource.GetData())

	// Step 2: Verify resource exists
	exists, err := service.ResourceExists(ctx, resourceID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Step 3: Retrieve the resource in original format
	retrievedResource, err := service.RetrieveResource(ctx, resourceID, "application/ld+json")
	require.NoError(t, err)
	assert.NotNil(t, retrievedResource)
	assert.Equal(t, resourceID, retrievedResource.ID())
	assert.Equal(t, "application/ld+json", retrievedResource.GetContentType())
	assert.Equal(t, jsonLDData, retrievedResource.GetData())

	// Step 4: Retrieve the resource in different format (Turtle)
	turtleResource, err := service.RetrieveResource(ctx, resourceID, "text/turtle")
	require.NoError(t, err)
	assert.NotNil(t, turtleResource)
	assert.Equal(t, resourceID, turtleResource.ID())
	assert.Equal(t, "text/turtle", turtleResource.GetContentType())
	assert.NotEqual(t, jsonLDData, turtleResource.GetData())         // Should be converted
	assert.Contains(t, string(turtleResource.GetData()), "John Doe") // Should contain semantic content

	// Step 5: Retrieve the resource in RDF/XML format
	rdfXMLResource, err := service.RetrieveResource(ctx, resourceID, "application/rdf+xml")
	require.NoError(t, err)
	assert.NotNil(t, rdfXMLResource)
	assert.Equal(t, resourceID, rdfXMLResource.ID())
	assert.Equal(t, "application/rdf+xml", rdfXMLResource.GetContentType())
	assert.Contains(t, string(rdfXMLResource.GetData()), "John Doe") // Should contain semantic content

	// Step 6: Update the resource
	updatedJSONLDData := []byte(`{
		"@context": {
			"name": "http://schema.org/name",
			"email": "http://schema.org/email",
			"Person": "http://schema.org/Person"
		},
		"@type": "Person",
		"name": "John Smith",
		"email": "john.smith@example.com"
	}`)

	updatedResource, err := service.StoreResource(ctx, resourceID, updatedJSONLDData, "application/ld+json")
	require.NoError(t, err)
	assert.NotNil(t, updatedResource)
	assert.Equal(t, updatedJSONLDData, updatedResource.GetData())

	// Step 7: Verify update
	retrievedUpdated, err := service.RetrieveResource(ctx, resourceID, "application/ld+json")
	require.NoError(t, err)
	assert.Equal(t, updatedJSONLDData, retrievedUpdated.GetData())
	assert.Contains(t, string(retrievedUpdated.GetData()), "John Smith")

	// Step 8: Delete the resource
	err = service.DeleteResource(ctx, resourceID)
	require.NoError(t, err)

	// Step 9: Verify resource is deleted
	exists, err = service.ResourceExists(ctx, resourceID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Step 10: Verify retrieval fails after deletion
	_, err = service.RetrieveResource(ctx, resourceID, "application/ld+json")
	assert.Error(t, err)
	assert.True(t, domain.IsResourceNotFound(err))
}

// Test complete binary resource lifecycle
func testCompleteBinaryResourceLifecycle(t *testing.T, ctx context.Context, service *application.StorageService) {
	// Test data - binary image data (PNG header)
	binaryData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 pixel
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // RGB, no compression
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, // compressed data
		0x00, 0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, // checksum
		0xBC, 0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, // IEND chunk
		0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	resourceID := "test-image.png"

	// Step 1: Store the binary resource
	storedResource, err := service.StoreResource(ctx, resourceID, binaryData, "image/png")
	require.NoError(t, err)
	assert.NotNil(t, storedResource)
	assert.Equal(t, resourceID, storedResource.ID())
	assert.Equal(t, "image/png", storedResource.GetContentType())
	assert.Equal(t, binaryData, storedResource.GetData())

	// Step 2: Verify resource exists
	exists, err := service.ResourceExists(ctx, resourceID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Step 3: Retrieve the binary resource
	retrievedResource, err := service.RetrieveResource(ctx, resourceID, "")
	require.NoError(t, err)
	assert.NotNil(t, retrievedResource)
	assert.Equal(t, resourceID, retrievedResource.ID())
	assert.Equal(t, "image/png", retrievedResource.GetContentType())
	assert.Equal(t, binaryData, retrievedResource.GetData())

	// Step 4: Test streaming access
	reader, contentType, err := service.StreamResource(ctx, resourceID, "")
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)

	streamedData := make([]byte, len(binaryData))
	n, err := reader.Read(streamedData)
	require.NoError(t, err)
	assert.Equal(t, len(binaryData), n)
	assert.Equal(t, binaryData, streamedData)
	reader.Close()

	// Step 5: Update the binary resource
	updatedBinaryData := append(binaryData, []byte{0xFF, 0xFF, 0xFF, 0xFF}...)
	updatedResource, err := service.StoreResource(ctx, resourceID, updatedBinaryData, "image/png")
	require.NoError(t, err)
	assert.Equal(t, updatedBinaryData, updatedResource.GetData())

	// Step 6: Verify update
	retrievedUpdated, err := service.RetrieveResource(ctx, resourceID, "")
	require.NoError(t, err)
	assert.Equal(t, updatedBinaryData, retrievedUpdated.GetData())

	// Step 7: Delete the resource
	err = service.DeleteResource(ctx, resourceID)
	require.NoError(t, err)

	// Step 8: Verify resource is deleted
	exists, err = service.ResourceExists(ctx, resourceID)
	require.NoError(t, err)
	assert.False(t, exists)
}

// Test multiple resources workflow
func testMultipleResourcesWorkflow(t *testing.T, ctx context.Context, service *application.StorageService) {
	// Create multiple resources of different types
	resources := map[string]struct {
		data        []byte
		contentType string
	}{
		"person-1": {
			data: []byte(`{
				"@context": "http://schema.org",
				"@type": "Person",
				"name": "Alice Johnson"
			}`),
			contentType: "application/ld+json",
		},
		"document-1": {
			data: []byte(`@prefix ex: <http://example.org/> .
ex:document ex:title "Test Document" ;
             ex:author "Bob Smith" .`),
			contentType: "text/turtle",
		},
		"image-1": {
			data:        []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			contentType: "image/png",
		},
	}

	// Step 1: Store all resources
	storedResources := make(map[string]*domain.Resource)
	for id, resource := range resources {
		stored, err := service.StoreResource(ctx, id, resource.data, resource.contentType)
		require.NoError(t, err)
		storedResources[id] = stored
	}

	// Step 2: Verify all resources exist
	for id := range resources {
		exists, err := service.ResourceExists(ctx, id)
		require.NoError(t, err)
		assert.True(t, exists, "Resource %s should exist", id)
	}

	// Step 3: Retrieve all resources and verify content
	for id, expectedResource := range resources {
		retrieved, err := service.RetrieveResource(ctx, id, "")
		require.NoError(t, err)
		assert.Equal(t, expectedResource.data, retrieved.GetData())
		assert.Equal(t, expectedResource.contentType, retrieved.GetContentType())
	}

	// Step 4: Test format conversion for RDF resources
	rdfResources := []string{"person-1", "document-1"}
	for _, id := range rdfResources {
		// Convert to different formats
		formats := []string{"application/ld+json", "text/turtle", "application/rdf+xml"}
		for _, format := range formats {
			converted, err := service.RetrieveResource(ctx, id, format)
			require.NoError(t, err)
			assert.Equal(t, format, converted.GetContentType())
			assert.NotEmpty(t, converted.GetData())
		}
	}

	// Step 5: Delete all resources
	for id := range resources {
		err := service.DeleteResource(ctx, id)
		require.NoError(t, err)
	}

	// Step 6: Verify all resources are deleted
	for id := range resources {
		exists, err := service.ResourceExists(ctx, id)
		require.NoError(t, err)
		assert.False(t, exists, "Resource %s should be deleted", id)
	}
}

// Test resource update workflow
func testResourceUpdateWorkflow(t *testing.T, ctx context.Context, service *application.StorageService) {
	resourceID := "updateable-resource"

	// Initial data
	initialData := []byte(`{
		"@context": "http://schema.org",
		"@type": "Article",
		"title": "Initial Title",
		"version": 1
	}`)

	// Step 1: Create initial resource
	initial, err := service.StoreResource(ctx, resourceID, initialData, "application/ld+json")
	require.NoError(t, err)
	assert.Contains(t, string(initial.GetData()), "Initial Title")

	// Step 2: Update the resource multiple times
	updates := [][]byte{
		[]byte(`{
			"@context": "http://schema.org",
			"@type": "Article",
			"title": "Updated Title v2",
			"version": 2
		}`),
		[]byte(`{
			"@context": "http://schema.org",
			"@type": "Article",
			"title": "Final Title v3",
			"version": 3,
			"author": "John Doe"
		}`),
	}

	for i, updateData := range updates {
		updated, err := service.StoreResource(ctx, resourceID, updateData, "application/ld+json")
		require.NoError(t, err)
		assert.Equal(t, updateData, updated.GetData())

		// Verify the update persisted
		retrieved, err := service.RetrieveResource(ctx, resourceID, "application/ld+json")
		require.NoError(t, err)
		assert.Equal(t, updateData, retrieved.GetData())
		assert.Contains(t, string(retrieved.GetData()), fmt.Sprintf("version\": %d", i+2))
	}

	// Step 3: Verify final state
	final, err := service.RetrieveResource(ctx, resourceID, "application/ld+json")
	require.NoError(t, err)
	assert.Contains(t, string(final.GetData()), "Final Title v3")
	assert.Contains(t, string(final.GetData()), "John Doe")
}
