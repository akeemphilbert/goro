package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// FileSystemContainerRepository implements ContainerRepository using filesystem storage
type FileSystemContainerRepository struct {
	*FileSystemRepository // Inherits base filesystem operations
	indexer               MembershipIndexer
}

// NewFileSystemContainerRepository creates a new FileSystemContainerRepository
func NewFileSystemContainerRepository(basePath string, indexer MembershipIndexer) (*FileSystemContainerRepository, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path cannot be empty")
	}

	if indexer == nil {
		return nil, fmt.Errorf("membership indexer cannot be nil")
	}

	// Create base filesystem repository
	baseRepo, err := NewFileSystemRepository(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create base repository: %w", err)
	}

	// Ensure containers directory exists
	containersPath := filepath.Join(basePath, "containers")
	if err := os.MkdirAll(containersPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create containers directory: %w", err)
	}

	return &FileSystemContainerRepository{
		FileSystemRepository: baseRepo,
		indexer:              indexer,
	}, nil
}

// CreateContainer creates a new container in the filesystem
func (r *FileSystemContainerRepository) CreateContainer(ctx context.Context, container domain.ContainerResource) error {
	if container == nil {
		return domain.WrapStorageError(
			fmt.Errorf("container cannot be nil"),
			domain.ErrInvalidResource.Code,
			"container cannot be nil",
		).WithOperation("CreateContainer")
	}

	if container.ID() == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("CreateContainer")
	}

	// Check if container already exists
	exists, err := r.ContainerExists(ctx, container.ID())
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	if exists {
		return domain.WrapStorageError(
			fmt.Errorf("container already exists"),
			domain.ErrResourceAlreadyExists.Code,
			"container already exists",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	// Create container directory
	containerDir := r.getContainerPath(container.ID())
	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to create container directory",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	// Store container metadata
	if err := r.storeContainerMetadata(container); err != nil {
		// Clean up on error
		os.RemoveAll(containerDir)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to store container metadata",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	// Store the underlying resource
	var resource domain.Resource = container
	if err := r.FileSystemRepository.Store(ctx, resource); err != nil {
		// Clean up on error
		os.RemoveAll(containerDir)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to store container resource",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	// Insert container into database for foreign key constraints
	if err := r.insertContainerIntoDatabase(ctx, container); err != nil {
		// Clean up on error
		os.RemoveAll(containerDir)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to insert container into database",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	// Index container in membership indexer
	if err := r.indexContainer(ctx, container); err != nil {
		// Clean up on error
		os.RemoveAll(containerDir)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to index container",
		).WithOperation("CreateContainer").WithContext("containerID", container.ID())
	}

	return nil
}

// GetContainer retrieves a container from the filesystem
func (r *FileSystemContainerRepository) GetContainer(ctx context.Context, id string) (domain.ContainerResource, error) {
	if id == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("GetContainer")
	}

	// Check if container exists
	exists, err := r.ContainerExists(ctx, id)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	if !exists {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container not found"),
			domain.ErrResourceNotFound.Code,
			"container not found",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	// Load container metadata
	metadata, err := r.loadContainerMetadata(id)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to load container metadata",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	// Retrieve the underlying resource
	resource, err := r.FileSystemRepository.Retrieve(ctx, id)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve container resource",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	// Create container from metadata and resource
	// Type assert to get the underlying BasicResource
	basicResource, ok := resource.(*domain.BasicResource)
	if !ok {
		return nil, domain.WrapStorageError(
			fmt.Errorf("resource is not a BasicResource"),
			domain.ErrInvalidResource.Code,
			"invalid resource type",
		).WithOperation("GetContainer").WithContext("containerID", id)
	}

	container := &domain.Container{
		BasicResource: basicResource,
		Members:       metadata.Members,
		ParentID:      metadata.ParentID,
		ContainerType: domain.ContainerType(metadata.ContainerType),
	}

	// Restore container-specific metadata
	container.SetMetadata("type", "Container")
	container.SetMetadata("containerType", metadata.ContainerType)
	container.SetMetadata("parentID", metadata.ParentID)
	container.SetMetadata("title", metadata.Title)
	container.SetMetadata("description", metadata.Description)
	container.SetMetadata("createdAt", metadata.CreatedAt)
	container.SetMetadata("updatedAt", metadata.UpdatedAt)

	return container, nil
}

// UpdateContainer updates an existing container
func (r *FileSystemContainerRepository) UpdateContainer(ctx context.Context, container domain.ContainerResource) error {
	if container == nil {
		return domain.WrapStorageError(
			fmt.Errorf("container cannot be nil"),
			domain.ErrInvalidResource.Code,
			"container cannot be nil",
		).WithOperation("UpdateContainer")
	}

	// Check if container exists
	exists, err := r.ContainerExists(ctx, container.ID())
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("UpdateContainer").WithContext("containerID", container.ID())
	}

	if !exists {
		return domain.WrapStorageError(
			fmt.Errorf("container not found"),
			domain.ErrResourceNotFound.Code,
			"container not found",
		).WithOperation("UpdateContainer").WithContext("containerID", container.ID())
	}

	// Update container metadata
	if err := r.storeContainerMetadata(container); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to update container metadata",
		).WithOperation("UpdateContainer").WithContext("containerID", container.ID())
	}

	// Update the underlying resource
	var resource domain.Resource = container
	if err := r.FileSystemRepository.Store(ctx, resource); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to update container resource",
		).WithOperation("UpdateContainer").WithContext("containerID", container.ID())
	}

	return nil
}

