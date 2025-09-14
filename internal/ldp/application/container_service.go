package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// ContainerListing represents a paginated list of container members
type ContainerListing struct {
	ContainerID string                   `json:"containerId"`
	Members     []string                 `json:"members"`
	Pagination  domain.PaginationOptions `json:"pagination"`
	TotalCount  int                      `json:"totalCount,omitempty"`
}

// EnhancedContainerListing represents a paginated list with filtering and sorting
type EnhancedContainerListing struct {
	ContainerID   string                      `json:"containerId"`
	Members       []infrastructure.MemberInfo `json:"members"`
	Pagination    domain.PaginationOptions    `json:"pagination"`
	Filter        domain.FilterOptions        `json:"filter,omitempty"`
	Sort          domain.SortOptions          `json:"sort"`
	TotalCount    int                         `json:"totalCount"`
	FilteredCount int                         `json:"filteredCount"`
}

// BreadcrumbItem represents a single item in a breadcrumb navigation trail
type BreadcrumbItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// ContainerInfo represents basic information about a container
type ContainerInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	ParentID string `json:"parentId"`
}

// ContainerPathResolution represents the result of path-based container resolution
type ContainerPathResolution struct {
	Container   *ContainerInfo   `json:"container,omitempty"`
	Path        string           `json:"path"`
	Exists      bool             `json:"exists"`
	IsContainer bool             `json:"isContainer"`
	Breadcrumbs []BreadcrumbItem `json:"breadcrumbs,omitempty"`
}

// ContainerTypeInfo represents detailed type information about a container
type ContainerTypeInfo struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	MemberCount   int      `json:"memberCount"`
	ChildCount    int      `json:"childCount"`
	IsEmpty       bool     `json:"isEmpty"`
	AcceptedTypes []string `json:"acceptedTypes"`
	Capabilities  []string `json:"capabilities"`
}

// MemberInfo represents information about a container member
type MemberInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"` // "Container" or "Resource"
}

// ContainerStructureInfo represents hierarchical structure information
type ContainerStructureInfo struct {
	Container ContainerInfo            `json:"container"`
	Members   []MemberInfo             `json:"members"`
	Children  []ContainerStructureInfo `json:"children"`
	Depth     int                      `json:"depth"`
}

// ContainerService orchestrates container operations with business logic and event handling
type ContainerService struct {
	containerRepo      domain.ContainerRepository
	unitOfWorkFactory  func() pericarpdomain.UnitOfWork
	rdfConverter       *infrastructure.ContainerRDFConverter
	timestampManager   *domain.TimestampManager
	corruptionDetector *domain.MetadataCorruptionDetector
	mu                 sync.RWMutex // For concurrent access handling
}

// NewContainerService creates a new container service instance
func NewContainerService(
	containerRepo domain.ContainerRepository,
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	rdfConverter *infrastructure.ContainerRDFConverter,
) *ContainerService {
	timestampManager := domain.NewTimestampManager()
	corruptionDetector := domain.NewMetadataCorruptionDetector(timestampManager)

	return &ContainerService{
		containerRepo:      containerRepo,
		unitOfWorkFactory:  unitOfWorkFactory,
		rdfConverter:       rdfConverter,
		timestampManager:   timestampManager,
		corruptionDetector: corruptionDetector,
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

// GetContainerWithFormat retrieves a container and converts it to the specified RDF format
func (s *ContainerService) GetContainerWithFormat(ctx context.Context, id, format, baseURI string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if id == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainerWithFormat")
	}

	if format == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("format cannot be empty"),
			domain.ErrInvalidID.Code,
			"format cannot be empty",
		).WithOperation("GetContainerWithFormat")
	}

	if baseURI == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("base URI cannot be empty"),
			domain.ErrInvalidID.Code,
			"base URI cannot be empty",
		).WithOperation("GetContainerWithFormat")
	}

	// Get container from repository
	container, err := s.containerRepo.GetContainer(ctx, id)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GetContainerWithFormat").WithContext("containerID", id)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container",
		).WithOperation("GetContainerWithFormat").WithContext("containerID", id)
	}

	// Convert to requested format
	switch format {
	case "text/turtle":
		return s.rdfConverter.ConvertToTurtle(container, baseURI)
	case "application/ld+json":
		return s.rdfConverter.ConvertToJSONLD(container, baseURI)
	case "application/rdf+xml":
		return s.rdfConverter.ConvertToRDFXML(container, baseURI)
	default:
		return nil, domain.WrapStorageError(
			fmt.Errorf("unsupported format: %s", format),
			domain.ErrInvalidFormat.Code,
			"unsupported RDF format",
		).WithOperation("GetContainerWithFormat").WithContext("format", format)
	}
}

