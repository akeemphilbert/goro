//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"

	"github.com/akeemphilbert/goro/internal/conf"
)

// wireApp init kratos application.
func wireApp(*conf.Server, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(ProviderSet, newApp))
}

// ProviderSet is the provider set for Wire dependency injection
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	wire.FieldsOf(new(*conf.Server), "HTTP", "GRPC"),
)

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(c *conf.HTTP, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Network(c.Network),
		http.Address(c.Addr),
	}
	if c.Timeout != 0 {
		opts = append(opts, http.Timeout(c.Timeout))
	}

	srv := http.NewServer(opts...)
	return srv
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(c *conf.GRPC, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Network(c.Network),
		grpc.Address(c.Addr),
	}
	if c.Timeout != 0 {
		opts = append(opts, grpc.Timeout(c.Timeout))
	}

	srv := grpc.NewServer(opts...)
	return srv
}
