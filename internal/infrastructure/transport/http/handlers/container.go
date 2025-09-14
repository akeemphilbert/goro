package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/segmentio/ksuid"
)

// ContainerHandler handles HTTP operations for LDP containers
type ContainerHandler struct {
	containerService ContainerServiceInterface
	storageService   StorageServiceInterface
	logger           log.Logger
}

// NewContainerHandler creates a new container handler
func NewContainerHandler(containerService ContainerServiceInterface, storageService StorageServiceInterface, logger log.Logger) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
		storageService:   storageService,
		logger:           logger,
	}
}

// ContainerMetadataUpdate represents the structure for container metadata updates
type ContainerMetadataUpdate struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// GetContainer handles GET requests for container retrieval with member listing
func (h *ContainerHandler) GetContainer(ctx khttp.Context) error {
	// Extract container ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Container ID is required")
	}

	// Get Accept header for content negotiation
	acceptHeader := ctx.Request().Header.Get("Accept")
	acceptFormat := h.negotiateContentType(acceptHeader)

	// Retrieve container
	container, err := h.containerService.GetContainer(context.Background(), id)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	// Get container members with pagination
	pagination := h.parsePaginationOptions(ctx.Request())
	listing, err := h.containerService.ListContainerMembers(context.Background(), id, pagination)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	// Build container response
	response := h.buildContainerResponse(container, listing, acceptFormat)

	// Set LDP-specific headers
	h.setLDPHeaders(ctx, container)
	ctx.Response().Header().Set("Content-Type", h.getResponseContentType(acceptFormat))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateContainerETag(container)))

	return ctx.JSON(http.StatusOK, response)
}

// PostResource handles POST requests for resource creation in containers
func (h *ContainerHandler) PostResource(ctx khttp.Context) error {
	// Extract container ID from path parameters
	vars := ctx.Vars()
	containerID := ""
	if len(vars["id"]) > 0 {
		containerID = vars["id"][0]
	}

	if containerID == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Container ID is required")
	}

	// Check if container exists
	exists, err := h.containerService.ContainerExists(context.Background(), containerID)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	if !exists {
		return h.writeErrorResponse(ctx, http.StatusNotFound, "CONTAINER_NOT_FOUND", "Container not found")
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

	// Generate resource ID
	resourceID := h.generateResourceID()

	// Store the resource
	resource, err := h.storageService.StoreResource(context.Background(), resourceID, body, contentType)
	if err != nil {
		return h.handleStorageError(ctx, err)
	}

	// Add resource to container
	err = h.containerService.AddResource(context.Background(), containerID, resourceID, resource)
	if err != nil {
		// If adding to container fails, we should clean up the resource
		if deleteErr := h.storageService.DeleteResource(context.Background(), resourceID); deleteErr != nil {
			h.logger.Log(log.LevelWarn, "msg", "Failed to cleanup resource after container add failure",
				"resourceID", resourceID, "error", deleteErr.Error())
		}
		return h.handleContainerError(ctx, err)
	}

	// Set response headers
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("Location", fmt.Sprintf("/resources/%s", resource.ID()))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateResourceETag(resource)))

	// Build response
	response := map[string]interface{}{
		"id":          resource.ID(),
		"contentType": resource.GetContentType(),
		"size":        resource.GetSize(),
		"containerID": containerID,
		"message":     "Resource created in container successfully",
	}

	return ctx.JSON(http.StatusCreated, response)
}

// PutContainer handles PUT requests for container metadata updates
func (h *ContainerHandler) PutContainer(ctx khttp.Context) error {
	// Extract container ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Container ID is required")
	}

	// Read request body
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_BODY", "Failed to read request body")
	}

	// Parse metadata update
	var update ContainerMetadataUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
	}

	// Retrieve existing container
	container, err := h.containerService.GetContainer(context.Background(), id)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	// Update container metadata
	if update.Title != "" {
		container.SetTitle(update.Title)
	}
	if update.Description != "" {
		container.SetDescription(update.Description)
	}

	// Save updated container
	err = h.containerService.UpdateContainer(context.Background(), container)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	// Set response headers
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateContainerETag(container)))

	// Build response
	response := map[string]interface{}{
		"id":          container.ID(),
		"title":       container.GetTitle(),
		"description": container.GetDescription(),
		"message":     "Container updated successfully",
	}

	return ctx.JSON(http.StatusOK, response)
}