// DeleteContainer deletes a container from the filesystem
func (r *FileSystemContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	if id == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("DeleteContainer")
	}

	// Check if container exists
	exists, err := r.ContainerExists(ctx, id)
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container existence",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	if !exists {
		return domain.WrapStorageError(
			fmt.Errorf("container not found"),
			domain.ErrResourceNotFound.Code,
			"container not found",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Check if container is empty
	members, err := r.ListMembers(ctx, id, domain.PaginationOptions{Limit: 1})
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container members",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	if len(members) > 0 {
		return domain.WrapStorageError(
			fmt.Errorf("container is not empty"),
			domain.ErrContainerNotEmpty.Code,
			"container contains resources and cannot be deleted",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Remove container directory
	containerDir := r.getContainerPath(id)
	if err := os.RemoveAll(containerDir); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to delete container directory",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Delete the underlying resource
	if err := r.FileSystemRepository.Delete(ctx, id); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to delete container resource",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	// Remove container from database
	if err := r.removeContainerFromDatabase(ctx, id); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to remove container from database",
		).WithOperation("DeleteContainer").WithContext("containerID", id)
	}

	return nil
}

// ContainerExists checks if a container exists
func (r *FileSystemContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ContainerExists")
	}

	containerDir := r.getContainerPath(id)
	metadataPath := filepath.Join(containerDir, "container.json")

	// Check if container metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to check container metadata file",
		).WithOperation("ContainerExists").WithContext("containerID", id)
	}

	return true, nil
}

// AddMember adds a member to a container
func (r *FileSystemContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	if containerID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("AddMember")
	}

	if memberID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("member ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"member ID cannot be empty",
		).WithOperation("AddMember")
	}

	// Get container
	container, err := r.GetContainer(ctx, containerID)
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container",
		).WithOperation("AddMember").WithContext("containerID", containerID)
	}

	// Get the resource to add as member
	resource, err := r.FileSystemRepository.Retrieve(ctx, memberID)
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to retrieve member resource",
		).WithOperation("AddMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	// Add member to container
	if err := container.AddMember(ctx, resource); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to add member to container",
		).WithOperation("AddMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	// Update container
	if err := r.UpdateContainer(ctx, container); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to update container after adding member",
		).WithOperation("AddMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	// Index membership
	if err := r.indexer.IndexMembership(ctx, containerID, memberID); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to index membership",
		).WithOperation("AddMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	return nil
}

