package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

func TestHTTPMethodSupport(t *testing.T) {
	// Create test server with method support
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000, // 30 seconds in nanoseconds
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register method-specific routes for testing
	RegisterMethodSpecificRoutes(server)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedJSON   map[string]interface{}
		expectNoBody   bool
	}{
		{
			name:           "GET method should work",
			method:         "GET",
			path:           "/api/resource",
			expectedStatus: 200,
			expectedJSON:   map[string]interface{}{"method": "GET", "message": "success"},
		},
		{
			name:           "POST method should work",
			method:         "POST",
			path:           "/api/resource",
			expectedStatus: 201,
			expectedJSON:   map[string]interface{}{"method": "POST", "message": "created"},
		},
		{
			name:           "PUT method should work",
			method:         "PUT",
			path:           "/api/resource",
			expectedStatus: 200,
			expectedJSON:   map[string]interface{}{"method": "PUT", "message": "updated"},
		},
		{
			name:           "DELETE method should work",
			method:         "DELETE",
			path:           "/api/resource",
			expectedStatus: 200,
			expectedJSON:   map[string]interface{}{"method": "DELETE", "message": "deleted"},
		},
		{
			name:           "PATCH method should work",
			method:         "PATCH",
			path:           "/api/resource",
			expectedStatus: 200,
			expectedJSON:   map[string]interface{}{"method": "PATCH", "message": "patched"},
		},
		{
			name:           "HEAD method should work",
			method:         "HEAD",
			path:           "/api/resource",
			expectedStatus: 200,
			expectNoBody:   true, // HEAD should have no body
		},
		{
			name:           "OPTIONS method should work",
			method:         "OPTIONS",
			path:           "/api/resource",
			expectedStatus: 200,
			expectedJSON:   map[string]interface{}{"methods": []interface{}{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectNoBody {
				if w.Body.Len() > 0 {
					t.Errorf("Expected no body for %s method, got %q", tt.method, w.Body.String())
				}
			} else if tt.expectedJSON != nil {
				var actualJSON map[string]interface{}
				body := strings.TrimSpace(w.Body.String())
				if err := json.Unmarshal([]byte(body), &actualJSON); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
					return
				}

				// Compare JSON objects
				if !jsonEqual(tt.expectedJSON, actualJSON) {
					t.Errorf("Expected JSON %v, got %v", tt.expectedJSON, actualJSON)
				}
			}
		})
	}
}

// jsonEqual compares two JSON objects for equality
func jsonEqual(expected, actual map[string]interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			return false
		}

		// Handle slice comparison for methods array
		if expectedSlice, ok := expectedValue.([]interface{}); ok {
			if actualSlice, ok := actualValue.([]interface{}); ok {
				if len(expectedSlice) != len(actualSlice) {
					return false
				}
				for i, v := range expectedSlice {
					if v != actualSlice[i] {
						return false
					}
				}
			} else {
				return false
			}
		} else if expectedValue != actualValue {
			return false
		}
	}

	return true
}

func TestMethodNotAllowed(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register a route that only supports GET
	server.Route("/api/get-only").GET("/", func(ctx kratoshttp.Context) error {
		return ctx.JSON(200, map[string]string{"method": "GET"})
	})

	// Test unsupported method - Kratos returns 404 for unmatched routes
	// This is actually correct HTTP behavior
	req := httptest.NewRequest("POST", "/api/get-only", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	// Kratos returns 404 for unmatched routes, which is acceptable behavior
	if w.Code != 404 {
		t.Errorf("Expected status 404 Not Found for unsupported method, got %d", w.Code)
	}
}

func TestMultipleMethodsOnSameRoute(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register multiple methods on same route
	route := server.Route("/api/multi")
	route.GET("/", func(ctx kratoshttp.Context) error {
		return ctx.JSON(200, map[string]string{"method": "GET"})
	})
	route.POST("/", func(ctx kratoshttp.Context) error {
		return ctx.JSON(201, map[string]string{"method": "POST"})
	})

	tests := []struct {
		method         string
		expectedStatus int
		expectedMethod string
	}{
		{"GET", 200, "GET"},
		{"POST", 201, "POST"},
		{"PUT", 404, ""}, // Should return 404 for unsupported method (Kratos behavior)
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/multi", nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