// DeleteContainer handles DELETE requests for container deletion with empty validation
func (h *ContainerHandler) DeleteContainer(ctx khttp.Context) error {
	// Extract container ID from path parameters
	vars := ctx.Vars()
	id := ""
	if len(vars["id"]) > 0 {
		id = vars["id"][0]
	}

	if id == "" {
		return h.writeErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Container ID is required")
	}

	// Delete the container (service will validate it's empty)
	err := h.containerService.DeleteContainer(context.Background(), id)
	if err != nil {
		return h.handleContainerError(ctx, err)
	}

	// Build response
	response := map[string]interface{}{
		"id":      id,
		"message": "Container deleted successfully",
	}

	return ctx.JSON(http.StatusOK, response)
}

// HeadContainer handles HEAD requests for container metadata
func (h *ContainerHandler) HeadContainer(ctx khttp.Context) error {
	// Extract container ID from path parameters
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

	// Retrieve container
	container, err := h.containerService.GetContainer(context.Background(), id)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			ctx.Response().WriteHeader(http.StatusNotFound)
			return nil
		}
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	// Set response headers (same as GET but no body)
	h.setLDPHeaders(ctx, container)
	ctx.Response().Header().Set("Content-Type", h.getResponseContentType(acceptFormat))
	ctx.Response().Header().Set("ETag", fmt.Sprintf(`"%s"`, h.generateContainerETag(container)))

	ctx.Response().WriteHeader(http.StatusOK)
	return nil
}

// OptionsContainer handles OPTIONS requests for container endpoints
func (h *ContainerHandler) OptionsContainer(ctx khttp.Context) error {
	// Set CORS headers
	ctx.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
	ctx.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	ctx.Response().Header().Set("Access-Control-Max-Age", "86400")

	// Set LDP headers
	ctx.Response().Header().Set("Link", `<http://www.w3.org/ns/ldp#BasicContainer>; rel="type"`)

	// Return allowed methods and formats
	response := map[string]interface{}{
		"methods": []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		"formats": []string{"application/ld+json", "text/turtle", "application/rdf+xml"},
		"ldpType": "BasicContainer",
	}

	return ctx.JSON(http.StatusOK, response)
}

// Helper methods

// parsePaginationOptions extracts pagination parameters from request
func (h *ContainerHandler) parsePaginationOptions(req *http.Request) domain.PaginationOptions {
	pagination := domain.GetDefaultPagination()

	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			pagination.Limit = limit
		}
	}

	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			pagination.Offset = offset
		}
	}

	return pagination
}

// buildContainerResponse builds the container response based on format
func (h *ContainerHandler) buildContainerResponse(container *domain.Container, listing *application.ContainerListing, format string) map[string]interface{} {
	response := map[string]interface{}{
		"@context": map[string]interface{}{
			"ldp":     "http://www.w3.org/ns/ldp#",
			"dcterms": "http://purl.org/dc/terms/",
		},
		"@id":          container.ID(),
		"@type":        []string{"ldp:BasicContainer", "ldp:Container"},
		"ldp:contains": listing.Members,
	}

	// Add metadata if available
	if title := container.GetTitle(); title != "" {
		response["dcterms:title"] = title
	}
	if description := container.GetDescription(); description != "" {
		response["dcterms:description"] = description
	}

	// Add timestamps from metadata
	if metadata := container.GetMetadata(); metadata != nil {
		if createdAt, exists := metadata["createdAt"]; exists {
			if t, ok := createdAt.(time.Time); ok {
				response["dcterms:created"] = t.Format(time.RFC3339)
			}
		}
		if updatedAt, exists := metadata["updatedAt"]; exists {
			if t, ok := updatedAt.(time.Time); ok {
				response["dcterms:modified"] = t.Format(time.RFC3339)
			}
		}
	}

	// Add member count
	response["ldp:memberCount"] = len(listing.Members)

	// Add pagination info if applicable
	if listing.Pagination.Limit != domain.GetDefaultPagination().Limit || listing.Pagination.Offset != 0 {
		response["pagination"] = map[string]interface{}{
			"limit":  listing.Pagination.Limit,
			"offset": listing.Pagination.Offset,
		}
	}

	return response
}

