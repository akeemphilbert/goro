package infrastructure

import (
	"github.com/akeemphilbert/goro/internal/ldp/application"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/google/wire"
)

// InfrastructureSet provides all infrastructure dependencies
var InfrastructureSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	NewEventDispatcher,
	NewOptimizedFileSystemRepositoryProvider,
	NewRDFConverter,
	NewUnitOfWorkFactory,
	// Bind interfaces to implementations
	wire.Bind(new(application.FormatConverter), new(*RDFConverter)),
	wire.Bind(new(pericarpdomain.EventStore), new(*pericarpinfra.GormEventStore)),
)

// OptimizedInfrastructureSet provides optimized infrastructure dependencies with caching and indexing
var OptimizedInfrastructureSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	NewEventDispatcher,
	NewOptimizedFileSystemRepositoryProvider,
	NewRDFConverter,
	NewUnitOfWorkFactory,
	// Bind interfaces to implementations
	wire.Bind(new(application.FormatConverter), new(*RDFConverter)),
	wire.Bind(new(pericarpdomain.EventStore), new(*pericarpinfra.GormEventStore)),
)

// NewUnitOfWorkFactory creates a factory function for creating UnitOfWork instances
func NewUnitOfWorkFactory(
	eventStore pericarpdomain.EventStore,
	eventDispatcher pericarpdomain.EventDispatcher,
) func() pericarpdomain.UnitOfWork {
	return func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}
}
