package main

import (
	"fmt"

	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/go-kratos/kratos/v2/log"
)

// NewResourceHandlerWithDependencies creates a ResourceHandler with all its dependencies
func NewResourceHandlerWithDependencies(logger log.Logger) (*handlers.ResourceHandler, error) {
	// Create database
	db, err := infrastructure.DatabaseProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Create event store
	eventStore, err := infrastructure.EventStoreProvider(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create event store: %w", err)
	}

	// Create event dispatcher
	eventDispatcher, err := infrastructure.NewEventDispatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create event dispatcher: %w", err)
	}

	// Create unit of work factory
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create repository
	repo, err := infrastructure.NewFileSystemRepositoryProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create converter
	converter := infrastructure.NewRDFConverter()

	// Create storage service
	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)

	// Register event handlers
	registrar := application.NewEventHandlerRegistrar(eventDispatcher)
	if err := registrar.RegisterAllHandlers(repo); err != nil {
		return nil, fmt.Errorf("failed to register event handlers: %w", err)
	}

	// Create resource handler
	resourceHandler := handlers.NewResourceHandler(storageService, logger)

	return resourceHandler, nil
}
