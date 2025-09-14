package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerErrors_Predefined(t *testing.T) {
	// Test that all predefined container errors are properly defined
	assert.Equal(t, "CONTAINER_NOT_FOUND", ErrContainerNotFound.Code)
	assert.Equal(t, "container not found", ErrContainerNotFound.Message)

	assert.Equal(t, "CONTAINER_NOT_EMPTY", ErrContainerNotEmpty.Code)
	assert.Equal(t, "container contains resources", ErrContainerNotEmpty.Message)

	assert.Equal(t, "CIRCULAR_REFERENCE", ErrCircularReference.Code)
	assert.Equal(t, "circular container reference", ErrCircularReference.Message)

	assert.Equal(t, "INVALID_HIERARCHY", ErrInvalidHierarchy.Code)
	assert.Equal(t, "invalid container hierarchy", ErrInvalidHierarchy.Message)

	assert.Equal(t, "MEMBERSHIP_CONFLICT", ErrMembershipConflict.Code)
	assert.Equal(t, "membership already exists", ErrMembershipConflict.Message)

	assert.Equal(t, "INVALID_CONTAINER_TYPE", ErrInvalidContainerType.Code)
	assert.Equal(t, "unsupported container type", ErrInvalidContainerType.Message)
}

func TestContainerErrors_ErrorInterface(t *testing.T) {
	// Test that container errors implement the error interface correctly
	err := ErrContainerNotFound
	assert.Equal(t, "CONTAINER_NOT_FOUND: container not found", err.Error())

	err = ErrContainerNotEmpty
	assert.Equal(t, "CONTAINER_NOT_EMPTY: container contains resources", err.Error())

	err = ErrCircularReference
	assert.Equal(t, "CIRCULAR_REFERENCE: circular container reference", err.Error())
}

func TestContainerErrors_WithContext(t *testing.T) {
	err := ErrContainerNotFound.WithContext("containerID", "test-container-123")

	assert.Equal(t, "CONTAINER_NOT_FOUND", err.Code)
	assert.Equal(t, "container not found", err.Message)
	assert.Equal(t, "test-container-123", err.Context["containerID"])
}

func TestContainerErrors_WithOperation(t *testing.T) {
	err := ErrContainerNotEmpty.WithOperation("DeleteContainer")

	assert.Equal(t, "CONTAINER_NOT_EMPTY", err.Code)
	assert.Equal(t, "container contains resources", err.Message)
	assert.Equal(t, "DeleteContainer", err.Operation)
}

func TestContainerErrors_WithCause(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := WrapStorageError(originalErr, "CONTAINER_OPERATION_FAILED", "failed to perform container operation")

	assert.Equal(t, "CONTAINER_OPERATION_FAILED", err.Code)
	assert.Equal(t, "failed to perform container operation", err.Message)
	assert.Equal(t, originalErr, err.Cause)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestContainerErrors_Unwrap(t *testing.T) {
	originalErr := errors.New("underlying error")
	err := WrapStorageError(originalErr, "CONTAINER_ERROR", "container operation failed")

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, originalErr, unwrapped)
}

func TestIsContainerNotFound(t *testing.T) {
	// Test with container not found error
	err := ErrContainerNotFound
	assert.True(t, IsContainerNotFound(err))

	// Test with wrapped container not found error
	wrappedErr := WrapStorageError(errors.New("db error"), "CONTAINER_NOT_FOUND", "container not found")
	assert.True(t, IsContainerNotFound(wrappedErr))

	// Test with different error
	err = ErrContainerNotEmpty
	assert.False(t, IsContainerNotFound(err))

	// Test with non-storage error
	err = &StorageError{Code: "OTHER_ERROR", Message: "other error"}
	assert.False(t, IsContainerNotFound(err))
}

func TestIsContainerNotEmpty(t *testing.T) {
	// Test with container not empty error
	err := ErrContainerNotEmpty
	assert.True(t, IsContainerNotEmpty(err))

	// Test with wrapped container not empty error
	wrappedErr := WrapStorageError(errors.New("validation error"), "CONTAINER_NOT_EMPTY", "container contains resources")
	assert.True(t, IsContainerNotEmpty(wrappedErr))

	// Test with different error
	err = ErrContainerNotFound
	assert.False(t, IsContainerNotEmpty(err))
}

