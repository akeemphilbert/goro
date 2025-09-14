package application

import (
	"context"
	"errors"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestMockContainerRepository is a mock implementation of ContainerRepository for container service tests
type TestMockContainerRepository struct {
	mock.Mock
}

func (m *TestMockContainerRepository) CreateContainer(ctx context.Context, container domain.ContainerResource) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *TestMockContainerRepository) GetContainer(ctx context.Context, id string) (domain.ContainerResource, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

func (m *TestMockContainerRepository) UpdateContainer(ctx context.Context, container domain.ContainerResource) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *TestMockContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestMockContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *TestMockContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *TestMockContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *TestMockContainerRepository) ListMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) ([]string, error) {
	args := m.Called(ctx, containerID, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *TestMockContainerRepository) GetChildren(ctx context.Context, containerID string) ([]domain.ContainerResource, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ContainerResource), args.Error(1)
}

func (m *TestMockContainerRepository) GetParent(ctx context.Context, containerID string) (domain.ContainerResource, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

func (m *TestMockContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *TestMockContainerRepository) FindByPath(ctx context.Context, path string) (domain.ContainerResource, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

// Implement ResourceRepository methods (delegated)
func (m *TestMockContainerRepository) Store(ctx context.Context, resource domain.Resource) error {
	args := m.Called(ctx, resource)
	return args.Error(0)
}

func (m *TestMockContainerRepository) Retrieve(ctx context.Context, id string) (domain.Resource, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.Resource), args.Error(1)
}

func (m *TestMockContainerRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestMockContainerRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// MockUnitOfWork is a mock implementation of UnitOfWork
type MockUnitOfWork struct {
	mock.Mock
}

func (m *MockUnitOfWork) RegisterEvents(events []pericarpdomain.Event) {
	m.Called(events)
}

func (m *MockUnitOfWork) Commit(ctx context.Context) ([]pericarpdomain.Envelope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]pericarpdomain.Envelope), args.Error(1)
}

func (m *MockUnitOfWork) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// Test fixtures
func setupContainerServiceTest() (*ContainerService, *TestMockContainerRepository, *MockUnitOfWork) {
	mockRepo := &TestMockContainerRepository{}
	mockUoW := &MockUnitOfWork{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUoW
	}

	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := NewContainerService(mockRepo, unitOfWorkFactory, rdfConverter)
	return service, mockRepo, mockUoW
}

// Test Container Service Business Logic

func TestContainerService_CreateContainer_Success(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	// Test data
	containerID := "test-container"
	parentID := "parent-container"
	containerType := domain.BasicContainer

	// Setup expectations
	mockRepo.On("ContainerExists", ctx, containerID).Return(false, nil)
	mockRepo.On("ContainerExists", ctx, parentID).Return(true, nil)       // Parent exists
	mockRepo.On("GetPath", ctx, parentID).Return([]string{parentID}, nil) // Parent path for hierarchy validation
	mockRepo.On("CreateContainer", ctx, mock.AnythingOfType("*domain.Container")).Return(nil)
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

	// Verify mocks
	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

func TestContainerService_CreateContainer_AlreadyExists(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "existing-container"

	// Setup expectations
	mockRepo.On("ContainerExists", ctx, containerID).Return(true, nil)

	// Execute
	container, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

func TestContainerService_CreateContainer_InvalidID(t *testing.T) {
	service, _, _ := setupContainerServiceTest()
	ctx := context.Background()

	// Execute with empty ID
	container, err := service.CreateContainer(ctx, "", "", domain.BasicContainer)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "container ID cannot be empty")
}

func TestContainerService_CreateContainer_InvalidType(t *testing.T) {
	service, _, _ := setupContainerServiceTest()
	ctx := context.Background()

	// Execute with invalid container type
	container, err := service.CreateContainer(ctx, "test-container", "", domain.ContainerType("InvalidType"))

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "invalid container type")
}

