package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple tests that focus on business logic without complex event handling

func TestContainerService_CreateContainer_ValidationLogic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	tests := []struct {
		name        string
		containerID string
		parentID    string
		cType       domain.ContainerType
		setupMocks  func()
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty container ID",
			containerID: "",
			parentID:    "",
			cType:       domain.BasicContainer,
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "container ID cannot be empty",
		},
		{
			name:        "invalid container type",
			containerID: "test",
			parentID:    "",
			cType:       domain.ContainerType("invalid"),
			setupMocks:  func() {},
			expectError: true,
			errorMsg:    "unsupported container type",
		},
		{
			name:        "container already exists",
			containerID: "existing",
			parentID:    "",
			cType:       domain.BasicContainer,
			setupMocks: func() {
				mockRepo.On("ContainerExists", ctx, "existing").Return(true, nil)
			},
			expectError: true,
			errorMsg:    "already exists",
		},
		{
			name:        "parent container not found",
			containerID: "child",
			parentID:    "nonexistent-parent",
			cType:       domain.BasicContainer,
			setupMocks: func() {
				mockRepo.On("ContainerExists", ctx, "child").Return(false, nil)
				mockRepo.On("ContainerExists", ctx, "nonexistent-parent").Return(false, nil)
				mockRepo.On("GetContainer", ctx, "nonexistent-parent").Return(nil, domain.ErrContainerNotFound)
			},
			expectError: true,
			errorMsg:    "parent container not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			// Setup test-specific mocks
			tt.setupMocks()

			// Execute
			container, err := service.CreateContainer(ctx, tt.containerID, tt.parentID, tt.cType)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, container)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, container)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestContainerService_GetContainer_Logic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("empty ID validation", func(t *testing.T) {
		container, err := service.GetContainer(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, container)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("container not found", func(t *testing.T) {
		mockRepo.On("GetContainer", ctx, "nonexistent").Return(nil, domain.ErrResourceNotFound)

		container, err := service.GetContainer(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, container)
		assert.True(t, domain.IsResourceNotFound(err))

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful retrieval", func(t *testing.T) {
		expectedContainer := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
		mockRepo.On("GetContainer", ctx, "test-container").Return(expectedContainer, nil)

		container, err := service.GetContainer(ctx, "test-container")
		require.NoError(t, err)
		assert.NotNil(t, container)
		assert.Equal(t, "test-container", container.ID())

		mockRepo.AssertExpectations(t)
	})
}

func TestContainerService_DeleteContainer_Logic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("empty ID validation", func(t *testing.T) {
		err := service.DeleteContainer(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("container not found", func(t *testing.T) {
		mockRepo.On("GetContainer", ctx, "nonexistent").Return(nil, domain.ErrResourceNotFound)

		err := service.DeleteContainer(ctx, "nonexistent")
		assert.Error(t, err)
		assert.True(t, domain.IsResourceNotFound(err))

		mockRepo.AssertExpectations(t)
	})

	t.Run("container not empty", func(t *testing.T) {
		container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
		container.AddMember(ctx, domain.NewResource(ctx, "member1", "text/plain", []byte("Hello, World!"))) // Make container non-empty

		mockRepo.On("GetContainer", ctx, "test-container").Return(container, nil)

		err := service.DeleteContainer(ctx, "test-container")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not empty")

		mockRepo.AssertExpectations(t)
	})
}

func TestContainerService_MembershipOperations_Logic(t *testing.T) {
	service, _, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("add resource - empty container ID", func(t *testing.T) {
		resource := domain.NewResource(ctx, "resource1", "text/plain", []byte("Hello, World!"))
		err := service.AddResource(ctx, "", "resource1", resource)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("add resource - empty resource ID", func(t *testing.T) {
		resource := domain.NewResource(ctx, "resource1", "text/plain", []byte("Hello, World!"))
		err := service.AddResource(ctx, "container1", "", resource)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource ID cannot be empty")
	})

	t.Run("remove resource - empty container ID", func(t *testing.T) {
		err := service.RemoveResource(ctx, "", "resource1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("remove resource - empty resource ID", func(t *testing.T) {
		err := service.RemoveResource(ctx, "container1", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource ID cannot be empty")
	})
}

func TestContainerService_ListMembers_Logic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("empty container ID", func(t *testing.T) {
		listing, err := service.ListContainerMembers(ctx, "", domain.PaginationOptions{})
		assert.Error(t, err)
		assert.Nil(t, listing)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("successful listing", func(t *testing.T) {
		expectedMembers := []string{"member1", "member2"}
		pagination := domain.PaginationOptions{Limit: 10, Offset: 0}

		mockRepo.On("ListMembers", ctx, "test-container", pagination).Return(expectedMembers, nil)

		listing, err := service.ListContainerMembers(ctx, "test-container", pagination)
		require.NoError(t, err)
		assert.NotNil(t, listing)
		assert.Equal(t, "test-container", listing.ContainerID)
		assert.Equal(t, expectedMembers, listing.Members)
		assert.Equal(t, pagination, listing.Pagination)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid pagination uses defaults", func(t *testing.T) {
		expectedMembers := []string{"member1"}
		invalidPagination := domain.PaginationOptions{Limit: -1, Offset: -1}
		defaultPagination := domain.GetDefaultPagination()

		mockRepo.On("ListMembers", ctx, "test-container", defaultPagination).Return(expectedMembers, nil)

		listing, err := service.ListContainerMembers(ctx, "test-container", invalidPagination)
		require.NoError(t, err)
		assert.Equal(t, defaultPagination, listing.Pagination)

		mockRepo.AssertExpectations(t)
	})
}

func TestContainerService_HierarchyNavigation_Logic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("get path - empty container ID", func(t *testing.T) {
		path, err := service.GetContainerPath(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, path)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("get path - successful", func(t *testing.T) {
		expectedPath := []string{"root", "parent", "child"}
		mockRepo.On("GetPath", ctx, "child").Return(expectedPath, nil)

		path, err := service.GetContainerPath(ctx, "child")
		require.NoError(t, err)
		assert.Equal(t, expectedPath, path)

		mockRepo.AssertExpectations(t)
	})

	t.Run("find by path - empty path", func(t *testing.T) {
		container, err := service.FindContainerByPath(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, container)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	t.Run("find by path - successful", func(t *testing.T) {
		expectedContainer := domain.NewContainer(context.Background(), "child", "parent", domain.BasicContainer)
		mockRepo.On("FindByPath", ctx, "/root/parent/child").Return(expectedContainer, nil)

		container, err := service.FindContainerByPath(ctx, "/root/parent/child")
		require.NoError(t, err)
		assert.NotNil(t, container)
		assert.Equal(t, "child", container.ID())

		mockRepo.AssertExpectations(t)
	})

	t.Run("get children - empty container ID", func(t *testing.T) {
		children, err := service.GetChildren(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, children)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("get parent - empty container ID", func(t *testing.T) {
		parent, err := service.GetParent(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, parent)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})
}

func TestContainerService_ContainerExists_Logic(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	t.Run("empty container ID", func(t *testing.T) {
		exists, err := service.ContainerExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "container ID cannot be empty")
	})

	t.Run("container exists", func(t *testing.T) {
		mockRepo.On("ContainerExists", ctx, "existing").Return(true, nil)

		exists, err := service.ContainerExists(ctx, "existing")
		require.NoError(t, err)
		assert.True(t, exists)

		mockRepo.AssertExpectations(t)
	})

	t.Run("container does not exist", func(t *testing.T) {
		mockRepo.On("ContainerExists", ctx, "nonexistent").Return(false, nil)

		exists, err := service.ContainerExists(ctx, "nonexistent")
		require.NoError(t, err)
		assert.False(t, exists)

		mockRepo.AssertExpectations(t)
	})
}
