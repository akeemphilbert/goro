package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// ResourceEventHandler handles resource events and updates the repository accordingly
type ResourceEventHandler struct {
	repo                  domain.ResourceRepository
	eventLogPath          string
	enableFilePersistence bool
}

// NewResourceEventHandler creates a new resource event handler
func NewResourceEventHandler(repo domain.ResourceRepository) *ResourceEventHandler {
	return &ResourceEventHandler{
		repo:                  repo,
		eventLogPath:          "pod-data/events",
		enableFilePersistence: true,
	}
}

// NewResourceEventHandlerWithConfig creates a new resource event handler with custom configuration
func NewResourceEventHandlerWithConfig(repo domain.ResourceRepository, eventLogPath string, enableFilePersistence bool) *ResourceEventHandler {
	return &ResourceEventHandler{
		repo:                  repo,
		eventLogPath:          eventLogPath,
		enableFilePersistence: enableFilePersistence,
	}
}

// EventTypes returns the list of event types this handler can process
func (h *ResourceEventHandler) EventTypes() []string {
	return []string{
		"resource.created",
		"resource.updated",
		"resource.deleted",
	}
}

// Handle processes resource events and updates the repository
func (h *ResourceEventHandler) Handle(ctx context.Context, envelope pericarpdomain.Envelope) error {
	event := envelope.Event()

	// Check if this is a resource event
	if entityEvent, ok := event.(*pericarpdomain.EntityEvent); ok {
		if entityEvent.EntityType != "resource" {
			// Not a resource event, ignore
			return nil
		}

		switch entityEvent.Type {
		case domain.EventTypeResourceCreated:
			return h.handleResourceCreated(ctx, entityEvent)
		case domain.EventTypeResourceUpdated:
			return h.handleResourceUpdated(ctx, entityEvent)
		case domain.EventTypeResourceDeleted:
			return h.handleResourceDeleted(ctx, entityEvent)
		default:
			// Unknown event type, log and ignore
			fmt.Printf("Unknown resource event type: %s\n", entityEvent.Type)
			return nil
		}
	}

	return nil
}

// handleResourceCreated handles resource created events
func (h *ResourceEventHandler) handleResourceCreated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Persist event to file system first
	if err := h.persistEventToFileSystem(event); err != nil {
		return fmt.Errorf("failed to persist created event to file system: %w", err)
	}

	// Extract resource data from event payload
	resource, err := h.reconstructResourceFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to reconstruct resource from created event: %w", err)
	}

	// Store the resource in the repository
	if err := h.repo.Store(ctx, resource); err != nil {
		return fmt.Errorf("failed to store resource in repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s created\n", event.AggregateID())
	return nil
}

// handleResourceUpdated handles resource updated events
func (h *ResourceEventHandler) handleResourceUpdated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Persist event to file system first
	if err := h.persistEventToFileSystem(event); err != nil {
		return fmt.Errorf("failed to persist updated event to file system: %w", err)
	}

	// Extract resource data from event payload
	resource, err := h.reconstructResourceFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to reconstruct resource from updated event: %w", err)
	}

	// Update the resource in the repository
	if err := h.repo.Store(ctx, resource); err != nil {
		return fmt.Errorf("failed to update resource in repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s updated\n", event.AggregateID())
	return nil
}

