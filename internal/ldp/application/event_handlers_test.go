package application

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

func TestResourceEventHandler_Handle(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)
	ctx := context.Background()

	// Test resource created event
	createdEvent := domain.NewResourceCreatedEvent("test-resource", map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        100,
		"createdAt":   time.Now(),
	})

	// Create a mock envelope
	envelope := &mockEnvelope{
		event:     createdEvent,
		eventID:   "test-event-1",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event
	err := handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling created event: %v", err)
	}

	// Verify resource was stored in repository
	exists, err := repo.Exists(ctx, "test-resource")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should have been stored in repository")
	}
}

func TestResourceEventHandler_HandleUpdated(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)
	ctx := context.Background()

	// First create a resource
	initialResource := domain.NewResource("update-test", "text/plain", []byte("initial"))
	repo.Store(ctx, initialResource)

	// Test resource updated event
	updatedEvent := domain.NewResourceUpdatedEvent("update-test", map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        200,
		"updatedAt":   time.Now(),
	})

	envelope := &mockEnvelope{
		event:     updatedEvent,
		eventID:   "test-event-2",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event
	err := handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling updated event: %v", err)
	}

	// Verify resource still exists (updated, not deleted)
	exists, err := repo.Exists(ctx, "update-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should still exist after update")
	}
}

func TestResourceEventHandler_HandleDeleted(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)
	ctx := context.Background()

	// First create a resource
	initialResource := domain.NewResource("delete-test", "text/plain", []byte("to be deleted"))
	repo.Store(ctx, initialResource)

	// Verify it exists
	exists, err := repo.Exists(ctx, "delete-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Fatal("Resource should exist before deletion")
	}

	// Test resource deleted event
	deletedEvent := domain.NewResourceDeletedEvent("delete-test", map[string]interface{}{
		"deletedAt": time.Now(),
	})

	envelope := &mockEnvelope{
		event:     deletedEvent,
		eventID:   "test-event-3",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event
	err = handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling deleted event: %v", err)
	}

	// Verify resource was deleted from repository
	exists, err = repo.Exists(ctx, "delete-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if exists {
		t.Error("Resource should have been deleted from repository")
	}
}

func TestResourceEventHandler_IgnoreNonResourceEvents(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)
	ctx := context.Background()

	// Create a non-resource event
	nonResourceEvent := pericarpdomain.NewEntityEvent("user", "created", "user-123", "", "", map[string]interface{}{
		"name": "John Doe",
	})

	envelope := &mockEnvelope{
		event:     nonResourceEvent,
		eventID:   "test-event-4",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event - should not error and should ignore it
	err := handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling non-resource event: %v", err)
	}

	// Verify no resources were created
	exists, err := repo.Exists(ctx, "user-123")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if exists {
		t.Error("No resource should have been created for non-resource event")
	}
}

func TestEventHandlerRegistrar_RegisterResourceEventHandler(t *testing.T) {
	mockDispatcher := &mockEventDispatcher{}
	registrar := NewEventHandlerRegistrar(mockDispatcher)
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)

	err := registrar.RegisterResourceEventHandler(handler)
	if err != nil {
		t.Fatalf("Failed to register resource event handler: %v", err)
	}

	// Verify handlers were registered (this is a basic test since mockEventDispatcher doesn't track subscriptions)
	if len(registrar.handlers) != 1 {
		t.Errorf("Expected 1 handler to be registered, got %d", len(registrar.handlers))
	}
}

func TestEventHandlerRegistrar_RegisterAllHandlers(t *testing.T) {
	mockDispatcher := &mockEventDispatcher{}
	registrar := NewEventHandlerRegistrar(mockDispatcher)
	repo := newMockRepository()

	err := registrar.RegisterAllHandlers(repo)
	if err != nil {
		t.Fatalf("Failed to register all handlers: %v", err)
	}

	// Verify handlers were registered
	if len(registrar.handlers) == 0 {
		t.Error("Expected handlers to be registered")
	}
}

// Test helper to create a mock event with proper payload
func createMockResourceEvent(eventType, resourceID string, payload map[string]interface{}) *pericarpdomain.EntityEvent {
	// Create event with the payload data - pericarp will handle JSON encoding
	event := pericarpdomain.NewEntityEvent("resource", eventType, resourceID, "", "", payload)
	return event
}