// RemoveMember removes a member from a container
func (r *FileSystemContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	if containerID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("RemoveMember")
	}

	if memberID == "" {
		return domain.WrapStorageError(
			fmt.Errorf("member ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"member ID cannot be empty",
		).WithOperation("RemoveMember")
	}

	// Get container
	container, err := r.GetContainer(ctx, containerID)
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container",
		).WithOperation("RemoveMember").WithContext("containerID", containerID)
	}

	// Remove member from container
	if err := container.RemoveMember(ctx, memberID); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to remove member from container",
		).WithOperation("RemoveMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	// Update container
	if err := r.UpdateContainer(ctx, container); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to update container after removing member",
		).WithOperation("RemoveMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	// Remove membership from index
	if err := r.indexer.RemoveMembership(ctx, containerID, memberID); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to remove membership from index",
		).WithOperation("RemoveMember").WithContext("containerID", containerID).WithContext("memberID", memberID)
	}

	return nil
}

// ListMembers lists all members of a container
func (r *FileSystemContainerRepository) ListMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) ([]string, error) {
	if containerID == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("container ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"container ID cannot be empty",
		).WithOperation("ListMembers")
	}

	// Get container to ensure it exists
	container, err := r.GetContainer(ctx, containerID)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to get container",
		).WithOperation("ListMembers").WithContext("containerID", containerID)
	}

	// Apply pagination to members list
	members := container.GetMembers()
	if pagination.Offset >= len(members) {
		return []string{}, nil
	}

	end := pagination.Offset + pagination.Limit
	if end > len(members) {
		end = len(members)
	}

	return members[pagination.Offset:end], nil
}

// GetChildren returns all child containers of a container
func (r *FileSystemContainerRepository) GetChildren(ctx context.Context, containerID string) ([]domain.ContainerResource, error) {
	// This would require scanning all containers to find children
	// For now, return empty slice - this can be optimized with indexing
	return []domain.ContainerResource{}, nil
}

// GetParent returns the parent container of a container
func (r *FileSystemContainerRepository) GetParent(ctx context.Context, containerID string) (domain.ContainerResource, error) {
	container, err := r.GetContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	if container.GetParentID() == "" {
		return nil, nil // No parent (root container)
	}

	return r.GetContainer(ctx, container.GetParentID())
}

// GetPath returns the path to a container as a slice of container IDs
func (r *FileSystemContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	var path []string
	currentID := containerID

	for currentID != "" {
		container, err := r.GetContainer(ctx, currentID)
		if err != nil {
			return nil, err
		}

		path = append([]string{currentID}, path...) // Prepend to build path from root
		currentID = container.GetParentID()
	}

	return path, nil
}

// FindByPath finds a container by its path
func (r *FileSystemContainerRepository) FindByPath(ctx context.Context, path string) (domain.ContainerResource, error) {
	// Simple implementation - assumes path is just the container ID
	// A more sophisticated implementation would parse hierarchical paths
	return r.GetContainer(ctx, path)
}

// Helper methods

// getContainerPath returns the filesystem path for a container
func (r *FileSystemContainerRepository) getContainerPath(id string) string {
	sanitizedID := r.sanitizeID(id)
	return filepath.Join(r.basePath, "containers", sanitizedID)
}

