package application

import (
	"fmt"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the application layer
var ProviderSet = wire.NewSet(
	NewStorageServiceProvider,
	NewContainerServiceProvider,
	NewEventHandlerRegistrarProvider,
)

// NewStorageServiceProvider creates a StorageService with all dependencies and registers event handlers
func NewStorageServiceProvider(
	repo domain.StreamingResourceRepository,
	converter domain.FormatConverter,
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	eventDispatcher pericarpdomain.EventDispatcher,
) (*StorageService, error) {
	// Create the storage service
	service := NewStorageService(repo, converter, unitOfWorkFactory)

	// Register event handlers to update repository after events are committed
	registrar := NewEventHandlerRegistrar(eventDispatcher)
	if err := registrar.RegisterAllHandlers(repo); err != nil {
		return nil, fmt.Errorf("failed to register event handlers: %w", err)
	}

	return service, nil
}

// NewContainerServiceProvider creates a ContainerService with all dependencies and registers event handlers
func NewContainerServiceProvider(
	containerRepo domain.ContainerRepository,
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	eventDispatcher pericarpdomain.EventDispatcher,
) (*ContainerService, error) {
	// Create the container service
	service := NewContainerService(containerRepo, unitOfWorkFactory)

	// Register container event handlers to update repository after events are committed
	registrar := NewEventHandlerRegistrar(eventDispatcher)
	if err := registrar.RegisterContainerEventHandler(NewContainerEventHandler(containerRepo)); err != nil {
		return nil, fmt.Errorf("failed to register container event handlers: %w", err)
	}

	return service, nil
}

// NewEventHandlerRegistrarProvider creates an event handler registrar
func NewEventHandlerRegistrarProvider(eventDispatcher pericarpdomain.EventDispatcher) *EventHandlerRegistrar {
	return NewEventHandlerRegistrar(eventDispatcher)
}