func TestResourceEventHandler_ReconstructResourceFromEvent(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)

	// Create a mock event with proper payload
	payload := map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        float64(150), // JSON numbers are float64
		"createdAt":   time.Now().Format(time.RFC3339),
	}

	event := createMockResourceEvent("created", "reconstruct-test", payload)

	// Test reconstruction
	resource, err := handler.reconstructResourceFromEvent(event)
	if err != nil {
		t.Fatalf("Failed to reconstruct resource: %v", err)
	}

	if resource.ID() != "reconstruct-test" {
		t.Errorf("Expected resource ID 'reconstruct-test', got %s", resource.ID())
	}

	if resource.GetContentType() != "application/ld+json" {
		t.Errorf("Expected content type 'application/ld+json', got %s", resource.GetContentType())
	}

	if resource.GetSize() != 150 {
		t.Errorf("Expected size 150, got %d", resource.GetSize())
	}
}

func TestResourceEventHandler_FilePersistence(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	eventLogPath := filepath.Join(tempDir, "events")

	repo := newMockRepository()
	handler := NewResourceEventHandlerWithConfig(repo, eventLogPath, true)
	ctx := context.Background()

	// Test resource created event with file persistence
	createdEvent := domain.NewResourceCreatedEvent("persistence-test", map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        100,
		"createdAt":   time.Now(),
	})

	envelope := &mockEnvelope{
		event:     createdEvent,
		eventID:   "test-event-persistence",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event
	err := handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling created event: %v", err)
	}

	// Verify event was persisted to file system
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(eventLogPath, today, "events.log")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Event log file should exist at %s", logFile)
	}

	// Read and verify log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read event log file: %v", err)
	}

	// Parse the JSON event entry
	var eventEntry map[string]interface{}
	if err := json.Unmarshal(content, &eventEntry); err != nil {
		t.Fatalf("Failed to parse event log entry: %v", err)
	}

	// Verify event entry fields
	if eventEntry["entityType"] != "resource" {
		t.Errorf("Expected entityType 'resource', got %v", eventEntry["entityType"])
	}
	if eventEntry["eventType"] != "created" {
		t.Errorf("Expected eventType 'created', got %v", eventEntry["eventType"])
	}
	if eventEntry["aggregateID"] != "persistence-test" {
		t.Errorf("Expected aggregateID 'persistence-test', got %v", eventEntry["aggregateID"])
	}
}

func TestResourceEventHandler_FilePersistenceDisabled(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	eventLogPath := filepath.Join(tempDir, "events")

	repo := newMockRepository()
	handler := NewResourceEventHandlerWithConfig(repo, eventLogPath, false) // Disable persistence
	ctx := context.Background()

	// Test resource created event without file persistence
	createdEvent := domain.NewResourceCreatedEvent("no-persistence-test", map[string]interface{}{
		"contentType": "text/plain",
		"size":        50,
		"createdAt":   time.Now(),
	})

	envelope := &mockEnvelope{
		event:     createdEvent,
		eventID:   "test-event-no-persistence",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	// Handle the event
	err := handler.Handle(ctx, envelope)
	if err != nil {
		t.Fatalf("Unexpected error handling created event: %v", err)
	}

	// Verify no event log file was created
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(eventLogPath, today, "events.log")

	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("Event log file should not exist when persistence is disabled")
	}

	// Verify resource was still stored in repository
	exists, err := repo.Exists(ctx, "no-persistence-test")
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should have been stored in repository even without file persistence")
	}
}

