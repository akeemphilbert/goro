package handlers

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test container handler creation
func TestNewContainerHandler(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)

	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockContainerService, handler.containerService)
	assert.Equal(t, mockStorageService, handler.storageService)
	assert.Equal(t, logger, handler.logger)
}

// Test container response building
func TestContainerHandler_BuildContainerResponse(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)
	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	// Create test container
	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container")
	container.AddMember("resource-1")
	container.AddMember("resource-2")

	// Create test listing
	listing := &application.ContainerListing{
		ContainerID: "test-container",
		Members:     []string{"resource-1", "resource-2"},
		Pagination:  domain.GetDefaultPagination(),
	}

	// Build response
	response := handler.buildContainerResponse(container, listing, "application/ld+json")

	// Verify response structure
	assert.Equal(t, "test-container", response["@id"])
	assert.Contains(t, response["@type"], "ldp:BasicContainer")
	assert.Contains(t, response["@type"], "ldp:Container")
	assert.Equal(t, "Test Container", response["dcterms:title"])
	assert.Equal(t, "A test container", response["dcterms:description"])
	assert.Equal(t, []string{"resource-1", "resource-2"}, response["ldp:contains"])
	assert.Equal(t, 2, response["ldp:memberCount"])

	// Verify context
	context := response["@context"].(map[string]interface{})
	assert.Equal(t, "http://www.w3.org/ns/ldp#", context["ldp"])
	assert.Equal(t, "http://purl.org/dc/terms/", context["dcterms"])
}

// Test content negotiation
func TestContainerHandler_NegotiateContentType(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)
	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	tests := []struct {
		name         string
		acceptHeader string
		expected     string
	}{
		{
			name:         "empty accept header defaults to JSON-LD",
			acceptHeader: "",
			expected:     "application/ld+json",
		},
		{
			name:         "JSON-LD exact match",
			acceptHeader: "application/ld+json",
			expected:     "application/ld+json",
		},
		{
			name:         "JSON alias to JSON-LD",
			acceptHeader: "application/json",
			expected:     "application/ld+json",
		},
		{
			name:         "Turtle format",
			acceptHeader: "text/turtle",
			expected:     "text/turtle",
		},
		{
			name:         "RDF/XML format",
			acceptHeader: "application/rdf+xml",
			expected:     "application/rdf+xml",
		},
		{
			name:         "wildcard defaults to JSON-LD",
			acceptHeader: "*/*",
			expected:     "application/ld+json",
		},
		{
			name:         "unsupported format defaults to JSON-LD",
			acceptHeader: "text/html",
			expected:     "application/ld+json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.negotiateContentType(tt.acceptHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test response content type mapping
func TestContainerHandler_GetResponseContentType(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)
	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "JSON-LD format",
			format:   "application/ld+json",
			expected: "application/ld+json",
		},
		{
			name:     "Turtle format",
			format:   "text/turtle",
			expected: "text/turtle",
		},
		{
			name:     "RDF/XML format",
			format:   "application/rdf+xml",
			expected: "application/rdf+xml",
		},
		{
			name:     "unknown format defaults to JSON-LD",
			format:   "unknown",
			expected: "application/ld+json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.getResponseContentType(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test ETag generation
func TestContainerHandler_GenerateContainerETag(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)
	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	container.AddMember("resource-1")
	container.AddMember("resource-2")

	etag := handler.generateContainerETag(container)
	expected := "test-container-2" // ID + member count

	assert.Equal(t, expected, etag)
}

// Test pagination parsing
func TestContainerHandler_ParsePaginationOptions(t *testing.T) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(io.Discard)
	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

	tests := []struct {
		name           string
		queryString    string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "no parameters uses defaults",
			queryString:    "",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name:           "valid limit and offset",
			queryString:    "limit=25&offset=10",
			expectedLimit:  25,
			expectedOffset: 10,
		},
		{
			name:           "invalid limit uses default",
			queryString:    "limit=invalid&offset=5",
			expectedLimit:  50,
			expectedOffset: 5,
		},
		{
			name:           "limit too high uses default",
			queryString:    "limit=2000",
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name:           "negative offset uses default",
			queryString:    "offset=-5",
			expectedLimit:  50,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a real HTTP request with query parameters
			req, err := http.NewRequest("GET", "http://example.com/containers/test?"+tt.queryString, nil)
			assert.NoError(t, err)

			result := handler.parsePaginationOptions(req)

			assert.Equal(t, tt.expectedLimit, result.Limit)
			assert.Equal(t, tt.expectedOffset, result.Offset)
		})
	}
}

