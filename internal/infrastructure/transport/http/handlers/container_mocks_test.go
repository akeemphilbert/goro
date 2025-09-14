package handlers

import (
	"context"
	"io"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/mock"
)

// MockContainerService is a mock implementation of ContainerServiceInterface
type MockContainerService struct {
	mock.Mock
}

func (m *MockContainerService) CreateContainer(ctx context.Context, id, parentID string, containerType domain.ContainerType) (*domain.Container, error) {
	args := m.Called(ctx, id, parentID, containerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Container), args.Error(1)
}

func (m *MockContainerService) GetContainer(ctx context.Context, id string) (*domain.Container, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Container), args.Error(1)
}

func (m *MockContainerService) UpdateContainer(ctx context.Context, container *domain.Container) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockContainerService) DeleteContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerService) AddResource(ctx context.Context, containerID, resourceID string) error {
	args := m.Called(ctx, containerID, resourceID)
	return args.Error(0)
}

func (m *MockContainerService) RemoveResource(ctx context.Context, containerID, resourceID string) error {
	args := m.Called(ctx, containerID, resourceID)
	return args.Error(0)
}

func (m *MockContainerService) ListContainerMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) (*application.ContainerListing, error) {
	args := m.Called(ctx, containerID, pagination)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.ContainerListing), args.Error(1)
}

func (m *MockContainerService) GetContainerPath(ctx context.Context, containerID string) ([]string, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContainerService) FindContainerByPath(ctx context.Context, path string) (*domain.Container, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Container), args.Error(1)
}

func (m *MockContainerService) GetChildren(ctx context.Context, containerID string) ([]*domain.Container, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Container), args.Error(1)
}

func (m *MockContainerService) GetParent(ctx context.Context, containerID string) (*domain.Container, error) {
	args := m.Called(ctx, containerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Container), args.Error(1)
}

func (m *MockContainerService) ContainerExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// MockContainerStorageService is a mock implementation of StorageServiceInterface for container tests
type MockContainerStorageService struct {
	mock.Mock
}

func (m *MockContainerStorageService) StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error) {
	args := m.Called(ctx, id, data, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockContainerStorageService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error) {
	args := m.Called(ctx, id, acceptFormat)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockContainerStorageService) DeleteResource(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContainerStorageService) ResourceExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockContainerStorageService) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	args := m.Called(ctx, id, acceptFormat)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.String(1), args.Error(2)
}

func (m *MockContainerStorageService) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) (*domain.Resource, error) {
	args := m.Called(ctx, id, reader, contentType, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}
