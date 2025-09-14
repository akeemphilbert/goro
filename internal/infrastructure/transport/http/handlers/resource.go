package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/segmentio/ksuid"
)

// StorageServiceInterface defines the interface for storage operations
type StorageServiceInterface interface {
	StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error)
	RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error)
	DeleteResource(ctx context.Context, id string) error
	ResourceExists(ctx context.Context, id string) (bool, error)
	StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error)
	StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string) (*domain.Resource, error)
}

// ResourceHandler handles HTTP storage operations for resources
type ResourceHandler struct {
	storageService StorageServiceInterface
	logger         log.Logger
}

// NewResourceHandler creates a new resource handler
func NewResourceHandler(storageService StorageServiceInterface, logger log.Logger) *ResourceHandler {
	return &ResourceHandler{
		storageService: storageService,
		logger:         logger,
	}
}

// GetResource handles GET requests for resource retrieval
func (h *ResourceHandler) GetResource(ctx khttp.Context) error {
	// Extract resource ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Resource ID is required")
	}

	// Get Accept header for content negotiation
	acceptHeader := ctx.Request().Header.Get("Accept")
	acceptFormat := h.negotiateContentType(acceptHeader)

	// Retrieve the resource
	resource, err := h.storageService.RetrieveResource(context.Background(), id, acceptFormat)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Set response headers
	ctx.Response().Header().Set("Content-Type", resource.GetContentType())
	ctx.Response().Header().Set("Content-Length", strconv.Itoa(resource.GetSize()))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateETag(resource)))

	// Write response body
	ctx.Response().WriteHeader(http.StatusOK)
	_, err = ctx.Response().Write(resource.GetData())
	return err
}

// PostResource handles POST requests for resource creation
func (h *ResourceHandler) PostResource(ctx khttp.Context) error {
	// Extract resource ID from path parameters (optional for POST)
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	// Generate ID if not provided (for POST to collection)
	if id == "" {
		id = h.generateResourceID()
	}

	// Get content type from request
	contentType := ctx.Request().Header.Get("Content-Type")
	if contentType == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "MISSING_CONTENT_TYPE", "Content-Type header is required")
	}

	// Read request body
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_BODY", "Failed to read request body")
	}

	if len(body) == 0 {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "EMPTY_BODY", "Request body cannot be empty")
	}

	// Store the resource
	resource, err := h.storageService.StoreResource(context.Background(), id, body, contentType)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Set response headers
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("Location", fmt.Sprintf("/resources/%s", resource.ID()))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateETag(resource)))

	// Write response
	response := map[string]interface{}{
		"id":          resource.ID(),
		"contentType": resource.GetContentType(),
		"size":        resource.GetSize(),
		"message":     "Resource created successfully",
	}

	return ctx.JSON(http.StatusCreated, response)
}

// PutResource handles PUT requests for resource creation/update
func (h *ResourceHandler) PutResource(ctx khttp.Context) error {
	// Extract resource ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Resource ID is required")
	}

	// Get content type from request
	contentType := ctx.Request().Header.Get("Content-Type")
	if contentType == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "MISSING_CONTENT_TYPE", "Content-Type header is required")
	}

	// Read request body
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_BODY", "Failed to read request body")
	}

	if len(body) == 0 {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "EMPTY_BODY", "Request body cannot be empty")
	}

	// Check if resource exists to determine response status
	exists, err := h.storageService.ResourceExists(context.Background(), id)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Store the resource
	resource, err := h.storageService.StoreResource(context.Background(), id, body, contentType)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Set response headers
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateETag(resource)))

	// Determine response status and message
	var status int
	var message string
	if exists {
		status = http.StatusOK
		message = "Resource updated successfully"
	} else {
		status = http.StatusCreated
		message = "Resource created successfully"
		ctx.Response().Header().Set("Location", fmt.Sprintf("/resources/%s", resource.ID()))
	}

	// Write response
	response := map[string]interface{}{
		"id":          resource.ID(),
		"contentType": resource.GetContentType(),
		"size":        resource.GetSize(),
		"message":     message,
	}

	return ctx.JSON(status, response)
}

// DeleteResource handles DELETE requests for resource deletion
func (h *ResourceHandler) DeleteResource(ctx khttp.Context) error {
	// Extract resource ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Resource ID is required")
	}

	// Delete the resource
	err := h.storageService.DeleteResource(context.Background(), id)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Write response
	response := map[string]interface{}{
		"id":      id,
		"message": "Resource deleted successfully",
	}

	return ctx.JSON(http.StatusOK, response)
}

// HeadResource handles HEAD requests for resource metadata
func (h *ResourceHandler) HeadResource(ctx khttp.Context) error {
	// Extract resource ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		ctx.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	// Get Accept header for content negotiation
	acceptHeader := ctx.Request().Header.Get("Accept")
	acceptFormat := h.negotiateContentType(acceptHeader)

	// Retrieve the resource
	resource, err := h.storageService.RetrieveResource(context.Background(), id, acceptFormat)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			ctx.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		if domain.IsUnsupportedFormat(err) {
			ctx.Response().WriteHeader(http.StatusNotAcceptable)
			return nil
		}
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	// Set response headers (same as GET but no body)
	ctx.Response().Header().Set("Content-Type", resource.GetContentType())
	ctx.Response().Header().Set("Content-Length", strconv.Itoa(resource.GetSize()))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateETag(resource)))

	ctx.Response().WriteHeader(http.StatusOK)
	return nil
}

