package handlers

import (
	"context"
	"io"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService is a mock implementation of StorageServiceInterface
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) StoreResource(ctx context.Context, id string, data []byte, contentType string) (*domain.Resource, error) {
	args := m.Called(ctx, id, data, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockStorageService) RetrieveResource(ctx context.Context, id string, acceptFormat string) (*domain.Resource, error) {
	args := m.Called(ctx, id, acceptFormat)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func (m *MockStorageService) DeleteResource(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorageService) ResourceExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorageService) StreamResource(ctx context.Context, id string, acceptFormat string) (io.ReadCloser, string, error) {
	args := m.Called(ctx, id, acceptFormat)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.String(1), args.Error(2)
}

func (m *MockStorageService) StoreResourceStream(ctx context.Context, id string, reader io.Reader, contentType string) (*domain.Resource, error) {
	args := m.Called(ctx, id, reader, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resource), args.Error(1)
}

func TestNewResourceHandler(t *testing.T) {
	mockService := new(MockStorageService)
	logger := log.NewStdLogger(io.Discard)

	handler := NewResourceHandler(mockService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.storageService)
	assert.Equal(t, logger, handler.logger)
}

func TestResourceHandler_ContentNegotiation(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedFormat string
	}{
		{
			name:           "JSON-LD exact match",
			acceptHeader:   "application/ld+json",
			expectedFormat: "application/ld+json",
		},
		{
			name:           "Turtle exact match",
			acceptHeader:   "text/turtle",
			expectedFormat: "text/turtle",
		},
		{
			name:           "RDF/XML exact match",
			acceptHeader:   "application/rdf+xml",
			expectedFormat: "application/rdf+xml",
		},
		{
			name:           "JSON alias to JSON-LD",
			acceptHeader:   "application/json",
			expectedFormat: "application/ld+json",
		},
		{
			name:           "Plain text alias to Turtle",
			acceptHeader:   "text/plain",
			expectedFormat: "text/turtle",
		},
		{
			name:           "XML alias to RDF/XML",
			acceptHeader:   "application/xml",
			expectedFormat: "application/rdf+xml",
		},
		{
			name:           "Quality values - prefer JSON-LD",
			acceptHeader:   "text/turtle;q=0.8, application/ld+json;q=0.9",
			expectedFormat: "application/ld+json",
		},
		{
			name:           "Quality values - prefer Turtle",
			acceptHeader:   "application/ld+json;q=0.7, text/turtle;q=0.9",
			expectedFormat: "text/turtle",
		},
		{
			name:           "Wildcard acceptance",
			acceptHeader:   "*/*",
			expectedFormat: "application/ld+json", // Should return preferred format
		},
		{
			name:           "Application wildcard",
			acceptHeader:   "application/*",
			expectedFormat: "application/ld+json",
		},
		{
			name:           "Unsupported format",
			acceptHeader:   "application/unsupported",
			expectedFormat: "",
		},
		{
			name:           "Empty accept header",
			acceptHeader:   "",
			expectedFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockStorageService)
			logger := log.NewStdLogger(io.Discard)
			handler := NewResourceHandler(mockService, logger)

			// Test content negotiation
			result := handler.negotiateContentType(tt.acceptHeader)

			// Assert
			assert.Equal(t, tt.expectedFormat, result)
		})
	}
}

func TestResourceHandler_ParseAcceptHeader(t *testing.T) {
	tests := []struct {
		name            string
		acceptHeader    string
		expectedCount   int
		expectedFirst   string
		expectedQuality float64
	}{
		{
			name:            "single type",
			acceptHeader:    "application/ld+json",
			expectedCount:   1,
			expectedFirst:   "application/ld+json",
			expectedQuality: 1.0,
		},
		{
			name:            "multiple types without quality",
			acceptHeader:    "application/ld+json, text/turtle",
			expectedCount:   2,
			expectedFirst:   "application/ld+json",
			expectedQuality: 1.0,
		},
		{
			name:            "multiple types with quality - sorted",
			acceptHeader:    "text/turtle;q=0.8, application/ld+json;q=0.9",
			expectedCount:   2,
			expectedFirst:   "application/ld+json", // Higher quality should be first
			expectedQuality: 0.9,
		},
		{
			name:            "wildcard type",
			acceptHeader:    "*/*",
			expectedCount:   1,
			expectedFirst:   "*/*",
			expectedQuality: 1.0,
		},
		{
			name:            "empty header",
			acceptHeader:    "",
			expectedCount:   0,
			expectedFirst:   "",
			expectedQuality: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockStorageService)
			logger := log.NewStdLogger(io.Discard)
			handler := NewResourceHandler(mockService, logger)

			// Test accept header parsing
			result := handler.parseAcceptHeader(tt.acceptHeader)

			// Assert
			assert.Equal(t, tt.expectedCount, len(result))

			if tt.expectedCount > 0 {
				assert.Equal(t, tt.expectedFirst, result[0].mediaType)
				assert.InDelta(t, tt.expectedQuality, result[0].quality, 0.001)
			}
		})
	}
}

func TestResourceHandler_MatchesMediaType(t *testing.T) {
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
			name:       "no match",
			acceptType: "application/pdf",
			format:     "text/turtle",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockStorageService)
			logger := log.NewStdLogger(io.Discard)
			handler := NewResourceHandler(mockService, logger)

			// Test media type matching
			result := handler.matchesMediaType(tt.acceptType, tt.format)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceHandler_GenerateETag(t *testing.T) {
	// Setup
	mockService := new(MockStorageService)
	logger := log.NewStdLogger(io.Discard)
	handler := NewResourceHandler(mockService, logger)

	// Create test resource
	resource := domain.NewResource("test-id", "application/ld+json", []byte(`{"test": "data"}`))

	// Test ETag generation
	etag := handler.generateETag(resource)

	// Assert
	assert.NotEmpty(t, etag)
	assert.Contains(t, etag, "test-id")

	// Test that same resource generates same ETag
	etag2 := handler.generateETag(resource)
	assert.Equal(t, etag, etag2)

	// Test that different resource generates different ETag
	resource2 := domain.NewResource("test-id-2", "application/ld+json", []byte(`{"test": "different"}`))
	etag3 := handler.generateETag(resource2)
	assert.NotEqual(t, etag, etag3)
}
