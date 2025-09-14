package handlers

import (
	"context"
	"io"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// StorageServiceInterface defines the interface for storage operations
type StorageServiceInterface interface {
	StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error)
	RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error)
	DeleteResource(ctx context.Context, id string) error
	ResourceExists(ctx context.Context, id string) (bool, error)
	StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error)
	StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) (*domain.Resource, error)
}

// ContainerServiceInterface defines the interface for container operations
type ContainerServiceInterface interface {
	CreateContainer(ctx context.Context, id, parentID string, containerType domain.ContainerType) (*domain.Container, error)
	GetContainer(ctx context.Context, id string) (*domain.Container, error)
	UpdateContainer(ctx context.Context, container *domain.Container) error
	DeleteContainer(ctx context.Context, id string) error
	AddResource(ctx context.Context, containerID, resourceID string) error
	RemoveResource(ctx context.Context, containerID, resourceID string) error
	ListContainerMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) (*application.ContainerListing, error)
	GetContainerPath(ctx context.Context, containerID string) ([]string, error)
	FindContainerByPath(ctx context.Context, path string) (*domain.Container, error)
	GetChildren(ctx context.Context, containerID string) ([]*domain.Container, error)
	GetParent(ctx context.Context, containerID string) (*domain.Container, error)
	ContainerExists(ctx context.Context, id string) (bool, error)
}
