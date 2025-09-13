package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
)

func TestOPTIONSMethodDiscovery(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register routes with different method combinations
	RegisterOPTIONSTestRoutes(server)

	tests := []struct {
		name            string
		path            string
		expectedMethods []string
	}{
		{
			name:            "Full CRUD resource should return all methods",
			path:            "/api/crud-resource",
			expectedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		},
		{
			name:            "Read-only resource should return limited methods",
			path:            "/api/readonly-resource",
			expectedMethods: []string{"GET", "HEAD", "OPTIONS"},
		},
		{
			name:            "Write-only resource should return write methods",
			path:            "/api/writeonly-resource",
			expectedMethods: []string{"POST", "PUT", "OPTIONS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", tt.path, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Parse JSON response
			var response map[string]interface{}
			body := strings.TrimSpace(w.Body.String())
			if err := json.Unmarshal([]byte(body), &response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			// Check methods array
			methodsInterface, exists := response["methods"]
			if !exists {
				t.Error("Response should contain 'methods' field")
				return
			}

			methods, ok := methodsInterface.([]interface{})
			if !ok {
				t.Error("Methods field should be an array")
				return
			}

			// Convert to string slice for comparison
			var actualMethods []string
			for _, method := range methods {
				if methodStr, ok := method.(string); ok {
					actualMethods = append(actualMethods, methodStr)
				}
			}

			// Check that all expected methods are present
			for _, expectedMethod := range tt.expectedMethods {
				found := false
				for _, actualMethod := range actualMethods {
					if actualMethod == expectedMethod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected method %s not found in response. Got: %v", expectedMethod, actualMethods)
				}
			}
		})
	}
}

func TestOPTIONSCORSPreflight(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register test routes
	RegisterOPTIONSTestRoutes(server)

	tests := []struct {
		name                   string
		path                   string
		requestMethod          string
		requestHeaders         string
		expectCORSHeaders      bool
		expectedAllowedMethods string
		expectedAllowedHeaders string
	}{
		{
			name:                   "CORS preflight request should be handled by middleware",
			path:                   "/api/crud-resource",
			requestMethod:          "POST",
			requestHeaders:         "Content-Type,Authorization",
			expectCORSHeaders:      true,
			expectedAllowedMethods: "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
			expectedAllowedHeaders: "Content-Type,Authorization,X-Requested-With",
		},
		{
			name:                   "Simple OPTIONS without CORS headers should go to handler",
			path:                   "/api/crud-resource",
			requestMethod:          "",
			requestHeaders:         "",
			expectCORSHeaders:      true, // CORS middleware still adds headers
			expectedAllowedMethods: "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
			expectedAllowedHeaders: "Content-Type,Authorization,X-Requested-With",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", tt.path, nil)

			// Add CORS preflight headers if specified
			if tt.requestMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.requestMethod)
			}
			if tt.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			// Check CORS headers are present
			if tt.expectCORSHeaders {
				allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if allowOrigin == "" {
					t.Error("Expected Access-Control-Allow-Origin header")
				}

				allowMethods := w.Header().Get("Access-Control-Allow-Methods")
				if allowMethods != tt.expectedAllowedMethods {
					t.Errorf("Expected Access-Control-Allow-Methods %q, got %q", tt.expectedAllowedMethods, allowMethods)
				}

				allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
				if allowHeaders != tt.expectedAllowedHeaders {
					t.Errorf("Expected Access-Control-Allow-Headers %q, got %q", tt.expectedAllowedHeaders, allowHeaders)
				}
			}

			// For preflight requests, expect 204 No Content
			if tt.requestMethod != "" {
				if w.Code != 204 {
					t.Errorf("Expected status 204 for preflight request, got %d", w.Code)
				}
			} else {
				// For non-preflight OPTIONS, expect 200 with JSON response
				if w.Code != 200 {
					t.Errorf("Expected status 200 for OPTIONS request, got %d", w.Code)
				}
			}
		})
	}
}

func TestOPTIONSWithPathParameters(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)

	// Register parameterized routes
	RegisterParameterizedOPTIONSRoutes(server)

	tests := []struct {
		name            string
		path            string
		expectedMethods []string
	}{
		{
			name:            "Resource with ID should return item-specific methods",
			path:            "/api/items/123",
			expectedMethods: []string{"GET", "PUT", "DELETE", "HEAD", "OPTIONS"},
		},
		{
			name:            "Collection should return collection methods",
			path:            "/api/items",
			expectedMethods: []string{"GET", "POST", "HEAD", "OPTIONS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", tt.path, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Parse JSON response
			var response map[string]interface{}
			body := strings.TrimSpace(w.Body.String())
			if err := json.Unmarshal([]byte(body), &response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			// Check methods array
			methodsInterface, exists := response["methods"]
			if !exists {
				t.Error("Response should contain 'methods' field")
				return
			}

			methods, ok := methodsInterface.([]interface{})
			if !ok {
				t.Error("Methods field should be an array")
				return
			}

			// Convert to string slice for comparison
			var actualMethods []string
			for _, method := range methods {
				if methodStr, ok := method.(string); ok {
					actualMethods = append(actualMethods, methodStr)
				}
			}

			// Check that all expected methods are present
			for _, expectedMethod := range tt.expectedMethods {
				found := false
				for _, actualMethod := range actualMethods {
					if actualMethod == expectedMethod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected method %s not found in response. Got: %v", expectedMethod, actualMethods)
				}
			}
		})
	}
}
