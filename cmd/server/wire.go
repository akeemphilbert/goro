//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/akeemphilbert/goro/internal/conf"
	httpServer "github.com/akeemphilbert/goro/internal/infrastructure/transport/http"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
)

// wireApp init kratos application.
func wireApp(*conf.Server, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(ProviderSet, newAppWithCleanup))
}

// newAppWithCleanup creates both the app and cleanup function
func newAppWithCleanup(logger log.Logger, hs *http.Server, gs *grpc.Server, config *conf.Server) (*kratos.App, func()) {
	app := newApp(logger, hs, gs, config)
	cleanup := func() {
		log.Info("Executing resource cleanup...")
		// Add any specific cleanup logic here
		// For now, just log that cleanup was called
		log.Info("Resource cleanup completed")
	}
	return app, cleanup
}

// ProviderSet is the provider set for Wire dependency injection
var ProviderSet = wire.NewSet(
	httpServer.NewHTTPServer,
	handlers.NewHealthHandler,
	handlers.NewRequestResponseHandler,
	NewGRPCServer,
	wire.FieldsOf(new(*conf.Server), "HTTP", "GRPC"),
)

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(c *conf.GRPC, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Network(c.Network),
		grpc.Address(c.Addr),
	}
	if c.Timeout != 0 {
		opts = append(opts, grpc.Timeout(time.Duration(c.Timeout)))
	}

	srv := grpc.NewServer(opts...)
	return srv
}
