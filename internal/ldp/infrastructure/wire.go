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
