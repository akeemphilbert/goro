package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// ContainerListing represents a paginated list of container members
type ContainerListing struct {
	ContainerID string                   `json:"containerId"`
	Members     []string                 `json:"members"`
	Pagination  domain.PaginationOptions `json:"pagination"`
	TotalCount  int                      `json:"totalCount,omitempty"`
}

// ContainerService orchestrates container operations with business logic and event handling
type ContainerService struct {
	containerRepo     domain.ContainerRepository
	unitOfWorkFactory func() pericarpdomain.UnitOfWork
	mu                sync.RWMutex // For concurrent access handling
}

// NewContainerService creates a new container service instance
func NewContainerService(
	containerRepo domain.ContainerRepository,
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
) *ContainerService {
	return &ContainerService{
		containerRepo:     containerRepo,
		unitOfWorkFactory: unitOfWorkFactory,
	}
}

// CreateContainer creates a new container with validation and event handling
func (s *ContainerService) CreateContainer(ctx context.Context, id, parentID string, containerType domain.ContainerType) (*domain.Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if id == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("CreateContainer")
	}

	if !containerType.IsValid() {
		return nil, domain.WrapStorageError(
			fmt.Errorf("invalid container type: %s", containerType),
			domain.ErrInvalidResource.Code,
			"invalid container type",
		).WithOperation("CreateContainer").WithContext("containerType", containerType.String())
	}

	// Check if container already exists
	exists, err := s.containerRepo.ContainerExists(ctx, id)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("CreateContainer").WithContext("containerID", id)
	}

	if exists {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container already exists"),
			domain.ErrResourceAlreadyExists.Code,
			"container already exists",
		).WithOperation("CreateContainer").WithContext("containerID", id)
	}

	// Validate parent container exists if parentID is provided
	if parentID != "" {
		parentExists, err := s.containerRepo.ContainerExists(ctx, parentID)
		if err != nil {
			return nil, domain.WrapStorageError(
				err,
				domain.ErrStorageOperation.Code,
				"failed to check parent container existence",
			).WithOperation("CreateContainer").WithContext("parentID", parentID)
		}

		if !parentExists {
			return nil, domain.WrapStorageError(
				fmt.Errorf("parent container not found"),
				domain.ErrResourceNotFound.Code,
				"parent container not found",
			).WithOperation("CreateContainer").WithContext("parentID", parentID)
		}
	}

	// Create container entity
	container := domain.NewContainer(id, parentID, containerType)

	// Validate hierarchy to prevent circular references
	if parentID != "" {
		path, err := s.containerRepo.GetPath(ctx, parentID)
		if err != nil {
			return nil, domain.WrapStorageError(
				err,
				domain.ErrStorageOperation.Code,
				"failed to get parent path for hierarchy validation",
			).WithOperation("CreateContainer").WithContext("parentID", parentID)
		}

		if err := container.ValidateHierarchy(path); err != nil {
			return nil, domain.WrapStorageError(
				err,
				domain.ErrInvalidHierarchy.Code,
				"hierarchy validation failed",
			).WithOperation("CreateContainer").WithContext("containerID", id)
		}
	}

	// Create unit of work for event handling
	unitOfWork := s.unitOfWorkFactory()

	// Register container events
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing - this will trigger event handlers to update repository
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback unit of work on commit failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to commit container creation events",
		).WithOperation("CreateContainer").WithContext("containerID", id)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	// Log successful event processing
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for container creation %s\n", len(envelopes), id)
	}

	return container, nil
}

// GetContainer retrieves a container by ID
func (s *ContainerService) GetContainer(ctx context.Context, id string) (*domain.Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if id == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainer")
	}

	// Retrieve container from repository
	container, err := s.containerRepo.GetContainer(ctx, id)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GetContainer").WithContext("containerID", id)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	return container, nil
}

// UpdateContainer updates an existing container
func (s *ContainerService) UpdateContainer(ctx context.Context, container *domain.Container) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if container == nil {
		return domain.WrapStorageError(
			fmt.Errorf("container cannot be nil"),
			domain.ErrInvalidResource.Code,
			"container cannot be nil",
		).WithOperation("UpdateContainer")
	}

	if container.ID() == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("UpdateContainer")
	}

	// Create unit of work for event handling
	unitOfWork := s.unitOfWorkFactory()

	// Register container events
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing - this will trigger event handlers to update repository
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback unit of work on commit failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to commit container update events",
		).WithOperation("UpdateContainer").WithContext("containerID", container.ID())
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	// Log successful event processing
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for container update %s\n", len(envelopes), container.ID())
	}

	return nil
}

// DeleteContainer deletes a container after validating it's empty
func (s *ContainerService) DeleteContainer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if id == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("DeleteContainer")
	}

	// Retrieve container to validate deletion
	container, err := s.containerRepo.GetContainer(ctx, id)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return domain.ErrResourceNotFound.WithOperation("DeleteContainer").WithContext("containerID", id)
		}
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for deletion",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Validate container can be deleted (must be empty)
	if err := container.Delete(); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrContainerNotEmpty.Code,
			"container cannot be deleted",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Create unit of work for event handling
	unitOfWork := s.unitOfWorkFactory()

	// Register delete events
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing - this will trigger event handlers to delete from repository
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback unit of work on commit failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to commit container deletion events",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	// Log successful event processing
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for container deletion %s\n", len(envelopes), id)
	}

	return nil
}

