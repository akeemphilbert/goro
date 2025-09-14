package application

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// EventPersistenceHandler handles persistence of all events to file system
// This is designed to be extensible for future activity stream integration
type EventPersistenceHandler struct {
	eventLogPath          string
	enableFilePersistence bool
	activityStreamEnabled bool // For future activity stream integration
}

// EventPersistenceConfig configures the event persistence handler
type EventPersistenceConfig struct {
	EventLogPath          string
	EnableFilePersistence bool
	ActivityStreamEnabled bool
}

// NewEventPersistenceHandler creates a new event persistence handler with default configuration
func NewEventPersistenceHandler() *EventPersistenceHandler {
	return &EventPersistenceHandler{
		eventLogPath:          "pod-data/events",
		enableFilePersistence: true,
		activityStreamEnabled: false, // For future use
	}
}

// NewEventPersistenceHandlerWithConfig creates a new event persistence handler with custom configuration
func NewEventPersistenceHandlerWithConfig(config EventPersistenceConfig) *EventPersistenceHandler {
	return &EventPersistenceHandler{
		eventLogPath:          config.EventLogPath,
		enableFilePersistence: config.EnableFilePersistence,
		activityStreamEnabled: config.ActivityStreamEnabled,
	}
}

// EventTypes returns empty slice as this handler subscribes to all events
func (h *EventPersistenceHandler) EventTypes() []string {
	return []string{} // Will subscribe to all events
}

// Handle processes any event and persists it
func (h *EventPersistenceHandler) Handle(ctx context.Context, envelope pericarpdomain.Envelope) error {
	event := envelope.Event()

	// Handle different event types
	if entityEvent, ok := event.(*pericarpdomain.EntityEvent); ok {
		return h.handleEntityEvent(ctx, entityEvent)
	}

	// Handle other event types as needed
	return h.handleGenericEvent(ctx, event)
}

// handleEntityEvent handles entity events (resources, containers, etc.)
func (h *EventPersistenceHandler) handleEntityEvent(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Persist to file system
	if err := h.persistEventToFileSystem(event); err != nil {
		return fmt.Errorf("failed to persist entity event to file system: %w", err)
	}

	// Future: Add activity stream persistence
	if h.activityStreamEnabled {
		if err := h.persistToActivityStream(ctx, event); err != nil {
			// Log error but don't fail the event processing
			fmt.Printf("Failed to persist event to activity stream: %v\n", err)
		}
	}

	return nil
}

// handleGenericEvent handles non-entity events
func (h *EventPersistenceHandler) handleGenericEvent(ctx context.Context, event pericarpdomain.Event) error {
	// For now, just log generic events
	fmt.Printf("Generic event received: Type=%T, ID=%s\n", event, event.GetID())

	// Future: Implement generic event persistence
	return nil
}

// persistEventToFileSystem persists an event to the file system for audit trail
func (h *EventPersistenceHandler) persistEventToFileSystem(event *pericarpdomain.EntityEvent) error {
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

	// Create event log entry with additional metadata for activity streams
	eventEntry := map[string]interface{}{
		"eventID":     fmt.Sprintf("%s-%d", event.AggregateID(), now.UnixNano()),
		"timestamp":   now.Format(time.RFC3339),
		"entityType":  event.EntityType,
		"eventType":   event.Type,
		"aggregateID": event.AggregateID(),
		"payload":     json.RawMessage(event.Payload()),

		// Additional metadata for activity stream compatibility
		"version":     "1.0",
		"source":      "ldp-server",
		"dataVersion": event.GetVersion(), // If available
	}

	// Marshal event to JSON (single line for easier parsing)
	eventJSON, err := json.Marshal(eventEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	// Determine log file based on entity type
	var logFile string
	switch event.EntityType {
	case "resource":
		logFile = filepath.Join(eventDir, "resource-events.log")
	case "container":
		logFile = filepath.Join(eventDir, "container-events.log")
	default:
		logFile = filepath.Join(eventDir, "events.log")
	}

	// Write to event log file (append mode)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open event log file %s: %w", logFile, err)
	}
	defer file.Close()

	// Write event entry with newline
	if _, err := file.Write(append(eventJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write event to log file: %w", err)
	}

	fmt.Printf("Event persisted: %s/%s -> %s\n", event.EntityType, event.Type, logFile)
	return nil
}

// persistToActivityStream persists an event to activity stream (future implementation)
func (h *EventPersistenceHandler) persistToActivityStream(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Future implementation for Activity Streams 2.0 format
	// This would convert the event to Activity Streams format and store it
	// in a format suitable for ActivityPub or other activity stream consumers

	activityStreamEntry := h.convertToActivityStreamFormat(event)

	// For now, just log what would be persisted
	activityJSON, _ := json.MarshalIndent(activityStreamEntry, "", "  ")
	fmt.Printf("Activity Stream Entry (not yet persisted):\n%s\n", activityJSON)

	return nil
}

// convertToActivityStreamFormat converts an entity event to Activity Streams 2.0 format
func (h *EventPersistenceHandler) convertToActivityStreamFormat(event *pericarpdomain.EntityEvent) map[string]interface{} {
	// Convert to Activity Streams 2.0 format
	// See: https://www.w3.org/TR/activitystreams-core/

	var activityType string
	switch event.Type {
	case "resource.created", "container.created":
		activityType = "Create"
	case "resource.updated", "container.updated":
		activityType = "Update"
	case "resource.deleted", "container.deleted":
		activityType = "Delete"
	case "container.member_added":
		activityType = "Add"
	case "container.member_removed":
		activityType = "Remove"
	default:
		activityType = "Activity"
	}

	return map[string]interface{}{
		"@context": []interface{}{
			"https://www.w3.org/ns/activitystreams",
			map[string]interface{}{
				"ldp": "http://www.w3.org/ns/ldp#",
			},
		},
		"type":      activityType,
		"id":        fmt.Sprintf("urn:event:%s", event.GetID()),
		"published": time.Now().Format(time.RFC3339),
		"actor": map[string]interface{}{
			"type": "Service",
			"name": "LDP Server",
		},
		"object": map[string]interface{}{
			"type": event.EntityType,
			"id":   event.AggregateID(),
		},
		"ldp:eventType":   event.Type,
		"ldp:entityType":  event.EntityType,
		"ldp:aggregateID": event.AggregateID(),
		"ldp:payload":     json.RawMessage(event.Payload()),
	}
}

// GetEventLogPath returns the current event log path
func (h *EventPersistenceHandler) GetEventLogPath() string {
	return h.eventLogPath
}

// SetEventLogPath sets the event log path
func (h *EventPersistenceHandler) SetEventLogPath(path string) {
	h.eventLogPath = path
}

// IsFilePersistenceEnabled returns whether file persistence is enabled
func (h *EventPersistenceHandler) IsFilePersistenceEnabled() bool {
	return h.enableFilePersistence
}

// SetFilePersistenceEnabled enables or disables file persistence
func (h *EventPersistenceHandler) SetFilePersistenceEnabled(enabled bool) {
	h.enableFilePersistence = enabled
}

// IsActivityStreamEnabled returns whether activity stream persistence is enabled
func (h *EventPersistenceHandler) IsActivityStreamEnabled() bool {
	return h.activityStreamEnabled
}

// SetActivityStreamEnabled enables or disables activity stream persistence
func (h *EventPersistenceHandler) SetActivityStreamEnabled(enabled bool) {
	h.activityStreamEnabled = enabled
}
