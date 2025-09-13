package handlers

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// RequestResponseHandler demonstrates request/response processing with Kratos patterns
type RequestResponseHandler struct {
	logger log.Logger
}

// NewRequestResponseHandler creates a new request/response handler
func NewRequestResponseHandler(logger log.Logger) *RequestResponseHandler {
	return &RequestResponseHandler{
		logger: logger,
	}
}

// GetWithPathParams demonstrates path parameter extraction using ctx.Vars()
func (h *RequestResponseHandler) GetWithPathParams(ctx http.Context) error {
	vars := ctx.Vars()

	response := make(map[string]interface{})

	// Extract path parameters
	for key, values := range vars {
		if len(values) > 0 {
			response[key] = values[0]
		}
	}

	return ctx.JSON(200, response)
}

// GetWithQueryParams demonstrates query parameter handling using ctx.Query()
func (h *RequestResponseHandler) GetWithQueryParams(ctx http.Context) error {
	// Extract query parameters from request
	query := ctx.Request().URL.Query()
	name := query.Get("name")
	age := query.Get("age")
	active := query.Get("active")

	response := map[string]interface{}{
		"name":   name,
		"age":    age,
		"active": active,
	}

	return ctx.JSON(200, response)
}

// GetJSONResponse demonstrates JSON response handling using ctx.JSON()
func (h *RequestResponseHandler) GetJSONResponse(ctx http.Context) error {
	response := map[string]interface{}{
		"message": "success",
		"data": map[string]interface{}{
			"id":   1,
			"name": "example",
		},
	}

	return ctx.JSON(200, response)
}