// ListContainerMembersWithFormat lists container members and returns them in the specified RDF format
func (s *ContainerService) ListContainerMembersWithFormat(ctx context.Context, containerID, format, baseURI string, pagination domain.PaginationOptions) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ListContainerMembersWithFormat")
	}

	if format == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("format cannot be empty"),
			domain.ErrInvalidID.Code,
			"format cannot be empty",
		).WithOperation("ListContainerMembersWithFormat")
	}

	if baseURI == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("base URI cannot be empty"),
			domain.ErrInvalidID.Code,
			"base URI cannot be empty",
		).WithOperation("ListContainerMembersWithFormat")
	}

	// Validate pagination options
	if !pagination.IsValid() {
		pagination = domain.GetDefaultPagination()
	}

	// Get container from repository to access its members
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("ListContainerMembersWithFormat").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for member listing",
		).WithOperation("ListContainerMembersWithFormat").WithContext("containerID", containerID)
	}

	// Convert to requested format (this will include membership triples)
	switch format {
	case "text/turtle":
		return s.rdfConverter.ConvertToTurtle(container, baseURI)
	case "application/ld+json":
		return s.rdfConverter.ConvertToJSONLD(container, baseURI)
	case "application/rdf+xml":
		return s.rdfConverter.ConvertToRDFXML(container, baseURI)
	default:
		return nil, domain.WrapStorageError(
			fmt.Errorf("unsupported format: %s", format),
			domain.ErrInvalidFormat.Code,
			"unsupported RDF format",
		).WithOperation("ListContainerMembersWithFormat").WithContext("format", format)
	}
}

// GenerateMembershipTriples generates LDP membership triples for a container
func (s *ContainerService) GenerateMembershipTriples(ctx context.Context, containerID, baseURI string) ([]infrastructure.ContainerTriple, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GenerateMembershipTriples")
	}

	if baseURI == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("base URI cannot be empty"),
			domain.ErrInvalidID.Code,
			"base URI cannot be empty",
		).WithOperation("GenerateMembershipTriples")
	}

	// Get container from repository
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GenerateMembershipTriples").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for membership triples",
		).WithOperation("GenerateMembershipTriples").WithContext("containerID", containerID)
	}

	// Generate membership triples using the RDF converter
	triples := s.rdfConverter.GenerateMembershipTriples(container, baseURI)
	return triples, nil
}

// ListContainerMembersEnhanced lists container members with filtering, sorting, and enhanced pagination
func (s *ContainerService) ListContainerMembersEnhanced(ctx context.Context, containerID string, options domain.ListingOptions) (*EnhancedContainerListing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ListContainerMembersEnhanced")
	}

	// Validate and set defaults for options
	if !options.IsValid() {
		options = domain.GetDefaultListingOptions()
	}

	// Check if container exists
	exists, err := s.containerRepo.ContainerExists(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("ListContainerMembersEnhanced").WithContext("containerID", containerID)
	}

	if !exists {
		return nil, domain.ErrResourceNotFound.WithOperation("ListContainerMembersEnhanced").WithContext("containerID", containerID)
	}

	// For now, use basic listing (enhanced filtering would require repository interface updates)
	// This is a placeholder for the enhanced implementation
	basicListing, err := s.ListContainerMembers(ctx, containerID, options.Pagination)
	if err != nil {
		return nil, err
	}

	// Convert to enhanced listing format
	// In a full implementation, this would use the membership indexer with filtering
	members := make([]infrastructure.MemberInfo, len(basicListing.Members))
	for i, memberID := range basicListing.Members {
		members[i] = infrastructure.MemberInfo{
			ID:          memberID,
			Type:        infrastructure.ResourceTypeResource, // Default type
			ContentType: "application/octet-stream",
			Size:        0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	enhancedListing := &EnhancedContainerListing{
		ContainerID:   containerID,
		Members:       members,
		Pagination:    options.Pagination,
		Filter:        options.Filter,
		Sort:          options.Sort,
		TotalCount:    len(basicListing.Members),
		FilteredCount: len(basicListing.Members),
	}

	return enhancedListing, nil
}

// GetContainerStats returns statistics about a container
func (s *ContainerService) GetContainerStats(ctx context.Context, containerID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainerStats")
	}

	// Get container to ensure it exists
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GetContainerStats").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for stats",
		).WithOperation("GetContainerStats").WithContext("containerID", containerID)
	}

	// Build basic stats
	stats := make(map[string]interface{})
	stats["container_id"] = containerID
	stats["container_type"] = container.ContainerType.String()
	stats["parent_id"] = container.ParentID
	stats["member_count"] = container.GetMemberCount()
	stats["is_empty"] = container.IsEmpty()
	stats["created_at"] = container.GetMetadata()["createdAt"]
	stats["updated_at"] = container.GetMetadata()["updatedAt"]

	// Add title and description if available
	if title := container.GetTitle(); title != "" {
		stats["title"] = title
	}
	if description := container.GetDescription(); description != "" {
		stats["description"] = description
	}

	return stats, nil
}

