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

// MockStorageServiceWithLimits simulates storage limitations
type MockStorageServiceWithLimits struct {
	simulateInsufficientStorage bool
	simulateDataCorruption      bool
	simulateFormatError         bool
}

// MockUnsupportedFormatService always returns unsupported format errors
type MockUnsupportedFormatService struct{}

func (m *MockUnsupportedFormatService) StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error) {
	return nil, &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"}
}

func (m *MockUnsupportedFormatService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error) {
	return nil, &domain.StorageError{
		Code:      "UNSUPPORTED_FORMAT",
		Message:   "unsupported format requested",
		Operation: "retrieve",
		Context: map[string]any{
			"resourceID": id,
			"format":     acceptFormat,
		},
	}
}

func (m *MockUnsupportedFormatService) DeleteResource(ctx context.Context, id string) error {
	return &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"}
}

func (m *MockUnsupportedFormatService) ResourceExists(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *MockUnsupportedFormatService) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	return nil, "", &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"}
}

func (m *MockUnsupportedFormatService) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string) (*domain.Resource, error) {
	return nil, &domain.StorageError{Code: "RESOURCE_NOT_FOUND", Message: "not found"}
}

func (m *MockStorageServiceWithLimits) StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error) {
	if m.simulateInsufficientStorage {
		return nil, &domain.StorageError{
			Code:      "INSUFFICIENT_STORAGE",
			Message:   "insufficient storage space",
			Operation: "store",
			Context: map[string]any{
				"resourceID":     id,
				"requiredSpace":  len(data),
				"availableSpace": 0,
			},
		}
	}

	if m.simulateDataCorruption {
		return nil, &domain.StorageError{
			Code:      "DATA_CORRUPTION",
			Message:   "data corruption detected during storage",
			Operation: "store",
			Context: map[string]any{
				"resourceID": id,
				"checksum":   "failed",
			},
		}
	}

	if m.simulateFormatError {
		return nil, &domain.StorageError{
			Code:      "FORMAT_CONVERSION_FAILED",
			Message:   "failed to convert format",
			Operation: "store",
			Context: map[string]any{
				"resourceID": id,
				"format":     contentType,
			},
		}
	}

	return domain.NewResource(id, contentType, data), nil
}

func (m *MockStorageServiceWithLimits) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error) {
	return nil, &domain.StorageError{
		Code:      "RESOURCE_NOT_FOUND",
		Message:   "resource not found",
		Operation: "retrieve",
		Context: map[string]any{
			"resourceID": id,
		},
	}
}

func (m *MockStorageServiceWithLimits) DeleteResource(ctx context.Context, id string) error {
	return &domain.StorageError{
		Code:      "RESOURCE_NOT_FOUND",
		Message:   "resource not found",
		Operation: "delete",
		Context: map[string]any{
			"resourceID": id,
		},
	}
}

func (m *MockStorageServiceWithLimits) ResourceExists(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (m *MockStorageServiceWithLimits) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	return nil, "", &domain.StorageError{
		Code:      "RESOURCE_NOT_FOUND",
		Message:   "resource not found",
		Operation: "stream",
		Context: map[string]any{
			"resourceID": id,
		},
	}
}

