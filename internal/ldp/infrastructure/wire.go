package infrastructure

import (
	"github.com/google/wire"
)

// InfrastructureSet provides all infrastructure dependencies
var InfrastructureSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	NewEventDispatcher,
	NewFileSystemRepositoryProvider,
	NewRDFConverter,
)

// OptimizedInfrastructureSet provides optimized infrastructure dependencies with caching and indexing
var OptimizedInfrastructureSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	NewEventDispatcher,
	NewOptimizedFileSystemRepositoryProvider,
	NewRDFConverter,
)