// StreamContainerMembers streams container members for large containers
func (s *ContainerService) StreamContainerMembers(ctx context.Context, containerID string, options domain.ListingOptions) (<-chan infrastructure.MemberInfo, <-chan error, error) {
	// Validate input
	if containerID == "" {
		return nil, nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("StreamContainerMembers")
	}

	// Check if container exists
	exists, err := s.containerRepo.ContainerExists(ctx, containerID)
	if err != nil {
		return nil, nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("StreamContainerMembers").WithContext("containerID", containerID)
	}

	if !exists {
		return nil, nil, domain.ErrResourceNotFound.WithOperation("StreamContainerMembers").WithContext("containerID", containerID)
	}

	// Create channels for streaming
	memberChan := make(chan infrastructure.MemberInfo, 100) // Buffered channel
	errorChan := make(chan error, 10)

	// Start streaming goroutine
	go func() {
		defer close(memberChan)
		defer close(errorChan)

		// Stream members in pages
		pageSize := 100
		offset := 0

		for {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
				// Get next page
				pagination := domain.PaginationOptions{Limit: pageSize, Offset: offset}
				listing, err := s.ListContainerMembers(ctx, containerID, pagination)
				if err != nil {
					errorChan <- err
					return
				}

				// Stream members from this page
				for _, memberID := range listing.Members {
					member := infrastructure.MemberInfo{
						ID:          memberID,
						Type:        infrastructure.ResourceTypeResource,
						ContentType: "application/octet-stream",
						Size:        0,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}

					select {
					case memberChan <- member:
					case <-ctx.Done():
						errorChan <- ctx.Err()
						return
					}
				}

				// Check if we've reached the end
				if len(listing.Members) < pageSize {
					return
				}

				offset += pageSize
			}
		}
	}()

	return memberChan, errorChan, nil
}

// SetContainerDublinCoreMetadata sets Dublin Core metadata on a container
func (s *ContainerService) SetContainerDublinCoreMetadata(ctx context.Context, containerID string, dc domain.DublinCoreMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Set Dublin Core metadata
	container.SetDublinCoreMetadata(dc)

	// Update timestamp
	s.timestampManager.UpdateTimestamp(container)

	// Update the container in repository
	if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to update container with Dublin Core metadata: %w", err)
	}

	// Register and commit events
	unitOfWork := s.unitOfWorkFactory()
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
		}
		return fmt.Errorf("failed to commit Dublin Core metadata events: %w", err)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	return nil
}

// GetContainerDublinCoreMetadata retrieves Dublin Core metadata from a container
func (s *ContainerService) GetContainerDublinCoreMetadata(ctx context.Context, containerID string) (domain.DublinCoreMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return domain.DublinCoreMetadata{}, fmt.Errorf("failed to get container: %w", err)
	}

	// Return Dublin Core metadata
	return container.GetDublinCoreMetadata(), nil
}

