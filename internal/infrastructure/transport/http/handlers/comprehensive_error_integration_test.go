package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveErrorResponsesIntegration tests all error scenarios end-to-end
// This test validates Requirements: 1.5, 2.4, 2.5, 5.5
func TestComprehensiveErrorResponsesIntegration(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)

	t.Run("HTTP Status Code Mapping", func(t *testing.T) {
		// Test all HTTP status codes are correctly mapped from storage errors
		testCases := []struct {
			name           string
			storageError   *domain.StorageError
			expectedStatus int
			expectedCode   string
		}{
			{
				name:           "404 Not Found",
				storageError:   &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"},
				expectedStatus: http.StatusNotFound,
				expectedCode:   "RESOURCE_NOT_FOUND",
			},
			{
				name:           "406 Not Acceptable",
				storageError:   &domain.StorageError{Code: "UNSUPPORTED_FORMAT", Message: "unsupported"},
				expectedStatus: http.StatusNotAcceptable,
				expectedCode:   "UNSUPPORTED_FORMAT",
			},
			{
				name:           "507 Insufficient Storage",
				storageError:   &domain.StorageError{Code: "INSUFFICIENT_STORAGE", Message: "insufficient"},
				expectedStatus: http.StatusInsufficientStorage,
				expectedCode:   "INSUFFICIENT_STORAGE",
			},
			{
				name:           "422 Unprocessable Entity - Data Corruption",
				storageError:   &domain.StorageError{Code: "DATA_CORRUPTION", Message: "corrupted"},
				expectedStatus: http.StatusUnprocessableEntity,
				expectedCode:   "DATA_CORRUPTION",
			},
			{
				name:           "422 Unprocessable Entity - Checksum Mismatch",
				storageError:   &domain.StorageError{Code: "CHECKSUM_MISMATCH", Message: "checksum failed"},
				expectedStatus: http.StatusUnprocessableEntity,
				expectedCode:   "CHECKSUM_MISMATCH",
			},
			{
				name:           "400 Bad Request - Format Conversion",
				storageError:   &domain.StorageError{Code: "FORMAT_CONVERSION_FAILED", Message: "conversion failed"},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "FORMAT_CONVERSION_FAILED",
			},
			{
				name:           "400 Bad Request - Invalid ID",
				storageError:   &domain.StorageError{Code: "INVALID_ID", Message: "invalid id"},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "INVALID_ID",
			},
			{
				name:           "400 Bad Request - Invalid Resource",
				storageError:   &domain.StorageError{Code: "INVALID_RESOURCE", Message: "invalid resource"},
				expectedStatus: http.StatusBadRequest,
				expectedCode:   "INVALID_RESOURCE",
			},
			{
				name:           "409 Conflict - Resource Exists",
				storageError:   &domain.StorageError{Code: "RESOURCE_EXISTS", Message: "already exists"},
				expectedStatus: http.StatusConflict,
				expectedCode:   "RESOURCE_EXISTS",
			},
			{
				name:           "500 Internal Server Error - Storage Operation Failed",
				storageError:   &domain.StorageError{Code: "STORAGE_OPERATION_FAILED", Message: "operation failed"},
				expectedStatus: http.StatusInternalServerError,
				expectedCode:   "STORAGE_OPERATION_FAILED",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := &MockErrorStorageService{
					retrieveError: tc.storageError,
				}
				handler := NewResourceHandler(mockService, logger)

				w := httptest.NewRecorder()
				ctx := &testContext{
					response: w,
					vars:     map[string]string{"id": "test-resource"},
				}

				err := handler.GetResource(ctx)
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedStatus, w.Code, "Status code should match for %s", tc.name)

				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				errorObj := response["error"].(map[string]interface{})
				assert.Equal(t, tc.expectedCode, errorObj["code"])
				assert.Equal(t, float64(tc.expectedStatus), errorObj["status"])
			})
		}
	})

	t.Run("406 Not Acceptable with Supported Formats", func(t *testing.T) {
		// Test that 406 responses include supported formats list
		mockService := &MockErrorStorageService{
			retrieveError: &domain.StorageError{
				Code:    "UNSUPPORTED_FORMAT",
				Message: "unsupported format",
				Context: map[string]any{
					"format": "application/unsupported",
				},
			},
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-resource"},
		}

		req := httptest.NewRequest("GET", "/resources/test-resource", nil)
		req.Header.Set("Accept", "application/unsupported")
		ctx.request = req

		err := handler.GetResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusNotAcceptable, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "UNSUPPORTED_FORMAT", errorObj["code"])

		// Verify supported formats are included
		supportedFormats, ok := errorObj["supportedFormats"].([]interface{})
		require.True(t, ok, "supportedFormats should be present")
		assert.Len(t, supportedFormats, 3)
		assert.Contains(t, supportedFormats, "application/ld+json")
		assert.Contains(t, supportedFormats, "text/turtle")
		assert.Contains(t, supportedFormats, "application/rdf+xml")

		// Verify helpful message
		assert.Contains(t, errorObj["message"], "Supported formats")
	})

	t.Run("507 Insufficient Storage with Helpful Information", func(t *testing.T) {
		// Test that 507 responses include helpful suggestions
		mockService := &MockErrorStorageService{
			storeError: &domain.StorageError{
				Code:      "INSUFFICIENT_STORAGE",
				Message:   "insufficient storage space",
				Operation: "store",
				Context: map[string]any{
					"resourceID":     "large-resource",
					"requiredSpace":  1024 * 1024,
					"availableSpace": 512,
				},
			},
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "large-resource"},
		}

		req := httptest.NewRequest("PUT", "/resources/large-resource", strings.NewReader("large data"))
		req.Header.Set("Content-Type", "application/ld+json")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInsufficientStorage, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "INSUFFICIENT_STORAGE", errorObj["code"])
		assert.Contains(t, errorObj["suggestion"], "reducing the size")
		assert.Contains(t, errorObj["message"], "Insufficient storage space")

		// Verify context information is included safely
		if context, ok := errorObj["context"].(map[string]interface{}); ok {
			assert.Equal(t, "large-resource", context["resourceID"])
		}
	})

	t.Run("Meaningful Error Messages and Logging", func(t *testing.T) {
		// Test that error messages are meaningful and logging works correctly
		var logOutput strings.Builder
		testLogger := log.NewStdLogger(&logOutput)

		storageErr := &domain.StorageError{
			Code:      "DATA_CORRUPTION",
			Message:   "data corruption detected",
			Operation: "store",
			Context: map[string]any{
				"resourceID":       "corrupted-resource",
				"expectedChecksum": "abc123",
				"actualChecksum":   "def456",
			},
		}

		mockService := &MockErrorStorageService{
			storeError: storageErr,
		}
		handler := NewResourceHandler(mockService, testLogger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "corrupted-resource"},
		}

		req := httptest.NewRequest("PUT", "/resources/corrupted-resource", strings.NewReader("corrupted data"))
		req.Header.Set("Content-Type", "application/ld+json")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		// Verify HTTP response
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "DATA_CORRUPTION", errorObj["code"])
		assert.Contains(t, errorObj["message"], "Data corruption detected")
		assert.Contains(t, errorObj["suggestion"], "try uploading the resource again")

		// Verify logging occurred
		logStr := logOutput.String()
		assert.Contains(t, logStr, "Storage operation error")
		assert.Contains(t, logStr, "DATA_CORRUPTION")
		assert.Contains(t, logStr, "corrupted-resource")
		assert.Contains(t, logStr, "store")
	})

	t.Run("Error Response Structure Consistency", func(t *testing.T) {
		// Test that all error responses have consistent structure
		mockService := &MockErrorStorageService{
			retrieveError: &domain.StorageError{
				Code:      "RESOURCE_NOT_FOUND",
				Message:   "resource not found",
				Operation: "retrieve",
				Context: map[string]any{
					"resourceID": "missing-resource",
				},
			},
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "missing-resource"},
		}

		err := handler.GetResource(ctx)
		assert.NoError(t, err)

		// Verify response headers
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify top-level structure
		assert.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})

		// Verify required fields are present
		requiredFields := []string{"code", "message", "status", "timestamp"}
		for _, field := range requiredFields {
			assert.Contains(t, errorObj, field, "Error response should contain field: %s", field)
		}

		// Verify field types and values
		assert.IsType(t, "", errorObj["code"])
		assert.IsType(t, "", errorObj["message"])
		assert.IsType(t, float64(0), errorObj["status"])
		assert.IsType(t, "", errorObj["timestamp"])

		// Verify operation context is included
		assert.Equal(t, "retrieve", errorObj["operation"])
	})

	t.Run("Context Information Safety", func(t *testing.T) {
		// Test that sensitive information is filtered from error responses
		storageErr := &domain.StorageError{
			Code:      "STORAGE_OPERATION_FAILED",
			Message:   "storage operation failed",
			Operation: "store",
			Context: map[string]any{
				// Safe fields that should be included
				"resourceID":  "safe-resource",
				"contentType": "application/ld+json",
				"format":      "json-ld",
				"size":        1024,
				"operation":   "store",
				// Sensitive fields that should be filtered out
				"password":     "secret123",
				"internalPath": "/var/data/secret",
				"systemError":  "database connection failed",
				"apiKey":       "sk-1234567890",
			},
		}

		mockService := &MockErrorStorageService{
			storeError: storageErr,
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "safe-resource"},
		}

		req := httptest.NewRequest("PUT", "/resources/safe-resource", strings.NewReader("test data"))
		req.Header.Set("Content-Type", "application/ld+json")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		context := errorObj["context"].(map[string]interface{})

		// Verify safe fields are present
		assert.Equal(t, "safe-resource", context["resourceID"])
		assert.Equal(t, "application/ld+json", context["contentType"])
		assert.Equal(t, "json-ld", context["format"])
		assert.Equal(t, float64(1024), context["size"])
		assert.Equal(t, "store", context["operation"])

		// Verify sensitive fields are filtered out
		assert.NotContains(t, context, "password")
		assert.NotContains(t, context, "internalPath")
		assert.NotContains(t, context, "systemError")
		assert.NotContains(t, context, "apiKey")
	})

	t.Run("Error Response Headers Validation", func(t *testing.T) {
		// Test that all error responses have correct headers
		mockService := &MockErrorStorageService{
			retrieveError: &domain.StorageError{
				Code:    "RESOURCE_NOT_FOUND",
				Message: "resource not found",
			},
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-resource"},
		}

		err := handler.GetResource(ctx)
		assert.NoError(t, err)

		// Verify response headers are set correctly
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))

		// Verify response structure
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		errorObj := response["error"].(map[string]interface{})
		assert.Contains(t, errorObj, "timestamp")

		// Verify timestamp is a valid Unix timestamp string
		timestamp, ok := errorObj["timestamp"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, timestamp)
	})
}

