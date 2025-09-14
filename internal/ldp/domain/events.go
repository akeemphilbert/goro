package domain

import (
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Re-export pericarp types for convenience
type EntityEvent = pericarpdomain.EntityEvent
type EventDispatcher = pericarpdomain.EventDispatcher
type EventHandler = pericarpdomain.EventHandler

// Event types for resource operations
const (
	EventTypeResourceCreated              = "created"
	EventTypeResourceUpdated              = "updated"
	EventTypeResourceDeleted              = "deleted"
	EventTypeResourceCreatedWithRelations = "created_with_relations"
	EventTypeResourceUpdatedWithRelations = "updated_with_relations"
	EventTypeResourceLinked               = "linked"
	EventTypeResourceRelationshipUpdated  = "relationship_updated"
)

// Event types for container operations
const (
	EventTypeContainerCreated = "container_created"
	EventTypeContainerUpdated = "container_updated"
	EventTypeContainerDeleted = "container_deleted"
	EventTypeMemberAdded      = "member_added"
	EventTypeMemberRemoved    = "member_removed"
)

// NewResourceCreatedEvent creates a new resource created event
func NewResourceCreatedEvent(resourceID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceCreated, resourceID, "", "", data)
}

// NewResourceUpdatedEvent creates a new resource updated event
func NewResourceUpdatedEvent(resourceID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceUpdated, resourceID, "", "", data)
}

// NewResourceDeletedEvent creates a new resource deleted event
func NewResourceDeletedEvent(resourceID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceDeleted, resourceID, "", "", data)
}

// NewResourceCreatedWithRelationsEvent creates a new resource created event with RDF relationships
func NewResourceCreatedWithRelationsEvent(resourceID string, data *ResourceEventData) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceCreatedWithRelations, resourceID, "", "", data)
}

// NewResourceUpdatedWithRelationsEvent creates a new resource updated event with RDF relationships
func NewResourceUpdatedWithRelationsEvent(resourceID string, data *ResourceEventData) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceUpdated, resourceID, "", "", data)
}

// NewResourceLinkedEvent creates a new resource linked event
func NewResourceLinkedEvent(resourceID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceLinked, resourceID, "", "", data)
}

// NewResourceRelationshipUpdatedEvent creates a new resource relationship updated event
func NewResourceRelationshipUpdatedEvent(resourceID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("resource", EventTypeResourceRelationshipUpdated, resourceID, "", "", data)
}

// Container event constructors

// NewContainerCreatedEvent creates a new container created event
func NewContainerCreatedEvent(containerID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("container", EventTypeContainerCreated, containerID, "", "", data)
}

// NewContainerUpdatedEvent creates a new container updated event
func NewContainerUpdatedEvent(containerID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("container", EventTypeContainerUpdated, containerID, "", "", data)
}

// NewContainerDeletedEvent creates a new container deleted event
func NewContainerDeletedEvent(containerID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("container", EventTypeContainerDeleted, containerID, "", "", data)
}

// NewMemberAddedEvent creates a new member added event
func NewMemberAddedEvent(containerID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("container", EventTypeMemberAdded, containerID, "", "", data)
}

// NewMemberRemovedEvent creates a new member removed event
func NewMemberRemovedEvent(containerID string, data interface{}) *EntityEvent {
	return pericarpdomain.NewEntityEvent("container", EventTypeMemberRemoved, containerID, "", "", data)
}