func TestIsCircularReference(t *testing.T) {
	// Test with circular reference error
	err := ErrCircularReference
	assert.True(t, IsCircularReference(err))

	// Test with wrapped circular reference error
	wrappedErr := WrapStorageError(errors.New("hierarchy error"), "CIRCULAR_REFERENCE", "circular container reference")
	assert.True(t, IsCircularReference(wrappedErr))

	// Test with different error
	err = ErrInvalidHierarchy
	assert.False(t, IsCircularReference(err))
}

func TestIsInvalidHierarchy(t *testing.T) {
	// Test with invalid hierarchy error
	err := ErrInvalidHierarchy
	assert.True(t, IsInvalidHierarchy(err))

	// Test with wrapped invalid hierarchy error
	wrappedErr := WrapStorageError(errors.New("validation error"), "INVALID_HIERARCHY", "invalid container hierarchy")
	assert.True(t, IsInvalidHierarchy(wrappedErr))

	// Test with different error
	err = ErrCircularReference
	assert.False(t, IsInvalidHierarchy(err))
}

func TestIsMembershipConflict(t *testing.T) {
	// Test with membership conflict error
	err := ErrMembershipConflict
	assert.True(t, IsMembershipConflict(err))

	// Test with wrapped membership conflict error
	wrappedErr := WrapStorageError(errors.New("constraint error"), "MEMBERSHIP_CONFLICT", "membership already exists")
	assert.True(t, IsMembershipConflict(wrappedErr))

	// Test with different error
	err = ErrInvalidContainerType
	assert.False(t, IsMembershipConflict(err))
}

func TestIsInvalidContainerType(t *testing.T) {
	// Test with invalid container type error
	err := ErrInvalidContainerType
	assert.True(t, IsInvalidContainerType(err))

	// Test with wrapped invalid container type error
	wrappedErr := WrapStorageError(errors.New("validation error"), "INVALID_CONTAINER_TYPE", "unsupported container type")
	assert.True(t, IsInvalidContainerType(wrappedErr))

	// Test with different error
	err = ErrMembershipConflict
	assert.False(t, IsInvalidContainerType(err))
}

func TestNewContainerError(t *testing.T) {
	err := NewContainerError("CUSTOM_ERROR", "custom error message")

	assert.Equal(t, "CUSTOM_ERROR", err.Code)
	assert.Equal(t, "custom error message", err.Message)
	assert.NotNil(t, err.Context)
	assert.Nil(t, err.Cause)
}

func TestWrapContainerError(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapContainerError(originalErr, "WRAPPED_ERROR", "wrapped error message")

	assert.Equal(t, "WRAPPED_ERROR", err.Code)
	assert.Equal(t, "wrapped error message", err.Message)
	assert.Equal(t, originalErr, err.Cause)
	assert.NotNil(t, err.Context)
}

func TestContainerError_ChainedContext(t *testing.T) {
	err := ErrContainerNotFound.
		WithContext("containerID", "test-123").
		WithContext("operation", "GetContainer").
		WithOperation("ContainerService.GetContainer")

	assert.Equal(t, "test-123", err.Context["containerID"])
	assert.Equal(t, "GetContainer", err.Context["operation"])
	assert.Equal(t, "ContainerService.GetContainer", err.Operation)
}

func TestContainerError_ErrorMessage_WithCause(t *testing.T) {
	originalErr := errors.New("database timeout")
	err := WrapStorageError(originalErr, "CONTAINER_TIMEOUT", "container operation timed out")

	expectedMessage := "CONTAINER_TIMEOUT: container operation timed out (caused by: database timeout)"
	assert.Equal(t, expectedMessage, err.Error())
}

func TestContainerError_ErrorMessage_WithoutCause(t *testing.T) {
	err := NewStorageError("CONTAINER_INVALID", "container validation failed")

	expectedMessage := "CONTAINER_INVALID: container validation failed"
	assert.Equal(t, expectedMessage, err.Error())
}
