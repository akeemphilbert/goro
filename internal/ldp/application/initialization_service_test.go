package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockContainerRepository is a mock implementation of domain.ContainerRepository
type MockContainerRepository struct {
	mock.Mock
}

func (m *MockContainerRepository) Store(ctx context.Context, resource domain.Resource) error {
	args := m.Called(ctx, resource)
	return args.Error(0)
}

func (m *MockContainerRepository) Retrieve(ctx context.Context, id string) (domain.Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Resource), args.Error(1)
}

func (m *MockContainerRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerRepository) CreateContainer(ctx context.Context, container domain.ContainerResource) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockContainerRepository) GetContainer(ctx context.Context, id string) (domain.ContainerResource, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

func (m *MockContainerRepository) UpdateContainer(ctx context.Context, container domain.ContainerResource) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *MockContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	args := m.Called(ctx, containerID, memberID)
	return args.Error(0)
}

func (m *MockContainerRepository) ListMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) ([]string, error) {
	args := m.Called(ctx, containerID, pagination)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContainerRepository) GetChildren(ctx context.Context, containerID string) ([]domain.ContainerResource, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).([]domain.ContainerResource), args.Error(1)
}

func (m *MockContainerRepository) GetParent(ctx context.Context, containerID string) (domain.ContainerResource, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

func (m *MockContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContainerRepository) FindByPath(ctx context.Context, path string) (domain.ContainerResource, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.ContainerResource), args.Error(1)
}

func TestNewInitializationService(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())

	service := NewInitializationService(mockRepo, logger)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.containerRepo)
	assert.NotNil(t, service.logger)
}

func TestInitializationService_EnsureRootContainer_DoesNotExist(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock that root container doesn't exist
	mockRepo.On("ContainerExists", ctx, "/").Return(false, nil)

	// Mock successful creation
	mockRepo.On("CreateContainer", ctx, mock.AnythingOfType("*domain.Container")).Return(nil)

	err := service.EnsureRootContainer(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_EnsureRootContainer_AlreadyExists(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock that root container already exists
	mockRepo.On("ContainerExists", ctx, "/").Return(true, nil)

	err := service.EnsureRootContainer(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Ensure CreateContainer was not called
	mockRepo.AssertNotCalled(t, "CreateContainer")
}

func TestInitializationService_EnsureRootContainer_CheckExistenceError(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock error when checking existence
	mockRepo.On("ContainerExists", ctx, "/").Return(false, assert.AnError)

	err := service.EnsureRootContainer(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check root container existence")
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_EnsureRootContainer_CreateError(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock that root container doesn't exist
	mockRepo.On("ContainerExists", ctx, "/").Return(false, nil)

	// Mock creation error
	mockRepo.On("CreateContainer", ctx, mock.AnythingOfType("*domain.Container")).Return(assert.AnError)

	err := service.EnsureRootContainer(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create root container")
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_GetRootContainer(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Create a mock root container
	rootContainer := domain.NewContainer(ctx, "/", "", domain.BasicContainer)
	rootContainer.SetTitle("Root Container")

	// Mock successful retrieval
	mockRepo.On("GetContainer", ctx, "/").Return(rootContainer, nil)

	retrieved, err := service.GetRootContainer(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "/", retrieved.ID())
	assert.Equal(t, "Root Container", retrieved.GetTitle())
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_GetRootContainer_Error(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock error when retrieving
	mockRepo.On("GetContainer", ctx, "/").Return(nil, assert.AnError)

	_, err := service.GetRootContainer(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve root container")
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_ValidateSystemState_Valid(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Create a proper root container
	rootContainer := domain.NewContainer(ctx, "/", "", domain.BasicContainer)

	// Mock successful existence check
	mockRepo.On("ContainerExists", ctx, "/").Return(true, nil)

	// Mock successful retrieval
	mockRepo.On("GetContainer", ctx, "/").Return(rootContainer, nil)

	err := service.ValidateSystemState(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_ValidateSystemState_RootDoesNotExist(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock that root container doesn't exist
	mockRepo.On("ContainerExists", ctx, "/").Return(false, nil)

	err := service.ValidateSystemState(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root container does not exist")
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_ValidateSystemState_InvalidRoot(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Create an invalid root container (has parent)
	invalidRoot := domain.NewContainer(ctx, "/", "parent", domain.BasicContainer)

	// Mock successful existence check
	mockRepo.On("ContainerExists", ctx, "/").Return(true, nil)

	// Mock successful retrieval but with invalid root
	mockRepo.On("GetContainer", ctx, "/").Return(invalidRoot, nil)

	err := service.ValidateSystemState(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "root container should not have a parent")
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_Initialize(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Mock that root container doesn't exist and needs to be created
	mockRepo.On("ContainerExists", ctx, "/").Return(false, nil)
	mockRepo.On("CreateContainer", ctx, mock.AnythingOfType("*domain.Container")).Return(nil)

	err := service.Initialize(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_CreateRootContainer(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Test the private method indirectly by checking the container properties
	// created by EnsureRootContainer
	mockRepo.On("ContainerExists", ctx, "/").Return(false, nil)
	mockRepo.On("CreateContainer", ctx, mock.MatchedBy(func(container domain.ContainerResource) bool {
		// Verify the created container has the expected properties
		return container.ID() == "/" &&
			container.GetParentID() == "" &&
			container.GetContainerType() == domain.BasicContainer &&
			container.GetTitle() == "Root Container" &&
			container.GetDescription() == "Root container for all resources in this pod" &&
			container.GetMetadata()["isRoot"] == true
	})).Return(nil)

	err := service.EnsureRootContainer(ctx)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInitializationService_MigrateFromLegacyStorage(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Currently a no-op, but should not error
	err := service.MigrateFromLegacyStorage(ctx)
	assert.NoError(t, err)
}

func TestInitializationService_RepairInconsistencies(t *testing.T) {
	mockRepo := &MockContainerRepository{}
	logger := log.NewStdLogger(log.NewWriter())
	service := NewInitializationService(mockRepo, logger)

	ctx := context.Background()

	// Currently a no-op, but should not error
	err := service.RepairInconsistencies(ctx)
	assert.NoError(t, err)
}
