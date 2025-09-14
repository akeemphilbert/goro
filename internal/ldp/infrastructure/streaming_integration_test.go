package infrastructure

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemRepository_StreamingOperations(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "streaming_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create repository
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("StoreStream_SmallFile", func(t *testing.T) {
		// Test data
		testData := []byte("Hello, streaming world!")
		reader := bytes.NewReader(testData)

		// Store using streaming
		err := repo.StoreStream(ctx, "test-small", reader, "text/plain", int64(len(testData)))
		assert.NoError(t, err)

		// Verify file exists
		exists, err := repo.Exists(ctx, "test-small")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Retrieve and verify content
		resource, err := repo.Retrieve(ctx, "test-small")
		assert.NoError(t, err)
		assert.Equal(t, testData, resource.GetData())
		assert.Equal(t, "text/plain", resource.GetContentType())
	})

	t.Run("StoreStream_LargeFile", func(t *testing.T) {
		// Generate large test data (5MB)
		largeData := make([]byte, 5*1024*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		reader := bytes.NewReader(largeData)

		// Store using streaming
		start := time.Now()
		err = repo.StoreStream(ctx, "test-large", reader, "application/octet-stream", int64(len(largeData)))
		duration := time.Since(start)

		assert.NoError(t, err)
		t.Logf("Stored 5MB file in %v", duration)

		// Verify file exists
		exists, err := repo.Exists(ctx, "test-large")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Retrieve and verify content
		resource, err := repo.Retrieve(ctx, "test-large")
		assert.NoError(t, err)
		assert.Equal(t, largeData, resource.GetData())
		assert.Equal(t, "application/octet-stream", resource.GetContentType())
		assert.Equal(t, len(largeData), resource.GetSize())
	})

	t.Run("RetrieveStream_SmallFile", func(t *testing.T) {
		// First store a file
		testData := []byte("Stream retrieval test")
		reader := bytes.NewReader(testData)
		err := repo.StoreStream(ctx, "test-retrieve", reader, "text/plain", int64(len(testData)))
		require.NoError(t, err)

		// Retrieve using streaming
		streamReader, metadata, err := repo.RetrieveStream(ctx, "test-retrieve")
		assert.NoError(t, err)
		defer streamReader.Close()

		// Verify metadata
		assert.Equal(t, "test-retrieve", metadata.ID)
		assert.Equal(t, "text/plain", metadata.ContentType)
		assert.Equal(t, int64(len(testData)), metadata.Size)

		// Read streamed data
		retrievedData, err := io.ReadAll(streamReader)
		assert.NoError(t, err)
		assert.Equal(t, testData, retrievedData)
	})

	t.Run("RetrieveStream_LargeFile", func(t *testing.T) {
		// Generate large test data (10MB)
		largeData := make([]byte, 10*1024*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		// Store the large file
		reader := bytes.NewReader(largeData)
		err = repo.StoreStream(ctx, "test-large-retrieve", reader, "application/octet-stream", int64(len(largeData)))
		require.NoError(t, err)

		// Retrieve using streaming
		start := time.Now()
		streamReader, metadata, err := repo.RetrieveStream(ctx, "test-large-retrieve")
		assert.NoError(t, err)
		defer streamReader.Close()

		// Verify metadata
		assert.Equal(t, "test-large-retrieve", metadata.ID)
		assert.Equal(t, "application/octet-stream", metadata.ContentType)
		assert.Equal(t, int64(len(largeData)), metadata.Size)

		// Read streamed data in chunks to simulate real streaming
		var retrievedData []byte
		buffer := make([]byte, 64*1024) // 64KB buffer
		for {
			n, err := streamReader.Read(buffer)
			if n > 0 {
				retrievedData = append(retrievedData, buffer[:n]...)
			}
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		t.Logf("Streamed 10MB file in %v", duration)

		// Verify content
		assert.Equal(t, largeData, retrievedData)
	})

	t.Run("StreamingChecksumValidation", func(t *testing.T) {
		// Store a file
		testData := []byte("Checksum validation test")
		reader := bytes.NewReader(testData)
		err := repo.StoreStream(ctx, "test-checksum", reader, "text/plain", int64(len(testData)))
		require.NoError(t, err)

		// Manually corrupt the content file to test checksum validation
		resourcePath := filepath.Join(tempDir, "resources", "test-checksum", "content")
		corruptedData := []byte("Corrupted data")
		err = os.WriteFile(resourcePath, corruptedData, 0644)
		require.NoError(t, err)

		// Try to retrieve - should fail with checksum error
		streamReader, _, err := repo.RetrieveStream(ctx, "test-checksum")
		require.NoError(t, err) // Opening should succeed
		defer streamReader.Close()

		// Reading should fail with checksum error
		_, err = io.ReadAll(streamReader)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum mismatch")
	})

	t.Run("StreamingErrorHandling", func(t *testing.T) {
		// Test storing with nil reader
		err := repo.StoreStream(ctx, "test-nil", nil, "text/plain", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reader cannot be nil")

		// Test storing with empty ID
		reader := bytes.NewReader([]byte("test"))
		err = repo.StoreStream(ctx, "", reader, "text/plain", 4)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource ID cannot be empty")

		// Test retrieving non-existent resource
		_, _, err = repo.RetrieveStream(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})
}

func TestOptimizedFileSystemRepository_StreamingOperations(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "optimized_streaming_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create optimized repository
	cacheConfig := CacheConfig{
		MaxSize:    50 * 1024 * 1024, // 50MB
		MaxEntries: 100,
		TTL:        5 * time.Minute,
	}
	repo, err := NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("OptimizedStoreStream", func(t *testing.T) {
		// Test data
		testData := []byte("Optimized streaming test")
		reader := bytes.NewReader(testData)

		// Store using streaming
		err := repo.StoreStream(ctx, "test-optimized", reader, "text/plain", int64(len(testData)))
		assert.NoError(t, err)

		// Verify indexing was updated
		stats := repo.indexer.GetStats()
		totalResources, ok := stats["totalResources"].(int)
		assert.True(t, ok)
		assert.Greater(t, totalResources, 0)

		// Verify file exists
		exists, err := repo.Exists(ctx, "test-optimized")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("OptimizedRetrieveStream", func(t *testing.T) {
		// First store a file
		testData := []byte("Optimized stream retrieval test")
		reader := bytes.NewReader(testData)
		err := repo.StoreStream(ctx, "test-opt-retrieve", reader, "text/plain", int64(len(testData)))
		require.NoError(t, err)

		// Retrieve using streaming (should bypass cache)
		streamReader, metadata, err := repo.RetrieveStream(ctx, "test-opt-retrieve")
		assert.NoError(t, err)
		defer streamReader.Close()

		// Verify metadata
		assert.Equal(t, "test-opt-retrieve", metadata.ID)
		assert.Equal(t, "text/plain", metadata.ContentType)
		assert.Equal(t, int64(len(testData)), metadata.Size)

		// Read streamed data
		retrievedData, err := io.ReadAll(streamReader)
		assert.NoError(t, err)
		assert.Equal(t, testData, retrievedData)
	})

	t.Run("StreamingPerformanceComparison", func(t *testing.T) {
		// Generate test data (1MB)
		testData := make([]byte, 1024*1024)
		_, err := rand.Read(testData)
		require.NoError(t, err)

		// Test streaming store performance
		reader := bytes.NewReader(testData)
		start := time.Now()
		err = repo.StoreStream(ctx, "perf-test", reader, "application/octet-stream", int64(len(testData)))
		streamingStoreDuration := time.Since(start)
		assert.NoError(t, err)

		// Test streaming retrieve performance
		start = time.Now()
		streamReader, _, err := repo.RetrieveStream(ctx, "perf-test")
		assert.NoError(t, err)

		retrievedData, err := io.ReadAll(streamReader)
		streamingRetrieveDuration := time.Since(start)
		streamReader.Close()

		assert.NoError(t, err)
		assert.Equal(t, testData, retrievedData)

		t.Logf("Streaming store: %v, retrieve: %v", streamingStoreDuration, streamingRetrieveDuration)

		// Both operations should complete in reasonable time (< 1 second for 1MB)
		assert.Less(t, streamingStoreDuration, time.Second)
		assert.Less(t, streamingRetrieveDuration, time.Second)
	})
}

// Benchmark streaming operations
func BenchmarkStreamingOperations(b *testing.B) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "streaming_bench_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Create repository
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(b, err)

	ctx := context.Background()

	// Test data sizes
	sizes := []int{
		1024,            // 1KB
		64 * 1024,       // 64KB
		1024 * 1024,     // 1MB
		5 * 1024 * 1024, // 5MB
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("StoreStream_%dB", size), func(b *testing.B) {
			testData := make([]byte, size)
			rand.Read(testData)

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(testData)
				err := repo.StoreStream(ctx, fmt.Sprintf("bench-%d-%d", size, i), reader, "application/octet-stream", int64(size))
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("RetrieveStream_%dB", size), func(b *testing.B) {
			// Pre-store test data
			testData := make([]byte, size)
			rand.Read(testData)
			reader := bytes.NewReader(testData)
			err := repo.StoreStream(ctx, fmt.Sprintf("bench-retrieve-%d", size), reader, "application/octet-stream", int64(size))
			require.NoError(b, err)

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				streamReader, _, err := repo.RetrieveStream(ctx, fmt.Sprintf("bench-retrieve-%d", size))
				if err != nil {
					b.Fatal(err)
				}

				_, err = io.ReadAll(streamReader)
				if err != nil {
					b.Fatal(err)
				}
				streamReader.Close()
			}
		})
	}
}
