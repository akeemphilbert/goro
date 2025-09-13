package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareIntegration(t *testing.T) {
	// Test that all middleware work together properly
	var logBuffer bytes.Buffer
	logger := &testLogger{buffer: &logBuffer}

	// Create middleware chain
	corsFilter := CORS()
	timeoutMiddleware := Timeout(100 * time.Millisecond)
	loggingMiddleware := StructuredLogging(logger)

	// Create a test handler that takes some time
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify correlation ID is set
		correlationID := GetCorrelationID(ctx)
		assert.NotEmpty(t, correlationID)

		// Verify request info is available
		requestInfo := GetRequestInfo(ctx)
		assert.Equal(t, "GET", requestInfo.Method)
		assert.Equal(t, "/integration-test", requestInfo.Path)

		// Simulate some work (but within timeout)
		time.Sleep(50 * time.Millisecond)

		return "integration success", nil
	}

	// Chain all middleware
	wrappedHandler := timeoutMiddleware(loggingMiddleware(handler))

	// Create HTTP request for CORS testing
	req := httptest.NewRequest("GET", "/integration-test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// Test CORS filter
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create context with request info for middleware testing
		ctx := context.Background()
		ctx = WithRequestInfo(ctx, RequestInfo{
			Method: r.Method,
			Path:   r.URL.Path,
		})

		// Execute middleware chain
		result, err := wrappedHandler(ctx, "integration-test-request")

		require.NoError(t, err)
		assert.Equal(t, "integration success", result)

		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	corsHandler := corsFilter(next)
	corsHandler.ServeHTTP(w, req)

	// Verify CORS headers were set
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Equal(t, 200, w.Code)

	// Verify logging occurred
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "correlation_id")
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "path=/integration-test")
	assert.Contains(t, logOutput, "duration")
	assert.Contains(t, logOutput, "INFO")
}

func TestMiddlewareIntegrationWithTimeout(t *testing.T) {
	// Test timeout middleware with other middleware
	var logBuffer bytes.Buffer
	logger := &testLogger{buffer: &logBuffer}

	// Create middleware chain with short timeout
	timeoutMiddleware := Timeout(50 * time.Millisecond)
	loggingMiddleware := StructuredLogging(logger)

	// Create a handler that exceeds timeout and respects context cancellation
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// This should timeout - use select to respect context cancellation
		select {
		case <-time.After(100 * time.Millisecond):
			return "should not reach here", nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Chain middleware
	wrappedHandler := timeoutMiddleware(loggingMiddleware(handler))

	// Create context with request info
	ctx := context.Background()
	ctx = WithRequestInfo(ctx, RequestInfo{
		Method: "POST",
		Path:   "/timeout-test",
	})

	// Execute and expect timeout
	start := time.Now()
	result, err := wrappedHandler(ctx, "timeout-test-request")
	duration := time.Since(start)

	// Debug output
	t.Logf("Duration: %v, Error: %v, Result: %v", duration, err, result)

	// Verify timeout occurred
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Nil(t, result)
	assert.Less(t, duration, 80*time.Millisecond, "Should timeout before handler completes")

	// Verify error was logged
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "ERROR")
	assert.Contains(t, logOutput, "context deadline exceeded")
	assert.Contains(t, logOutput, "method=POST")
	assert.Contains(t, logOutput, "path=/timeout-test")
}

func TestMiddlewareIntegrationWithCORSPreflight(t *testing.T) {
	// Test CORS preflight with other middleware
	var logBuffer bytes.Buffer
	logger := &testLogger{buffer: &logBuffer}

	corsFilter := CORS()
	loggingMiddleware := StructuredLogging(logger)

	// Create handler (should not be called for OPTIONS)
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "should not be called", nil
	}

	wrappedHandler := loggingMiddleware(handler)

	// Create OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/preflight-test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be called for OPTIONS requests
		ctx := context.Background()
		ctx = WithRequestInfo(ctx, RequestInfo{
			Method: r.Method,
			Path:   r.URL.Path,
		})

		wrappedHandler(ctx, "preflight-test-request")
		w.WriteHeader(200)
	})

	corsHandler := corsFilter(next)
	corsHandler.ServeHTTP(w, req)

	// Verify CORS preflight response
	assert.Equal(t, 204, w.Code) // No Content for OPTIONS
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
	assert.False(t, handlerCalled, "Handler should not be called for OPTIONS requests")
}
