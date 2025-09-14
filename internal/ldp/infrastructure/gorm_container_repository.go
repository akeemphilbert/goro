package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"gorm.io/gorm"
)

// GORMContainerRepository implements ContainerRepository using GORM
type GORMContainerRepository struct {
	db *gorm.DB
}

// NewGORMContainerRepository creates a new GORM-based container repository
func NewGORMContainerRepository(db *gorm.DB) (*GORMContainerRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	// Auto-migrate the models
	if err := db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate models: %w", err)
	}

	return &GORMContainerRepository{
		db: db,
	}, nil
}

// Store implements ResourceRepository.Store
func (r *GORMContainerRepository) Store(ctx context.Context, resource domain.Resource) error {
	if resource == nil {
		return fmt.Errorf("resource cannot be nil")
	}

	// Convert domain resource to GORM model
	if containerResource, ok := resource.(domain.ContainerResource); ok {
		return r.CreateContainer(ctx, containerResource)
	}

	// Handle basic resource
	return r.storeBasicResource(ctx, resource)
}

// storeBasicResource stores a basic resource
func (r *GORMContainerRepository) storeBasicResource(ctx context.Context, resource domain.Resource) error {
	// Serialize metadata
	metadataJSON, err := json.Marshal(resource.GetMetadata())
	if err != nil {
		return fmt.Errorf("failed to serialize metadata: %w", err)
	}

	model := &ResourceModel{
		ID:          resource.ID(),
		ContentType: resource.GetContentType(),
		Size:        int64(resource.GetSize()),
		Metadata:    string(metadataJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// TODO: Implement file storage and set FilePath
	// For now, we'll store small resources in the database
	// In production, large resources should be stored on filesystem

	return r.db.WithContext(ctx).Create(model).Error
}

// Retrieve implements ResourceRepository.Retrieve
func (r *GORMContainerRepository) Retrieve(ctx context.Context, id string) (domain.Resource, error) {
	if id == "" {
		return nil, fmt.Errorf("resource ID cannot be empty")
	}

	// First try to find as container
	container, err := r.GetContainer(ctx, id)
	if err == nil {
		return container, nil
	}

	// If not found as container, try as basic resource
	var model ResourceModel
	err = r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("resource not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	// Convert to domain resource
	return r.resourceModelToDomain(&model)
}

// Delete implements ResourceRepository.Delete
func (r *GORMContainerRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	// Try to delete as container first
	err := r.DeleteContainer(ctx, id)
	if err == nil {
		return nil
	}

	// If not a container, delete as basic resource
	result := r.db.WithContext(ctx).Delete(&ResourceModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete resource: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("resource not found: %s", id)
	}

	return nil
}

// Exists implements ResourceRepository.Exists
func (r *GORMContainerRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("resource ID cannot be empty")
	}

	// Check if exists as container
	exists, err := r.ContainerExists(ctx, id)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// Check if exists as basic resource
	var count int64
	err = r.db.WithContext(ctx).Model(&ResourceModel{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check resource existence: %w", err)
	}

	return count > 0, nil
}

// CreateContainer implements ContainerRepository.CreateContainer
func (r *GORMContainerRepository) CreateContainer(ctx context.Context, container domain.ContainerResource) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	// Check if container already exists
	exists, err := r.ContainerExists(ctx, container.ID())
	if err != nil {
		return fmt.Errorf("failed to check container existence: %w", err)
	}
	if exists {
		return fmt.Errorf("container already exists: %s", container.ID())
	}

	// Convert domain container to GORM model
	model := r.containerDomainToModel(container)

	// Create container in transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create container
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}

		// Create membership record if it has a parent
		if container.GetParentID() != "" {
			membership := &MembershipModel{
				ContainerID: container.GetParentID(),
				MemberID:    container.ID(),
				MemberType:  "Container",
				CreatedAt:   time.Now(),
			}
			if err := tx.Create(membership).Error; err != nil {
				return fmt.Errorf("failed to create membership: %w", err)
			}
		}

		return nil
	})
}

