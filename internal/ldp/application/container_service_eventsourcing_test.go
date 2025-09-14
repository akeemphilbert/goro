package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Tests for event sourcing behavior - containers should only be persisted via event handlers

func TestContainerService_EventSourcing_CreateContainer(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	parentID := "parent-container"
	containerType := domain.BasicContainer

	// Setup expectations - only for validation, NO repository persistence calls
	mockRepo.On("ContainerExists", ctx, containerID).Return(false, nil)
	mockRepo.On("ContainerExists", ctx, parentID).Return(true, nil)
	mockRepo.On("GetPath", ctx, parentID).Return([]string{parentID}, nil)

	// Mock for hierarchy validation
	parentContainer := domain.NewContainer(parentID, "", domain.BasicContainer)
	mockRepo.On("GetContainer", ctx, parentID).Return(parentContainer, nil)

	// Event sourcing expectations
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	container, err := service.CreateContainer(ctx, containerID, parentID, containerType)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.ID())
	assert.Equal(t, parentID, container.ParentID)
	assert.Equal(t, containerType, container.ContainerType)

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "CreateContainer")

	// Verify event sourcing methods were called
	mockUoW.AssertCalled(t, "RegisterEvents", mock.Anything)
	mockUoW.AssertCalled(t, "Commit", ctx)
}

func TestContainerService_EventSourcing_UpdateContainer(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	container.SetTitle("Updated Title")

	// Setup expectations - NO repository persistence calls
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.UpdateContainer(ctx, container)

	// Assert
	require.NoError(t, err)

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "UpdateContainer")

	// Verify event sourcing methods were called
	mockUoW.AssertCalled(t, "RegisterEvents", mock.Anything)
	mockUoW.AssertCalled(t, "Commit", ctx)
}

func TestContainerService_EventSourcing_DeleteContainer(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	container := domain.NewContainer(containerID, "", domain.BasicContainer)

	// Setup expectations - only for validation, NO repository persistence calls
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)
	mockRepo.On("GetChildren", ctx, containerID).Return([]*domain.Container{}, nil) // Empty container can be deleted

	// Event sourcing expectations
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.DeleteContainer(ctx, containerID)

	// Assert
	require.NoError(t, err)

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "DeleteContainer")

	// Verify event sourcing methods were called
	mockUoW.AssertCalled(t, "RegisterEvents", mock.Anything)
	mockUoW.AssertCalled(t, "Commit", ctx)
}

func TestContainerService_EventSourcing_AddResource(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	resourceID := "test-resource"

	// Setup expectations - validation mocks, NO repository persistence calls
	container := domain.NewContainer(containerID, "", domain.BasicContainer)
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)
	mockRepo.On("GetContainer", ctx, resourceID).Return(nil, domain.ErrContainerNotFound) // Resource is not a container

	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.AddResource(ctx, containerID, resourceID)

	// Assert
	require.NoError(t, err)

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "AddMember")

	// Verify event sourcing methods were called
	mockUoW.AssertCalled(t, "RegisterEvents", mock.Anything)
	mockUoW.AssertCalled(t, "Commit", ctx)
}

func TestContainerService_EventSourcing_RemoveResource(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	resourceID := "test-resource"

	// Setup expectations - validation mocks, NO repository persistence calls
	container := domain.NewContainer(containerID, "", domain.BasicContainer)
	container.AddMember(resourceID) // Add the resource so it can be removed
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)

	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.RemoveResource(ctx, containerID, resourceID)

	// Assert
	require.NoError(t, err)

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "RemoveMember")

	// Verify event sourcing methods were called
	mockUoW.AssertCalled(t, "RegisterEvents", mock.Anything)
	mockUoW.AssertCalled(t, "Commit", ctx)
}

func TestContainerService_EventSourcing_EventCommitFailure(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"

	// Setup expectations - validation passes but commit fails
	mockRepo.On("ContainerExists", ctx, containerID).Return(false, nil)
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return(nil, assert.AnError)
	mockUoW.On("Rollback").Return(nil)

	// Execute
	container, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to commit")

	// Verify that NO repository persistence methods were called
	mockRepo.AssertNotCalled(t, "CreateContainer")

	// Verify rollback was called
	mockUoW.AssertCalled(t, "Rollback")
}

// Test that read operations still use the repository directly (not event sourced)
func TestContainerService_ReadOperations_UseRepositoryDirectly(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	expectedContainer := domain.NewContainer(containerID, "", domain.BasicContainer)

	// Setup expectations for read operations
	mockRepo.On("GetContainer", ctx, containerID).Return(expectedContainer, nil)
	mockRepo.On("ListMembers", ctx, containerID, mock.Anything).Return([]string{"member1"}, nil)
	mockRepo.On("GetPath", ctx, containerID).Return([]string{containerID}, nil)
	mockRepo.On("FindByPath", ctx, "/path").Return(expectedContainer, nil)
	mockRepo.On("GetChildren", ctx, containerID).Return([]*domain.Container{}, nil)
	mockRepo.On("GetParent", ctx, containerID).Return(nil, nil)
	mockRepo.On("ContainerExists", ctx, containerID).Return(true, nil)

	// Test GetContainer
	container, err := service.GetContainer(ctx, containerID)
	require.NoError(t, err)
	assert.Equal(t, containerID, container.ID())

	// Test ListContainerMembers
	listing, err := service.ListContainerMembers(ctx, containerID, domain.GetDefaultPagination())
	require.NoError(t, err)
	assert.Equal(t, containerID, listing.ContainerID)

	// Test GetContainerPath
	path, err := service.GetContainerPath(ctx, containerID)
	require.NoError(t, err)
	assert.Equal(t, []string{containerID}, path)

	// Test FindContainerByPath
	foundContainer, err := service.FindContainerByPath(ctx, "/path")
	require.NoError(t, err)
	assert.Equal(t, containerID, foundContainer.ID())

	// Test GetChildren
	children, err := service.GetChildren(ctx, containerID)
	require.NoError(t, err)
	assert.Empty(t, children)

	// Test GetParent
	parent, err := service.GetParent(ctx, containerID)
	require.NoError(t, err)
	assert.Nil(t, parent)

	// Test ContainerExists
	exists, err := service.ContainerExists(ctx, containerID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify all repository read methods were called
	mockRepo.AssertExpectations(t)
}
