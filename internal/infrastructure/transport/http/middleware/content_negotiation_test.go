package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHandler implements http.Handler for testing
type mockHandler struct {
	called bool
	req    *http.Request
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.req = r
}

func TestContentNegotiation(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		acceptHeader       string
		expectedNegotiated string
		shouldSetHeader    bool
	}{
		{
			name:               "GET request with JSON-LD accept",
			method:             "GET",
			acceptHeader:       "application/ld+json",
			expectedNegotiated: "application/ld+json",
			shouldSetHeader:    true,
		},
		{
			name:               "GET request with Turtle accept",
			method:             "GET",
			acceptHeader:       "text/turtle",
			expectedNegotiated: "text/turtle",
			shouldSetHeader:    true,
		},
		{
			name:               "GET request with JSON alias",
			method:             "GET",
			acceptHeader:       "application/json",
			expectedNegotiated: "application/ld+json",
			shouldSetHeader:    true,
		},
		{
			name:               "HEAD request with RDF/XML accept",
			method:             "HEAD",
			acceptHeader:       "application/rdf+xml",
			expectedNegotiated: "application/rdf+xml",
			shouldSetHeader:    true,
		},
		{
			name:               "POST request (should not negotiate)",
			method:             "POST",
			acceptHeader:       "application/ld+json",
			expectedNegotiated: "",
			shouldSetHeader:    false,
		},
		{
			name:               "PUT request (should not negotiate)",
			method:             "PUT",
			acceptHeader:       "text/turtle",
			expectedNegotiated: "",
			shouldSetHeader:    false,
		},
		{
			name:               "GET request with wildcard",
			method:             "GET",
			acceptHeader:       "*/*",
			expectedNegotiated: "application/ld+json", // Default format
			shouldSetHeader:    true,
		},
		{
			name:               "GET request with unsupported format",
			method:             "GET",
			acceptHeader:       "application/unsupported",
			expectedNegotiated: "",
			shouldSetHeader:    false,
		},
		{
			name:               "GET request without Accept header",
			method:             "GET",
			acceptHeader:       "",
			expectedNegotiated: "",
			shouldSetHeader:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}

			w := httptest.NewRecorder()

			handler := &mockHandler{}
			middleware := ContentNegotiation()
			wrappedHandler := middleware(handler)

			// Execute
			wrappedHandler.ServeHTTP(w, req)

			// Assert
			assert.True(t, handler.called)

			if tt.shouldSetHeader {
				negotiatedFormat := req.Header.Get("X-Negotiated-Format")
				assert.Equal(t, tt.expectedNegotiated, negotiatedFormat)
			} else {
				negotiatedFormat := req.Header.Get("X-Negotiated-Format")
				assert.Empty(t, negotiatedFormat)
			}
		})
	}
}

func TestContentNegotiationWithConfig(t *testing.T) {
	// Custom configuration with different supported formats
	config := ContentNegotiationConfig{
		SupportedFormats: []string{"text/turtle", "application/rdf+xml"},
		DefaultFormat:    "text/turtle",
	}

	tests := []struct {
		name               string
		acceptHeader       string
		expectedNegotiated string
	}{
		{
			name:               "Turtle supported",
			acceptHeader:       "text/turtle",
			expectedNegotiated: "text/turtle",
		},
		{
			name:               "RDF/XML supported",
			acceptHeader:       "application/rdf+xml",
			expectedNegotiated: "application/rdf+xml",
		},
		{
			name:               "JSON-LD not supported",
			acceptHeader:       "application/ld+json",
			expectedNegotiated: "",
		},
		{
			name:               "Wildcard uses default",
			acceptHeader:       "*/*",
			expectedNegotiated: "text/turtle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept", tt.acceptHeader)

			w := httptest.NewRecorder()

			handler := &mockHandler{}
			middleware := ContentNegotiationWithConfig(config)
			wrappedHandler := middleware(handler)

			// Execute
			wrappedHandler.ServeHTTP(w, req)

			// Assert
			assert.True(t, handler.called)

			negotiatedFormat := req.Header.Get("X-Negotiated-Format")
			assert.Equal(t, tt.expectedNegotiated, negotiatedFormat)
		})
	}
}

