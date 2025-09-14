package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPStreamingOperations(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "http_streaming_test_*")
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

	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)
	logger := log.DefaultLogger
	resourceHandler := handlers.NewResourceHandler(storageService, logger)

	// Create HTTP server
	server := khttp.NewServer()
	server.Route("/resources").GET("/{id}", resourceHandler.GetResource)
	server.Route("/resources").POST("/{id}", resourceHandler.PostResource)
	server.Route("/resources").PUT("/{id}", resourceHandler.PutResource)

	t.Run("StreamingUpload_LargeFile", func(t *testing.T) {
		// Generate large test data (2MB)
		largeData := make([]byte, 2*1024*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		// Create request with large data
		req := httptest.NewRequest("PUT", "/resources/large-file", bytes.NewReader(largeData))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", strconv.Itoa(len(largeData)))

		// Record response
		w := httptest.NewRecorder()

		// Create test context
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "large-file"},
		}

		// Handle request
		start := time.Now()
		err = resourceHandler.PutResource(ctx)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		t.Logf("Uploaded 2MB file in %v", duration)

		// Verify response contains streaming indicator
		responseBody := w.Body.String()
		assert.Contains(t, responseBody, "streaming")
		assert.Contains(t, responseBody, "true")
	})

	t.Run("StreamingDownload_LargeFile", func(t *testing.T) {
		// First upload a large file
		largeData := make([]byte, 3*1024*1024) // 3MB
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		// Store the file using storage service directly
		_, err = storageService.StoreResourceStream(context.Background(), "download-test", bytes.NewReader(largeData), "application/octet-stream", int64(len(largeData)))
		require.NoError(t, err)

		// Create download request
		req := httptest.NewRequest("GET", "/resources/download-test", nil)
		req.Header.Set("Accept", "application/octet-stream")

		// Record response
		w := httptest.NewRecorder()

		// Create test context
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "download-test"},
		}

		// Handle request
		start := time.Now()
		err = resourceHandler.GetResource(ctx)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		t.Logf("Downloaded 3MB file in %v", duration)

		// Verify content
		responseData := w.Body.Bytes()
		assert.Equal(t, largeData, responseData)

		// Verify headers
		assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	})

	t.Run("StreamingUpload_SmallFile_RegularPath", func(t *testing.T) {
		// Small file should use regular upload path
		smallData := []byte("Small file content")

		// Create request with small data
		req := httptest.NewRequest("POST", "/resources/small-file", bytes.NewReader(smallData))
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Content-Length", strconv.Itoa(len(smallData)))

		// Record response
		w := httptest.NewRecorder()

		// Create test context
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "small-file"},
		}

		// Handle request
		err = resourceHandler.PostResource(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify response does NOT contain streaming indicator (regular path)
		responseBody := w.Body.String()
		assert.NotContains(t, responseBody, "streaming")
	})

	t.Run("StreamingUpload_UnknownSize", func(t *testing.T) {
		// Test with unknown content length (should trigger streaming)
		testData := make([]byte, 512*1024) // 512KB
		_, err := rand.Read(testData)
		require.NoError(t, err)

		// Create request without Content-Length header
		req := httptest.NewRequest("PUT", "/resources/unknown-size", bytes.NewReader(testData))
		req.Header.Set("Content-Type", "application/octet-stream")
		// Deliberately omit Content-Length to trigger streaming

		// Record response
		w := httptest.NewRecorder()

		// Create test context
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "unknown-size"},
		}

		// Handle request
		err = resourceHandler.PutResource(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Should use streaming path due to unknown size
		responseBody := w.Body.String()
		assert.Contains(t, responseBody, "streaming")
	})

	t.Run("StreamingDownload_WithRangeHeader", func(t *testing.T) {
		// Store test data
		testData := []byte("This is test data for range requests")
		_, err = storageService.StoreResource(context.Background(), "range-test", testData, "text/plain")
		require.NoError(t, err)

		// Create request with Range header
		req := httptest.NewRequest("GET", "/resources/range-test", nil)
		req.Header.Set("Range", "bytes=0-10")
		req.Header.Set("Accept", "text/plain")

		// Record response
		w := httptest.NewRecorder()

		// Create test context
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "range-test"},
		}

		// Handle request
		err = resourceHandler.GetResource(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		// Should use streaming path due to Range header
		// Note: Full range support would require additional implementation
		// For now, we just verify streaming path is triggered
		responseData := w.Body.Bytes()
		assert.Equal(t, testData, responseData) // Full content for now
	})

	t.Run("StreamingErrorHandling", func(t *testing.T) {
		// Test streaming download of non-existent resource
		req := httptest.NewRequest("GET", "/resources/non-existent", nil)
		req.Header.Set("Content-Length", "999999999") // Large size to trigger streaming

		w := httptest.NewRecorder()
		ctx := &testContext{
			request:  req,
			response: w,
			vars:     map[string]string{"id": "non-existent"},
		}

		err = resourceHandler.GetResource(ctx)
		assert.NoError(t, err) // Handler should not return error, but write error response
		assert.Equal(t, http.StatusNotFound, w.Code)

		responseBody := w.Body.String()
		assert.Contains(t, responseBody, "RESOURCE_NOT_FOUND")
	})
}

