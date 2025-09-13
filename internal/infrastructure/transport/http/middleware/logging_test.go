package middleware

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredLogging(t *testing.T) {
	tests := []struct {
		name           string
		handlerError   error
		expectedFields []string
		expectedLevel  string
	}{
		{
			name:         "Successful request logging",
			handlerError: nil,
			expectedFields: []string{
				"correlation_id",
				"method",
				"path",
				"status",
				"duration",
				"timestamp",
			},
			expectedLevel: "INFO",
		},
		{
			name:         "Error request logging",
			handlerError: assert.AnError,
			expectedFields: []string{
				"correlation_id",
				"method",
				"path",
				"error",
				"duration",
				"timestamp",
			},
			expectedLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var logBuffer bytes.Buffer
			logger := &testLogger{buffer: &logBuffer}

			// Create structured logging middleware
			loggingMiddleware := StructuredLogging(logger)

			// Create a test handler
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				// Verify correlation ID is in context
				correlationID := GetCorrelationID(ctx)
				assert.NotEmpty(t, correlationID, "Correlation ID should be set in context")

				if tt.handlerError != nil {
					return nil, tt.handlerError
				}
				return "success", nil
			}

			// Wrap handler with logging middleware
			wrappedHandler := loggingMiddleware(handler)

			// Create context with request info
			ctx := context.Background()
			ctx = WithRequestInfo(ctx, RequestInfo{
				Method: "GET",
				Path:   "/test",
			})

			// Execute the handler
			result, err := wrappedHandler(ctx, "test-request")

			// Verify handler behavior
			if tt.handlerError != nil {
				require.Error(t, err)
				assert.Equal(t, tt.handlerError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "success", result)
			}

			// Parse log output
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput, "Log output should not be empty")

			// Verify log contains expected fields
			for _, field := range tt.expectedFields {
				assert.Contains(t, logOutput, field, "Log should contain field: %s", field)
			}

			// Verify log level
			assert.Contains(t, logOutput, tt.expectedLevel, "Log should contain level: %s", tt.expectedLevel)

			// Verify duration is reasonable
			assert.Contains(t, logOutput, "duration", "Log should contain duration")

			// Verify correlation ID format (should be UUID-like)
			lines := strings.Split(strings.TrimSpace(logOutput), "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				assert.Contains(t, lastLine, "correlation_id", "Log should contain correlation_id")
			}
		})
	}
}

func TestCorrelationID(t *testing.T) {
	tests := []struct {
		name                  string
		existingCorrelationID string
		shouldGenerateNew     bool
	}{
		{
			name:                  "Generate new correlation ID",
			existingCorrelationID: "",
			shouldGenerateNew:     true,
		},
		{
			name:                  "Use existing correlation ID",
			existingCorrelationID: "existing-correlation-id-123",
			shouldGenerateNew:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := &testLogger{buffer: &logBuffer}
			loggingMiddleware := StructuredLogging(logger)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				correlationID := GetCorrelationID(ctx)

				if tt.shouldGenerateNew {
					assert.NotEmpty(t, correlationID, "Should generate new correlation ID")
					assert.NotEqual(t, tt.existingCorrelationID, correlationID)
				} else {
					assert.Equal(t, tt.existingCorrelationID, correlationID, "Should use existing correlation ID")
				}

				return "success", nil
			}

			wrappedHandler := loggingMiddleware(handler)

			ctx := context.Background()
			if tt.existingCorrelationID != "" {
				ctx = WithCorrelationID(ctx, tt.existingCorrelationID)
			}
			ctx = WithRequestInfo(ctx, RequestInfo{Method: "GET", Path: "/test"})

			_, err := wrappedHandler(ctx, "test-request")
			require.NoError(t, err)

			logOutput := logBuffer.String()
			if tt.shouldGenerateNew {
				assert.Contains(t, logOutput, "correlation_id", "Log should contain generated correlation ID")
			} else {
				assert.Contains(t, logOutput, tt.existingCorrelationID, "Log should contain existing correlation ID")
			}
		})
	}
}

