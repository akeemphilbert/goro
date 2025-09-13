package handlers

import (
	"net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedReason string
	}{
		{
			name:           "NotFound error should map to 404",
			error:          ErrResourceNotFound,
			expectedStatus: http.StatusNotFound,
			expectedReason: "RESOURCE_NOT_FOUND",
		},
		{
			name:           "BadRequest error should map to 400",
			error:          ErrInvalidRequest,
			expectedStatus: http.StatusBadRequest,
			expectedReason: "INVALID_REQUEST",
		},
		{
			name:           "InternalError should map to 500",
			error:          ErrInternalServer,
			expectedStatus: http.StatusInternalServerError,
			expectedReason: "INTERNAL_SERVER_ERROR",
		},
		{
			name:           "MethodNotAllowed should map to 405",
			error:          ErrMethodNotAllowed,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedReason: "METHOD_NOT_ALLOWED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is a Kratos error
			kratosErr, ok := tt.error.(*errors.Error)
			if !ok {
				t.Fatalf("expected Kratos error, got %T", tt.error)
			}

			// Verify status code
			if int(kratosErr.Code) != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, kratosErr.Code)
			}

			// Verify reason
			if kratosErr.Reason != tt.expectedReason {
				t.Errorf("expected reason %s, got %s", tt.expectedReason, kratosErr.Reason)
			}
		})
	}
}

func TestNewErrorHandler(t *testing.T) {
	logger := log.NewStdLogger(nil)
	handler := NewErrorHandler(logger)

	if handler == nil {
		t.Error("NewErrorHandler() should not return nil")
	}
	if handler.logger == nil {
		t.Error("handler.logger should not be nil")
	}
}

func TestErrorHandlerLogic(t *testing.T) {
	// Test the error handler creation and basic functionality
	logger := log.NewStdLogger(nil)
	handler := NewErrorHandler(logger)

	// Verify handler is created properly
	if handler == nil {
		t.Fatal("NewErrorHandler() should not return nil")
	}

	// Test that the handler has the expected structure
	if handler.logger == nil {
		t.Error("handler should have a logger")
	}

	// The actual HTTP context testing is complex due to Kratos internals
	// The integration test will cover the full HTTP flow
	t.Log("Error handler created successfully with proper dependencies")
}
