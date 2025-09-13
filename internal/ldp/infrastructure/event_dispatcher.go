package infrastructure

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
)

// NewEventDispatcher creates a new WatermillEventDispatcher instance using pericarp
func NewEventDispatcher() (domain.EventDispatcher, error) {
	// Create pericarp's WatermillEventDispatcher with a no-op logger
	dispatcher, err := infrastructure.NewWatermillEventDispatcher(watermill.NopLogger{})
	if err != nil {
		return nil, err
	}

	return dispatcher, nil
}
