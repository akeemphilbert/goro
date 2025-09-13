package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceCreatedEvent(t *testing.T) {
	// Arrange
	resourceID := "test-resource-123"
	data := map[string]interface{}{
		"contentType": "text/turtle",
		"size":        1024,
	}

	// Act
	event := NewResourceCreatedEvent(resourceID, data)

	// Assert
	require.NotNil(t, event)
	assert.Equal(t, resourceID, event.AggregateID())
	assert.Equal(t, "resource", event.EntityType)
	assert.Equal(t, EventTypeResourceCreated, event.Type)
	assert.Equal(t, "resource.created", event.EventType())
	assert.WithinDuration(t, time.Now(), event.CreatedAt(), time.Second)
}

func TestNewResourceUpdatedEvent(t *testing.T) {
	// Arrange
	resourceID := "test-resource-456"
	data := map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        2048,
		"updated":     true,
	}

	// Act
	event := NewResourceUpdatedEvent(resourceID, data)

	// Assert
	require.NotNil(t, event)
	assert.Equal(t, resourceID, event.AggregateID())
	assert.Equal(t, "resource", event.EntityType)
	assert.Equal(t, EventTypeResourceUpdated, event.Type)
	assert.Equal(t, "resource.updated", event.EventType())
	assert.WithinDuration(t, time.Now(), event.CreatedAt(), time.Second)
}

func TestNewResourceDeletedEvent(t *testing.T) {
	// Arrange
	resourceID := "test-resource-789"
	data := map[string]interface{}{
		"reason": "user_requested",
	}

	// Act
	event := NewResourceDeletedEvent(resourceID, data)

	// Assert
	require.NotNil(t, event)
	assert.Equal(t, resourceID, event.AggregateID())
	assert.Equal(t, "resource", event.EntityType)
	assert.Equal(t, EventTypeResourceDeleted, event.Type)
	assert.Equal(t, "resource.deleted", event.EventType())
	assert.WithinDuration(t, time.Now(), event.CreatedAt(), time.Second)
}

func TestEventConstants(t *testing.T) {
	// Test that event type constants are properly defined
	assert.Equal(t, "created", EventTypeResourceCreated)
	assert.Equal(t, "updated", EventTypeResourceUpdated)
	assert.Equal(t, "deleted", EventTypeResourceDeleted)
}

func TestEventDataIntegrity(t *testing.T) {
	// Test that event data is properly preserved through JSON serialization
	resourceID := "integrity-test"
	originalData := map[string]interface{}{
		"string":  "test",
		"number":  42,
		"boolean": true,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	event := NewResourceCreatedEvent(resourceID, originalData)

	// Verify event structure
	require.NotNil(t, event)
	assert.Equal(t, resourceID, event.AggregateID())
	assert.Equal(t, "resource", event.EntityType)
	assert.Equal(t, EventTypeResourceCreated, event.Type)

	// Verify payload is not empty (data is JSON serialized)
	assert.NotEmpty(t, event.Payload())
}

func TestEventMetadata(t *testing.T) {
	// Test that events support metadata
	resourceID := "metadata-test"
	data := map[string]interface{}{"test": "data"}

	event := NewResourceCreatedEvent(resourceID, data)

	// Add metadata
	event.SetMetadata("source", "test")
	event.SetMetadata("version", "1.0")

	// Verify metadata
	assert.Equal(t, "test", event.GetMetadata("source"))
	assert.Equal(t, "1.0", event.GetMetadata("version"))
	assert.Nil(t, event.GetMetadata("nonexistent"))
}

func TestEventSequenceNumber(t *testing.T) {
	// Test that events support sequence numbers
	resourceID := "sequence-test"
	data := map[string]interface{}{"test": "data"}

	event := NewResourceCreatedEvent(resourceID, data)

	// Initially should be 0
	assert.Equal(t, int64(0), event.SequenceNo())

	// Set sequence number
	event.SetSequenceNo(42)
	assert.Equal(t, int64(42), event.SequenceNo())
}
