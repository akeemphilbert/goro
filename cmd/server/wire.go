//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/akeemphilbert/goro/internal/conf"
	httpServer "github.com/akeemphilbert/goro/internal/infrastructure/transport/http"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
)

// wireApp init kratos application.
func wireApp(*conf.Server, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(ProviderSet, newAppWithCleanup))
}

// newAppWithCleanup creates both the app and cleanup function
func newAppWithCleanup(logger log.Logger, hs *http.Server, gs *grpc.Server, config *conf.Server, initService *application.InitializationService) (*kratos.App, func()) {
	// Initialize the system (create root container, etc.)
	ctx := context.Background()
	if err := initService.Initialize(ctx); err != nil {
		log.Errorf("Failed to initialize system: %v", err)
		panic(err)
	}

	app := newApp(logger, hs, gs, config)
	cleanup := func() {
		// Simple cleanup function for testing
		// In production, this would handle resource cleanup
	}
	return app, cleanup
}

// ProviderSet is the provider set for Wire dependency injection
var ProviderSet = wire.NewSet(
	handlers.NewHealthHandler,
	handlers.NewRequestResponseHandler,
	handlers.ProviderSet,
	application.ProviderSet,
	infrastructure.InfrastructureSet,

	// User management providers (basic set for now)
	// userInfrastructure.UserManagementProviderSet,
	// userApplication.UserApplicationProviderSet,

	NewGRPCServer,
	NewHTTPServerProvider,
	wire.FieldsOf(new(*conf.Server), "HTTP", "GRPC", "Container"),
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

// NewHTTPServerProvider creates an HTTP server with all handlers properly wired
func NewHTTPServerProvider(
	c *conf.HTTP,
	logger log.Logger,
	healthHandler *handlers.HealthHandler,
	requestResponseHandler *handlers.RequestResponseHandler,
	resourceHandler *handlers.ResourceHandler,
	containerHandler *handlers.ContainerHandler,
	// userHandler *handlers.UserHandler,
	// accountHandler *handlers.AccountHandler,
) *http.Server {
	return httpServer.NewHTTPServer(c, logger, healthHandler, requestResponseHandler, resourceHandler, containerHandler, nil, nil)
}
