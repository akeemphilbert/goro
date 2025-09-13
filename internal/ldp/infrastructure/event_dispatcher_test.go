package infrastructure

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventDispatcher(t *testing.T) {
	// Act
	dispatcher, err := NewEventDispatcher()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dispatcher)

	// Verify it implements the domain interface
	var _ domain.EventDispatcher = dispatcher

	// Cleanup - cast to concrete type to access Close method
	if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
		err = closer.Close()
		assert.NoError(t, err)
	}
}

func TestEventDispatcherIntegration(t *testing.T) {
	// Arrange
	dispatcher, err := NewEventDispatcher()
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
			closer.Close()
		}
	}()

	// Create a test event
	event := domain.NewResourceCreatedEvent("test-resource", map[string]interface{}{
		"contentType": "text/turtle",
		"size":        1024,
	})

	// Assert - Just verify the dispatcher was created successfully and event was created
	assert.NotNil(t, dispatcher)
	assert.NotNil(t, event)
	assert.Equal(t, "test-resource", event.AggregateID())
	assert.Equal(t, "resource.created", event.EventType())
}

func TestEventDispatcherClose(t *testing.T) {
	// Arrange
	dispatcher, err := NewEventDispatcher()
	require.NoError(t, err)

	// Act - cast to concrete type to access Close method
	closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher)
	require.True(t, ok, "Expected WatermillEventDispatcher")

	err = closer.Close()

	// Assert
	assert.NoError(t, err)
}