// MockErrorStorageService for testing error scenarios
type MockErrorStorageService struct {
	storeError    error
	retrieveError error
	deleteError   error
	existsError   error
}

func (m *MockErrorStorageService) StoreResource(ctx context.Context, id string, data []byte, contentType string) (domain.Resource, error) {
	if m.storeError != nil {
		return nil, m.storeError
	}
	return domain.NewResource(ctx, id, contentType, data), nil
}

func (m *MockErrorStorageService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (domain.Resource, error) {
	if m.retrieveError != nil {
		return nil, m.retrieveError
	}
	return domain.NewResource(ctx, id, "application/ld+json", []byte(`{"test": "data"}`)), nil
}

func (m *MockErrorStorageService) DeleteResource(ctx context.Context, id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	return nil
}

func (m *MockErrorStorageService) ResourceExists(ctx context.Context, id string) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	return false, nil
}

func (m *MockErrorStorageService) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	if m.retrieveError != nil {
		return nil, "", m.retrieveError
	}
	return io.NopCloser(strings.NewReader(`{"test": "data"}`)), "application/ld+json", nil
}

func (m *MockErrorStorageService) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) (domain.Resource, error) {
	if m.storeError != nil {
		return nil, m.storeError
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return domain.NewResource(ctx, id, contentType, data), nil
}