// DetectContainerCorruption analyzes a container for metadata corruption
func (s *ContainerService) DetectContainerCorruption(ctx context.Context, containerID string) (*domain.CorruptionReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container for corruption detection: %w", err)
	}

	// Detect corruption
	report := s.corruptionDetector.DetectCorruption(container)
	return report, nil
}

// RepairContainerCorruption attempts to repair detected corruption in a container
func (s *ContainerService) RepairContainerCorruption(ctx context.Context, containerID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to get container for corruption repair: %w", err)
	}

	// Detect corruption first
	report := s.corruptionDetector.DetectCorruption(container)
	if !report.IsCorrupted {
		return false, nil // Nothing to repair
	}

	// Attempt repair
	repaired, err := s.corruptionDetector.RepairCorruption(container, report)
	if err != nil {
		return false, fmt.Errorf("failed to repair container corruption: %w", err)
	}

	if repaired {
		// Update the container in repository
		if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
			return false, fmt.Errorf("failed to update repaired container: %w", err)
		}

		// Register and commit repair events
		unitOfWork := s.unitOfWorkFactory()
		events := container.UncommittedEvents()
		if len(events) > 0 {
			unitOfWork.RegisterEvents(events)
		}

		// Commit unit of work for event processing
		_, err := unitOfWork.Commit(ctx)
		if err != nil {
			if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
				return false, fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
			}
			return false, fmt.Errorf("failed to commit repair events: %w", err)
		}

		// Mark events as committed
		container.MarkEventsAsCommitted()
	}

	return repaired, nil
}

// ValidateContainerTimestamps validates that container timestamps are consistent
func (s *ContainerService) ValidateContainerTimestamps(ctx context.Context, containerID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container for timestamp validation: %w", err)
	}

	// Validate timestamps
	return s.timestampManager.ValidateTimestamps(container)
}

// RepairContainerTimestamps repairs missing or invalid timestamps in a container
func (s *ContainerService) RepairContainerTimestamps(ctx context.Context, containerID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return false, fmt.Errorf("failed to get container for timestamp repair: %w", err)
	}

	// Repair timestamps
	repaired := s.timestampManager.RepairTimestamps(container)

	if repaired {
		// Update the container in repository
		if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
			return false, fmt.Errorf("failed to update container with repaired timestamps: %w", err)
		}

		// Emit update event
		event := domain.NewContainerUpdatedEvent(container.ID(), map[string]interface{}{
			"timestampsRepaired": true,
			"updatedAt":          time.Now(),
		})
		container.AddEvent(event)

		// Register and commit timestamp repair events
		unitOfWork := s.unitOfWorkFactory()
		events := container.UncommittedEvents()
		if len(events) > 0 {
			unitOfWork.RegisterEvents(events)
		}

		// Commit unit of work for event processing
		_, err := unitOfWork.Commit(ctx)
		if err != nil {
			if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
				return false, fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
			}
			return false, fmt.Errorf("failed to commit timestamp repair events: %w", err)
		}

		// Mark events as committed
		container.MarkEventsAsCommitted()
	}

	return repaired, nil
}

// UpdateContainerWithTimestamp updates a container and automatically manages timestamps
func (s *ContainerService) UpdateContainerWithTimestamp(ctx context.Context, containerID, title, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Update with timestamp management
	if title != "" {
		container.SetTitleWithTimestamp(title, s.timestampManager)
	}
	if description != "" {
		container.SetDescriptionWithTimestamp(description, s.timestampManager)
	}

	// Update the container in repository
	if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	// Register and commit events
	unitOfWork := s.unitOfWorkFactory()
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
		}
		return fmt.Errorf("failed to commit update events: %w", err)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	return nil
}

// AddResourceWithTimestamp adds a resource to a container with timestamp management
func (s *ContainerService) AddResourceWithTimestamp(ctx context.Context, containerID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Add member with timestamp management
	if err := container.AddMemberWithTimestamp(resourceID, s.timestampManager); err != nil {
		return fmt.Errorf("failed to add member with timestamp: %w", err)
	}

	// Update the container in repository
	if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	// Also update membership index
	if err := s.containerRepo.AddMember(ctx, containerID, resourceID); err != nil {
		return fmt.Errorf("failed to update membership index: %w", err)
	}

	// Register and commit events
	unitOfWork := s.unitOfWorkFactory()
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
		}
		return fmt.Errorf("failed to commit add member events: %w", err)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	return nil
}

