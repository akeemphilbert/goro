package handlers

import (
	"context"
	"io"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestContainerHandlerSimpleIntegration(t *testing.T) {
	t.Run("handler with mocked services integration", func(t *testing.T) {
		// Create handler with mocked services
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Test complete container retrieval flow
		container := domain.NewContainer("integration-test-container", "", domain.BasicContainer)
		container.SetTitle("Integration Test Container")
		container.SetDescription("A container for integration testing")
		container.AddMember("resource-1")
		container.AddMember("resource-2")

		listing := &application.ContainerListing{
			ContainerID: "integration-test-container",
			Members:     []string{"resource-1", "resource-2"},
			Pagination:  domain.GetDefaultPagination(),
		}

		// Setup expectations
		mockContainerService.On("GetContainer", mock.Anything, "integration-test-container").Return(container, nil)
		mockContainerService.On("ListContainerMembers", mock.Anything, "integration-test-container", mock.AnythingOfType("domain.PaginationOptions")).Return(listing, nil)

		// Test service calls through handler
		retrievedContainer, err := handler.containerService.GetContainer(ctx, "integration-test-container")
		assert.NoError(t, err)
		assert.Equal(t, "integration-test-container", retrievedContainer.ID())
		assert.Equal(t, "Integration Test Container", retrievedContainer.GetTitle())
		assert.Equal(t, "A container for integration testing", retrievedContainer.GetDescription())
		assert.Equal(t, 2, retrievedContainer.GetMemberCount())

		retrievedListing, err := handler.containerService.ListContainerMembers(ctx, "integration-test-container", domain.GetDefaultPagination())
		assert.NoError(t, err)
		assert.Equal(t, "integration-test-container", retrievedListing.ContainerID)
		assert.Equal(t, []string{"resource-1", "resource-2"}, retrievedListing.Members)

		// Test response building
		response := handler.buildContainerResponse(retrievedContainer, retrievedListing, "application/ld+json")
		assert.Equal(t, "integration-test-container", response["@id"])
		assert.Contains(t, response["@type"], "ldp:BasicContainer")
		assert.Equal(t, "Integration Test Container", response["dcterms:title"])
		assert.Equal(t, "A container for integration testing", response["dcterms:description"])
		assert.Equal(t, []string{"resource-1", "resource-2"}, response["ldp:contains"])
		assert.Equal(t, 2, response["ldp:memberCount"])

		// Test content negotiation
		contentType := handler.negotiateContentType("application/ld+json")
		assert.Equal(t, "application/ld+json", contentType)

		responseContentType := handler.getResponseContentType(contentType)
		assert.Equal(t, "application/ld+json", responseContentType)

		// Test ETag generation
		etag := handler.generateContainerETag(retrievedContainer)
		assert.Equal(t, "integration-test-container-2", etag)

		mockContainerService.AssertExpectations(t)
	})

	t.Run("resource creation in container flow", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Setup expectations for resource creation flow
		mockContainerService.On("ContainerExists", mock.Anything, "test-container").Return(true, nil)

		resource := domain.NewResource("new-resource", "application/json", []byte(`{"test": "data"}`))
		mockStorageService.On("StoreResource", mock.Anything, "new-resource", mock.AnythingOfType("[]uint8"), "application/json").Return(resource, nil)

		mockContainerService.On("AddResource", mock.Anything, "test-container", "new-resource").Return(nil)

		// Test the flow
		exists, err := handler.containerService.ContainerExists(ctx, "test-container")
		assert.NoError(t, err)
		assert.True(t, exists)

		storedResource, err := handler.storageService.StoreResource(ctx, "new-resource", []byte(`{"test": "data"}`), "application/json")
		assert.NoError(t, err)
		assert.Equal(t, "new-resource", storedResource.ID())
		assert.Equal(t, "application/json", storedResource.GetContentType())

		err = handler.containerService.AddResource(ctx, "test-container", "new-resource")
		assert.NoError(t, err)

		// Test resource ETag generation
		resourceETag := handler.generateResourceETag(storedResource)
		assert.Contains(t, resourceETag, "new-resource")

		mockContainerService.AssertExpectations(t)
		mockStorageService.AssertExpectations(t)
	})

	t.Run("container metadata update flow", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Create container for update
		container := domain.NewContainer("update-test-container", "", domain.BasicContainer)

		// Setup expectations
		mockContainerService.On("GetContainer", mock.Anything, "update-test-container").Return(container, nil)
		mockContainerService.On("UpdateContainer", mock.Anything, mock.AnythingOfType("*domain.Container")).Return(nil)

		// Test the flow
		retrievedContainer, err := handler.containerService.GetContainer(ctx, "update-test-container")
		assert.NoError(t, err)

		// Update metadata
		retrievedContainer.SetTitle("Updated Title")
		retrievedContainer.SetDescription("Updated Description")

		err = handler.containerService.UpdateContainer(ctx, retrievedContainer)
		assert.NoError(t, err)

		// Verify updates
		assert.Equal(t, "Updated Title", retrievedContainer.GetTitle())
		assert.Equal(t, "Updated Description", retrievedContainer.GetDescription())

		mockContainerService.AssertExpectations(t)
	})

	t.Run("container deletion flow", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Setup expectations for successful deletion
		mockContainerService.On("DeleteContainer", mock.Anything, "delete-test-container").Return(nil)

		// Test the flow
		err := handler.containerService.DeleteContainer(ctx, "delete-test-container")
		assert.NoError(t, err)

		mockContainerService.AssertExpectations(t)
	})

	t.Run("error handling integration", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Test container not found
		mockContainerService.On("GetContainer", mock.Anything, "nonexistent-container").Return(nil, domain.ErrResourceNotFound)

		_, err := handler.containerService.GetContainer(ctx, "nonexistent-container")
		assert.Error(t, err)
		assert.True(t, domain.IsResourceNotFound(err))

		// Test container not empty error
		mockContainerService.On("DeleteContainer", mock.Anything, "non-empty-container").Return(domain.ErrContainerNotEmpty)

		err = handler.containerService.DeleteContainer(ctx, "non-empty-container")
		assert.Error(t, err)
		assert.True(t, domain.IsContainerNotEmpty(err))

		mockContainerService.AssertExpectations(t)
	})

	t.Run("pagination integration", func(t *testing.T) {
		mockContainerService := &MockContainerService{}
		mockStorageService := &MockContainerStorageService{}
		logger := log.NewStdLogger(io.Discard)
		handler := NewContainerHandler(mockContainerService, mockStorageService, logger)

		ctx := context.Background()

		// Test pagination with different options
		paginationOptions := domain.PaginationOptions{
			Limit:  25,
			Offset: 10,
		}

		listing := &application.ContainerListing{
			ContainerID: "paginated-container",
			Members:     []string{"resource-1", "resource-2", "resource-3"},
			Pagination:  paginationOptions,
		}

		mockContainerService.On("ListContainerMembers", mock.Anything, "paginated-container", paginationOptions).Return(listing, nil)

		retrievedListing, err := handler.containerService.ListContainerMembers(ctx, "paginated-container", paginationOptions)
		assert.NoError(t, err)
		assert.Equal(t, "paginated-container", retrievedListing.ContainerID)
		assert.Equal(t, 25, retrievedListing.Pagination.Limit)
		assert.Equal(t, 10, retrievedListing.Pagination.Offset)
		assert.Equal(t, 3, len(retrievedListing.Members))

		mockContainerService.AssertExpectations(t)
	})
}