// GetContainer implements ContainerRepository.GetContainer
func (r *GORMContainerRepository) GetContainer(ctx context.Context, id string) (domain.ContainerResource, error) {
	if id == "" {
		return nil, fmt.Errorf("container ID cannot be empty")
	}

	var model ContainerModel
	err := r.db.WithContext(ctx).
		Preload("Children").
		Preload("Resources").
		First(&model, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found: %s", id)
		}
		return nil, fmt.Errorf("failed to retrieve container: %w", err)
	}

	// Convert to domain container
	return r.containerModelToDomain(&model)
}

// UpdateContainer implements ContainerRepository.UpdateContainer
func (r *GORMContainerRepository) UpdateContainer(ctx context.Context, container domain.ContainerResource) error {
	if container == nil {
		return fmt.Errorf("container cannot be nil")
	}

	model := r.containerDomainToModel(container)
	model.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).
		Model(&ContainerModel{}).
		Where("id = ?", container.ID()).
		Updates(model)

	if result.Error != nil {
		return fmt.Errorf("failed to update container: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("container not found: %s", container.ID())
	}

	return nil
}

// DeleteContainer implements ContainerRepository.DeleteContainer
func (r *GORMContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete memberships
		if err := tx.Delete(&MembershipModel{}, "container_id = ? OR member_id = ?", id, id).Error; err != nil {
			return fmt.Errorf("failed to delete memberships: %w", err)
		}

		// Delete container
		result := tx.Delete(&ContainerModel{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("failed to delete container: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("container not found: %s", id)
		}

		return nil
	})
}

// ContainerExists implements ContainerRepository.ContainerExists
func (r *GORMContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("container ID cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&ContainerModel{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check container existence: %w", err)
	}

	return count > 0, nil
}

// AddMember implements ContainerRepository.AddMember
func (r *GORMContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	if containerID == "" || memberID == "" {
		return fmt.Errorf("container ID and member ID cannot be empty")
	}

	// Determine member type
	memberType := "Resource"
	exists, err := r.ContainerExists(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to check member type: %w", err)
	}
	if exists {
		memberType = "Container"
	}

	membership := &MembershipModel{
		ContainerID: containerID,
		MemberID:    memberID,
		MemberType:  memberType,
		CreatedAt:   time.Now(),
	}

	return r.db.WithContext(ctx).Create(membership).Error
}

// RemoveMember implements ContainerRepository.RemoveMember
func (r *GORMContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	if containerID == "" || memberID == "" {
		return fmt.Errorf("container ID and member ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&MembershipModel{}, "container_id = ? AND member_id = ?", containerID, memberID)
	if result.Error != nil {
		return fmt.Errorf("failed to remove membership: %w", result.Error)
	}

	return nil
}

// ListMembers implements ContainerRepository.ListMembers
func (r *GORMContainerRepository) ListMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) ([]string, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID cannot be empty")
	}

	var memberIDs []string
	query := r.db.WithContext(ctx).
		Model(&MembershipModel{}).
		Where("container_id = ?", containerID).
		Select("member_id")

	if pagination.Limit > 0 {
		query = query.Limit(pagination.Limit)
	}
	if pagination.Offset > 0 {
		query = query.Offset(pagination.Offset)
	}

	err := query.Find(&memberIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	return memberIDs, nil
}

// GetChildren implements ContainerRepository.GetChildren
func (r *GORMContainerRepository) GetChildren(ctx context.Context, containerID string) ([]domain.ContainerResource, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID cannot be empty")
	}

	var models []ContainerModel
	err := r.db.WithContext(ctx).Where("parent_id = ?", containerID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}

	var children []domain.ContainerResource
	for _, model := range models {
		child, err := r.containerModelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert child container: %w", err)
		}
		children = append(children, child)
	}

	return children, nil
}

