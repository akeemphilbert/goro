package application

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// testEnvelope is a simple implementation of Envelope for testing
type testEnvelope struct {
	event     pericarpdomain.Event
	timestamp time.Time
	metadata  map[string]interface{}
	eventID   string
}

func (e *testEnvelope) Event() pericarpdomain.Event {
	return e.event
}

func (e *testEnvelope) Metadata() map[string]interface{} {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	return e.metadata
}

func (e *testEnvelope) EventID() string {
	if e.eventID == "" {
		e.eventID = "test-event-id"
	}
	return e.eventID
}

func (e *testEnvelope) Timestamp() time.Time {
	return e.timestamp
}

func TestContainerEventHandler_HandleContainerCreated(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)
	handler.SetFilePersistenceEnabled(false) // Disable file persistence for testing

	ctx := context.Background()
	containerID := "test-container-id"
	parentID := "parent-container-id"
	containerType := domain.BasicContainer

	// Create event payload
	payload := map[string]interface{}{
		"parentID":      parentID,
		"containerType": containerType.String(),
		"createdAt":     time.Now(),
	}

	// Create container created event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerCreated, containerID, "", "", payload)

	// Create a simple envelope implementation for testing
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Setup mock expectations
	mockRepo.On("CreateContainer", ctx, mock.MatchedBy(func(container *domain.Container) bool {
		return container.ID() == containerID &&
			container.ParentID == parentID &&
			container.ContainerType == containerType
	})).Return(nil)

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerEventHandler_HandleContainerUpdated(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)
	handler.SetFilePersistenceEnabled(false) // Disable file persistence for testing

	ctx := context.Background()
	containerID := "test-container-id"
	newTitle := "Updated Container Title"

	// Create existing container
	existingContainer := domain.NewContainer(containerID, "", domain.BasicContainer)
	existingContainer.MarkEventsAsCommitted() // Clear creation events

	// Create event payload
	payload := map[string]interface{}{
		"title":     newTitle,
		"updatedAt": time.Now(),
	}

	// Create container updated event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerUpdated, containerID, "", "", payload)
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Setup mock expectations
	mockRepo.On("GetContainer", ctx, containerID).Return(existingContainer, nil)
	mockRepo.On("UpdateContainer", ctx, mock.MatchedBy(func(container *domain.Container) bool {
		return container.ID() == containerID && container.GetTitle() == newTitle
	})).Return(nil)

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerEventHandler_HandleContainerDeleted(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)
	handler.SetFilePersistenceEnabled(false) // Disable file persistence for testing

	ctx := context.Background()
	containerID := "test-container-id"

	// Create event payload
	payload := map[string]interface{}{
		"deletedAt": time.Now(),
	}

	// Create container deleted event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerDeleted, containerID, "", "", payload)
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Setup mock expectations
	mockRepo.On("DeleteContainer", ctx, containerID).Return(nil)

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerEventHandler_HandleMemberAdded(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)
	handler.SetFilePersistenceEnabled(false) // Disable file persistence for testing

	ctx := context.Background()
	containerID := "test-container-id"
	memberID := "test-member-id"

	// Create event payload
	payload := map[string]interface{}{
		"memberID":   memberID,
		"memberType": "Resource",
		"addedAt":    time.Now(),
	}

	// Create member added event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeMemberAdded, containerID, "", "", payload)
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Setup mock expectations
	mockRepo.On("AddMember", ctx, containerID, memberID).Return(nil)

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerEventHandler_HandleMemberRemoved(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)
	handler.SetFilePersistenceEnabled(false) // Disable file persistence for testing

	ctx := context.Background()
	containerID := "test-container-id"
	memberID := "test-member-id"

	// Create event payload
	payload := map[string]interface{}{
		"memberID":   memberID,
		"memberType": "Resource",
		"removedAt":  time.Now(),
	}

	// Create member removed event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeMemberRemoved, containerID, "", "", payload)
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Setup mock expectations
	mockRepo.On("RemoveMember", ctx, containerID, memberID).Return(nil)

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestContainerEventHandler_HandleNonContainerEvent(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	ctx := context.Background()

	// Create non-container event
	event := pericarpdomain.NewEntityEvent("resource", "created", "test-id", "", "", []byte("{}"))
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert - should handle gracefully without calling repository
	assert.NoError(t, err)
	mockRepo.AssertNotCalled(t, "CreateContainer")
	mockRepo.AssertNotCalled(t, "UpdateContainer")
	mockRepo.AssertNotCalled(t, "DeleteContainer")
}

