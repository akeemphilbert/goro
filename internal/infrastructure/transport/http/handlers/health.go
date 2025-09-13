package handlers

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger log.Logger
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(logger log.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// Check handles health check requests and returns server status
func (h *HealthHandler) Check(ctx http.Context) error {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}

	return ctx.JSON(200, response)
}

// Status handles status requests with error handling demonstration
func (h *HealthHandler) Status(ctx http.Context) error {
	// Example of using path parameters with error handling
	vars := ctx.Vars()
	idSlice, exists := vars["id"]
	if !exists || len(idSlice) == 0 || idSlice[0] == "" {
		return ErrInvalidRequest
	}

	id := idSlice[0]

	// Example of resource not found error
	if id == "notfound" {
		return ErrResourceNotFound
	}

	response := map[string]interface{}{
		"id":     id,
		"status": "running",
	}

	return ctx.JSON(200, response)
}
