package application

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
)

func TestStorageServiceIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "storage_service_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real infrastructure components
	repo, err := infrastructure.NewFileSystemRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	converter := infrastructure.NewRDFConverter()

	// Create database and event infrastructure
	db, err := infrastructure.DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	eventStore, err := infrastructure.EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	if err != nil {
		t.Fatalf("Failed to create event dispatcher: %v", err)
	}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create storage service (using hybrid approach - direct repository updates + events)
	service := NewStorageService(repo, converter, unitOfWorkFactory)

	ctx := context.Background()

	// Test storing a resource
	testData := []byte(`{
		"@context": "http://schema.org",
		"@type": "Person",
		"name": "John Doe",
		"email": "john@example.com"
	}`)

	resource, err := service.StoreResource(ctx, "integration-test", testData, "application/ld+json")
	if err != nil {
		t.Fatalf("Failed to store resource: %v", err)
	}

	if resource.ID() != "integration-test" {
		t.Errorf("Expected resource ID 'integration-test', got %s", resource.ID())
	}

	// Test retrieving the resource
	retrievedResource, err := service.RetrieveResource(ctx, "integration-test", "")
	if err != nil {
		t.Fatalf("Failed to retrieve resource: %v", err)
	}

	if string(retrievedResource.GetData()) != string(testData) {
		t.Errorf("Retrieved data doesn't match stored data")
	}

	// Test format conversion
	convertedResource, err := service.RetrieveResource(ctx, "integration-test", "text/turtle")
	if err != nil {
		t.Fatalf("Failed to retrieve resource with format conversion: %v", err)
	}

	if convertedResource.GetContentType() != "text/turtle" {
		t.Errorf("Expected content type 'text/turtle', got %s", convertedResource.GetContentType())
	}

	// Test resource existence
	exists, err := service.ResourceExists(ctx, "integration-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should exist")
	}

	// Test deleting the resource
	err = service.DeleteResource(ctx, "integration-test")
	if err != nil {
		t.Fatalf("Failed to delete resource: %v", err)
	}

	// Verify resource no longer exists
	exists, err = service.ResourceExists(ctx, "integration-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence after deletion: %v", err)
	}
	if exists {
		t.Error("Resource should not exist after deletion")
	}

	// Verify file was actually deleted from filesystem
	resourcePath := filepath.Join(tempDir, "integration-test")
	if _, err := os.Stat(resourcePath); !os.IsNotExist(err) {
		t.Error("Resource file should have been deleted from filesystem")
	}
}

func TestStorageServiceStreamingIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "storage_service_stream_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real infrastructure components
	repo, err := infrastructure.NewFileSystemRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	converter := infrastructure.NewRDFConverter()

	// Create database and event infrastructure
	db, err := infrastructure.DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	eventStore, err := infrastructure.EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	if err != nil {
		t.Fatalf("Failed to create event dispatcher: %v", err)
	}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create storage service
	service := NewStorageService(repo, converter, unitOfWorkFactory)

	ctx := context.Background()

	// Test large data streaming
	largeData := make([]byte, 1024*1024) // 1MB of data
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// Store large resource
	_, err = service.StoreResource(ctx, "large-resource", largeData, "application/octet-stream")
	if err != nil {
		t.Fatalf("Failed to store large resource: %v", err)
	}

	// Test streaming retrieval
	reader, contentType, err := service.StreamResource(ctx, "large-resource", "")
	if err != nil {
		t.Fatalf("Failed to stream resource: %v", err)
	}
	defer reader.Close()

	if contentType != "application/octet-stream" {
		t.Errorf("Expected content type 'application/octet-stream', got %s", contentType)
	}

	// Read streamed data
	streamedData := make([]byte, len(largeData))
	n, err := reader.Read(streamedData)
	if err != nil {
		t.Fatalf("Failed to read streamed data: %v", err)
	}

	if n != len(largeData) {
		t.Errorf("Expected to read %d bytes, got %d", len(largeData), n)
	}

	// Verify data integrity
	for i, b := range streamedData {
		if b != largeData[i] {
			t.Errorf("Data mismatch at byte %d: expected %d, got %d", i, largeData[i], b)
			break
		}
	}
}

func TestStorageServiceErrorHandlingIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "storage_service_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create real infrastructure components
	repo, err := infrastructure.NewFileSystemRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	converter := infrastructure.NewRDFConverter()

	// Create database and event infrastructure
	db, err := infrastructure.DatabaseProvider()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	eventStore, err := infrastructure.EventStoreProvider(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	if err != nil {
		t.Fatalf("Failed to create event dispatcher: %v", err)
	}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create storage service
	service := NewStorageService(repo, converter, unitOfWorkFactory)

	ctx := context.Background()

	// Test retrieving non-existent resource
	_, err = service.RetrieveResource(ctx, "nonexistent", "")
	if err == nil {
		t.Fatal("Expected error when retrieving non-existent resource")
	}

	if !domain.IsResourceNotFound(err) {
		t.Errorf("Expected ResourceNotFound error, got %T: %v", err, err)
	}

	// Test deleting non-existent resource
	err = service.DeleteResource(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent resource")
	}

	if !domain.IsResourceNotFound(err) {
		t.Errorf("Expected ResourceNotFound error, got %T: %v", err, err)
	}

	// Test invalid format conversion
	testData := []byte("not valid RDF data")
	_, err = service.StoreResource(ctx, "invalid-rdf", testData, "application/ld+json")
	if err != nil {
		t.Fatalf("Failed to store resource: %v", err)
	}

	// Try to convert to unsupported format
	_, err = service.RetrieveResource(ctx, "invalid-rdf", "application/n-triples")
	if err == nil {
		t.Fatal("Expected error when converting to unsupported format")
	}

	if !domain.IsFormatConversion(err) && !domain.IsUnsupportedFormat(err) {
		t.Errorf("Expected UnsupportedFormat or FormatConversion error, got %T: %v", err, err)
	}
}