func (m *MockStorageServiceWithLimits) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string) (*domain.Resource, error) {
	if m.simulateInsufficientStorage {
		return nil, &domain.StorageError{
			Code:      "INSUFFICIENT_STORAGE",
			Message:   "insufficient storage space for stream",
			Operation: "stream_store",
			Context: map[string]any{
				"resourceID": id,
			},
		}
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return domain.NewResource(id, contentType, data), nil
}

// TestStorageLimitErrors tests specific storage limitation error scenarios
func TestStorageLimitErrors(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)

	t.Run("507 Insufficient Storage", func(t *testing.T) {
		mockService := &MockStorageServiceWithLimits{
			simulateInsufficientStorage: true,
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-507"},
		}

		// Simulate PUT request with large data
		largeData := strings.Repeat("x", 1024*1024) // 1MB of data
		req := httptest.NewRequest("PUT", "/resources/test-507", strings.NewReader(largeData))
		req.Header.Set("Content-Type", "application/ld+json")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusInsufficientStorage, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "INSUFFICIENT_STORAGE", errorObj["code"])
		assert.Equal(t, float64(http.StatusInsufficientStorage), errorObj["status"])
		assert.Contains(t, errorObj["message"], "Insufficient storage space")
		assert.Contains(t, errorObj["suggestion"], "reducing the size")

		// Check context information
		if context, ok := errorObj["context"].(map[string]interface{}); ok {
			assert.Equal(t, "test-507", context["resourceID"])
		}
	})

	t.Run("422 Data Corruption", func(t *testing.T) {
		mockService := &MockStorageServiceWithLimits{
			simulateDataCorruption: true,
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-422"},
		}

		req := httptest.NewRequest("PUT", "/resources/test-422", strings.NewReader("test data"))
		req.Header.Set("Content-Type", "application/ld+json")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "DATA_CORRUPTION", errorObj["code"])
		assert.Equal(t, float64(http.StatusUnprocessableEntity), errorObj["status"])
		assert.Contains(t, errorObj["message"], "Data corruption detected")
		assert.Contains(t, errorObj["suggestion"], "try uploading the resource again")
	})

	t.Run("400 Format Conversion Failed", func(t *testing.T) {
		mockService := &MockStorageServiceWithLimits{
			simulateFormatError: true,
		}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-format"},
		}

		req := httptest.NewRequest("PUT", "/resources/test-format", strings.NewReader("invalid rdf"))
		req.Header.Set("Content-Type", "text/turtle")
		ctx.request = req

		err := handler.PutResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		assert.Equal(t, "FORMAT_CONVERSION_FAILED", errorObj["code"])
		assert.Equal(t, float64(http.StatusBadRequest), errorObj["status"])
		assert.Contains(t, errorObj["message"], "Failed to convert between the requested formats")
	})

	t.Run("406 Not Acceptable with Detailed Information", func(t *testing.T) {
		// Create a mock that simulates unsupported format error
		mockService := &MockUnsupportedFormatService{}
		handler := NewResourceHandler(mockService, logger)

		w := httptest.NewRecorder()
		ctx := &testContext{
			response: w,
			vars:     map[string]string{"id": "test-406"},
		}

		req := httptest.NewRequest("GET", "/resources/test-406", nil)
		req.Header.Set("Accept", "application/unsupported")
		ctx.request = req

		err := handler.GetResource(ctx)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusNotAcceptable, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj, ok := response["error"].(map[string]interface{})
		require.True(t, ok, "Response should have error object")

		assert.Equal(t, "UNSUPPORTED_FORMAT", errorObj["code"])
		assert.Equal(t, float64(http.StatusNotAcceptable), errorObj["status"])

		// Check that supported formats are listed
		if supportedFormats, ok := errorObj["supportedFormats"].([]interface{}); ok {
			assert.Len(t, supportedFormats, 3)
			assert.Contains(t, supportedFormats, "application/ld+json")
			assert.Contains(t, supportedFormats, "text/turtle")
			assert.Contains(t, supportedFormats, "application/rdf+xml")
		} else {
			t.Errorf("supportedFormats field is missing or not the expected type")
		}
	})

	t.Run("Error Context Safety", func(t *testing.T) {
		// Test that sensitive information is not exposed in error responses
		mockService := &MockStorageServiceWithLimits{
			simulateInsufficientStorage: true,
		}
		handler := NewResourceHandler(mockService, logger)

		// Create a storage error with both safe and sensitive context
		storageErr := &domain.StorageError{
			Code:      "INSUFFICIENT_STORAGE",
			Message:   "insufficient storage",
			Operation: "store",
			Context: map[string]any{
				"resourceID":   "safe-to-expose",
				"contentType":  "application/ld+json",
				"size":         1024,
				"format":       "json-ld",
				"operation":    "store",
				"password":     "should-not-appear",
				"internalPath": "/secret/path",
				"systemError":  "internal system error",
			},
		}

		w := httptest.NewRecorder()
		ctx := &testContext{response: w}

		err := handler.writeDetailedErrorResponse(ctx, http.StatusInsufficientStorage, "INSUFFICIENT_STORAGE", "test message", storageErr)
		assert.NoError(t, err)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		errorObj := response["error"].(map[string]interface{})
		context := errorObj["context"].(map[string]interface{})

		// Safe fields should be present
		assert.Equal(t, "safe-to-expose", context["resourceID"])
		assert.Equal(t, "application/ld+json", context["contentType"])
		assert.Equal(t, float64(1024), context["size"])
		assert.Equal(t, "json-ld", context["format"])
		assert.Equal(t, "store", context["operation"])

		// Sensitive fields should be filtered out
		assert.NotContains(t, context, "password")
		assert.NotContains(t, context, "internalPath")
		assert.NotContains(t, context, "systemError")
	})

	t.Run("Error Response Consistency", func(t *testing.T) {
		// Test that all error responses have consistent structure
		errorScenarios := []struct {
			name           string
			setupMock      func(*MockStorageServiceWithLimits)
			method         string
			expectedStatus int
			expectedCode   string
		}{
			{
				name:           "Resource Not Found",
				setupMock:      func(m *MockStorageServiceWithLimits) {},
				method:         "GET",
				expectedStatus: http.StatusNotFound,
				expectedCode:   "RESOURCE_NOT_FOUND",
			},
			{
				name: "Insufficient Storage",
				setupMock: func(m *MockStorageServiceWithLimits) {
					m.simulateInsufficientStorage = true
				},
				method:         "PUT",
				expectedStatus: http.StatusInsufficientStorage,
				expectedCode:   "INSUFFICIENT_STORAGE",
			},
			{
				name: "Data Corruption",
				setupMock: func(m *MockStorageServiceWithLimits) {
					m.simulateDataCorruption = true
				},
				method:         "PUT",
				expectedStatus: http.StatusUnprocessableEntity,
				expectedCode:   "DATA_CORRUPTION",
			},
		}

		for _, scenario := range errorScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				mockService := &MockStorageServiceWithLimits{}
				scenario.setupMock(mockService)
				testHandler := NewResourceHandler(mockService, logger)

				w := httptest.NewRecorder()
				ctx := &testContext{
					response: w,
					vars:     map[string]string{"id": "consistency-test"},
				}

				if scenario.method == "PUT" {
					req := httptest.NewRequest("PUT", "/resources/consistency-test", strings.NewReader("test"))
					req.Header.Set("Content-Type", "application/ld+json")
					ctx.request = req
					testHandler.PutResource(ctx)
				} else {
					req := httptest.NewRequest("GET", "/resources/consistency-test", nil)
					ctx.request = req
					testHandler.GetResource(ctx)
				}

				assert.Equal(t, scenario.expectedStatus, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Validate consistent structure
				assert.Contains(t, response, "error")
				errorObj := response["error"].(map[string]interface{})

				// All errors should have these fields
				requiredFields := []string{"code", "message", "status", "timestamp"}
				for _, field := range requiredFields {
					assert.Contains(t, errorObj, field, "Error %s should contain field: %s", scenario.name, field)
				}

				assert.Equal(t, scenario.expectedCode, errorObj["code"])
				assert.Equal(t, float64(scenario.expectedStatus), errorObj["status"])
			})
		}
	})
}