// setLDPHeaders sets LDP-specific response headers
func (h *ContainerHandler) setLDPHeaders(ctx khttp.Context, container *domain.Container) {
	ctx.Response().Header().Set("Link", `<http://www.w3.org/ns/ldp#BasicContainer>; rel="type"`)
	ctx.Response().Header().Set("Accept-Post", "text/turtle, application/ld+json, application/rdf+xml")
	ctx.Response().Header().Set("Allow", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
}

// negotiateContentType performs content negotiation for containers
func (h *ContainerHandler) negotiateContentType(acceptHeader string) string {
	if acceptHeader == "" {
		return "application/ld+json" // Default for containers
	}

	// Simple content negotiation - prefer JSON-LD for containers
	if strings.Contains(acceptHeader, "application/ld+json") || strings.Contains(acceptHeader, "application/json") {
		return "application/ld+json"
	}
	if strings.Contains(acceptHeader, "text/turtle") {
		return "text/turtle"
	}
	if strings.Contains(acceptHeader, "application/rdf+xml") {
		return "application/rdf+xml"
	}
	if strings.Contains(acceptHeader, "*/*") {
		return "application/ld+json"
	}

	return "application/ld+json"
}

// getResponseContentType returns the appropriate content type for response
func (h *ContainerHandler) getResponseContentType(format string) string {
	switch format {
	case "text/turtle":
		return "text/turtle"
	case "application/rdf+xml":
		return "application/rdf+xml"
	default:
		return "application/ld+json"
	}
}

// generateContainerETag generates an ETag for a container
func (h *ContainerHandler) generateContainerETag(container *domain.Container) string {
	// Generate ETag based on container ID and updated timestamp
	metadata := container.GetMetadata()
	if updatedAt, exists := metadata["updatedAt"]; exists {
		if t, ok := updatedAt.(time.Time); ok {
			return fmt.Sprintf("%s-%d", container.ID(), t.Unix())
		}
	}
	// Fallback to just container ID if no timestamp
	return container.ID()
}

// generateResourceETag generates an ETag for a resource
func (h *ContainerHandler) generateResourceETag(resource *domain.Resource) string {
	return fmt.Sprintf("%s-%d", resource.ID(), len(resource.GetData()))
}

// generateResourceID generates a unique resource ID
func (h *ContainerHandler) generateResourceID() string {
	return ksuid.New().String()
}

// handleContainerError converts container service errors to HTTP responses
func (h *ContainerHandler) handleContainerError(ctx khttp.Context, err error) error {
	// Extract storage error details if available
	storageErr, isStorageErr := domain.GetStorageError(err)

	// Log error with context
	h.logError(err, storageErr)

	// Handle specific container error types
	if domain.IsResourceNotFound(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusNotFound, "CONTAINER_NOT_FOUND",
			"The requested container could not be found", storageErr)
	}

	if domain.IsContainerNotEmpty(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusConflict, "CONTAINER_NOT_EMPTY",
			"Container cannot be deleted because it contains resources", storageErr)
	}

	if domain.IsInvalidHierarchy(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_HIERARCHY",
			"Invalid container hierarchy or circular reference detected", storageErr)
	}

	if storageErr != nil && storageErr.Code == domain.ErrResourceAlreadyExists.Code {
		return h.writeDetailedErrorResponse(ctx, http.StatusConflict, "CONTAINER_EXISTS",
			"A container with this ID already exists", storageErr)
	}

	// Handle other storage error types
	if isStorageErr {
		switch storageErr.Code {
		case "INVALID_ID":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_ID",
				"The provided container ID is invalid or malformed", storageErr)
		case "INVALID_RESOURCE":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_CONTAINER",
				"The container data is invalid or cannot be processed", storageErr)
		case "STORAGE_OPERATION_FAILED":
			return h.writeDetailedErrorResponse(ctx, http.StatusInternalServerError, "STORAGE_OPERATION_FAILED",
				"The container operation could not be completed", storageErr)
		}
	}

	// Generic server error for unexpected errors
	return h.writeDetailedErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR",
		"An unexpected error occurred while processing the container request", storageErr)
}

