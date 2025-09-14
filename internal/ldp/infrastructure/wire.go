package infrastructure

import (
	"fmt"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
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
	NewFileSystemContainerRepositoryProvider,
	NewRDFConverter,
	NewContainerRDFConverter,
	NewUnitOfWorkFactory,
	// Bind interfaces to implementations
	wire.Bind(new(domain.FormatConverter), new(*RDFConverter)),
	wire.Bind(new(pericarpdomain.EventStore), new(*pericarpinfra.GormEventStore)),
)

// OptimizedInfrastructureSet provides optimized infrastructure dependencies with caching and indexing
var OptimizedInfrastructureSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	NewEventDispatcher,
	NewOptimizedFileSystemRepositoryProvider,
	NewFileSystemContainerRepositoryProvider,
	NewRDFConverter,
	NewContainerRDFConverter,
	NewUnitOfWorkFactory,
	// Bind interfaces to implementations
	wire.Bind(new(domain.FormatConverter), new(*RDFConverter)),
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

// NewFileSystemContainerRepositoryProvider provides a FileSystemContainerRepository for Wire dependency injection
func NewFileSystemContainerRepositoryProvider(config *conf.Container) (domain.ContainerRepository, error) {
	// Set defaults if config is nil
	if config == nil {
		config = &conf.Container{}
		config.SetDefaults()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid container configuration: %w", err)
	}

	// Create membership indexer
	indexer, err := MembershipIndexerProvider(config.IndexPath)
	if err != nil {
		return nil, err
	}

	return NewFileSystemContainerRepository(config.StoragePath, indexer)
}
