package application

import (
	"fmt"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the application layer
var ProviderSet = wire.NewSet(
	NewStorageService,
	wire.Bind(new(FormatConverter), new(*infrastructure.RDFConverter)),
)

// NewStorageServiceProvider creates a StorageService with all dependencies and registers event handlers
func NewStorageServiceProvider(
	repo domain.ResourceRepository,
	converter FormatConverter,
	eventStore pericarpdomain.EventStore,
	eventDispatcher pericarpdomain.EventDispatcher,
) (*StorageService, error) {
	// Create a factory that creates new UnitOfWork instances
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create the storage service
	service := NewStorageService(repo, converter, unitOfWorkFactory)

	// Register event handlers to update repository after events are committed
	registrar := NewEventHandlerRegistrar(eventDispatcher)
	if err := registrar.RegisterAllHandlers(repo); err != nil {
		return nil, fmt.Errorf("failed to register event handlers: %w", err)
	}

	return service, nil
}

// NewEventHandlerRegistrarProvider creates an event handler registrar
func NewEventHandlerRegistrarProvider(eventDispatcher pericarpdomain.EventDispatcher) *EventHandlerRegistrar {
	return NewEventHandlerRegistrar(eventDispatcher)
}
