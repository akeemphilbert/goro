//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/google/wire"
)

// EventInfrastructureSet provides all event-related dependencies
var EventInfrastructureSet = wire.NewSet(
	NewEventDispatcher,
)

// RDFInfrastructureSet provides all RDF-related dependencies
var RDFInfrastructureSet = wire.NewSet(
	NewRDFConverter,
)

// StorageInfrastructureSet provides all storage-related dependencies
var StorageInfrastructureSet = wire.NewSet(
	NewFileSystemRepositoryProvider,
)

// InfrastructureSet provides all infrastructure dependencies
var InfrastructureSet = wire.NewSet(
	EventInfrastructureSet,
	RDFInfrastructureSet,
	StorageInfrastructureSet,
)
