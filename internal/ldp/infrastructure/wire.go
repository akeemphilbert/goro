//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/google/wire"
)

// DatabaseSet provides database-related dependencies
var DatabaseSet = wire.NewSet(
	DatabaseProvider,
	EventStoreProvider,
	wire.Bind(new(pericarpdomain.EventStore), new(*infrastructure.GormEventStore)),
)

// EventInfrastructureSet provides all event-related dependencies
var EventInfrastructureSet = wire.NewSet(
	NewEventDispatcher,
	infrastructure.UnitOfWorkProvider,
	wire.Bind(new(pericarpdomain.UnitOfWork), new(*infrastructure.UnitOfWorkImpl)),
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
	DatabaseSet,
	EventInfrastructureSet,
	RDFInfrastructureSet,
	StorageInfrastructureSet,
)
