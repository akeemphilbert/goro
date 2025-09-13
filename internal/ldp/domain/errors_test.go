package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestStorageError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StorageError
		expected string
	}{
		{
			name: "error without cause",
			err: &StorageError{
				Code:    "TEST_ERROR",
				Message: "test error message",
			},
			expected: "TEST_ERROR: test error message",
		},
		{
			name: "error with cause",
			err: &StorageError{
				Code:    "TEST_ERROR",
				Message: "test error message",
				Cause:   errors.New("underlying error"),
			},
			expected: "TEST_ERROR: test error message (caused by: underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("StorageError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStorageError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &StorageError{
		Code:    "TEST_ERROR",
		Message: "test error message",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("StorageError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestStorageError_WithContext(t *testing.T) {
	err := &StorageError{
		Code:    "TEST_ERROR",
		Message: "test error message",
	}

	result := err.WithContext("resourceID", "test-123")

	if result.Context["resourceID"] != "test-123" {
		t.Errorf("WithContext() did not set context correctly, got %v", result.Context)
	}

	// Test chaining
	result = result.WithContext("operation", "store")
	if result.Context["operation"] != "store" {
		t.Errorf("WithContext() chaining failed, got %v", result.Context)
	}
}

func TestStorageError_WithOperation(t *testing.T) {
	err := &StorageError{
		Code:    "TEST_ERROR",
		Message: "test error message",
	}

	result := err.WithOperation("retrieve")

	if result.Operation != "retrieve" {
		t.Errorf("WithOperation() = %v, want %v", result.Operation, "retrieve")
	}
}

func TestNewStorageError(t *testing.T) {
	err := NewStorageError("CUSTOM_ERROR", "custom message")

	if err.Code != "CUSTOM_ERROR" {
		t.Errorf("NewStorageError() code = %v, want %v", err.Code, "CUSTOM_ERROR")
	}

	if err.Message != "custom message" {
		t.Errorf("NewStorageError() message = %v, want %v", err.Message, "custom message")
	}

	if err.Context == nil {
		t.Error("NewStorageError() should initialize Context map")
	}
}

func TestWrapStorageError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapStorageError(originalErr, "WRAPPED_ERROR", "wrapped message")

	if wrappedErr.Code != "WRAPPED_ERROR" {
		t.Errorf("WrapStorageError() code = %v, want %v", wrappedErr.Code, "WRAPPED_ERROR")
	}

	if wrappedErr.Message != "wrapped message" {
		t.Errorf("WrapStorageError() message = %v, want %v", wrappedErr.Message, "wrapped message")
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("WrapStorageError() cause = %v, want %v", wrappedErr.Cause, originalErr)
	}

	if wrappedErr.Context == nil {
		t.Error("WrapStorageError() should initialize Context map")
	}
}

func TestIsStorageError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "storage error",
			err:      &StorageError{Code: "TEST", Message: "test"},
			expected: true,
		},
		{
			name:     "wrapped storage error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: "TEST", Message: "test"}),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStorageError(tt.err); got != tt.expected {
				t.Errorf("IsStorageError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetStorageError(t *testing.T) {
	storageErr := &StorageError{Code: "TEST", Message: "test"}

	tests := []struct {
		name        string
		err         error
		expectedErr *StorageError
		expectedOk  bool
	}{
		{
			name:        "direct storage error",
			err:         storageErr,
			expectedErr: storageErr,
			expectedOk:  true,
		},
		{
			name:        "wrapped storage error",
			err:         fmt.Errorf("wrapped: %w", storageErr),
			expectedErr: storageErr,
			expectedOk:  true,
		},
		{
			name:        "regular error",
			err:         errors.New("regular error"),
			expectedErr: nil,
			expectedOk:  false,
		},
		{
			name:        "nil error",
			err:         nil,
			expectedErr: nil,
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr, gotOk := GetStorageError(tt.err)
			if gotOk != tt.expectedOk {
				t.Errorf("GetStorageError() ok = %v, want %v", gotOk, tt.expectedOk)
			}
			if gotErr != tt.expectedErr {
				t.Errorf("GetStorageError() err = %v, want %v", gotErr, tt.expectedErr)
			}
		})
	}
}

func TestIsResourceNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "resource not found error",
			err:      &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"},
			expected: true,
		},
		{
			name:     "wrapped resource not found error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"}),
			expected: true,
		},
		{
			name:     "different storage error",
			err:      &StorageError{Code: ErrUnsupportedFormat.Code, Message: "unsupported"},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsResourceNotFound(tt.err); got != tt.expected {
				t.Errorf("IsResourceNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsUnsupportedFormat(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unsupported format error",
			err:      &StorageError{Code: ErrUnsupportedFormat.Code, Message: "unsupported"},
			expected: true,
		},
		{
			name:     "wrapped unsupported format error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: ErrUnsupportedFormat.Code, Message: "unsupported"}),
			expected: true,
		},
		{
			name:     "different storage error",
			err:      &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnsupportedFormat(tt.err); got != tt.expected {
				t.Errorf("IsUnsupportedFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsInsufficientStorage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "insufficient storage error",
			err:      &StorageError{Code: ErrInsufficientStorage.Code, Message: "insufficient"},
			expected: true,
		},
		{
			name:     "wrapped insufficient storage error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: ErrInsufficientStorage.Code, Message: "insufficient"}),
			expected: true,
		},
		{
			name:     "different storage error",
			err:      &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInsufficientStorage(tt.err); got != tt.expected {
				t.Errorf("IsInsufficientStorage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsDataCorruption(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "data corruption error",
			err:      &StorageError{Code: ErrDataCorruption.Code, Message: "corrupted"},
			expected: true,
		},
		{
			name:     "wrapped data corruption error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: ErrDataCorruption.Code, Message: "corrupted"}),
			expected: true,
		},
		{
			name:     "different storage error",
			err:      &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDataCorruption(tt.err); got != tt.expected {
				t.Errorf("IsDataCorruption() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsFormatConversion(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "format conversion error",
			err:      &StorageError{Code: ErrFormatConversion.Code, Message: "conversion failed"},
			expected: true,
		},
		{
			name:     "wrapped format conversion error",
			err:      fmt.Errorf("wrapped: %w", &StorageError{Code: ErrFormatConversion.Code, Message: "conversion failed"}),
			expected: true,
		},
		{
			name:     "different storage error",
			err:      &StorageError{Code: ErrResourceNotFound.Code, Message: "not found"},
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFormatConversion(tt.err); got != tt.expected {
				t.Errorf("IsFormatConversion() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *StorageError
		code string
	}{
		{"ErrResourceNotFound", ErrResourceNotFound, "RESOURCE_NOT_FOUND"},
		{"ErrUnsupportedFormat", ErrUnsupportedFormat, "UNSUPPORTED_FORMAT"},
		{"ErrInsufficientStorage", ErrInsufficientStorage, "INSUFFICIENT_STORAGE"},
		{"ErrDataCorruption", ErrDataCorruption, "DATA_CORRUPTION"},
		{"ErrFormatConversion", ErrFormatConversion, "FORMAT_CONVERSION_FAILED"},
		{"ErrInvalidResource", ErrInvalidResource, "INVALID_RESOURCE"},
		{"ErrStorageOperation", ErrStorageOperation, "STORAGE_OPERATION_FAILED"},
		{"ErrResourceExists", ErrResourceExists, "RESOURCE_EXISTS"},
		{"ErrInvalidID", ErrInvalidID, "INVALID_ID"},
		{"ErrChecksumMismatch", ErrChecksumMismatch, "CHECKSUM_MISMATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("%s code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
			if tt.err.Message == "" {
				t.Errorf("%s should have a non-empty message", tt.name)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	// Test complex error chaining scenario
	originalErr := errors.New("filesystem error")
	storageErr := WrapStorageError(originalErr, "STORAGE_OPERATION_FAILED", "failed to write file")
	storageErr = storageErr.WithContext("resourceID", "test-123").WithOperation("store")

	// Test error unwrapping
	if !errors.Is(storageErr, originalErr) {
		t.Error("errors.Is should find the original error in the chain")
	}

	// Test storage error detection
	if !IsStorageError(storageErr) {
		t.Error("IsStorageError should detect the storage error")
	}

	// Test context preservation
	if storageErr.Context["resourceID"] != "test-123" {
		t.Error("Context should be preserved")
	}

	if storageErr.Operation != "store" {
		t.Error("Operation should be preserved")
	}
}
