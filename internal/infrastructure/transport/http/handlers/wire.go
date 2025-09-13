package handlers

import (
	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for HTTP handlers
var ProviderSet = wire.NewSet(
	NewResourceHandler,
	wire.Bind(new(StorageServiceInterface), new(*application.StorageService)),
)

// NewResourceHandlerProvider creates a ResourceHandler with proper dependency injection
func NewResourceHandlerProvider(storageService *application.StorageService, logger log.Logger) *ResourceHandler {
	return NewResourceHandler(storageService, logger)
}