// RemoveResourceWithTimestamp removes a resource from a container with timestamp management
func (s *ContainerService) RemoveResourceWithTimestamp(ctx context.Context, containerID, resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Remove member with timestamp management
	if err := container.RemoveMemberWithTimestamp(resourceID, s.timestampManager); err != nil {
		return fmt.Errorf("failed to remove member with timestamp: %w", err)
	}

	// Update the container in repository
	if err := s.containerRepo.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}

	// Also update membership index
	if err := s.containerRepo.RemoveMember(ctx, containerID, resourceID); err != nil {
		return fmt.Errorf("failed to update membership index: %w", err)
	}

	// Register and commit events
	unitOfWork := s.unitOfWorkFactory()
	events := container.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work for event processing
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback unit of work: %v", rollbackErr)
		}
		return fmt.Errorf("failed to commit remove member events: %w", err)
	}

	// Mark events as committed
	container.MarkEventsAsCommitted()

	return nil
}

// GenerateBreadcrumbs generates breadcrumb navigation for container hierarchies
func (s *ContainerService) GenerateBreadcrumbs(ctx context.Context, containerID string) ([]BreadcrumbItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GenerateBreadcrumbs")
	}

	// Get the container to ensure it exists
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GenerateBreadcrumbs").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for breadcrumbs",
		).WithOperation("GenerateBreadcrumbs").WithContext("containerID", containerID)
	}

	// Get the path to the container
	path, err := s.containerRepo.GetPath(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container path for breadcrumbs",
		).WithOperation("GenerateBreadcrumbs").WithContext("containerID", containerID)
	}

	// Build breadcrumbs from path
	breadcrumbs := make([]BreadcrumbItem, 0, len(path))
	currentPath := ""

	for i, pathSegment := range path {
		if i == 0 {
			currentPath = "/" + pathSegment
		} else {
			currentPath = currentPath + "/" + pathSegment
		}

		// Get container info for this path segment
		var title string
		if pathSegment == containerID {
			// Use the container we already have
			title = container.GetTitle()
			if title == "" {
				title = pathSegment
			}
		} else {
			// Get the container for this path segment
			segmentContainer, err := s.containerRepo.GetContainer(ctx, pathSegment)
			if err != nil {
				// If we can't get the container, use the ID as title
				title = pathSegment
			} else {
				title = segmentContainer.GetTitle()
				if title == "" {
					title = pathSegment
				}
			}
		}

		breadcrumbs = append(breadcrumbs, BreadcrumbItem{
			ID:    pathSegment,
			Title: title,
			Path:  currentPath,
		})
	}

	return breadcrumbs, nil
}

// ResolveContainerPath resolves a path to container information with breadcrumbs
func (s *ContainerService) ResolveContainerPath(ctx context.Context, path string) (*ContainerPathResolution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if path == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("path cannot be empty"),
			domain.ErrInvalidID.Code,
			"path cannot be empty",
		).WithOperation("ResolveContainerPath")
	}

	// Try to find the container by path
	container, err := s.containerRepo.FindByPath(ctx, path)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			// Path doesn't exist, return resolution with exists=false
			return &ContainerPathResolution{
				Container:   nil,
				Path:        path,
				Exists:      false,
				IsContainer: false,
				Breadcrumbs: nil,
			}, nil
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to resolve container path",
		).WithOperation("ResolveContainerPath").WithContext("path", path)
	}

	// Generate breadcrumbs for the resolved container
	breadcrumbs, err := s.GenerateBreadcrumbs(ctx, container.ID())
	if err != nil {
		// Don't fail the resolution if breadcrumbs fail, just log and continue
		breadcrumbs = nil
	}

	// Build container info
	containerInfo := &ContainerInfo{
		ID:       container.ID(),
		Title:    container.GetTitle(),
		Type:     container.ContainerType.String(),
		ParentID: container.ParentID,
	}

	return &ContainerPathResolution{
		Container:   containerInfo,
		Path:        path,
		Exists:      true,
		IsContainer: true,
		Breadcrumbs: breadcrumbs,
	}, nil
}

