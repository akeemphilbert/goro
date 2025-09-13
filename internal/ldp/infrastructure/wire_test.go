package infrastructure

import (
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventDispatcher_WireProvider(t *testing.T) {
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

func TestEventInfrastructureSetIntegration(t *testing.T) {
	// This test verifies that the Wire set is properly configured
	// In a real Wire-generated application, this would be tested through
	// the generated wire_gen.go file, but we can test the provider directly

	// Act
	dispatcher, err := NewEventDispatcher()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dispatcher)

	// Cleanup - cast to concrete type to access Close method
	if closer, ok := dispatcher.(*infrastructure.WatermillEventDispatcher); ok {
		err = closer.Close()
		assert.NoError(t, err)
	}
}
