package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceHandler_HandleStorageError(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	mockService := &MockStorageService{}
	handler := NewResourceHandler(mockService, logger)

	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
		expectedFields []string // Fields that should be present in response
	}{
		{
			name:           "Resource Not Found",
			error:          &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "RESOURCE_NOT_FOUND",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Unsupported Format",
			error:          &domain.StorageError{Code: "UNSUPPORTED_FORMAT", Message: "unsupported"},
			expectedStatus: http.StatusNotAcceptable,
			expectedCode:   "UNSUPPORTED_FORMAT",
			expectedFields: []string{"code", "message", "status", "timestamp", "supportedFormats"},
		},
		{
			name:           "Insufficient Storage",
			error:          &domain.StorageError{Code: "INSUFFICIENT_STORAGE", Message: "insufficient"},
			expectedStatus: http.StatusInsufficientStorage,
			expectedCode:   "INSUFFICIENT_STORAGE",
			expectedFields: []string{"code", "message", "status", "timestamp", "suggestion"},
		},
		{
			name:           "Data Corruption",
			error:          &domain.StorageError{Code: "DATA_CORRUPTION", Message: "corrupted"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "DATA_CORRUPTION",
			expectedFields: []string{"code", "message", "status", "timestamp", "suggestion"},
		},
		{
			name:           "Format Conversion Failed",
			error:          &domain.StorageError{Code: "FORMAT_CONVERSION_FAILED", Message: "conversion failed"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "FORMAT_CONVERSION_FAILED",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Invalid ID",
			error:          &domain.StorageError{Code: "INVALID_ID", Message: "invalid id"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_ID",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Invalid Resource",
			error:          &domain.StorageError{Code: "INVALID_RESOURCE", Message: "invalid resource"},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_RESOURCE",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Resource Exists",
			error:          &domain.StorageError{Code: "RESOURCE_EXISTS", Message: "already exists"},
			expectedStatus: http.StatusConflict,
			expectedCode:   "RESOURCE_EXISTS",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Checksum Mismatch",
			error:          &domain.StorageError{Code: "CHECKSUM_MISMATCH", Message: "checksum failed"},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "CHECKSUM_MISMATCH",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Storage Operation Failed",
			error:          &domain.StorageError{Code: "STORAGE_OPERATION_FAILED", Message: "operation failed"},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "STORAGE_OPERATION_FAILED",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
		{
			name:           "Unexpected Error",
			error:          errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedFields: []string{"code", "message", "status", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := &testContext{response: w}

			err := handler.handleStorageError(ctx, tt.error)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			errorObj, ok := response["error"].(map[string]interface{})
			require.True(t, ok, "Response should have error object")

			assert.Equal(t, tt.expectedCode, errorObj["code"])
			assert.Equal(t, float64(tt.expectedStatus), errorObj["status"])

			// Check that expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, errorObj, field, "Response should contain field: %s", field)
			}

			// Verify timestamp is a valid number
			if timestamp, ok := errorObj["timestamp"].(string); ok {
				assert.NotEmpty(t, timestamp)
			}
		})
	}
}

func TestResourceHandler_WriteDetailedErrorResponse(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	mockService := &MockStorageService{}
	handler := NewResourceHandler(mockService, logger)

	t.Run("Error with context", func(t *testing.T) {
		storageErr := &domain.StorageError{
			Code:      "TEST_ERROR",
			Message:   "test error",
			Operation: "test_operation",
			Context: map[string]any{
				"resourceID":  "test-123",
				"contentType": "application/ld+json",
				"size":        1024,
				"sensitive":   "should-not-appear", // This should be filtered out
			},
		}

		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "TEST_ERROR", "test message", storageErr)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "TEST_ERROR", errorObj["code"])
		assert.Equal(t, "test message", errorObj["message"])
		assert.Equal(t, "test_operation", errorObj["operation"])

		// Check context filtering
		context := errorObj["context"].(map[string]interface{})
		assert.Equal(t, "test-123", context["resourceID"])
		assert.Equal(t, "application/ld+json", context["contentType"])
		assert.Equal(t, float64(1024), context["size"])
		assert.NotContains(t, context, "sensitive")
	})

	t.Run("Unsupported format with supported formats", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusNotAcceptable, "UNSUPPORTED_FORMAT", "unsupported format", nil)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		supportedFormats := errorObj["supportedFormats"].([]interface{})

		assert.Contains(t, supportedFormats, "application/ld+json")
		assert.Contains(t, supportedFormats, "text/turtle")
		assert.Contains(t, supportedFormats, "application/rdf+xml")
	})

	t.Run("Insufficient storage with suggestion", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusInsufficientStorage, "INSUFFICIENT_STORAGE", "insufficient storage", nil)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Contains(t, errorObj["suggestion"], "reducing the size")
	})

	t.Run("Data corruption with suggestion", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusUnprocessableEntity, "DATA_CORRUPTION", "data corruption", nil)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Contains(t, errorObj["suggestion"], "try uploading the resource again")
	})

	t.Run("Response headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "TEST_ERROR", "test message", nil)
		assert.NoError(t, err)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	})
}