// AddResource adds a resource to a container
func (s *ContainerService) AddResource(ctx context.Context, containerID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if containerID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("AddResource")
	}

	if resourceID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("AddResource")
	}

	// Create unit of work for event handling
	unitOfWork := s.unitOfWorkFactory()

	// Create member added event
	event := domain.NewMemberAddedEvent(containerID, map[string]interface{}{
		"memberID":   resourceID,
		"memberType": "Resource",
		"addedAt":    time.Now(),
	})
	unitOfWork.RegisterEvents([]pericarpdomain.Event{event})

	// Commit unit of work for event processing - this will trigger event handlers to update repository
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback unit of work on commit failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to commit add resource events",
		).WithOperation("AddResource").WithContext("containerID", containerID).WithContext("resourceID", resourceID)
	}

	// Log successful event processing
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for adding resource %s to container %s\n", len(envelopes), resourceID, containerID)
	}

	return nil
}

// RemoveResource removes a resource from a container
func (s *ContainerService) RemoveResource(ctx context.Context, containerID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input
	if containerID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("RemoveResource")
	}

	if resourceID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("RemoveResource")
	}

	// Create unit of work for event handling
	unitOfWork := s.unitOfWorkFactory()

	// Create member removed event
	event := domain.NewMemberRemovedEvent(containerID, map[string]interface{}{
		"memberID":   resourceID,
		"memberType": "Resource",
		"removedAt":  time.Now(),
	})
	unitOfWork.RegisterEvents([]pericarpdomain.Event{event})

	// Commit unit of work for event processing - this will trigger event handlers to update repository
	envelopes, err := unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback unit of work on commit failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback unit of work: %v\n", rollbackErr)
		}
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to commit remove resource events",
		).WithOperation("RemoveResource").WithContext("containerID", containerID).WithContext("resourceID", resourceID)
	}

	// Log successful event processing
	if len(envelopes) > 0 {
		fmt.Printf("Successfully processed %d events for removing resource %s from container %s\n", len(envelopes), resourceID, containerID)
	}

	return nil
}

// ListContainerMembers lists all members of a container with pagination
func (s *ContainerService) ListContainerMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) (*ContainerListing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ListContainerMembers")
	}

	// Validate pagination options
	if !pagination.IsValid() {
		pagination = domain.GetDefaultPagination()
	}

	// List members from repository
	members, err := s.containerRepo.ListMembers(ctx, containerID, pagination)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to list container members",
		).WithOperation("ListContainerMembers").WithContext("containerID", containerID)
	}

	// Create container listing
	listing := &ContainerListing{
		ContainerID: containerID,
		Members:     members,
		Pagination:  pagination,
	}

	return listing, nil
}

// GetContainerPath returns the hierarchical path to a container
func (s *ContainerService) GetContainerPath(ctx context.Context, containerID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainerPath")
	}

	// Get path from repository
	path, err := s.containerRepo.GetPath(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container path",
		).WithOperation("GetContainerPath").WithContext("containerID", containerID)
	}

	return path, nil
}

// FindContainerByPath finds a container by its hierarchical path
func (s *ContainerService) FindContainerByPath(ctx context.Context, path string) (*domain.Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if path == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("path cannot be empty"),
			domain.ErrInvalidID.Code,
			"path cannot be empty",
		).WithOperation("FindContainerByPath")
	}

	// Find container by path from repository
	container, err := s.containerRepo.FindByPath(ctx, path)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("FindContainerByPath").WithContext("path", path)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to find container by path",
		).WithOperation("FindContainerByPath").WithContext("path", path)
	}

	return container, nil
}

// GetChildren returns all child containers of a container
func (s *ContainerService) GetChildren(ctx context.Context, containerID string) ([]*domain.Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetChildren")
	}

	// Get children from repository
	children, err := s.containerRepo.GetChildren(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container children",
		).WithOperation("GetChildren").WithContext("containerID", containerID)
	}

	return children, nil
}

// GetParent returns the parent container of a container
func (s *ContainerService) GetParent(ctx context.Context, containerID string) (*domain.Container, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetParent")
	}

	// Get parent from repository
	parent, err := s.containerRepo.GetParent(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container parent",
		).WithOperation("GetParent").WithContext("containerID", containerID)
	}

	return parent, nil
}

// ContainerExists checks if a container exists
func (s *ContainerService) ContainerExists(ctx context.Context, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id == "" {
		return false, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ContainerExists")
	}

	exists, err := s.containerRepo.ContainerExists(ctx, id)
	if err != nil {
		return false, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("ContainerExists").WithContext("containerID", id)
	}

	return exists, nil
}
