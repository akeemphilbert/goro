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
	EventTypeResourceCreated = "created"
	EventTypeResourceUpdated = "updated"
	EventTypeResourceDeleted = "deleted"
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