// GetParent implements ContainerRepository.GetParent
func (r *GORMContainerRepository) GetParent(ctx context.Context, containerID string) (domain.ContainerResource, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID cannot be empty")
	}

	// First get the container to find its parent ID
	container, err := r.GetContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	parentID := container.GetParentID()
	if parentID == "" {
		return nil, fmt.Errorf("container has no parent: %s", containerID)
	}

	return r.GetContainer(ctx, parentID)
}

// GetPath implements ContainerRepository.GetPath
func (r *GORMContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	if containerID == "" {
		return nil, fmt.Errorf("container ID cannot be empty")
	}

	var path []string
	currentID := containerID

	for currentID != "" {
		path = append([]string{currentID}, path...)

		container, err := r.GetContainer(ctx, currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get container in path: %w", err)
		}

		currentID = container.GetParentID()
	}

	return path, nil
}

// FindByPath implements ContainerRepository.FindByPath
func (r *GORMContainerRepository) FindByPath(ctx context.Context, path string) (domain.ContainerResource, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// For now, assume path is just the container ID
	// In a more sophisticated implementation, you'd parse the path
	return r.GetContainer(ctx, path)
}

// Helper methods for domain model conversion

func (r *GORMContainerRepository) containerDomainToModel(container domain.ContainerResource) *ContainerModel {
	var parentID *string
	if pid := container.GetParentID(); pid != "" {
		parentID = &pid
	}

	return &ContainerModel{
		ID:          container.ID(),
		ParentID:    parentID,
		Type:        container.GetContainerType().String(),
		Title:       container.GetTitle(),
		Description: container.GetDescription(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (r *GORMContainerRepository) containerModelToDomain(model *ContainerModel) (domain.ContainerResource, error) {
	// Create the underlying basic resource
	ctx := context.Background()
	basicResource := domain.NewResource(ctx, model.ID, "application/ld+json", []byte("{}"))

	// Set metadata
	basicResource.SetMetadata("type", "Container")
	basicResource.SetMetadata("containerType", model.Type)
	basicResource.SetMetadata("title", model.Title)
	basicResource.SetMetadata("description", model.Description)
	basicResource.SetMetadata("createdAt", model.CreatedAt)
	basicResource.SetMetadata("updatedAt", model.UpdatedAt)

	// Determine parent ID
	parentID := ""
	if model.ParentID != nil {
		parentID = *model.ParentID
	}

	// Parse container type
	var containerType domain.ContainerType
	switch model.Type {
	case "DirectContainer":
		containerType = domain.DirectContainer
	default:
		containerType = domain.BasicContainer
	}

	// Create container
	container := &domain.Container{
		BasicResource: basicResource,
		ParentID:      parentID,
		ContainerType: containerType,
		Children:      make([]domain.Resource, 0),
	}

	// Convert child containers and resources to domain objects
	for _, childModel := range model.Children {
		child, err := r.containerModelToDomain(&childModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert child container: %w", err)
		}
		container.AddChild(child)
	}

	for _, resourceModel := range model.Resources {
		resource, err := r.resourceModelToDomain(&resourceModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert child resource: %w", err)
		}
		container.AddChild(resource)
	}

	return container, nil
}

func (r *GORMContainerRepository) resourceModelToDomain(model *ResourceModel) (domain.Resource, error) {
	ctx := context.Background()

	// TODO: Load data from filesystem if FilePath is set
	// For now, assume data is small and stored in database
	data := []byte{} // Placeholder

	resource := domain.NewResource(ctx, model.ID, model.ContentType, data)

	// Deserialize metadata
	if model.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err == nil {
			for key, value := range metadata {
				resource.SetMetadata(key, value)
			}
		}
	}

	resource.SetMetadata("createdAt", model.CreatedAt)
	resource.SetMetadata("updatedAt", model.UpdatedAt)

	return resource, nil
}
