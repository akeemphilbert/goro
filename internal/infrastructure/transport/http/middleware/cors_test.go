package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		origin          string
		requestHeaders  string
		expectedStatus  int
		expectedHeaders map[string]string
		shouldCallNext  bool
	}{
		{
			name:           "GET request with CORS headers",
			method:         "GET",
			origin:         "https://example.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
			},
			shouldCallNext: true,
		},
		{
			name:           "POST request with CORS headers",
			method:         "POST",
			origin:         "https://example.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
			},
			shouldCallNext: true,
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			origin:         "https://example.com",
			requestHeaders: "Content-Type,Authorization",
			expectedStatus: 204,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
				"Access-Control-Max-Age":       "86400",
			},
			shouldCallNext: false,
		},
		{
			name:           "OPTIONS request without origin",
			method:         "OPTIONS",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
			},
			shouldCallNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}

			// Create a response recorder
			w := httptest.NewRecorder()

			// Track if next handler was called
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(200)
				w.Write([]byte("OK"))
			})

			// Create CORS filter
			corsFilter := CORS()

			// Execute the filter
			handler := corsFilter(next)
			handler.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.shouldCallNext, nextCalled)

			// Check expected headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
			}
		})
	}
}

func TestCORSWithConfig(t *testing.T) {
	tests := []struct {
		name            string
		config          CORSConfig
		method          string
		origin          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name: "Custom allowed origins",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com", "https://test.com"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         3600,
			},
			method:         "GET",
			origin:         "https://example.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		},
		{
			name: "Disallowed origin",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			method:         "GET",
			origin:         "https://malicious.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "", // Should not be set
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write([]byte("OK"))
			})

			corsFilter := CORSWithConfig(tt.config)
			handler := corsFilter(next)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			for key, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(key)
				if expectedValue == "" {
					assert.Empty(t, actualValue, "Header %s should be empty", key)
				} else {
					assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
				}
			}
		})
	}
}
