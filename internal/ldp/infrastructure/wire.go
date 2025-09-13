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
