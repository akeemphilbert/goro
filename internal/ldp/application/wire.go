package application

import (
	"fmt"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
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
	rdfConverter *infrastructure.ContainerRDFConverter,
) (*ContainerService, error) {
	// Validate dependencies
	if containerRepo == nil {
		return nil, fmt.Errorf("container repository cannot be nil")
	}
	if unitOfWorkFactory == nil {
		return nil, fmt.Errorf("unit of work factory cannot be nil")
	}
	if eventDispatcher == nil {
		return nil, fmt.Errorf("event dispatcher cannot be nil")
	}
	if rdfConverter == nil {
		return nil, fmt.Errorf("RDF converter cannot be nil")
	}

	// Create the container service
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

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