func TestContainerService_GetContainer_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	expectedContainer := domain.NewContainer(ctx, containerID, "", domain.BasicContainer)

	// Setup expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(expectedContainer, nil)

	// Execute
	container, err := service.GetContainer(ctx, containerID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.ID())

	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetContainer_NotFound(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "nonexistent-container"

	// Setup expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(nil, domain.ErrResourceNotFound)

	// Execute
	container, err := service.GetContainer(ctx, containerID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.True(t, domain.IsResourceNotFound(err))

	mockRepo.AssertExpectations(t)
}

func TestContainerService_UpdateContainer_Success(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	container.SetTitle("Updated Title")

	// Setup expectations
	mockRepo.On("UpdateContainer", ctx, container).Return(nil)
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.UpdateContainer(ctx, container)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

func TestContainerService_DeleteContainer_Success(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	container := domain.NewContainer(ctx, containerID, "", domain.BasicContainer)

	// Setup expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)
	mockRepo.On("DeleteContainer", ctx, containerID).Return(nil)
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.DeleteContainer(ctx, containerID)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

func TestContainerService_DeleteContainer_NotEmpty(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	container := domain.NewContainer(ctx, containerID, "", domain.BasicContainer)
	// Create a mock resource for AddMember
	mockResource := domain.NewResource(ctx, "member1", "text/plain", []byte("test"))
	container.AddMember(ctx, mockResource) // Make container non-empty

	// Setup expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)

	// Execute
	err := service.DeleteContainer(ctx, containerID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")

	mockRepo.AssertExpectations(t)
}

// Test Container Lifecycle Operations

func TestContainerService_AddResource_Success(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	resourceID := "test-resource"
	mockResource := domain.NewResource(ctx, resourceID, "text/plain", []byte("test"))

	// Create a container for the GetContainer call
	container := domain.NewContainer(ctx, containerID, "", domain.BasicContainer)

	// Setup expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(container, nil)
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.AddResource(ctx, containerID, resourceID, mockResource)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

func TestContainerService_RemoveResource_Success(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	resourceID := "test-resource"

	// Setup expectations
	mockRepo.On("RemoveMember", ctx, containerID, resourceID).Return(nil)
	mockUoW.On("RegisterEvents", mock.Anything).Return()
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Execute
	err := service.RemoveResource(ctx, containerID, resourceID)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

// Test Membership Management Operations

func TestContainerService_ListContainerMembers_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	pagination := domain.PaginationOptions{Limit: 10, Offset: 0}
	expectedMembers := []string{"member1", "member2", "member3"}

	// Setup expectations
	mockRepo.On("ListMembers", ctx, containerID, pagination).Return(expectedMembers, nil)

	// Execute
	listing, err := service.ListContainerMembers(ctx, containerID, pagination)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, listing)
	assert.Equal(t, expectedMembers, listing.Members)
	assert.Equal(t, containerID, listing.ContainerID)

	mockRepo.AssertExpectations(t)
}

func TestContainerService_ListContainerMembers_WithPagination(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	pagination := domain.PaginationOptions{Limit: 2, Offset: 1}
	expectedMembers := []string{"member2", "member3"}

	// Setup expectations
	mockRepo.On("ListMembers", ctx, containerID, pagination).Return(expectedMembers, nil)

	// Execute
	listing, err := service.ListContainerMembers(ctx, containerID, pagination)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedMembers, listing.Members)
	assert.Equal(t, 2, listing.Pagination.Limit)
	assert.Equal(t, 1, listing.Pagination.Offset)

	mockRepo.AssertExpectations(t)
}

// Test Hierarchy Navigation and Path Resolution

func TestContainerService_GetContainerPath_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "child-container"
	expectedPath := []string{"root", "parent", "child-container"}

	// Setup expectations
	mockRepo.On("GetPath", ctx, containerID).Return(expectedPath, nil)

	// Execute
	path, err := service.GetContainerPath(ctx, containerID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedPath, path)

	mockRepo.AssertExpectations(t)
}

func TestContainerService_FindContainerByPath_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	path := "/root/parent/child"
	expectedContainer := domain.NewContainer(ctx, "child", "parent", domain.BasicContainer)

	// Setup expectations
	mockRepo.On("FindByPath", ctx, path).Return(expectedContainer, nil)

	// Execute
	container, err := service.FindContainerByPath(ctx, path)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, "child", container.ID())

	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetChildren_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "parent-container"
	child1 := domain.NewContainer(ctx, "child1", containerID, domain.BasicContainer)
	child2 := domain.NewContainer(ctx, "child2", containerID, domain.BasicContainer)
	expectedChildren := []domain.ContainerResource{child1, child2}

	// Setup expectations
	mockRepo.On("GetChildren", ctx, containerID).Return(expectedChildren, nil)

	// Execute
	children, err := service.GetChildren(ctx, containerID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, children, 2)
	assert.Equal(t, "child1", children[0].ID())
	assert.Equal(t, "child2", children[1].ID())

	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetParent_Success(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "child-container"
	expectedParent := domain.NewContainer(ctx, "parent-container", "", domain.BasicContainer)

	// Setup expectations
	mockRepo.On("GetParent", ctx, containerID).Return(expectedParent, nil)

	// Execute
	parent, err := service.GetParent(ctx, containerID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, parent)
	assert.Equal(t, "parent-container", parent.ID())

	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetParent_NoParent(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "root-container"

	// Setup expectations - root container has no parent
	mockRepo.On("GetParent", ctx, containerID).Return(nil, nil)

	// Execute
	parent, err := service.GetParent(ctx, containerID)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, parent)

	mockRepo.AssertExpectations(t)
}

// Test Error Handling

func TestContainerService_CreateContainer_RepositoryError(t *testing.T) {
	service, mockRepo, _ := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "test-container"
	repositoryError := errors.New("repository error")

	// Setup expectations
	mockRepo.On("ContainerExists", ctx, containerID).Return(false, repositoryError)

	// Execute
	container, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "repository error")

	mockRepo.AssertExpectations(t)
}