// OptionsResource handles OPTIONS requests for resource endpoints
func (h *ResourceHandler) OptionsResource(ctx khttp.Context) error {
	// Set CORS headers
	ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
	ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	ctx.Response().Header().Set("Access-Control-Max-Age", "86400")

	// Return allowed methods
	response := map[string]interface{}{
		"methods": []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		"formats": []string{"application/ld+json", "text/turtle", "application/rdf+xml"},
	}

	return ctx.JSON(http.StatusOK, response)
}

// negotiateContentType performs content negotiation based on Accept header
func (h *ResourceHandler) negotiateContentType(acceptHeader string) string {
	if acceptHeader == "" {
		return ""
	}

	// Parse Accept header and find the best match
	acceptTypes := h.parseAcceptHeader(acceptHeader)

	// Supported RDF formats in order of preference
	supportedFormats := []string{
		"application/ld+json",
		"text/turtle",
		"application/rdf+xml",
	}

	// Find the best match
	for _, acceptType := range acceptTypes {
		for _, format := range supportedFormats {
			if h.matchesMediaType(acceptType.mediaType, format) {
				return format
			}
		}
	}

	// Check for wildcard acceptance
	for _, acceptType := range acceptTypes {
		if acceptType.mediaType == "*/*" || acceptType.mediaType == "application/*" {
			return supportedFormats[0] // Return preferred format
		}
	}

	return ""
}

// acceptType represents a parsed Accept header entry
type acceptType struct {
	mediaType string
	quality   float64
}

// parseAcceptHeader parses the Accept header into media types with quality values
func (h *ResourceHandler) parseAcceptHeader(acceptHeader string) []acceptType {
	var acceptTypes []acceptType

	parts := strings.Split(acceptHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split media type and parameters
		segments := strings.Split(part, ";")
		mediaType := strings.TrimSpace(segments[0])

		quality := 1.0 // Default quality

		// Parse quality parameter if present
		for i := 1; i < len(segments); i++ {
			param := strings.TrimSpace(segments[i])
			if strings.HasPrefix(param, "q=") {
				if q, err := strconv.ParseFloat(param[2:], 64); err == nil {
					quality = q
				}
			}
		}

		acceptTypes = append(acceptTypes, acceptType{
			mediaType: mediaType,
			quality:   quality,
		})
	}

	// Sort by quality (highest first)
	for i := 0; i < len(acceptTypes)-1; i++ {
		for j := i + 1; j < len(acceptTypes); j++ {
			if acceptTypes[j].quality > acceptTypes[i].quality {
				acceptTypes[i], acceptTypes[j] = acceptTypes[j], acceptTypes[i]
			}
		}
	}

	return acceptTypes
}

// matchesMediaType checks if an accept type matches a supported format
func (h *ResourceHandler) matchesMediaType(acceptType, format string) bool {
	// Exact match
	if acceptType == format {
		return true
	}

	// Handle common aliases
	switch acceptType {
	case "application/json":
		return format == "application/ld+json"
	case "text/plain":
		return format == "text/turtle"
	case "application/xml":
		return format == "application/rdf+xml"
	}

	return false
}

// handleStorageError converts storage errors to appropriate HTTP responses
func (h *ResourceHandler) handleStorageError(ctx khttp.Context, err error) error {
	// Extract storage error details if available
	storageErr, isStorageErr := domain.GetStorageError(err)

	// Log error with context
	h.logError(err, storageErr)

	// Handle specific storage error types
	if domain.IsResourceNotFound(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusNotFound, "RESOURCE_NOT_FOUND",
			"The requested resource could not be found", storageErr)
	}

	if domain.IsUnsupportedFormat(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusNotAcceptable, "UNSUPPORTED_FORMAT",
			"The requested format is not supported. Supported formats: application/ld+json, text/turtle, application/rdf+xml", storageErr)
	}

	if domain.IsInsufficientStorage(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusInsufficientStorage, "INSUFFICIENT_STORAGE",
			"Insufficient storage space available to complete the operation", storageErr)
	}

	if domain.IsDataCorruption(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusUnprocessableEntity, "DATA_CORRUPTION",
			"Data corruption detected. The resource cannot be processed safely", storageErr)
	}

	if domain.IsFormatConversion(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "FORMAT_CONVERSION_FAILED",
			"Failed to convert between the requested formats", storageErr)
	}

	// Handle other storage error types
	if isStorageErr {
		switch storageErr.Code {
		case "INVALID_ID":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_ID",
				"The provided resource ID is invalid or malformed", storageErr)
		case "INVALID_RESOURCE":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_RESOURCE",
				"The resource data is invalid or cannot be processed", storageErr)
		case "RESOURCE_EXISTS":
			return h.writeDetailedErrorResponse(ctx, http.StatusConflict, "RESOURCE_EXISTS",
				"A resource with this ID already exists", storageErr)
		case "CHECKSUM_MISMATCH":
			return h.writeDetailedErrorResponse(ctx, http.StatusUnprocessableEntity, "CHECKSUM_MISMATCH",
				"Data integrity check failed. The resource may be corrupted", storageErr)
		case "STORAGE_OPERATION_FAILED":
			return h.writeDetailedErrorResponse(ctx, http.StatusInternalServerError, "STORAGE_OPERATION_FAILED",
				"The storage operation could not be completed", storageErr)
		}
	}

	// Generic server error for unexpected errors
	return h.writeDetailedErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR",
		"An unexpected error occurred while processing the request", storageErr)
}