func TestParseAcceptHeader(t *testing.T) {
	tests := []struct {
		name          string
		acceptHeader  string
		expectedTypes []AcceptType
	}{
		{
			name:         "single type",
			acceptHeader: "application/ld+json",
			expectedTypes: []AcceptType{
				{MediaType: "application/ld+json", Quality: 1.0},
			},
		},
		{
			name:         "multiple types without quality",
			acceptHeader: "application/ld+json, text/turtle",
			expectedTypes: []AcceptType{
				{MediaType: "application/ld+json", Quality: 1.0},
				{MediaType: "text/turtle", Quality: 1.0},
			},
		},
		{
			name:         "multiple types with quality",
			acceptHeader: "application/ld+json;q=0.9, text/turtle;q=0.8",
			expectedTypes: []AcceptType{
				{MediaType: "application/ld+json", Quality: 0.9},
				{MediaType: "text/turtle", Quality: 0.8},
			},
		},
		{
			name:         "quality sorting",
			acceptHeader: "text/turtle;q=0.8, application/ld+json;q=0.9, application/rdf+xml;q=0.7",
			expectedTypes: []AcceptType{
				{MediaType: "application/ld+json", Quality: 0.9},
				{MediaType: "text/turtle", Quality: 0.8},
				{MediaType: "application/rdf+xml", Quality: 0.7},
			},
		},
		{
			name:         "wildcard type",
			acceptHeader: "*/*",
			expectedTypes: []AcceptType{
				{MediaType: "*/*", Quality: 1.0},
			},
		},
		{
			name:         "application wildcard",
			acceptHeader: "application/*",
			expectedTypes: []AcceptType{
				{MediaType: "application/*", Quality: 1.0},
			},
		},
		{
			name:          "empty header",
			acceptHeader:  "",
			expectedTypes: []AcceptType{},
		},
		{
			name:         "complex header with parameters",
			acceptHeader: "application/ld+json;q=0.9;charset=utf-8, text/turtle;q=0.8",
			expectedTypes: []AcceptType{
				{MediaType: "application/ld+json", Quality: 0.9},
				{MediaType: "text/turtle", Quality: 0.8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptHeader(tt.acceptHeader)

			assert.Equal(t, len(tt.expectedTypes), len(result))

			for i, expected := range tt.expectedTypes {
				if i < len(result) {
					assert.Equal(t, expected.MediaType, result[i].MediaType)
					assert.InDelta(t, expected.Quality, result[i].Quality, 0.001)
				}
			}
		})
	}
}

func TestMatchesMediaType(t *testing.T) {
	tests := []struct {
		name       string
		acceptType string
		format     string
		expected   bool
	}{
		{
			name:       "exact match JSON-LD",
			acceptType: "application/ld+json",
			format:     "application/ld+json",
			expected:   true,
		},
		{
			name:       "exact match Turtle",
			acceptType: "text/turtle",
			format:     "text/turtle",
			expected:   true,
		},
		{
			name:       "JSON alias to JSON-LD",
			acceptType: "application/json",
			format:     "application/ld+json",
			expected:   true,
		},
		{
			name:       "plain text alias to Turtle",
			acceptType: "text/plain",
			format:     "text/turtle",
			expected:   true,
		},
		{
			name:       "XML alias to RDF/XML",
			acceptType: "application/xml",
			format:     "application/rdf+xml",
			expected:   true,
		},
		{
			name:       "application wildcard matches JSON-LD",
			acceptType: "application/*",
			format:     "application/ld+json",
			expected:   true,
		},
		{
			name:       "application wildcard matches RDF/XML",
			acceptType: "application/*",
			format:     "application/rdf+xml",
			expected:   true,
		},
		{
			name:       "text wildcard matches Turtle",
			acceptType: "text/*",
			format:     "text/turtle",
			expected:   true,
		},
		{
			name:       "no match",
			acceptType: "application/pdf",
			format:     "text/turtle",
			expected:   false,
		},
		{
			name:       "wildcard doesn't match different type",
			acceptType: "image/*",
			format:     "text/turtle",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesMediaType(tt.acceptType, tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateContentType(t *testing.T) {
	supportedFormats := []string{
		"application/ld+json",
		"text/turtle",
		"application/rdf+xml",
	}

	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "valid JSON-LD",
			contentType: "application/ld+json",
			expected:    true,
		},
		{
			name:        "valid Turtle",
			contentType: "text/turtle",
			expected:    true,
		},
		{
			name:        "valid RDF/XML",
			contentType: "application/rdf+xml",
			expected:    true,
		},
		{
			name:        "JSON alias",
			contentType: "application/json",
			expected:    true,
		},
		{
			name:        "plain text alias",
			contentType: "text/plain",
			expected:    true,
		},
		{
			name:        "XML alias",
			contentType: "application/xml",
			expected:    true,
		},
		{
			name:        "content type with charset",
			contentType: "application/ld+json; charset=utf-8",
			expected:    true,
		},
		{
			name:        "case insensitive",
			contentType: "APPLICATION/LD+JSON",
			expected:    true,
		},
		{
			name:        "unsupported format",
			contentType: "application/pdf",
			expected:    false,
		},
		{
			name:        "empty content type",
			contentType: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateContentType(tt.contentType, supportedFormats)
			assert.Equal(t, tt.expected, result)
		})
	}
}
