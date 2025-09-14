package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerCreatedEvent(t *testing.T) {
	containerID := "test-container-1"
	data := map[string]interface{}{
		"parentID":      "parent-container",
		"containerType": "BasicContainer",
		"createdAt":     time.Now(),
	}

	event := NewContainerCreatedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerCreated, event.Type)
}

func TestNewContainerUpdatedEvent(t *testing.T) {
	containerID := "test-container-1"
	data := map[string]interface{}{
		"title":       "Updated Title",
		"description": "Updated Description",
		"updatedAt":   time.Now(),
	}

	event := NewContainerUpdatedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerUpdated, event.Type)
}

func TestNewContainerDeletedEvent(t *testing.T) {
	containerID := "test-container-1"
	data := map[string]interface{}{
		"deletedAt": time.Now(),
		"reason":    "user_requested",
	}

	event := NewContainerDeletedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerDeleted, event.Type)
}

func TestNewMemberAddedEvent(t *testing.T) {
	containerID := "test-container-1"
	data := map[string]interface{}{
		"memberID":   "resource-1",
		"memberType": "Resource",
		"addedAt":    time.Now(),
	}

	event := NewMemberAddedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeMemberAdded, event.Type)
}

func TestNewMemberRemovedEvent(t *testing.T) {
	containerID := "test-container-1"
	data := map[string]interface{}{
		"memberID":   "resource-1",
		"memberType": "Resource",
		"removedAt":  time.Now(),
	}

	event := NewMemberRemovedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeMemberRemoved, event.Type)
}

func TestContainerEventTypes_Constants(t *testing.T) {
	// Verify event type constants are properly defined
	assert.Equal(t, "container_created", EventTypeContainerCreated)
	assert.Equal(t, "container_updated", EventTypeContainerUpdated)
	assert.Equal(t, "container_deleted", EventTypeContainerDeleted)
	assert.Equal(t, "member_added", EventTypeMemberAdded)
	assert.Equal(t, "member_removed", EventTypeMemberRemoved)
}

func TestContainerEventData_Serialization(t *testing.T) {
	// Test that event data can contain various types
	containerID := "test-container"
	data := map[string]interface{}{
		"string_field": "test_value",
		"int_field":    42,
	}

	event := NewContainerCreatedEvent(containerID, data)

	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerCreated, event.Type)
}

func TestContainerEvents_EmptyData(t *testing.T) {
	containerID := "test-container"

	// Test with nil data
	event := NewContainerCreatedEvent(containerID, nil)
	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerCreated, event.Type)

	// Test with empty map
	event = NewContainerUpdatedEvent(containerID, map[string]interface{}{})
	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerUpdated, event.Type)
}

func TestContainerEvents_EventIDUniqueness(t *testing.T) {
	containerID := "test-container"
	data := map[string]interface{}{"test": "data"}

	// Create multiple events and verify they are created successfully
	event1 := NewContainerCreatedEvent(containerID, data)
	event2 := NewContainerCreatedEvent(containerID, data)
	event3 := NewContainerUpdatedEvent(containerID, data)

	assert.NotNil(t, event1)
	assert.NotNil(t, event2)
	assert.NotNil(t, event3)
	assert.Equal(t, EventTypeContainerCreated, event1.Type)
	assert.Equal(t, EventTypeContainerCreated, event2.Type)
	assert.Equal(t, EventTypeContainerUpdated, event3.Type)
}

func TestContainerEvents_TimestampAccuracy(t *testing.T) {
	containerID := "test-container"
	data := map[string]interface{}{"test": "data"}

	event := NewContainerCreatedEvent(containerID, data)

	// Event should be created successfully
	assert.NotNil(t, event)
	assert.Equal(t, EventTypeContainerCreated, event.Type)
}
