package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	pericarpinfra "github.com/akeemphilbert/pericarp/pkg/infrastructure"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResourceEndpointsIntegration tests the complete resource storage HTTP endpoints
func TestResourceEndpointsIntegration(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	logger := log.NewStdLogger(os.Stdout)

	// Create test server with all dependencies
	server := createResourceTestServer(t, tempDir, logger)
	defer server.Stop(context.Background())

	// Start server
	go func() {
		if err := server.Start(context.Background()); err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Get server address
	serverAddr := "http://127.0.0.1:8080"

	t.Run("POST_Resource_Creation", func(t *testing.T) {
		testResourceCreation(t, serverAddr)
	})

	t.Run("GET_Resource_Retrieval", func(t *testing.T) {
		testResourceRetrieval(t, serverAddr)
	})

	t.Run("PUT_Resource_Update", func(t *testing.T) {
		testResourceUpdate(t, serverAddr)
	})

	t.Run("DELETE_Resource_Deletion", func(t *testing.T) {
		testResourceDeletion(t, serverAddr)
	})

	t.Run("HEAD_Resource_Metadata", func(t *testing.T) {
		testResourceMetadata(t, serverAddr)
	})

	t.Run("OPTIONS_Resource_Methods", func(t *testing.T) {
		testResourceOptions(t, serverAddr)
	})

	t.Run("Content_Negotiation", func(t *testing.T) {
		testContentNegotiation(t, serverAddr)
	})

	t.Run("Error_Handling", func(t *testing.T) {
		testErrorHandling(t, serverAddr)
	})

	t.Run("CORS_Headers", func(t *testing.T) {
		testCORSHeaders(t, serverAddr)
	})
}

// createResourceTestServer creates a test HTTP server with all dependencies
func createResourceTestServer(t *testing.T, tempDir string, logger log.Logger) *khttp.Server {
	// Create infrastructure dependencies
	db, err := infrastructure.DatabaseProvider()
	require.NoError(t, err)

	eventStore, err := infrastructure.EventStoreProvider(db)
	require.NoError(t, err)

	eventDispatcher, err := infrastructure.NewEventDispatcher()
	require.NoError(t, err)

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return pericarpinfra.UnitOfWorkProvider(eventStore, eventDispatcher)
	}

	// Create repository
	repo, err := infrastructure.NewFileSystemRepositoryWithPath(tempDir)
	require.NoError(t, err)

	// Create converter
	converter := infrastructure.NewRDFConverter()

	// Create storage service
	storageService := application.NewStorageService(repo, converter, unitOfWorkFactory)

	// Register event handlers
	registrar := application.NewEventHandlerRegistrar(eventDispatcher)
	err = registrar.RegisterAllHandlers(repo)
	require.NoError(t, err)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)
	resourceHandler := handlers.NewResourceHandler(storageService, logger)

	// Create HTTP server configuration
	httpConfig := &conf.HTTP{
		Addr:    "127.0.0.1:8080",
		Timeout: conf.Duration(30 * time.Second),
	}

	// Create server
	server := NewHTTPServer(httpConfig, logger, healthHandler, requestResponseHandler, resourceHandler)

	return server
}