// handleResourceDeleted handles resource deleted events
func (h *ResourceEventHandler) handleResourceDeleted(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Persist event to file system first
	if err := h.persistEventToFileSystem(event); err != nil {
		return fmt.Errorf("failed to persist deleted event to file system: %w", err)
	}

	// Delete the resource from the repository
	if err := h.repo.Delete(ctx, event.AggregateID()); err != nil {
		return fmt.Errorf("failed to delete resource from repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s deleted\n", event.AggregateID())
	return nil
}

// reconstructResourceFromEvent reconstructs a resource from an event payload
func (h *ResourceEventHandler) reconstructResourceFromEvent(event *pericarpdomain.EntityEvent) (*domain.Resource, error) {
	// The payload contains the resource data encoded as JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// Extract resource information from payload
	contentType, ok := payload["contentType"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid contentType in event payload")
	}

	size, ok := payload["size"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid size in event payload")
	}

	// For now, we'll create a placeholder resource since we don't have the actual data
	// In a real implementation, you might store the data in the event or have a separate mechanism
	// to retrieve the current state of the resource
	data := make([]byte, int(size))

	// Create the resource
	resource := domain.NewResource(event.AggregateID(), contentType, data)

	// Clear events since this is a reconstruction from events
	resource.MarkEventsAsCommitted()

	return resource, nil
}

// persistEventToFileSystem persists an event to the file system for audit trail
func (h *ResourceEventHandler) persistEventToFileSystem(event *pericarpdomain.EntityEvent) error {
	if !h.enableFilePersistence {
		return nil // Skip persistence if disabled
	}

	// Create event log directory structure: pod-data/events/{date}/
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	eventDir := filepath.Join(h.eventLogPath, dateDir)

	// Ensure directory exists
	if err := os.MkdirAll(eventDir, 0755); err != nil {
		return fmt.Errorf("failed to create event directory %s: %w", eventDir, err)
	}

	// Create event log entry
	eventEntry := map[string]interface{}{
		"eventID":     fmt.Sprintf("%s-%d", event.AggregateID(), now.UnixNano()),
		"timestamp":   now.Format(time.RFC3339),
		"entityType":  event.EntityType,
		"eventType":   event.Type,
		"aggregateID": event.AggregateID(),
		"payload":     json.RawMessage(event.Payload()),
	}

	// Marshal event to JSON (single line for easier parsing)
	eventJSON, err := json.Marshal(eventEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	// Write to event log file (append mode)
	logFile := filepath.Join(eventDir, "events.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open event log file %s: %w", logFile, err)
	}
	defer file.Close()

	// Write event entry with newline
	if _, err := file.Write(append(eventJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write event to log file: %w", err)
	}

	return nil
}

// GetEventLogPath returns the current event log path
func (h *ResourceEventHandler) GetEventLogPath() string {
	return h.eventLogPath
}

// SetEventLogPath sets the event log path
func (h *ResourceEventHandler) SetEventLogPath(path string) {
	h.eventLogPath = path
}

// IsFilePersistenceEnabled returns whether file persistence is enabled
func (h *ResourceEventHandler) IsFilePersistenceEnabled() bool {
	return h.enableFilePersistence
}

// SetFilePersistenceEnabled enables or disables file persistence
func (h *ResourceEventHandler) SetFilePersistenceEnabled(enabled bool) {
	h.enableFilePersistence = enabled
}

// EventHandlerRegistrar registers event handlers with the event dispatcher
type EventHandlerRegistrar struct {
	dispatcher pericarpdomain.EventDispatcher
	handlers   []pericarpdomain.EventHandler
}

// NewEventHandlerRegistrar creates a new event handler registrar
func NewEventHandlerRegistrar(dispatcher pericarpdomain.EventDispatcher) *EventHandlerRegistrar {
	return &EventHandlerRegistrar{
		dispatcher: dispatcher,
		handlers:   make([]pericarpdomain.EventHandler, 0),
	}
}

// RegisterResourceEventHandler registers the resource event handler for all resource events
func (r *EventHandlerRegistrar) RegisterResourceEventHandler(handler *ResourceEventHandler) error {
	// Register for all resource event types
	eventTypes := []string{
		"resource.created",
		"resource.updated",
		"resource.deleted",
	}

	for _, eventType := range eventTypes {
		if err := r.dispatcher.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to subscribe to event type %s: %w", eventType, err)
		}
	}

	r.handlers = append(r.handlers, handler)
	fmt.Printf("Registered resource event handler for events: %v\n", eventTypes)
	return nil
}

// RegisterAllHandlers registers all event handlers
func (r *EventHandlerRegistrar) RegisterAllHandlers(repo domain.ResourceRepository) error {
	// Create and register resource event handler
	resourceHandler := NewResourceEventHandler(repo)
	if err := r.RegisterResourceEventHandler(resourceHandler); err != nil {
		return fmt.Errorf("failed to register resource event handler: %w", err)
	}

	return nil
}
