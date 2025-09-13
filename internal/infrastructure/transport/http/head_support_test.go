package http

import (
	"net/http/httptest"
	"testing"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
)

func TestHEADMethodSupport(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)

	// Register HEAD test routes
	RegisterHEADTestRoutes(server)

	tests := []struct {
		name                string
		path                string
		expectedStatus      int
		expectedContentType string
		expectedHeaders     map[string]string
	}{
		{
			name:                "HEAD request should return same headers as GET",
			path:                "/api/head-test",
			expectedStatus:      200,
			expectedContentType: "application/json",
			expectedHeaders: map[string]string{
				"X-Resource-Type": "test-resource",
				"X-Version":       "1.0",
			},
		},
		{
			name:                "HEAD request with path parameter",
			path:                "/api/head-test/123",
			expectedStatus:      200,
			expectedContentType: "application/json",
			expectedHeaders: map[string]string{
				"X-Resource-Type": "test-resource",
				"X-Resource-ID":   "123",
			},
		},
		{
			name:           "HEAD request for non-existent resource",
			path:           "/api/head-test/notfound",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, make a GET request to compare headers
			getReq := httptest.NewRequest("GET", tt.path, nil)
			getW := httptest.NewRecorder()
			server.ServeHTTP(getW, getReq)

			// Then make a HEAD request
			headReq := httptest.NewRequest("HEAD", tt.path, nil)
			headW := httptest.NewRecorder()
			server.ServeHTTP(headW, headReq)

			// Check status code
			if headW.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, headW.Code)
			}

			// HEAD should have no body
			if headW.Body.Len() > 0 {
				t.Errorf("HEAD response should have no body, got %q", headW.Body.String())
			}

			// For successful responses, compare headers with GET
			if tt.expectedStatus == 200 && getW.Code == 200 {
				// Check Content-Type header
				if tt.expectedContentType != "" {
					headContentType := headW.Header().Get("Content-Type")
					getContentType := getW.Header().Get("Content-Type")

					if headContentType != getContentType {
						t.Errorf("HEAD Content-Type %q should match GET Content-Type %q", headContentType, getContentType)
					}

					if headContentType != tt.expectedContentType {
						t.Errorf("Expected Content-Type %q, got %q", tt.expectedContentType, headContentType)
					}
				}

				// Check custom headers
				for headerName, expectedValue := range tt.expectedHeaders {
					headValue := headW.Header().Get(headerName)
					getValue := getW.Header().Get(headerName)

					if headValue != getValue {
						t.Errorf("HEAD header %s=%q should match GET header %s=%q", headerName, headValue, headerName, getValue)
					}

					if headValue != expectedValue {
						t.Errorf("Expected header %s=%q, got %q", headerName, expectedValue, headValue)
					}
				}

				// Check Content-Length header if present in GET
				getContentLength := getW.Header().Get("Content-Length")
				headContentLength := headW.Header().Get("Content-Length")

				if getContentLength != "" && headContentLength != getContentLength {
					t.Errorf("HEAD Content-Length %q should match GET Content-Length %q", headContentLength, getContentLength)
				}
			}
		})
	}
}

func TestHEADMethodWithDifferentContentTypes(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)

	// Register content type test routes
	RegisterContentTypeTestRoutes(server)

	tests := []struct {
		name                string
		path                string
		expectedContentType string
	}{
		{
			name:                "JSON content type",
			path:                "/api/content/json",
			expectedContentType: "application/json",
		},
		{
			name:                "Plain text content type",
			path:                "/api/content/text",
			expectedContentType: "text/plain",
		},
		{
			name:                "XML content type",
			path:                "/api/content/xml",
			expectedContentType: "application/xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make GET request
			getReq := httptest.NewRequest("GET", tt.path, nil)
			getW := httptest.NewRecorder()
			server.ServeHTTP(getW, getReq)

			// Make HEAD request
			headReq := httptest.NewRequest("HEAD", tt.path, nil)
			headW := httptest.NewRecorder()
			server.ServeHTTP(headW, headReq)

			// Both should return 200
			if getW.Code != 200 {
				t.Errorf("GET request failed with status %d", getW.Code)
			}
			if headW.Code != 200 {
				t.Errorf("HEAD request failed with status %d", headW.Code)
			}

			// HEAD should have no body
			if headW.Body.Len() > 0 {
				t.Errorf("HEAD response should have no body, got %q", headW.Body.String())
			}

			// GET should have body
			if getW.Body.Len() == 0 {
				t.Error("GET response should have body")
			}

			// Content-Type should match
			headContentType := headW.Header().Get("Content-Type")
			getContentType := getW.Header().Get("Content-Type")

			if headContentType != getContentType {
				t.Errorf("HEAD Content-Type %q should match GET Content-Type %q", headContentType, getContentType)
			}

			if headContentType != tt.expectedContentType {
				t.Errorf("Expected Content-Type %q, got %q", tt.expectedContentType, headContentType)
			}
		})
	}
}

func TestHEADMethodErrorHandling(t *testing.T) {
	// Create test server
	logger := log.NewStdLogger(nil)
	config := &conf.HTTP{
		Addr:    ":0",
		Timeout: 30000000000,
	}

	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)

	// Register error test routes
	RegisterErrorTestRoutes(server)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "HEAD request for 404 error",
			path:           "/api/error/404",
			expectedStatus: 404,
		},
		{
			name:           "HEAD request for 500 error",
			path:           "/api/error/500",
			expectedStatus: 500,
		},
		{
			name:           "HEAD request for 400 error",
			path:           "/api/error/400",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make GET request to compare
			getReq := httptest.NewRequest("GET", tt.path, nil)
			getW := httptest.NewRecorder()
			server.ServeHTTP(getW, getReq)

			// Make HEAD request
			headReq := httptest.NewRequest("HEAD", tt.path, nil)
			headW := httptest.NewRecorder()
			server.ServeHTTP(headW, headReq)

			// Status codes should match
			if headW.Code != getW.Code {
				t.Errorf("HEAD status %d should match GET status %d", headW.Code, getW.Code)
			}

			if headW.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, headW.Code)
			}

			// HEAD should have no body even for errors
			if headW.Body.Len() > 0 {
				t.Errorf("HEAD response should have no body even for errors, got %q", headW.Body.String())
			}

			// Error headers should match
			headContentType := headW.Header().Get("Content-Type")
			getContentType := getW.Header().Get("Content-Type")

			if headContentType != getContentType {
				t.Errorf("HEAD error Content-Type %q should match GET error Content-Type %q", headContentType, getContentType)
			}
		})
	}
}