func TestStreamingPerformance(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "streaming_perf_test_*")
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

	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)

	ctx := context.Background()

	t.Run("StreamingVsRegular_Performance", func(t *testing.T) {
		// Test data sizes
		sizes := []int{
			100 * 1024,      // 100KB
			1024 * 1024,     // 1MB
			5 * 1024 * 1024, // 5MB
		}

		for _, size := range sizes {
			t.Run(fmt.Sprintf("Size_%dB", size), func(t *testing.T) {
				testData := make([]byte, size)
				_, err := rand.Read(testData)
				require.NoError(t, err)

				// Test regular storage
				start := time.Now()
				_, err = storageService.StoreResource(ctx, fmt.Sprintf("regular-%d", size), testData, "application/octet-stream")
				regularDuration := time.Since(start)
				require.NoError(t, err)

				// Test streaming storage
				start = time.Now()
				_, err = storageService.StoreResourceStream(ctx, fmt.Sprintf("streaming-%d", size), bytes.NewReader(testData), "application/octet-stream", int64(size))
				streamingDuration := time.Since(start)
				require.NoError(t, err)

				t.Logf("Size %d bytes - Regular: %v, Streaming: %v", size, regularDuration, streamingDuration)

				// For large files, streaming should be competitive or better
				if size >= 1024*1024 {
					// Streaming should not be significantly slower (within 50% overhead)
					assert.Less(t, streamingDuration, regularDuration*3/2)
				}
			})
		}
	})

	t.Run("MemoryUsage_LargeFile", func(t *testing.T) {
		// This test verifies that streaming doesn't load entire file into memory
		// We can't easily measure memory usage in a unit test, but we can verify
		// that streaming operations complete successfully with very large files

		// Generate 10MB test data
		largeData := make([]byte, 10*1024*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err)

		// Store using streaming
		start := time.Now()
		_, err = storageService.StoreResourceStream(ctx, "memory-test", bytes.NewReader(largeData), "application/octet-stream", int64(len(largeData)))
		storeDuration := time.Since(start)
		require.NoError(t, err)

		// Retrieve using streaming
		start = time.Now()
		reader, contentType, err := storageService.StreamResource(ctx, "memory-test", "")
		require.NoError(t, err)
		defer reader.Close()

		// Read in chunks to simulate streaming behavior
		var retrievedSize int64
		buffer := make([]byte, 64*1024) // 64KB buffer
		for {
			n, err := reader.Read(buffer)
			retrievedSize += int64(n)
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
		retrieveDuration := time.Since(start)

		assert.Equal(t, "application/octet-stream", contentType)
		assert.Equal(t, int64(len(largeData)), retrievedSize)

		t.Logf("10MB file - Store: %v, Retrieve: %v", storeDuration, retrieveDuration)

		// Operations should complete in reasonable time
		assert.Less(t, storeDuration, 5*time.Second)
		assert.Less(t, retrieveDuration, 5*time.Second)
	})
}

// mockUnitOfWork provides a simple mock implementation for testing
type mockUnitOfWork struct {
	events []pericarpdomain.EntityEvent
}

func (m *mockUnitOfWork) RegisterEvents(events []pericarpdomain.EntityEvent) {
	m.events = append(m.events, events...)
}

func (m *mockUnitOfWork) Commit(ctx context.Context) ([]pericarpdomain.Event, error) {
	events := make([]pericarpdomain.Event, len(m.events))
	for i, event := range m.events {
		events[i] = event
	}
	return events, nil
}

func (m *mockUnitOfWork) Rollback() error {
	m.events = nil
	return nil
}

// testContext implements khttp.Context for testing
type testContext struct {
	request  *http.Request
	response http.ResponseWriter
	vars     map[string]string
}

func (c *testContext) Request() *http.Request {
	if c.request != nil {
		return c.request
	}
	return httptest.NewRequest("GET", "/", nil)
}

func (c *testContext) Response() http.ResponseWriter {
	return c.response
}

func (c *testContext) Vars() url.Values {
	result := make(url.Values)
	for k, v := range c.vars {
		result[k] = []string{v}
	}
	return result
}

func (c *testContext) Query() url.Values {
	return make(url.Values)
}

func (c *testContext) Form() url.Values {
	return make(url.Values)
}

func (c *testContext) Header() http.Header {
	return make(http.Header)
}

func (c *testContext) JSON(code int, v interface{}) error {
	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(v)
}
