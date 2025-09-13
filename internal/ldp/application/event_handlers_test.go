package application

import (
	"context"
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