// GetContainerTypeInfo returns detailed type information about a container
func (s *ContainerService) GetContainerTypeInfo(ctx context.Context, containerID string) (*ContainerTypeInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainerTypeInfo")
	}

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GetContainerTypeInfo").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for type info",
		).WithOperation("GetContainerTypeInfo").WithContext("containerID", containerID)
	}

	// Get member count
	members, err := s.containerRepo.ListMembers(ctx, containerID, domain.PaginationOptions{Limit: 1000, Offset: 0})
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to list members for type info",
		).WithOperation("GetContainerTypeInfo").WithContext("containerID", containerID)
	}

	// Get child count
	children, err := s.containerRepo.GetChildren(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get children for type info",
		).WithOperation("GetContainerTypeInfo").WithContext("containerID", containerID)
	}

	// Build type info
	typeInfo := &ContainerTypeInfo{
		ID:            containerID,
		Type:          container.ContainerType.String(),
		Title:         container.GetTitle(),
		Description:   container.GetDescription(),
		MemberCount:   len(members),
		ChildCount:    len(children),
		IsEmpty:       container.IsEmpty(),
		AcceptedTypes: []string{"*/*"}, // BasicContainer accepts all types
		Capabilities:  []string{"create", "read", "update", "delete", "list"},
	}

	return typeInfo, nil
}

// GenerateStructureInfo generates machine-readable hierarchical structure information
func (s *ContainerService) GenerateStructureInfo(ctx context.Context, containerID string, maxDepth int) (*ContainerStructureInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.generateStructureInfoRecursive(ctx, containerID, 0, maxDepth)
}

// generateStructureInfoRecursive is a helper method for recursive structure generation
func (s *ContainerService) generateStructureInfoRecursive(ctx context.Context, containerID string, currentDepth, maxDepth int) (*ContainerStructureInfo, error) {
	// Validate input
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GenerateStructureInfo")
	}

	// Get the container
	container, err := s.containerRepo.GetContainer(ctx, containerID)
	if err != nil {
		if domain.IsResourceNotFound(err) {
			return nil, domain.ErrResourceNotFound.WithOperation("GenerateStructureInfo").WithContext("containerID", containerID)
		}
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container for structure info",
		).WithOperation("GenerateStructureInfo").WithContext("containerID", containerID)
	}

	// Build container info
	containerInfo := ContainerInfo{
		ID:       container.ID(),
		Title:    container.GetTitle(),
		Type:     container.ContainerType.String(),
		ParentID: container.ParentID,
	}

	// Get members
	memberIDs, err := s.containerRepo.ListMembers(ctx, containerID, domain.PaginationOptions{Limit: 1000, Offset: 0})
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to list members for structure info",
		).WithOperation("GenerateStructureInfo").WithContext("containerID", containerID)
	}

	// Convert member IDs to MemberInfo
	members := make([]MemberInfo, len(memberIDs))
	for i, memberID := range memberIDs {
		// For now, assume all members are resources
		// In a full implementation, we would check if the member is a container
		members[i] = MemberInfo{
			ID:   memberID,
			Type: "Resource",
		}
	}

	// Get children (only if we haven't reached max depth)
	var childrenInfo []ContainerStructureInfo
	if currentDepth < maxDepth {
		children, err := s.containerRepo.GetChildren(ctx, containerID)
		if err != nil {
			return nil, domain.WrapStorageError(
				err,
				domain.ErrStorageOperation.Code,
				"failed to get children for structure info",
			).WithOperation("GenerateStructureInfo").WithContext("containerID", containerID)
		}

		// Recursively generate structure info for children
		childrenInfo = make([]ContainerStructureInfo, len(children))
		for i, child := range children {
			childInfo, err := s.generateStructureInfoRecursive(ctx, child.ID(), currentDepth+1, maxDepth)
			if err != nil {
				return nil, err
			}
			childrenInfo[i] = *childInfo
		}
	} else {
		childrenInfo = []ContainerStructureInfo{}
	}

	return &ContainerStructureInfo{
		Container: containerInfo,
		Members:   members,
		Children:  childrenInfo,
		Depth:     currentDepth,
	}, nil
}