// testResourceCreation tests POST /resources endpoint
func testResourceCreation(t *testing.T, serverAddr string) {
	// Test JSON-LD resource creation
	jsonLDData := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"@type": "Note",
		"content": "This is a test note"
	}`

	resp, err := http.Post(
		serverAddr+"/resources",
		"application/ld+json",
		strings.NewReader(jsonLDData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Unexpected status %d, response body: %s", resp.StatusCode, string(body))
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.NotEmpty(t, resp.Header.Get("Location"))
	assert.NotEmpty(t, resp.Header.Get("ETag"))

	// Parse response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	resourceID, ok := response["id"].(string)
	assert.True(t, ok, "Response should contain string ID")
	assert.NotEmpty(t, resourceID, "Resource ID should not be empty")
	assert.Len(t, resourceID, 27, "KSUID should be 27 characters long")

	// Verify it's a valid KSUID format (base62 characters only)
	for _, char := range resourceID {
		assert.True(t,
			(char >= '0' && char <= '9') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z'),
			"KSUID should only contain base62 characters, found: %c", char)
	}

	assert.Equal(t, "application/ld+json", response["contentType"])
	assert.Contains(t, response["message"], "created successfully")
}

// testResourceRetrieval tests GET /resources/{id} endpoint
func testResourceRetrieval(t *testing.T, serverAddr string) {
	// First create a resource
	resourceID := createTestResource(t, serverAddr)

	// Test retrieval with different Accept headers
	testCases := []struct {
		name                string
		acceptHeader        string
		expectedContentType string
	}{
		{
			name:                "JSON-LD",
			acceptHeader:        "application/ld+json",
			expectedContentType: "application/ld+json",
		},
		{
			name:                "Turtle",
			acceptHeader:        "text/turtle",
			expectedContentType: "text/turtle",
		},
		{
			name:                "RDF/XML",
			acceptHeader:        "application/rdf+xml",
			expectedContentType: "application/rdf+xml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", serverAddr+"/resources/"+resourceID, nil)
			require.NoError(t, err)
			req.Header.Set("Accept", tc.acceptHeader)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.expectedContentType, resp.Header.Get("Content-Type"))
			assert.NotEmpty(t, resp.Header.Get("Content-Length"))
			assert.NotEmpty(t, resp.Header.Get("ETag"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}

// testResourceUpdate tests PUT /resources/{id} endpoint
func testResourceUpdate(t *testing.T, serverAddr string) {
	resourceID := "test-update-resource"

	// Test creating new resource with PUT
	turtleData := `@prefix ex: <http://example.org/> .
ex:resource ex:title "Updated Resource" .`

	req, err := http.NewRequest("PUT", serverAddr+"/resources/"+resourceID, strings.NewReader(turtleData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/turtle")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Location"))

	// Test updating existing resource
	updatedData := `@prefix ex: <http://example.org/> .
