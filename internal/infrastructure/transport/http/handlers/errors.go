package handlers

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Domain errors using Kratos errors package
var (
	ErrResourceNotFound = errors.NotFound("RESOURCE_NOT_FOUND", "resource not found")
	ErrInvalidRequest   = errors.BadRequest("INVALID_REQUEST", "invalid request parameters")
	ErrInternalServer   = errors.InternalServer("INTERNAL_SERVER_ERROR", "internal server error")
	ErrMethodNotAllowed = errors.New(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
)

// ErrorHandler handles HTTP errors with proper status codes and responses
type ErrorHandler struct {
	logger log.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger log.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleError handles errors and returns appropriate HTTP responses
func (h *ErrorHandler) HandleError(ctx khttp.Context, err error) error {
	// Log the error
	h.logger.Log(log.LevelError, "msg", "HTTP error occurred", "error", err.Error())

	// Check if it's a Kratos error
	if kratosErr := errors.FromError(err); kratosErr != nil {
		return ctx.JSON(int(kratosErr.Code), map[string]interface{}{
			"error":   kratosErr.Reason,
			"message": kratosErr.Message,
		})
	}

	// Handle generic errors as internal server error
	return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error":   "INTERNAL_SERVER_ERROR",
		"message": "internal server error",
	})
}
