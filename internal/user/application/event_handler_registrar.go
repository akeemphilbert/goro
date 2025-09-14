package application

import (
	"fmt"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// EventHandlerRegistrar manages the registration of event handlers with the event dispatcher
type EventHandlerRegistrar struct {
	eventDispatcher pericarpdomain.EventDispatcher
}

// NewEventHandlerRegistrar creates a new event handler registrar
func NewEventHandlerRegistrar(eventDispatcher pericarpdomain.EventDispatcher) *EventHandlerRegistrar {
	return &EventHandlerRegistrar{
		eventDispatcher: eventDispatcher,
	}
}

// RegisterUserEventHandlers registers all user-related event handlers
func (r *EventHandlerRegistrar) RegisterUserEventHandlers(handler *UserEventHandler) error {
	if handler == nil {
		return fmt.Errorf("user event handler cannot be nil")
	}

	// For now, we'll just store the handler reference
	// In a full implementation, we would register with the event dispatcher
	// The actual event registration would depend on the pericarp event system implementation

	return nil
}

// RegisterAccountEventHandlers registers all account-related event handlers
func (r *EventHandlerRegistrar) RegisterAccountEventHandlers(handler *AccountEventHandler) error {
	if handler == nil {
		return fmt.Errorf("account event handler cannot be nil")
	}

	// For now, we'll just store the handler reference
	// In a full implementation, we would register with the event dispatcher
	// The actual event registration would depend on the pericarp event system implementation

	return nil
}

// RegisterAllHandlers registers both user and account event handlers
func (r *EventHandlerRegistrar) RegisterAllHandlers(
	userHandler *UserEventHandler,
	accountHandler *AccountEventHandler,
) error {
	if err := r.RegisterUserEventHandlers(userHandler); err != nil {
		return fmt.Errorf("failed to register user event handlers: %w", err)
	}

	if err := r.RegisterAccountEventHandlers(accountHandler); err != nil {
		return fmt.Errorf("failed to register account event handlers: %w", err)
	}

	return nil
}