// writeErrorResponse writes a standardized error response
func (h *ResourceHandler) writeErrorResponse(ctx khttp.Context, status int, code, message string) error {
	return h.writeDetailedErrorResponse(ctx, status, code, message, nil)
}

// writeDetailedErrorResponse writes a comprehensive error response with additional context
func (h *ResourceHandler) writeDetailedErrorResponse(ctx khttp.Context, status int, code, message string, storageErr *domain.StorageError) error {
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("Cache-Control", "no-cache")

	// Build error response with comprehensive information
	errorResponse := map[string]interface{}{
		"code":      code,
		"message":   message,
		"status":    status,
		"timestamp": h.getCurrentTimestamp(),
	}

	// Add additional context from storage error if available
	if storageErr != nil {
		if storageErr.Operation != "" {
			errorResponse["operation"] = storageErr.Operation
		}

		if len(storageErr.Context) > 0 {
			// Only include safe context information (no sensitive data)
			safeContext := make(map[string]interface{})
			for key, value := range storageErr.Context {
				switch key {
				case "resourceID", "contentType", "format", "operation", "size":
					safeContext[key] = value
				}
			}
			if len(safeContext) > 0 {
				errorResponse["context"] = safeContext
			}
		}

		// Include cause information if it's safe to expose
		if storageErr.Cause != nil && h.shouldExposeCause(storageErr.Cause) {
			errorResponse["details"] = storageErr.Cause.Error()
		}
	}

	// Add helpful information for specific error types
	switch code {
	case "UNSUPPORTED_FORMAT":
		errorResponse["supportedFormats"] = []string{
			"application/ld+json",
			"text/turtle",
			"application/rdf+xml",
		}
	case "INSUFFICIENT_STORAGE":
		errorResponse["suggestion"] = "Try reducing the size of your request or contact the administrator"
	case "DATA_CORRUPTION":
		errorResponse["suggestion"] = "Please try uploading the resource again"
	}

	response := map[string]interface{}{
		"error": errorResponse,
	}

	return ctx.JSON(status, response)
}

// generateETag generates an ETag for a resource
func (h *ResourceHandler) generateETag(resource *domain.Resource) string {
	// Simple ETag generation based on resource ID and content hash
	// In production, this could be more sophisticated
	return fmt.Sprintf("%s-%d", resource.ID(), len(resource.GetData()))
}

// generateResourceID generates a unique resource ID using KSUID
func (h *ResourceHandler) generateResourceID() string {
	// Generate a KSUID which provides lexicographically sortable, globally unique identifiers
	// KSUIDs are 27-character base62-encoded strings that include a timestamp component
	return ksuid.New().String()
}

// logError logs errors with appropriate context and level
func (h *ResourceHandler) logError(err error, storageErr *domain.StorageError) {
	logLevel := log.LevelError

	// Adjust log level based on error type
	if storageErr != nil {
		switch storageErr.Code {
		case "RESOURCE_NOT_FOUND", "UNSUPPORTED_FORMAT", "INVALID_ID", "INVALID_RESOURCE":
			logLevel = log.LevelWarn // Client errors are warnings
		case "INSUFFICIENT_STORAGE", "DATA_CORRUPTION", "CHECKSUM_MISMATCH":
			logLevel = log.LevelError // System errors are errors
		}
	}

	// Build log context
	logContext := []interface{}{
		"msg", "Storage operation error",
		"error", err.Error(),
	}

	if storageErr != nil {
		logContext = append(logContext,
			"errorCode", storageErr.Code,
			"operation", storageErr.Operation,
		)

		// Add safe context information
		for key, value := range storageErr.Context {
			switch key {
			case "resourceID", "contentType", "format", "size":
				logContext = append(logContext, key, value)
			}
		}
	}

	h.logger.Log(logLevel, logContext...)
}

// shouldExposeCause determines if the underlying cause should be exposed to clients
func (h *ResourceHandler) shouldExposeCause(cause error) bool {
	// Only expose certain types of errors to avoid information leakage
	causeStr := cause.Error()

	// Expose format-related errors
	if strings.Contains(causeStr, "format") || strings.Contains(causeStr, "parse") {
		return true
	}

	// Expose validation errors
	if strings.Contains(causeStr, "invalid") || strings.Contains(causeStr, "validation") {
		return true
	}

	// Don't expose filesystem, network, or other system errors
	return false
}

// getCurrentTimestamp returns the current timestamp in ISO 8601 format
func (h *ResourceHandler) getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