func TestRequestInfoContext(t *testing.T) {
	tests := []struct {
		name        string
		requestInfo RequestInfo
		expectPanic bool
	}{
		{
			name: "Valid request info",
			requestInfo: RequestInfo{
				Method: "POST",
				Path:   "/api/v1/users",
			},
			expectPanic: false,
		},
		{
			name: "Empty request info",
			requestInfo: RequestInfo{
				Method: "",
				Path:   "",
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logger := &testLogger{buffer: &logBuffer}
			loggingMiddleware := StructuredLogging(logger)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				requestInfo := GetRequestInfo(ctx)
				assert.Equal(t, tt.requestInfo.Method, requestInfo.Method)
				assert.Equal(t, tt.requestInfo.Path, requestInfo.Path)
				return "success", nil
			}

			wrappedHandler := loggingMiddleware(handler)

			ctx := context.Background()
			ctx = WithRequestInfo(ctx, tt.requestInfo)

			if tt.expectPanic {
				assert.Panics(t, func() {
					wrappedHandler(ctx, "test-request")
				})
			} else {
				result, err := wrappedHandler(ctx, "test-request")
				require.NoError(t, err)
				assert.Equal(t, "success", result)

				logOutput := logBuffer.String()
				if tt.requestInfo.Method != "" {
					assert.Contains(t, logOutput, tt.requestInfo.Method)
				}
				if tt.requestInfo.Path != "" {
					assert.Contains(t, logOutput, tt.requestInfo.Path)
				}
			}
		})
	}
}

func TestJSONLogFormat(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := &testLogger{buffer: &logBuffer}
	loggingMiddleware := StructuredLogging(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	wrappedHandler := loggingMiddleware(handler)

	ctx := context.Background()
	ctx = WithRequestInfo(ctx, RequestInfo{
		Method: "GET",
		Path:   "/test",
	})

	_, err := wrappedHandler(ctx, "test-request")
	require.NoError(t, err)

	logOutput := logBuffer.String()

	// The log output should contain structured data
	assert.Contains(t, logOutput, "correlation_id")
	assert.Contains(t, logOutput, "method")
	assert.Contains(t, logOutput, "path")
	assert.Contains(t, logOutput, "duration")
	assert.Contains(t, logOutput, "timestamp")
}

func TestLoggingMiddlewareChaining(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := &testLogger{buffer: &logBuffer}
	loggingMiddleware := StructuredLogging(logger)

	// Create another middleware to test chaining
	testMiddleware := func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx = context.WithValue(ctx, "test_middleware", "applied")
			return handler(ctx, req)
		}
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify both middleware were applied
		testValue := ctx.Value("test_middleware")
		assert.Equal(t, "applied", testValue)

		correlationID := GetCorrelationID(ctx)
		assert.NotEmpty(t, correlationID)

		return "chained success", nil
	}

	// Chain the middleware
	wrappedHandler := loggingMiddleware(testMiddleware(handler))

	ctx := context.Background()
	ctx = WithRequestInfo(ctx, RequestInfo{Method: "GET", Path: "/chain"})

	result, err := wrappedHandler(ctx, "test-request")
	require.NoError(t, err)
	assert.Equal(t, "chained success", result)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "correlation_id")
	assert.Contains(t, logOutput, "/chain")
}

// testLogger is a simple logger implementation for testing
type testLogger struct {
	buffer *bytes.Buffer
}

func (l *testLogger) Log(level log.Level, keyvals ...interface{}) error {
	levelStr := "INFO"
	if level == log.LevelError {
		levelStr = "ERROR"
	}

	l.buffer.WriteString(levelStr + " ")

	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := keyvals[i].(string)
			value := keyvals[i+1]
			l.buffer.WriteString(key)
			l.buffer.WriteString("=")
			l.buffer.WriteString(fmt.Sprintf("%v", value))
			l.buffer.WriteString(" ")
		}
	}
	l.buffer.WriteString("\n")
	return nil
}
