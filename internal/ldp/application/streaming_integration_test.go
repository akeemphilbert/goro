package application

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService_StreamingOperations(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "storage_service_streaming_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up dependencies
	repo, err := infrastructure.NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	converter := infrastructure.NewRDFConverter()

	// Mock unit of work factory
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &mockUnitOfWork{}
	}

	storageService := NewStorageService(repo, converter, unitOfWorkFactory)
	ctx := context.Background()

	t.Run("StoreResourceStream_LargeFile", func(t *testing.T) {
		// Generate large test data (2MB)
		largeData := make([]byte, 2*1024*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		// Store using streaming
		start := time.Now()
		resource, err := storageService.StoreResourceStream(ctx, "large-stream", bytes.NewReader(largeData), "application/octet-stream", int64(len(largeData)))
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, "large-stream", resource.ID())
		assert.Equal(t, "application/octet-stream", resource.GetContentType())
		assert.Equal(t, len(largeData), resource.GetSize())

		t.Logf("Stored 2MB file via streaming in %v", duration)

		// Verify file exists
		exists, err := storageService.ResourceExists(ctx, "large-stream")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("StreamResource_LargeFile", func(t *testing.T) {
		// First store a large file
		largeData := make([]byte, 3*1024*1024) // 3MB
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		_, err = storageService.StoreResourceStream(ctx, "stream-retrieve", bytes.NewReader(largeData), "application/octet-stream", int64(len(largeData)))
		require.NoError(t, err)

		// Retrieve using streaming
		start := time.Now()
		reader, contentType, err := storageService.StreamResource(ctx, "stream-retrieve", "")
		assert.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, "application/octet-stream", contentType)

		// Read streamed data in chunks
		var retrievedData []byte
		buffer := make([]byte, 64*1024) // 64KB buffer
		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				retrievedData = append(retrievedData, buffer[:n]...)
			}
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
		}
		duration := time.Since(start)

		t.Logf("Streamed 3MB file in %v", duration)

		// Verify content
		assert.Equal(t, largeData, retrievedData)
	})

	t.Run("StreamResource_WithFormatConversion", func(t *testing.T) {
		// Store RDF data
		jsonLDData := []byte(`{
			"@context": "https://www.w3.org/ns/activitystreams",
			"@type": "Note",
			"content": "This is a test note"
		}`)

		_, err := storageService.StoreResource(ctx, "rdf-convert", jsonLDData, "application/ld+json")
		require.NoError(t, err)

		// Stream with format conversion (this will fall back to regular retrieval + conversion)
		reader, contentType, err := storageService.StreamResource(ctx, "rdf-convert", "text/turtle")
		assert.NoError(t, err)
		defer reader.Close()

		// Should get converted format
		assert.Equal(t, "text/turtle", contentType)

		// Read converted data
		convertedData, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.NotEmpty(t, convertedData)
		assert.NotEqual(t, jsonLDData, convertedData) // Should be different due to conversion
	})

	t.Run("StreamingErrorHandling", func(t *testing.T) {
		// Test streaming non-existent resource
		_, _, err := storageService.StreamResource(ctx, "non-existent", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")

		// Test streaming with invalid ID
		_, _, err = storageService.StreamResource(ctx, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "INVALID_ID")

		// Test storing with nil reader
		_, err = storageService.StoreResourceStream(ctx, "test", nil, "text/plain", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "INVALID_RESOURCE")
	})

	t.Run("StreamingPerformanceComparison", func(t *testing.T) {
		// Test data (1MB)
		testData := make([]byte, 1024*1024)
		_, err := rand.Read(testData)
		require.NoError(t, err)

		// Test regular storage
		start := time.Now()
		_, err = storageService.StoreResource(ctx, "perf-regular", testData, "application/octet-stream")
		regularDuration := time.Since(start)
		require.NoError(t, err)

		// Test streaming storage
		start = time.Now()
		_, err = storageService.StoreResourceStream(ctx, "perf-streaming", bytes.NewReader(testData), "application/octet-stream", int64(len(testData)))
		streamingDuration := time.Since(start)
		require.NoError(t, err)

		t.Logf("1MB file - Regular: %v, Streaming: %v", regularDuration, streamingDuration)

		// Both should complete in reasonable time
		assert.Less(t, regularDuration, 2*time.Second)
		assert.Less(t, streamingDuration, 2*time.Second)

		// For 1MB files, performance should be comparable
		// Streaming might have slight overhead but should not be dramatically slower
		assert.Less(t, streamingDuration, regularDuration*2)
	})
}
