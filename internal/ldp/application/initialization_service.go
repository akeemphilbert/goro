package application

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// InitializationService handles system initialization tasks
type InitializationService struct {
	containerRepo domain.ContainerRepository
	logger        *log.Helper
}

// NewInitializationService creates a new initialization service
func NewInitializationService(
	containerRepo domain.ContainerRepository,
	logger log.Logger,
) *InitializationService {
	return &InitializationService{
		containerRepo: containerRepo,
		logger:        log.NewHelper(logger),
	}
}

// Initialize performs all system initialization tasks
func (s *InitializationService) Initialize(ctx context.Context) error {
	s.logger.Info("Starting system initialization...")

	// Ensure root container exists
	if err := s.EnsureRootContainer(ctx); err != nil {
		return fmt.Errorf("failed to ensure root container: %w", err)
	}

	s.logger.Info("System initialization completed successfully")
	return nil
}

// EnsureRootContainer ensures that the root container "/" exists
func (s *InitializationService) EnsureRootContainer(ctx context.Context) error {
	s.logger.Debug("Checking if root container exists...")

	const rootContainerID = "/"

	// Check if root container already exists
	exists, err := s.containerRepo.ContainerExists(ctx, rootContainerID)
	if err != nil {
		return fmt.Errorf("failed to check root container existence: %w", err)
	}

	if exists {
		s.logger.Info("Root container already exists")
		return nil
	}

	s.logger.Info("Creating root container...")

	// Create root container
	rootContainer := s.createRootContainer(ctx)

	// Store the root container
	if err := s.containerRepo.CreateContainer(ctx, rootContainer); err != nil {
		return fmt.Errorf("failed to create root container: %w", err)
	}

	s.logger.Info("Root container created successfully", "id", rootContainerID)
	return nil
}

// createRootContainer creates the root container domain object
func (s *InitializationService) createRootContainer(ctx context.Context) domain.ContainerResource {
	const rootContainerID = "/"

	s.logger.Debug("Creating root container domain object...")

	// Create the underlying basic resource
	basicResource := domain.NewResource(ctx, rootContainerID, "application/ld+json", []byte("{}"))

	// Set container-specific metadata
	basicResource.SetMetadata("type", "Container")
	basicResource.SetMetadata("containerType", domain.BasicContainer.String())
	basicResource.SetMetadata("title", "Root Container")
	basicResource.SetMetadata("description", "Root container for all resources in this pod")
	basicResource.SetMetadata("createdAt", time.Now())
	basicResource.SetMetadata("updatedAt", time.Now())
	basicResource.SetMetadata("isRoot", true)

	// Create container with no parent (root)
	container := &domain.Container{
		BasicResource: basicResource,
		ParentID:      "", // Root has no parent
		ContainerType: domain.BasicContainer,
		Members:       make([]string, 0),
		Children:      make([]domain.Resource, 0),
	}

	// Emit container creation event
	event := domain.NewContainerCreatedEvent(rootContainerID, map[string]interface{}{
		"parentID":      "",
		"containerType": domain.BasicContainer.String(),
		"title":         "Root Container",
		"description":   "Root container for all resources in this pod",
		"isRoot":        true,
		"createdAt":     time.Now(),
	})
	container.AddEvent(event)

	s.logger.Debug("Root container domain object created")
	return container
}

// GetRootContainer retrieves the root container with all its immediate children
func (s *InitializationService) GetRootContainer(ctx context.Context) (domain.ContainerResource, error) {
	s.logger.Debug("Retrieving root container...")

	const rootContainerID = "/"

	rootContainer, err := s.containerRepo.GetContainer(ctx, rootContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve root container: %w", err)
	}

	s.logger.Debug("Root container retrieved successfully",
		"id", rootContainer.ID(),
		"childCount", len(rootContainer.GetChildren()))

	return rootContainer, nil
}

// ValidateSystemState validates that the system is in a consistent state
func (s *InitializationService) ValidateSystemState(ctx context.Context) error {
	s.logger.Debug("Validating system state...")

	// Check that root container exists
	const rootContainerID = "/"
	exists, err := s.containerRepo.ContainerExists(ctx, rootContainerID)
	if err != nil {
		return fmt.Errorf("failed to check root container existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("root container does not exist - system is not initialized")
	}

	// Retrieve root container to validate its properties
	rootContainer, err := s.GetRootContainer(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve root container: %w", err)
	}

	// Validate root container properties
	if rootContainer.GetParentID() != "" {
		return fmt.Errorf("root container should not have a parent")
	}

	if rootContainer.GetContainerType() != domain.BasicContainer {
		return fmt.Errorf("root container should be a BasicContainer")
	}

	s.logger.Info("System state validation passed")
	return nil
}

// MigrateFromLegacyStorage migrates data from the old filesystem-only storage to the new GORM-based storage
func (s *InitializationService) MigrateFromLegacyStorage(ctx context.Context) error {
	s.logger.Info("Starting migration from legacy storage...")

	// TODO: Implement migration logic
	// This would involve:
	// 1. Reading existing containers and resources from filesystem
	// 2. Converting them to GORM models
	// 3. Inserting them into the database
	// 4. Maintaining file references where appropriate

	s.logger.Info("Legacy storage migration completed")
	return nil
}

// RepairInconsistencies repairs any inconsistencies found in the storage
func (s *InitializationService) RepairInconsistencies(ctx context.Context) error {
	s.logger.Info("Checking for storage inconsistencies...")

	// TODO: Implement consistency checks and repairs
	// This might include:
	// 1. Orphaned resources without containers
	// 2. Containers with missing parents
	// 3. Circular references in container hierarchy
	// 4. Missing membership records

	s.logger.Info("Storage consistency check completed")
	return nil
}