ex:resource ex:title "Updated Resource Again" .`

	req, err = http.NewRequest("PUT", serverAddr+"/resources/"+resourceID, strings.NewReader(updatedData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/turtle")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// testResourceDeletion tests DELETE /resources/{id} endpoint
func testResourceDeletion(t *testing.T, serverAddr string) {
	// Create a resource to delete
	resourceID := createTestResource(t, serverAddr)

	// Delete the resource
	req, err := http.NewRequest("DELETE", serverAddr+"/resources/"+resourceID, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "deleted successfully")

	// Verify resource is deleted
	req, err = http.NewRequest("GET", serverAddr+"/resources/"+resourceID, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// testResourceMetadata tests HEAD /resources/{id} endpoint
func testResourceMetadata(t *testing.T, serverAddr string) {
	resourceID := createTestResource(t, serverAddr)

	req, err := http.NewRequest("HEAD", serverAddr+"/resources/"+resourceID, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/ld+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/ld+json", resp.Header.Get("Content-Type"))
	assert.NotEmpty(t, resp.Header.Get("Content-Length"))
	assert.NotEmpty(t, resp.Header.Get("ETag"))

	// Verify no body is returned
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Empty(t, body)
}

// testResourceOptions tests OPTIONS /resources/{id} endpoint
func testResourceOptions(t *testing.T, serverAddr string) {
	req, err := http.NewRequest("OPTIONS", serverAddr+"/resources/test", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "DELETE")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	methods, ok := response["methods"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, methods, "GET")
	assert.Contains(t, methods, "POST")
	assert.Contains(t, methods, "PUT")
	assert.Contains(t, methods, "DELETE")
}

// testContentNegotiation tests content negotiation functionality
func testContentNegotiation(t *testing.T, serverAddr string) {
	resourceID := createTestResource(t, serverAddr)

	testCases := []struct {
		name           string
		acceptHeader   string
		expectedFormat string
	}{
		{
			name:           "Prefer JSON-LD",
			acceptHeader:   "application/ld+json,text/turtle;q=0.8",
			expectedFormat: "application/ld+json",
		},
		{
			name:           "Prefer Turtle",
			acceptHeader:   "text/turtle,application/ld+json;q=0.5",
			expectedFormat: "text/turtle",
		},
		{
			name:           "Wildcard Accept",
			acceptHeader:   "*/*",
			expectedFormat: "application/ld+json", // Should return preferred format
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", serverAddr+"/resources/"+resourceID, nil)
			require.NoError(t, err)
			req.Header.Set("Accept", tc.acceptHeader)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.expectedFormat, resp.Header.Get("Content-Type"))
		})
	}
}

// testErrorHandling tests error response handling
func testErrorHandling(t *testing.T, serverAddr string) {
	client := &http.Client{}

	t.Run("Resource Not Found", func(t *testing.T) {
		resp, err := client.Get(serverAddr + "/resources/nonexistent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var errorResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)

		errorInfo, ok := errorResponse["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "RESOURCE_NOT_FOUND", errorInfo["code"])
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		resourceID := createTestResource(t, serverAddr)

		req, err := http.NewRequest("GET", serverAddr+"/resources/"+resourceID, nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "application/unsupported")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
	})

	t.Run("Missing Content Type", func(t *testing.T) {
		resp, err := http.Post(serverAddr+"/resources", "", strings.NewReader("test data"))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)

		errorInfo, ok := errorResponse["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "MISSING_CONTENT_TYPE", errorInfo["code"])
	})
}

// testCORSHeaders tests CORS header functionality
func testCORSHeaders(t *testing.T, serverAddr string) {
	client := &http.Client{}

	// Test preflight request
	req, err := http.NewRequest("OPTIONS", serverAddr+"/resources/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Max-Age"))

	// Test actual request with CORS headers
	jsonData := `{"@context": "https://www.w3.org/ns/activitystreams", "@type": "Note"}`
	req, err = http.NewRequest("POST", serverAddr+"/resources", strings.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Content-Type", "application/ld+json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

// createTestResource creates a test resource and returns its ID
func createTestResource(t *testing.T, serverAddr string) string {
	jsonLDData := `{
		"@context": "https://www.w3.org/ns/activitystreams",
		"@type": "Note",
		"content": "Test resource for integration testing"
	}`

	resp, err := http.Post(
		serverAddr+"/resources",
		"application/ld+json",
		strings.NewReader(jsonLDData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	resourceID, ok := response["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, resourceID)

	return resourceID
}

// TestMiddlewareIntegration tests middleware functionality
func TestMiddlewareIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewStdLogger(os.Stdout)

	server := createResourceTestServer(t, tempDir, logger)
	defer server.Stop(context.Background())

	// Start server
	go func() {
		if err := server.Start(context.Background()); err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	serverAddr := "http://127.0.0.1:8080"

	t.Run("Request Timeout", func(t *testing.T) {
		// This test would require a slow endpoint to properly test timeout
		// For now, just verify the middleware is applied
		client := &http.Client{Timeout: 1 * time.Second}

		resp, err := client.Get(serverAddr + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Structured Logging", func(t *testing.T) {
		// Create a resource to generate log entries
		jsonData := `{"@context": "https://www.w3.org/ns/activitystreams", "@type": "Note"}`

		resp, err := http.Post(
			serverAddr+"/resources",
			"application/ld+json",
			strings.NewReader(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		// Logging verification would require capturing log output
		// This test ensures the middleware doesn't break the request flow
	})

	t.Run("Content Negotiation Middleware", func(t *testing.T) {
		resourceID := createTestResource(t, serverAddr)

		req, err := http.NewRequest("GET", serverAddr+"/resources/"+resourceID, nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "text/turtle")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/turtle", resp.Header.Get("Content-Type"))
	})
}

// TestRouteParameterHandling tests route parameter extraction
func TestRouteParameterHandling(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewStdLogger(os.Stdout)

	server := createResourceTestServer(t, tempDir, logger)
	defer server.Stop(context.Background())

	go func() {
		if err := server.Start(context.Background()); err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	serverAddr := "http://127.0.0.1:8080"

	t.Run("Valid Resource ID", func(t *testing.T) {
		resourceID := "test-resource-123"

		// Create resource with specific ID
		turtleData := `@prefix ex: <http://example.org/> .
ex:resource ex:title "Test Resource" .`

		req, err := http.NewRequest("PUT", serverAddr+"/resources/"+resourceID, strings.NewReader(turtleData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/turtle")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Retrieve resource to verify ID handling
		resp, err = client.Get(serverAddr + "/resources/" + resourceID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Empty Resource ID", func(t *testing.T) {
		// Test endpoint without ID parameter
		resp, err := http.Get(serverAddr + "/resources/")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return method not allowed or not found
		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed)
	})
}