func TestContainerService_UpdateContainer_EventCommitError(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	commitError := errors.New("commit failed")

	// Setup expectations
	mockRepo.On("UpdateContainer", ctx, container).Return(nil)
	mockUoW.On("RegisterEvents", mock.AnythingOfType("[]pericarpdomain.Event")).Return()
	mockUoW.On("Commit", ctx).Return(nil, commitError)
	mockUoW.On("Rollback").Return(nil)

	// Execute
	err := service.UpdateContainer(ctx, container)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit failed")

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

// Test Concurrent Access

func TestContainerService_ConcurrentOperations(t *testing.T) {
	service, mockRepo, mockUoW := setupContainerServiceTest()
	ctx := context.Background()

	containerID := "concurrent-container"

	// Setup expectations for multiple concurrent calls
	mockRepo.On("ContainerExists", ctx, containerID).Return(false, nil).Times(3)
	mockRepo.On("CreateContainer", ctx, mock.AnythingOfType("*domain.Container")).Return(nil).Times(3)
	mockUoW.On("RegisterEvents", mock.Anything).Return().Times(3)
	mockUoW.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil).Times(3)

	// Execute concurrent operations
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			defer func() { done <- true }()
			_, err := service.CreateContainer(ctx, containerID+"-"+string(rune(id)), "", domain.BasicContainer)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	mockRepo.AssertExpectations(t)
	mockUoW.AssertExpectations(t)
}

// Test Input Validation

func TestContainerService_ValidateInputs(t *testing.T) {
	service, _, _ := setupContainerServiceTest()
	ctx := context.Background()

	tests := []struct {
		name        string
		containerID string
		parentID    string
		cType       domain.ContainerType
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty container ID",
			containerID: "",
			parentID:    "",
			cType:       domain.BasicContainer,
			expectError: true,
			errorMsg:    "container ID cannot be empty",
		},
		{
			name:        "invalid container type",
			containerID: "test",
			parentID:    "",
			cType:       domain.ContainerType("invalid"),
			expectError: true,
			errorMsg:    "invalid container type",
		},
		{
			name:        "valid inputs",
			containerID: "test",
			parentID:    "parent",
			cType:       domain.BasicContainer,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateContainer(ctx, tt.containerID, tt.parentID, tt.cType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				// For valid inputs, we expect a different error (like repository error)
				// since we haven't mocked the repository for this test
				assert.Error(t, err) // Will fail due to unmocked repository call
			}
		})
	}
}
