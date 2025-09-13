package infrastructure

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventInfrastructureUsageExample(t *testing.T) {
	// This test demonstrates how to use the event infrastructure

	// 1. Create the event dispatcher
	dispatcher, err := NewEventDispatcher()
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
			closer.Close()
		}
	}()

	// 2. Create domain events
	createdEvent := domain.NewResourceCreatedEvent("resource-123", map[string]interface{}{
		"contentType": "text/turtle",
		"size":        1024,
	})

	updatedEvent := domain.NewResourceUpdatedEvent("resource-123", map[string]interface{}{
		"contentType": "application/ld+json",
		"size":        2048,
	})

	deletedEvent := domain.NewResourceDeletedEvent("resource-123", map[string]interface{}{
		"reason": "user_requested",
	})

	// 3. Verify events are properly structured
	assert.Equal(t, "resource-123", createdEvent.AggregateID())
	assert.Equal(t, "resource.created", createdEvent.EventType())
	assert.Equal(t, "resource", createdEvent.EntityType)
	assert.Equal(t, "created", createdEvent.Type)

	assert.Equal(t, "resource-123", updatedEvent.AggregateID())
	assert.Equal(t, "resource.updated", updatedEvent.EventType())
	assert.Equal(t, "resource", updatedEvent.EntityType)
	assert.Equal(t, "updated", updatedEvent.Type)

	assert.Equal(t, "resource-123", deletedEvent.AggregateID())
	assert.Equal(t, "resource.deleted", deletedEvent.EventType())
	assert.Equal(t, "resource", deletedEvent.EntityType)
	assert.Equal(t, "deleted", deletedEvent.Type)

	// 4. Verify event constants
	assert.Equal(t, "created", domain.EventTypeResourceCreated)
	assert.Equal(t, "updated", domain.EventTypeResourceUpdated)
	assert.Equal(t, "deleted", domain.EventTypeResourceDeleted)

	// Note: To actually dispatch events, you would need to create proper envelopes
	// and use the Dispatch method. This example focuses on the setup and event creation.

	t.Log("Event infrastructure setup completed successfully")
	t.Logf("Created event: %s for aggregate: %s", createdEvent.EventType(), createdEvent.AggregateID())
	t.Logf("Updated event: %s for aggregate: %s", updatedEvent.EventType(), updatedEvent.AggregateID())
	t.Logf("Deleted event: %s for aggregate: %s", deletedEvent.EventType(), deletedEvent.AggregateID())
}
