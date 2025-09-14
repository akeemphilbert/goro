package domain

import (
	"errors"
	"fmt"
)

// StorageError represents a storage operation error with context
type StorageError struct {
	Code      string
	Message   string
	Cause     error
	Context   map[string]any
	Operation string
}

func (e *StorageError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *StorageError) WithContext(key string, value any) *StorageError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation that caused the error
func (e *StorageError) WithOperation(operation string) *StorageError {
	e.Operation = operation
	return e
}

// Predefined error types for storage operations
var (
	// ErrResourceNotFound indicates a resource was not found
	ErrResourceNotFound = &StorageError{
		Code:    "RESOURCE_NOT_FOUND",
		Message: "resource not found",
	}

	// ErrUnsupportedFormat indicates an unsupported format was requested
	ErrUnsupportedFormat = &StorageError{
		Code:    "UNSUPPORTED_FORMAT",
		Message: "unsupported format",
	}

	// ErrInsufficientStorage indicates insufficient storage space
	ErrInsufficientStorage = &StorageError{
		Code:    "INSUFFICIENT_STORAGE",
		Message: "insufficient storage space",
	}

	// ErrDataCorruption indicates data corruption was detected
	ErrDataCorruption = &StorageError{
		Code:    "DATA_CORRUPTION",
		Message: "data corruption detected",
	}

	// ErrFormatConversion indicates format conversion failed
	ErrFormatConversion = &StorageError{
		Code:    "FORMAT_CONVERSION_FAILED",
		Message: "format conversion failed",
	}

	// ErrInvalidResource indicates invalid resource data
	ErrInvalidResource = &StorageError{
		Code:    "INVALID_RESOURCE",
		Message: "invalid resource data",
	}

	// ErrStorageOperation indicates a general storage operation failure
	ErrStorageOperation = &StorageError{
		Code:    "STORAGE_OPERATION_FAILED",
		Message: "storage operation failed",
	}

	// ErrResourceExists indicates a resource already exists
	ErrResourceExists = &StorageError{
		Code:    "RESOURCE_EXISTS",
		Message: "resource already exists",
	}

	// ErrResourceAlreadyExists indicates a resource already exists (alias for ErrResourceExists)
	ErrResourceAlreadyExists = &StorageError{
		Code:    "RESOURCE_ALREADY_EXISTS",
		Message: "resource already exists",
	}

	// ErrInvalidID indicates an invalid resource ID
	ErrInvalidID = &StorageError{
		Code:    "INVALID_ID",
		Message: "invalid resource ID",
	}

	// ErrChecksumMismatch indicates checksum validation failed
	ErrChecksumMismatch = &StorageError{
		Code:    "CHECKSUM_MISMATCH",
		Message: "checksum validation failed",
	}

	// Container-specific errors

	// ErrContainerNotFound indicates a container was not found
	ErrContainerNotFound = &StorageError{
		Code:    "CONTAINER_NOT_FOUND",
		Message: "container not found",
	}

	// ErrContainerNotEmpty indicates a container contains resources and cannot be deleted
	ErrContainerNotEmpty = &StorageError{
		Code:    "CONTAINER_NOT_EMPTY",
		Message: "container contains resources",
	}

	// ErrCircularReference indicates a circular reference in container hierarchy
	ErrCircularReference = &StorageError{
		Code:    "CIRCULAR_REFERENCE",
		Message: "circular container reference",
	}

	// ErrInvalidHierarchy indicates an invalid container hierarchy
	ErrInvalidHierarchy = &StorageError{
		Code:    "INVALID_HIERARCHY",
		Message: "invalid container hierarchy",
	}

	// ErrMembershipConflict indicates a membership already exists
	ErrMembershipConflict = &StorageError{
		Code:    "MEMBERSHIP_CONFLICT",
		Message: "membership already exists",
	}

	// ErrInvalidContainerType indicates an unsupported container type
	ErrInvalidContainerType = &StorageError{
		Code:    "INVALID_CONTAINER_TYPE",
		Message: "unsupported container type",
	}
)

// NewStorageError creates a new storage error with the given code and message
func NewStorageError(code, message string) *StorageError {
	return &StorageError{
		Code:    code,
		Message: message,
		Context: make(map[string]any),
	}
}

// WrapStorageError wraps an existing error with storage error context
func WrapStorageError(err error, code, message string) *StorageError {
	return &StorageError{
		Code:    code,
		Message: message,
		Cause:   err,
		Context: make(map[string]any),
	}
}

// IsStorageError checks if an error is a StorageError
func IsStorageError(err error) bool {
	var storageErr *StorageError
	return errors.As(err, &storageErr)
}

// GetStorageError extracts a StorageError from an error chain
func GetStorageError(err error) (*StorageError, bool) {
	var storageErr *StorageError
	if errors.As(err, &storageErr) {
		return storageErr, true
	}
	return nil, false
}

// IsResourceNotFound checks if an error indicates a resource was not found
func IsResourceNotFound(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrResourceNotFound.Code
	}
	return false
}

// IsUnsupportedFormat checks if an error indicates an unsupported format
func IsUnsupportedFormat(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrUnsupportedFormat.Code
	}
	return false
}

// IsInsufficientStorage checks if an error indicates insufficient storage
func IsInsufficientStorage(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrInsufficientStorage.Code
	}
	return false
}

// IsDataCorruption checks if an error indicates data corruption
func IsDataCorruption(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrDataCorruption.Code
	}
	return false
}

// IsFormatConversion checks if an error indicates format conversion failure
func IsFormatConversion(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrFormatConversion.Code
	}
	return false
}

// Container error helper functions

// NewContainerError creates a new container-specific storage error
func NewContainerError(code, message string) *StorageError {
	return NewStorageError(code, message)
}

// WrapContainerError wraps an existing error with container error context
func WrapContainerError(err error, code, message string) *StorageError {
	return WrapStorageError(err, code, message)
}

// IsContainerNotFound checks if an error indicates a container was not found
func IsContainerNotFound(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrContainerNotFound.Code
	}
	return false
}

// IsContainerNotEmpty checks if an error indicates a container is not empty
func IsContainerNotEmpty(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrContainerNotEmpty.Code
	}
	return false
}

// IsCircularReference checks if an error indicates a circular reference
func IsCircularReference(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrCircularReference.Code
	}
	return false
}

// IsInvalidHierarchy checks if an error indicates an invalid hierarchy
func IsInvalidHierarchy(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrInvalidHierarchy.Code
	}
	return false
}

// IsMembershipConflict checks if an error indicates a membership conflict
func IsMembershipConflict(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrMembershipConflict.Code
	}
	return false
}

// IsInvalidContainerType checks if an error indicates an invalid container type
func IsInvalidContainerType(err error) bool {
	if storageErr, ok := GetStorageError(err); ok {
		return storageErr.Code == ErrInvalidContainerType.Code
	}
	return false
}