// Test container error types (without HTTP context)
func TestContainerHandler_ErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		inputError error
		isExpected bool
	}{
		{
			name:       "resource not found error",
			inputError: domain.ErrResourceNotFound,
			isExpected: true,
		},
		{
			name:       "container not empty error",
			inputError: domain.ErrContainerNotEmpty,
			isExpected: true,
		},
		{
			name:       "invalid hierarchy error",
			inputError: domain.ErrInvalidHierarchy,
			isExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that we can identify the error types correctly
			if tt.name == "resource not found error" {
				assert.True(t, domain.IsResourceNotFound(tt.inputError))
			}
			if tt.name == "container not empty error" {
				assert.True(t, domain.IsContainerNotEmpty(tt.inputError))
			}
			if tt.name == "invalid hierarchy error" {
				assert.True(t, domain.IsInvalidHierarchy(tt.inputError))
			}
		})
	}
}

// Test container service integration
func TestContainerHandler_ServiceIntegration(t *testing.T) {
	t.Run("successful container retrieval flow", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		// Setup expectations
		container := domain.NewContainer("test-container", "", domain.BasicContainer)
		container.AddMember("resource-1")

		mockContainerService.On("GetContainer", mock.Anything, "test-container").Return(container, nil)

		listing := &application.ContainerListing{
			ContainerID: "test-container",
			Members:     []string{"resource-1"},
			Pagination:  domain.GetDefaultPagination(),
		}
		mockContainerService.On("ListContainerMembers", mock.Anything, "test-container", mock.AnythingOfType("domain.PaginationOptions")).Return(listing, nil)

		// Test the service calls directly
		retrievedContainer, err := handler.containerService.GetContainer(context.Background(), "test-container")
		assert.NoError(t, err)
		assert.Equal(t, "test-container", retrievedContainer.ID())

		retrievedListing, err := handler.containerService.ListContainerMembers(context.Background(), "test-container", domain.GetDefaultPagination())
		assert.NoError(t, err)
		assert.Equal(t, "test-container", retrievedListing.ContainerID)
		assert.Equal(t, []string{"resource-1"}, retrievedListing.Members)

		mockContainerService.AssertExpectations(t)
	})

	t.Run("successful resource creation flow", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		// Setup expectations
		mockContainerService.On("ContainerExists", mock.Anything, "test-container").Return(true, nil)

		resource := domain.NewResource("new-resource", "application/json", []byte(`{"test": "data"}`))
		mockStorageService.On("StoreResource", mock.Anything, "new-resource", mock.AnythingOfType("[]uint8"), "application/json").Return(resource, nil)

		mockContainerService.On("AddResource", mock.Anything, "test-container", "new-resource").Return(nil)

		// Test the service calls directly
		exists, err := handler.containerService.ContainerExists(context.Background(), "test-container")
		assert.NoError(t, err)
		assert.True(t, exists)

		storedResource, err := handler.storageService.StoreResource(context.Background(), "new-resource", []byte(`{"test": "data"}`), "application/json")
		assert.NoError(t, err)
		assert.Equal(t, "new-resource", storedResource.ID())

		err = handler.containerService.AddResource(context.Background(), "test-container", "new-resource")
		assert.NoError(t, err)

		mockContainerService.AssertExpectations(t)
		mockStorageService.AssertExpectations(t)
	})
}