// sanitizeID sanitizes a container ID for safe filesystem usage
func (r *FileSystemContainerRepository) sanitizeID(id string) string {
	// Replace any potentially dangerous characters
	sanitized := strings.ReplaceAll(id, "..", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	return sanitized
}

// ContainerMetadata represents the metadata stored for a container
type ContainerMetadata struct {
	ID            string    `json:"id"`
	ParentID      string    `json:"parentId"`
	ContainerType string    `json:"containerType"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Members       []string  `json:"members"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// storeContainerMetadata stores container metadata as JSON
func (r *FileSystemContainerRepository) storeContainerMetadata(container domain.ContainerResource) error {
	containerDir := r.getContainerPath(container.ID())
	metadataPath := filepath.Join(containerDir, "container.json")

	// Create metadata structure
	metadata := ContainerMetadata{
		ID:            container.ID(),
		ParentID:      container.GetParentID(),
		ContainerType: container.GetContainerType().String(),
		Title:         container.GetTitle(),
		Description:   container.GetDescription(),
		Members:       container.GetMembers(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Extract timestamps from container metadata if available
	if createdAt, exists := container.GetMetadata()["createdAt"]; exists {
		if t, ok := createdAt.(time.Time); ok {
			metadata.CreatedAt = t
		}
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal container metadata: %w", err)
	}

	// Write to file
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write container metadata file: %w", err)
	}

	return nil
}

// loadContainerMetadata loads container metadata from JSON
func (r *FileSystemContainerRepository) loadContainerMetadata(id string) (*ContainerMetadata, error) {
	containerDir := r.getContainerPath(id)
	metadataPath := filepath.Join(containerDir, "container.json")

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read container metadata file: %w", err)
	}

	// Unmarshal JSON
	var metadata ContainerMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal container metadata: %w", err)
	}

	return &metadata, nil
}

// indexContainer indexes a container in the membership indexer
func (r *FileSystemContainerRepository) indexContainer(ctx context.Context, container domain.ContainerResource) error {
	// Index all existing members
	for _, memberID := range container.GetMembers() {
		if err := r.indexer.IndexMembership(ctx, container.ID(), memberID); err != nil {
			return fmt.Errorf("failed to index member %s: %w", memberID, err)
		}
	}

	return nil
}

// Implement remaining ResourceRepository methods by delegating to base repository

// Store stores a resource (delegates to base repository)
func (r *FileSystemContainerRepository) Store(ctx context.Context, resource domain.Resource) error {
	return r.FileSystemRepository.Store(ctx, resource)
}

// Retrieve retrieves a resource (delegates to base repository)
func (r *FileSystemContainerRepository) Retrieve(ctx context.Context, id string) (domain.Resource, error) {
	return r.FileSystemRepository.Retrieve(ctx, id)
}

// Delete deletes a resource (delegates to base repository)
func (r *FileSystemContainerRepository) Delete(ctx context.Context, id string) error {
	return r.FileSystemRepository.Delete(ctx, id)
}

// Exists checks if a resource exists (delegates to base repository)
func (r *FileSystemContainerRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.FileSystemRepository.Exists(ctx, id)
}

// StoreStream stores a resource from a stream (delegates to base repository)
func (r *FileSystemContainerRepository) StoreStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) error {
	return r.FileSystemRepository.StoreStream(ctx, id, reader, contentType, size)
}

// RetrieveStream retrieves a resource as a stream (delegates to base repository)
func (r *FileSystemContainerRepository) RetrieveStream(ctx context.Context, id string) (io.ReadCloser, *domain.ResourceMetadata, error) {
	return r.FileSystemRepository.RetrieveStream(ctx, id)
}

// insertContainerIntoDatabase inserts container metadata into the database
func (r *FileSystemContainerRepository) insertContainerIntoDatabase(ctx context.Context, container domain.ContainerResource) error {
	// Get database connection from indexer
	db, err := r.getDatabaseConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Handle parent_id - use NULL if empty
	var parentID interface{}
	if container.GetParentID() == "" {
		parentID = nil
	} else {
		parentID = container.GetParentID()
	}

	// Insert container into containers table
	query := `
		INSERT OR REPLACE INTO containers (id, parent_id, type, title, description, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err = db.ExecContext(ctx, query,
		container.ID(),
		parentID,
		container.GetContainerType().String(),
		container.GetTitle(),
		container.GetDescription(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert container into database: %w", err)
	}

	return nil
}

// getDatabaseConnection gets the database connection from the membership indexer
func (r *FileSystemContainerRepository) getDatabaseConnection() (*sql.DB, error) {
	// Type assert to get the SQLite indexer
	sqliteIndexer, ok := r.indexer.(*SQLiteMembershipIndexer)
	if !ok {
		return nil, fmt.Errorf("indexer is not a SQLiteMembershipIndexer")
	}

	return sqliteIndexer.GetDB(), nil
}

// removeContainerFromDatabase removes container from the database
func (r *FileSystemContainerRepository) removeContainerFromDatabase(ctx context.Context, containerID string) error {
	// Get database connection from indexer
	db, err := r.getDatabaseConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Remove container from containers table
	query := "DELETE FROM containers WHERE id = ?"
	_, err = db.ExecContext(ctx, query, containerID)
	if err != nil {
		return fmt.Errorf("failed to remove container from database: %w", err)
	}

	return nil
}
