package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContainerRepository is a mock implementation of ContainerRepository for testing
type MockContainerRepository struct {
	mock.Mock
}

func (m *MockContainerRepository) Store(ctx context.Context, resource *Resource) error {
	args := m.Called(ctx, resource)
	return args.Error(0)
}

func (m *MockContainerRepository) Retrieve(ctx context.Context, id string) (*Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Resource), args.Error(1)
}

func (m *MockContainerRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerRepository) CreateContainer(ctx context.Context, container *Container) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockContainerRepository) GetContainer(ctx context.Context, id string) (*Container, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Container), args.Error(1)
}

func (m *MockContainerRepository) UpdateContainer(ctx context.Context, container *Container) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *MockContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *MockContainerRepository) ListMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]string, error) {
	args := m.Called(ctx, containerID, pagination)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContainerRepository) GetChildren(ctx context.Context, containerID string) ([]*Container, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Container), args.Error(1)
}

func (m *MockContainerRepository) GetParent(ctx context.Context, containerID string) (*Container, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Container), args.Error(1)
}

func (m *MockContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContainerRepository) FindByPath(ctx context.Context, path string) (*Container, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Container), args.Error(1)
}

func (m *MockContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// Test ContainerRepository interface operations
func TestContainerRepository_CreateContainer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	container := NewContainer(ctx, "test-container", "parent", BasicContainer)

	mockRepo.On("CreateContainer", ctx, container).Return(nil)

	err := mockRepo.CreateContainer(ctx, container)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_CreateContainer_Error(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	container := NewContainer(ctx, "test-container", "parent", BasicContainer)

	expectedErr := NewStorageError("CONTAINER_CREATE_FAILED", "failed to create container")
	mockRepo.On("CreateContainer", ctx, container).Return(expectedErr)

	err := mockRepo.CreateContainer(ctx, container)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetContainer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	expectedContainer := NewContainer(ctx, "test-container", "parent", BasicContainer)

	mockRepo.On("GetContainer", ctx, "test-container").Return(expectedContainer, nil)

	container, err := mockRepo.GetContainer(ctx, "test-container")
	assert.NoError(t, err)
	assert.Equal(t, expectedContainer, container)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetContainer_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("GetContainer", ctx, "non-existent").Return(nil, ErrContainerNotFound)

	container, err := mockRepo.GetContainer(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Equal(t, ErrContainerNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_UpdateContainer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	container := NewContainer(ctx, "test-container", "parent", BasicContainer)

	mockRepo.On("UpdateContainer", ctx, container).Return(nil)

	err := mockRepo.UpdateContainer(ctx, container)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_DeleteContainer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("DeleteContainer", ctx, "test-container").Return(nil)

	err := mockRepo.DeleteContainer(ctx, "test-container")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_DeleteContainer_NotEmpty(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	expectedErr := ErrContainerNotEmpty
	mockRepo.On("DeleteContainer", ctx, "test-container").Return(expectedErr)

	err := mockRepo.DeleteContainer(ctx, "test-container")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_AddMember(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("AddMember", ctx, "container-1", "resource-1").Return(nil)

	err := mockRepo.AddMember(ctx, "container-1", "resource-1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_AddMember_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	expectedErr := ErrMembershipConflict
	mockRepo.On("AddMember", ctx, "container-1", "resource-1").Return(expectedErr)

	err := mockRepo.AddMember(ctx, "container-1", "resource-1")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_RemoveMember(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("RemoveMember", ctx, "container-1", "resource-1").Return(nil)

	err := mockRepo.RemoveMember(ctx, "container-1", "resource-1")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_RemoveMember_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	expectedErr := NewStorageError("MEMBER_NOT_FOUND", "member not found in container")
	mockRepo.On("RemoveMember", ctx, "container-1", "resource-1").Return(expectedErr)

	err := mockRepo.RemoveMember(ctx, "container-1", "resource-1")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_ListMembers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	pagination := PaginationOptions{Limit: 10, Offset: 0}
	expectedMembers := []string{"resource-1", "resource-2", "container-1"}

	mockRepo.On("ListMembers", ctx, "container-1", pagination).Return(expectedMembers, nil)

	members, err := mockRepo.ListMembers(ctx, "container-1", pagination)
	assert.NoError(t, err)
	assert.Equal(t, expectedMembers, members)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetChildren(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	expectedChildren := []*Container{
		NewContainer(ctx, "child-1", "parent", BasicContainer),
		NewContainer(ctx, "child-2", "parent", BasicContainer),
	}

	mockRepo.On("GetChildren", ctx, "parent").Return(expectedChildren, nil)

	children, err := mockRepo.GetChildren(ctx, "parent")
	assert.NoError(t, err)
	assert.Equal(t, expectedChildren, children)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetParent(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	expectedParent := NewContainer(ctx, "parent", "grandparent", BasicContainer)

	mockRepo.On("GetParent", ctx, "child").Return(expectedParent, nil)

	parent, err := mockRepo.GetParent(ctx, "child")
	assert.NoError(t, err)
	assert.Equal(t, expectedParent, parent)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetParent_Root(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("GetParent", ctx, "root").Return(nil, nil)

	parent, err := mockRepo.GetParent(ctx, "root")
	assert.NoError(t, err)
	assert.Nil(t, parent)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_GetPath(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	expectedPath := []string{"root", "parent", "child"}

	mockRepo.On("GetPath", ctx, "child").Return(expectedPath, nil)

	path, err := mockRepo.GetPath(ctx, "child")
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, path)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_FindByPath(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)
	expectedContainer := NewContainer(ctx, "child", "parent", BasicContainer)

	mockRepo.On("FindByPath", ctx, "/root/parent/child").Return(expectedContainer, nil)

	container, err := mockRepo.FindByPath(ctx, "/root/parent/child")
	assert.NoError(t, err)
	assert.Equal(t, expectedContainer, container)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_FindByPath_NotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("FindByPath", ctx, "/non/existent/path").Return(nil, ErrContainerNotFound)

	container, err := mockRepo.FindByPath(ctx, "/non/existent/path")
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Equal(t, ErrContainerNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerRepository_ContainerExists(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockContainerRepository)

	mockRepo.On("ContainerExists", ctx, "existing-container").Return(true, nil)
	mockRepo.On("ContainerExists", ctx, "non-existent").Return(false, nil)

	exists, err := mockRepo.ContainerExists(ctx, "existing-container")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = mockRepo.ContainerExists(ctx, "non-existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	mockRepo.AssertExpectations(t)
}

// Test PaginationOptions
func TestPaginationOptions_Validate(t *testing.T) {
	// Valid pagination
	pagination := PaginationOptions{Limit: 10, Offset: 0}
	assert.True(t, pagination.IsValid())

	// Invalid limit (too high)
	pagination = PaginationOptions{Limit: 1001, Offset: 0}
	assert.False(t, pagination.IsValid())

	// Invalid limit (zero)
	pagination = PaginationOptions{Limit: 0, Offset: 0}
	assert.False(t, pagination.IsValid())

	// Invalid offset (negative)
	pagination = PaginationOptions{Limit: 10, Offset: -1}
	assert.False(t, pagination.IsValid())
}

func TestPaginationOptions_GetDefaults(t *testing.T) {
	pagination := GetDefaultPagination()
	assert.Equal(t, 50, pagination.Limit)
	assert.Equal(t, 0, pagination.Offset)
	assert.True(t, pagination.IsValid())
}