func TestContainerEventHandler_HandleUnknownContainerEvent(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	ctx := context.Background()

	// Create unknown container event
	event := pericarpdomain.NewEntityEvent("container", "unknown_event", "test-id", "", "", []byte("{}"))
	envelope := &testEnvelope{event: event, timestamp: time.Now()}

	// Execute
	err := handler.Handle(ctx, envelope)

	// Assert - should handle gracefully without calling repository
	assert.NoError(t, err)
	mockRepo.AssertNotCalled(t, "CreateContainer")
	mockRepo.AssertNotCalled(t, "UpdateContainer")
	mockRepo.AssertNotCalled(t, "DeleteContainer")
}

func TestContainerEventHandler_EventTypes(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	// Execute
	eventTypes := handler.EventTypes()

	// Assert
	expectedTypes := []string{
		"container.created",
		"container.updated",
		"container.deleted",
		"container.member_added",
		"container.member_removed",
	}
	assert.Equal(t, expectedTypes, eventTypes)
}

func TestContainerEventHandler_ReconstructContainerFromEvent(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	containerID := "test-container-id"
	parentID := "parent-container-id"
	containerType := domain.BasicContainer
	title := "Test Container"
	description := "Test Description"

	// Create event payload with all metadata
	payload := map[string]interface{}{
		"parentID":      parentID,
		"containerType": containerType.String(),
		"title":         title,
		"description":   description,
		"createdAt":     time.Now(),
	}

	// Create container created event (NewEntityEvent will marshal the payload)
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerCreated, containerID, "", "", payload)

	// Execute
	container, err := handler.reconstructContainerFromEvent(event)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, container)
	assert.Equal(t, containerID, container.ID())
	assert.Equal(t, parentID, container.ParentID)
	assert.Equal(t, containerType, container.ContainerType)
	assert.Equal(t, title, container.GetTitle())
	assert.Equal(t, description, container.GetDescription())
}

func TestContainerEventHandler_ReconstructContainerFromEvent_InvalidPayload(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	containerID := "test-container-id"

	// Create event with invalid JSON payload
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerCreated, containerID, "", "", []byte("invalid json"))

	// Execute
	container, err := handler.reconstructContainerFromEvent(event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "failed to unmarshal event payload")
}

func TestContainerEventHandler_ReconstructContainerFromEvent_MissingContainerType(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	containerID := "test-container-id"

	// Create event payload without containerType
	payload := map[string]interface{}{
		"parentID":  "parent-id",
		"createdAt": time.Now(),
	}
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerCreated, containerID, "", "", payload)

	// Execute
	container, err := handler.reconstructContainerFromEvent(event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, container)
	assert.Contains(t, err.Error(), "missing or invalid containerType")
}

func TestContainerEventHandler_ApplyContainerUpdatesFromEvent(t *testing.T) {
	// Setup
	mockRepo := new(MockContainerRepository)
	handler := NewContainerEventHandler(mockRepo)

	// Create existing container
	container := domain.NewContainer("test-id", "", domain.BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	newTitle := "Updated Title"
	newDescription := "Updated Description"

	// Create event payload
	payload := map[string]interface{}{
		"title":       newTitle,
		"description": newDescription,
		"updatedAt":   time.Now(),
	}
	event := pericarpdomain.NewEntityEvent("container", domain.EventTypeContainerUpdated, "test-id", "", "", payload)

	// Execute
	err := handler.applyContainerUpdatesFromEvent(container, event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newTitle, container.GetTitle())
	assert.Equal(t, newDescription, container.GetDescription())
}