// handleStorageError handles storage service errors (reuse from resource handler)
func (h *ContainerHandler) handleStorageError(ctx khttp.Context, err error) error {
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
			"The requested format is not supported", storageErr)
	}

	if domain.IsInsufficientStorage(err) {
		return h.writeDetailedErrorResponse(ctx, http.StatusInsufficientStorage, "INSUFFICIENT_STORAGE",
			"Insufficient storage space available", storageErr)
	}

	// Handle other storage error types
	if isStorageErr {
		switch storageErr.Code {
		case "INVALID_ID":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_ID",
				"The provided resource ID is invalid", storageErr)
		case "INVALID_RESOURCE":
			return h.writeDetailedErrorResponse(ctx, http.StatusBadRequest, "INVALID_RESOURCE",
				"The resource data is invalid", storageErr)
		}
	}

	// Generic server error
	return h.writeDetailedErrorResponse(ctx, http.StatusInternalServerError, "INTERNAL_ERROR",
		"An unexpected error occurred", storageErr)
}

// writeErrorResponse writes a standardized error response
func (h *ContainerHandler) writeErrorResponse(ctx khttp.Context, status int, code, message string) error {
	return h.writeDetailedErrorResponse(ctx, status, code, message, nil)
}

// writeDetailedErrorResponse writes a comprehensive error response
func (h *ContainerHandler) writeDetailedErrorResponse(ctx khttp.Context, status int, code, message string, storageErr *domain.StorageError) error {
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().Header().Set("Cache-Control", "no-cache")

	// Build error response
	errorResponse := map[string]interface{}{
		"code":      code,
		"message":   message,
		"status":    status,
		"timestamp": time.Now().Unix(),
	}

	// Add additional context from storage error if available
	if storageErr != nil {
		if storageErr.Operation != "" {
			errorResponse["operation"] = storageErr.Operation
		}

		if len(storageErr.Context) > 0 {
			// Only include safe context information
			safeContext := make(map[string]interface{})
			for key, value := range storageErr.Context {
				switch key {
				case "containerID", "resourceID", "contentType", "format", "operation":
					safeContext[key] = value
				}
			}
			if len(safeContext) > 0 {
				errorResponse["context"] = safeContext
			}
		}
	}

	response := map[string]interface{}{
		"error": errorResponse,
	}

	return ctx.JSON(status, response)
}

// logError logs errors with appropriate context
func (h *ContainerHandler) logError(err error, storageErr *domain.StorageError) {
	logLevel := log.LevelError

	// Adjust log level based on error type
	if storageErr != nil {
		switch storageErr.Code {
		case "RESOURCE_NOT_FOUND", "CONTAINER_NOT_FOUND", "INVALID_ID":
			logLevel = log.LevelWarn // Client errors are warnings
		}
	}

	// Build log context
	logContext := []interface{}{
		"msg", "Container operation error",
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
			case "containerID", "resourceID", "contentType", "format":
				logContext = append(logContext, key, value)
			}
		}
	}

	h.logger.Log(logLevel, logContext...)
}