func TestResourceEventHandler_MultipleEventsFilePersistence(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	eventLogPath := filepath.Join(tempDir, "events")

	repo := newMockRepository()
	handler := NewResourceEventHandlerWithConfig(repo, eventLogPath, true)
	ctx := context.Background()

	// Create multiple events
	events := []struct {
		eventType  string
		resourceID string
	}{
		{"created", "multi-test-1"},
		{"updated", "multi-test-1"},
		{"created", "multi-test-2"},
		{"deleted", "multi-test-1"},
	}

	for _, eventData := range events {
		var event *pericarpdomain.EntityEvent
		switch eventData.eventType {
		case "created":
			event = domain.NewResourceCreatedEvent(eventData.resourceID, map[string]interface{}{
				"contentType": "application/ld+json",
				"size":        100,
			})
		case "updated":
			event = domain.NewResourceUpdatedEvent(eventData.resourceID, map[string]interface{}{
				"contentType": "text/turtle",
				"size":        150,
			})
		case "deleted":
			event = domain.NewResourceDeletedEvent(eventData.resourceID, map[string]interface{}{
				"deletedAt": time.Now(),
			})
		}

		envelope := &mockEnvelope{
			event:     event,
			eventID:   "test-event-multi",
			timestamp: time.Now(),
			metadata:  make(map[string]interface{}),
		}

		// Handle the event
		if err := handler.Handle(ctx, envelope); err != nil {
			t.Fatalf("Unexpected error handling %s event: %v", eventData.eventType, err)
		}
	}

	// Verify all events were persisted to file system
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(eventLogPath, today, "events.log")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("Event log file should exist at %s", logFile)
	}

	// Read and verify log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read event log file: %v", err)
	}

	// Count the number of event entries (each on a separate line)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(events) {
		t.Errorf("Expected %d event entries, got %d", len(events), len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var eventEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &eventEntry); err != nil {
			t.Errorf("Failed to parse event log entry %d: %v", i, err)
		}
	}
}

func TestResourceEventHandler_ConfigurationMethods(t *testing.T) {
	repo := newMockRepository()
	handler := NewResourceEventHandler(repo)

	// Test default configuration
	if !handler.IsFilePersistenceEnabled() {
		t.Error("File persistence should be enabled by default")
	}
	if handler.GetEventLogPath() != "pod-data/events" {
		t.Errorf("Expected default event log path 'pod-data/events', got %s", handler.GetEventLogPath())
	}

	// Test configuration setters
	handler.SetFilePersistenceEnabled(false)
	if handler.IsFilePersistenceEnabled() {
		t.Error("File persistence should be disabled after setting to false")
	}

	newPath := "/custom/event/path"
	handler.SetEventLogPath(newPath)
	if handler.GetEventLogPath() != newPath {
		t.Errorf("Expected event log path '%s', got %s", newPath, handler.GetEventLogPath())
	}
}

func TestResourceEventHandler_EventWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	eventLogPath := filepath.Join(tempDir, "events")

	repo := newMockRepository()
	handler := NewResourceEventHandlerWithConfig(repo, eventLogPath, true)
	ctx := context.Background()

	// Test complete workflow: create -> update -> delete
	resourceID := "workflow-test"

	// 1. Create resource
	createdEvent := domain.NewResourceCreatedEvent(resourceID, map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        100,
		"createdAt":   time.Now(),
	})

	envelope1 := &mockEnvelope{
		event:     createdEvent,
		eventID:   "workflow-create",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	if err := handler.Handle(ctx, envelope1); err != nil {
		t.Fatalf("Failed to handle create event: %v", err)
	}

	// Verify resource exists
	exists, err := repo.Exists(ctx, resourceID)
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should exist after creation")
	}

	// 2. Update resource
	updatedEvent := domain.NewResourceUpdatedEvent(resourceID, map[string]interface{}{
		"contentType": "text/turtle",
		"size":        150,
		"updatedAt":   time.Now(),
	})

	envelope2 := &mockEnvelope{
		event:     updatedEvent,
		eventID:   "workflow-update",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	if err := handler.Handle(ctx, envelope2); err != nil {
		t.Fatalf("Failed to handle update event: %v", err)
	}

	// Verify resource still exists
	exists, err = repo.Exists(ctx, resourceID)
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if !exists {
		t.Error("Resource should still exist after update")
	}

	// 3. Delete resource
	deletedEvent := domain.NewResourceDeletedEvent(resourceID, map[string]interface{}{
		"deletedAt": time.Now(),
	})

	envelope3 := &mockEnvelope{
		event:     deletedEvent,
		eventID:   "workflow-delete",
		timestamp: time.Now(),
		metadata:  make(map[string]interface{}),
	}

	if err := handler.Handle(ctx, envelope3); err != nil {
		t.Fatalf("Failed to handle delete event: %v", err)
	}

	// Verify resource no longer exists
	exists, err = repo.Exists(ctx, resourceID)
	if err != nil {
		t.Fatalf("Failed to check resource existence: %v", err)
	}
	if exists {
		t.Error("Resource should not exist after deletion")
	}

	// Verify all events were persisted
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(eventLogPath, today, "events.log")

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read event log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 event entries for workflow, got %d", len(lines))
	}
}
