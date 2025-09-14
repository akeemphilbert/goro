package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ContainerType represents the type of LDP container
type ContainerType string

const (
	BasicContainer  ContainerType = "BasicContainer"
	DirectContainer ContainerType = "DirectContainer"
)

// String returns the string representation of the container type
func (ct ContainerType) String() string {
	return string(ct)
}

// IsValid checks if the container type is valid
func (ct ContainerType) IsValid() bool {
	switch ct {
	case BasicContainer, DirectContainer:
		return true
	default:
		return false
	}
}

// Container represents a container resource that can hold other resources
type Container struct {
	*Resource                   // Inherits from Resource
	Members       []string      // Resource IDs contained in this container
	ParentID      string        // Parent container ID (empty for root)
	ContainerType ContainerType // BasicContainer, DirectContainer, etc.
}

// NewContainer creates a new Container entity
func NewContainer(id, parentID string, containerType ContainerType) *Container {
	if id == "" {
		id = uuid.New().String()
	}

	// Create the underlying resource with container-specific content type
	resource := NewResource(id, "application/ld+json", []byte("{}"))

	container := &Container{
		Resource:      resource,
		Members:       make([]string, 0),
		ParentID:      parentID,
		ContainerType: containerType,
	}

	// Set container-specific metadata
	container.SetMetadata("type", "Container")
	container.SetMetadata("containerType", containerType.String())
	container.SetMetadata("parentID", parentID)
	container.SetMetadata("createdAt", time.Now())

	// Emit container created event (resource created event will also be present)
	event := NewContainerCreatedEvent(id, map[string]interface{}{
		"parentID":      parentID,
		"containerType": containerType.String(),
		"createdAt":     time.Now(),
	})
	container.AddEvent(event)

	return container
}

// AddMember adds a resource to the container
func (c *Container) AddMember(memberID string) error {
	// Check if member already exists
	for _, member := range c.Members {
		if member == memberID {
			return fmt.Errorf("member already exists: %s", memberID)
		}
	}

	c.Members = append(c.Members, memberID)
	c.SetMetadata("updatedAt", time.Now())

	// Emit member added event
	event := NewMemberAddedEvent(c.ID(), map[string]interface{}{
		"memberID":   memberID,
		"memberType": "Resource", // Default to Resource, could be enhanced to detect type
		"addedAt":    time.Now(),
	})
	c.AddEvent(event)

	return nil
}

// RemoveMember removes a resource from the container
func (c *Container) RemoveMember(memberID string) error {
	for i, member := range c.Members {
		if member == memberID {
			// Remove member from slice
			c.Members = append(c.Members[:i], c.Members[i+1:]...)
			c.SetMetadata("updatedAt", time.Now())

			// Emit member removed event
			event := NewMemberRemovedEvent(c.ID(), map[string]interface{}{
				"memberID":   memberID,
				"memberType": "Resource",
				"removedAt":  time.Now(),
			})
			c.AddEvent(event)

			return nil
		}
	}

	return fmt.Errorf("member not found: %s", memberID)
}

// HasMember checks if a resource is a member of the container
func (c *Container) HasMember(memberID string) bool {
	for _, member := range c.Members {
		if member == memberID {
			return true
		}
	}
	return false
}

// GetMemberCount returns the number of members in the container
func (c *Container) GetMemberCount() int {
	return len(c.Members)
}

// IsEmpty checks if the container has no members
func (c *Container) IsEmpty() bool {
	return len(c.Members) == 0
}

// ValidateHierarchy validates the container hierarchy to prevent circular references
func (c *Container) ValidateHierarchy(ancestorPath []string) error {
	// Check for self-reference
	if c.ParentID == c.ID() {
		return fmt.Errorf("container cannot be its own parent")
	}

	// Check for circular reference in the ancestor path
	for _, ancestorID := range ancestorPath {
		if ancestorID == c.ID() {
			return fmt.Errorf("circular reference detected: container %s is already in the hierarchy path", c.ID())
		}
	}

	return nil
}

// SetTitle sets the container title
func (c *Container) SetTitle(title string) {
	c.SetMetadata("title", title)
	c.SetMetadata("updatedAt", time.Now())

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"title":     title,
		"updatedAt": time.Now(),
	})
	c.AddEvent(event)
}

// GetTitle returns the container title
func (c *Container) GetTitle() string {
	if title, exists := c.GetMetadata()["title"]; exists {
		if titleStr, ok := title.(string); ok {
			return titleStr
		}
	}
	return ""
}

// SetDescription sets the container description
func (c *Container) SetDescription(description string) {
	c.SetMetadata("description", description)
	c.SetMetadata("updatedAt", time.Now())

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"description": description,
		"updatedAt":   time.Now(),
	})
	c.AddEvent(event)
}

// GetDescription returns the container description
func (c *Container) GetDescription() string {
	if description, exists := c.GetMetadata()["description"]; exists {
		if descStr, ok := description.(string); ok {
			return descStr
		}
	}
	return ""
}

// GetPath returns the path representation of the container
func (c *Container) GetPath() string {
	if c.ParentID == "" {
		return c.ID()
	}
	return c.ParentID + "/" + c.ID()
}

// Delete marks the container as deleted and emits a delete event
func (c *Container) Delete() error {
	// Check if container is empty
	if !c.IsEmpty() {
		return fmt.Errorf("container is not empty: cannot delete container with %d members", len(c.Members))
	}

	// Emit delete event
	event := NewContainerDeletedEvent(c.ID(), map[string]interface{}{
		"deletedAt": time.Now(),
	})
	c.AddEvent(event)

	return nil
}

// PaginationOptions represents pagination parameters for listing operations
type PaginationOptions struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// IsValid validates the pagination options
func (p PaginationOptions) IsValid() bool {
	return p.Limit > 0 && p.Limit <= 1000 && p.Offset >= 0
}

// GetDefaultPagination returns default pagination options
func GetDefaultPagination() PaginationOptions {
	return PaginationOptions{
		Limit:  50,
		Offset: 0,
	}
}