func TestResourceHandler_LogError(t *testing.T) {
	// Create a logger that captures output
	var logOutput strings.Builder
	logger := log.NewStdLogger(&logOutput)
	mockService := &MockStorageService{}
	handler := NewResourceHandler(mockService, logger)

	t.Run("Log storage error with context", func(t *testing.T) {
		logOutput.Reset()

		storageErr := &domain.StorageError{
			Code:      "TEST_ERROR",
			Message:   "test error",
			Operation: "test_operation",
			Context: map[string]any{
				"resourceID":  "test-123",
				"contentType": "application/ld+json",
			},
		}

		handler.logError(storageErr, storageErr)

		logStr := logOutput.String()
		assert.Contains(t, logStr, "Storage operation error")
		assert.Contains(t, logStr, "TEST_ERROR")
		assert.Contains(t, logStr, "test_operation")
		assert.Contains(t, logStr, "test-123")
		assert.Contains(t, logStr, "application/ld+json")
	})

	t.Run("Log regular error", func(t *testing.T) {
		logOutput.Reset()

		regularErr := errors.New("regular error")
		handler.logError(regularErr, nil)

		logStr := logOutput.String()
		assert.Contains(t, logStr, "Storage operation error")
		assert.Contains(t, logStr, "regular error")
	})
}

func TestResourceHandler_ShouldExposeCause(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	mockService := &MockStorageService{}
	handler := NewResourceHandler(mockService, logger)

	tests := []struct {
		name     string
		cause    error
		expected bool
	}{
		{
			name:     "Format error should be exposed",
			cause:    errors.New("invalid format specified"),
			expected: true,
		},
		{
			name:     "Parse error should be exposed",
			cause:    errors.New("failed to parse JSON"),
			expected: true,
		},
		{
			name:     "Validation error should be exposed",
			cause:    errors.New("validation failed for field"),
			expected: true,
		},
		{
			name:     "Invalid error should be exposed",
			cause:    errors.New("invalid resource structure"),
			expected: true,
		},
		{
			name:     "Filesystem error should not be exposed",
			cause:    errors.New("permission denied: /var/data"),
			expected: false,
		},
		{
			name:     "Network error should not be exposed",
			cause:    errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "Database error should not be exposed",
			cause:    errors.New("database connection failed"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.shouldExposeCause(tt.cause)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceHandler_GetCurrentTimestamp(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	mockService := &MockStorageService{}
	handler := NewResourceHandler(mockService, logger)

	timestamp := handler.getCurrentTimestamp()
	assert.NotEmpty(t, timestamp)

	// Should be a valid Unix timestamp string
	assert.Regexp(t, `^\d+$`, timestamp)
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

func (c *testContext) XML(code int, v interface{}) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *testContext) String(code int, text string) error {
	c.response.WriteHeader(code)
	_, err := c.response.Write([]byte(text))
	return err
}

func (c *testContext) Blob(code int, contentType string, data []byte) error {
	c.response.Header().Set("Content-Type", contentType)
	c.response.WriteHeader(code)
	_, err := c.response.Write(data)
	return err
}

func (c *testContext) Stream(code int, contentType string, rd io.Reader) error {
	c.response.Header().Set("Content-Type", contentType)
	c.response.WriteHeader(code)
	_, err := io.Copy(c.response, rd)
	return err
}

func (c *testContext) Reset(http.ResponseWriter, *http.Request) {}

func (c *testContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *testContext) Done() <-chan struct{} {
	return nil
}

func (c *testContext) Err() error {
	return nil
}

func (c *testContext) Value(key interface{}) interface{} {
	return nil
}

func (c *testContext) Bind(v interface{}) error {
	return fmt.Errorf("bind not implemented in test context")
}

func (c *testContext) BindForm(v interface{}) error {
	return fmt.Errorf("BindForm not implemented in test context")
}

func (c *testContext) BindQuery(v interface{}) error {
	return fmt.Errorf("BindQuery not implemented in test context")
}

func (c *testContext) BindVars(v interface{}) error {
	return fmt.Errorf("BindVars not implemented in test context")
}

func (c *testContext) Middleware(h middleware.Handler) middleware.Handler {
	return h
}

func (c *testContext) Result(code int, v interface{}) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *testContext) Returns(v interface{}, err error) error {
	return err
}
