package domain

import (
	"context"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Resource defines the interface that all resource types must implement
type Resource interface {
	pericarpdomain.Entity
	// Identity
	ID() string

	// Content
	GetContentType() string
	GetData() []byte
	GetSize() int

	// Metadata
	GetMetadata() map[string]interface{}
	SetMetadata(key string, value interface{})

	// Validation and Error handling
	Reset()

	// Resource operations
	Update(ctx context.Context, data []byte, contentType string)
	Delete(ctx context.Context)

	// Format operations
	ToFormat(format string) ([]byte, error)
}

// ContainerResource defines the interface for container-specific operations
type ContainerResource interface {
	Resource

	// Container-specific properties
	GetParentID() string
	GetContainerType() ContainerType
	GetPath() string

	// Membership operations
	AddMember(ctx context.Context, resource Resource) error
	RemoveMember(ctx context.Context, resourceID string) error
	HasMember(memberID string) bool
	GetMembers() []string
	GetMemberCount() int
	IsEmpty() bool

	// Hierarchy operations
	ValidateHierarchy(ancestorPath []string) error

	// Container metadata operations
	SetTitle(title string)
	GetTitle() string
	SetDescription(description string)
	GetDescription() string
}

// ResourceType represents the type of resource for polymorphic operations
type ResourceType string

const (
	ResourceTypeBasic     ResourceType = "BasicResource"
	ResourceTypeContainer ResourceType = "Container"
)

// ResourceInfo provides basic information about a resource for projections
type ResourceInfo struct {
	ID          string                 `json:"id"`
	Type        ResourceType           `json:"type"`
	ContentType string                 `json:"contentType"`
	Size        int                    `json:"size"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// GetResourceInfo extracts basic information from any Resource
func GetResourceInfo(resource Resource) ResourceInfo {
	var resourceType ResourceType
	var createdAt, updatedAt time.Time

	// Determine resource type
	if _, ok := resource.(ContainerResource); ok {
		resourceType = ResourceTypeContainer
	} else {
		resourceType = ResourceTypeBasic
	}

	// Extract timestamps from metadata if available
	metadata := resource.GetMetadata()
	if created, exists := metadata["createdAt"]; exists {
		if t, ok := created.(time.Time); ok {
			createdAt = t
		}
	}
	if updated, exists := metadata["updatedAt"]; exists {
		if t, ok := updated.(time.Time); ok {
			updatedAt = t
		}
	}

	return ResourceInfo{
		ID:          resource.ID(),
		Type:        resourceType,
		ContentType: resource.GetContentType(),
		Size:        resource.GetSize(),
		Metadata:    metadata,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
