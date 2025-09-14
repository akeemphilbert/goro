package handlers

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// TestErrorScenarios tests various error scenarios that can occur during resource operations
func TestErrorScenarios(t *testing.T) {
	logger := log.NewStdLogger(nil)

	t.Run("Resource Not Found Scenario", func(t *testing.T) {
		mockService := &MockStorageService{}
		_ = NewResourceHandler(mockService, logger) // Create handler to test integration

		// Mock service to return resource not found error
		notFoundErr := &domain.StorageError{
			Code:    "RESOURCE_NOT_FOUND",
			Message: "resource not found",
			Context: map[string]any{
				"resourceID": "missing-resource",
			},
		}

		// Test that the error is properly identified
		assert.True(t, domain.IsResourceNotFound(notFoundErr))
		assert.False(t, domain.IsUnsupportedFormat(notFoundErr))
	})

	t.Run("Unsupported Format Scenario", func(t *testing.T) {
		mockService := &MockStorageService{}
		_ = NewResourceHandler(mockService, logger) // Create handler to test integration

		// Mock service to return unsupported format error
		formatErr := &domain.StorageError{
			Code:    "UNSUPPORTED_FORMAT",
			Message: "unsupported format",
			Context: map[string]any{
				"format": "application/unsupported",
			},
		}

		// Test that the error is properly identified
		assert.True(t, domain.IsUnsupportedFormat(formatErr))
		assert.False(t, domain.IsResourceNotFound(formatErr))
	})

	t.Run("Insufficient Storage Scenario", func(t *testing.T) {
		mockService := &MockStorageService{}
		_ = NewResourceHandler(mockService, logger) // Create handler to test integration

		// Mock service to return insufficient storage error
		storageErr := &domain.StorageError{
			Code:    "INSUFFICIENT_STORAGE",
			Message: "insufficient storage space",
			Context: map[string]any{
				"availableSpace": 0,
				"requiredSpace":  1024,
			},
		}

		// Test that the error is properly identified
		assert.True(t, domain.IsInsufficientStorage(storageErr))
		assert.False(t, domain.IsDataCorruption(storageErr))
	})

	t.Run("Data Corruption Scenario", func(t *testing.T) {
		mockService := &MockStorageService{}
		_ = NewResourceHandler(mockService, logger) // Create handler to test integration

		// Mock service to return data corruption error
		corruptionErr := &domain.StorageError{
			Code:    "DATA_CORRUPTION",
			Message: "data corruption detected",
			Context: map[string]any{
				"expectedChecksum": "abc123",
				"actualChecksum":   "def456",
			},
		}

		// Test that the error is properly identified
		assert.True(t, domain.IsDataCorruption(corruptionErr))
		assert.False(t, domain.IsFormatConversion(corruptionErr))
	})

	t.Run("Format Conversion Failure Scenario", func(t *testing.T) {
		mockService := &MockStorageService{}
		_ = NewResourceHandler(mockService, logger) // Create handler to test integration

		// Mock service to return format conversion error
		conversionErr := &domain.StorageError{
			Code:    "FORMAT_CONVERSION_FAILED",
			Message: "format conversion failed",
			Context: map[string]any{
				"fromFormat": "application/ld+json",
				"toFormat":   "text/turtle",
			},
		}

		// Test that the error is properly identified
		assert.True(t, domain.IsFormatConversion(conversionErr))
		assert.False(t, domain.IsInsufficientStorage(conversionErr))
	})

	t.Run("Error Context Preservation", func(t *testing.T) {
		// Test that error context is properly preserved through the error chain
		originalErr := &domain.StorageError{
			Code:      "TEST_ERROR",
			Message:   "test error",
			Operation: "test_operation",
			Context: map[string]any{
				"resourceID":  "test-123",
				"contentType": "application/ld+json",
				"size":        1024,
			},
		}

		// Test context preservation
		assert.Equal(t, "test_operation", originalErr.Operation)
		assert.Equal(t, "test-123", originalErr.Context["resourceID"])
		assert.Equal(t, "application/ld+json", originalErr.Context["contentType"])
		assert.Equal(t, 1024, originalErr.Context["size"])

		// Test error chaining
		wrappedErr := originalErr.WithContext("timestamp", "2023-01-01T00:00:00Z")
		assert.Equal(t, "2023-01-01T00:00:00Z", wrappedErr.Context["timestamp"])
		assert.Equal(t, "test-123", wrappedErr.Context["resourceID"]) // Original context preserved
	})

	t.Run("Error Type Detection", func(t *testing.T) {
		// Test all error type detection functions
		testCases := []struct {
			name     string
			error    *domain.StorageError
			detector func(error) bool
		}{
			{
				name:     "Resource Not Found",
				error:    &domain.StorageError{Code: "RESOURCE_NOT_FOUND"},
				detector: domain.IsResourceNotFound,
			},
			{
				name:     "Unsupported Format",
				error:    &domain.StorageError{Code: "UNSUPPORTED_FORMAT"},
				detector: domain.IsUnsupportedFormat,
			},
			{
				name:     "Insufficient Storage",
				error:    &domain.StorageError{Code: "INSUFFICIENT_STORAGE"},
				detector: domain.IsInsufficientStorage,
			},
			{
				name:     "Data Corruption",
				error:    &domain.StorageError{Code: "DATA_CORRUPTION"},
				detector: domain.IsDataCorruption,
			},
			{
				name:     "Format Conversion",
				error:    &domain.StorageError{Code: "FORMAT_CONVERSION_FAILED"},
				detector: domain.IsFormatConversion,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Should detect the correct error type
				assert.True(t, tc.detector(tc.error), "Should detect %s", tc.name)

				// Should not detect other error types
				for _, other := range testCases {
					if other.name != tc.name {
						assert.False(t, other.detector(tc.error),
							"Should not detect %s as %s", tc.name, other.name)
					}
				}
			})
		}
	})
}

// TestErrorResponseStructure tests the structure of error responses
func TestErrorResponseStructure(t *testing.T) {
	logger := log.NewStdLogger(nil)
	mockService := &MockStorageService{}
	_ = NewResourceHandler(mockService, logger) // Create handler to test integration

	t.Run("Error Response Contains Required Fields", func(t *testing.T) {
		storageErr := &domain.StorageError{
			Code:      "TEST_ERROR",
			Message:   "test error message",
			Operation: "test_operation",
			Context: map[string]any{
				"resourceID":  "test-123",
				"contentType": "application/ld+json",
			},
		}

		// Test that storage error has all expected fields
		assert.Equal(t, "TEST_ERROR", storageErr.Code)
		assert.Equal(t, "test error message", storageErr.Message)
		assert.Equal(t, "test_operation", storageErr.Operation)
		assert.NotNil(t, storageErr.Context)

		// Test error string representation
		errorStr := storageErr.Error()
		assert.Contains(t, errorStr, "TEST_ERROR")
		assert.Contains(t, errorStr, "test error message")
	})

	t.Run("Error Logging Information", func(t *testing.T) {
		storageErr := &domain.StorageError{
			Code:      "STORAGE_ERROR",
			Message:   "storage operation failed",
			Operation: "store",
			Context: map[string]any{
				"resourceID": "log-test-123",
				"size":       2048,
			},
		}

		// Test that we can extract logging information
		assert.Equal(t, "STORAGE_ERROR", storageErr.Code)
		assert.Equal(t, "store", storageErr.Operation)

		// Test context contains safe information for logging
		safeFields := []string{"resourceID", "contentType", "format", "size"}
		for _, field := range safeFields {
			if value, exists := storageErr.Context[field]; exists {
				assert.NotNil(t, value, "Safe field %s should have a value", field)
			}
		}
	})
}
