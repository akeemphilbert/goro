package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/application"
	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create a test container handler
func createTestContainerHandler() (*ContainerHandler, *MockContainerService, *MockContainerStorageService) {
	mockContainerService := &MockContainerService{}
	mockStorageService := &MockContainerStorageService{}
	logger := log.NewStdLogger(log.NewFilter(log.NewStdLogger(nil), log.LevelError))

	handler := NewContainerHandler(mockContainerService, mockStorageService, logger)
	return handler, mockContainerService, mockStorageService
}

// Helper function to create a test HTTP context
func createTestContext(method, path string, body []byte, vars map[string][]string) khttp.Context {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil && len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()

	// Create a mock context that implements khttp.Context
	ctx := &mockHTTPContext{
		request:  req,
		response: w,
		vars:     vars,
	}

	return ctx
}

// mockHTTPContext implements khttp.Context for testing
type mockHTTPContext struct {
	request  *http.Request
	response *httptest.ResponseRecorder
	vars     map[string][]string
}

func (m *mockHTTPContext) Request() *http.Request {
	return m.request
}

func (m *mockHTTPContext) Response() http.ResponseWriter {
	return m.response
}

func (m *mockHTTPContext) Vars() map[string][]string {
	return m.vars
}

func (m *mockHTTPContext) JSON(code int, v interface{}) error {
	m.response.Header().Set("Content-Type", "application/json")
	m.response.WriteHeader(code)
	return json.NewEncoder(m.response).Encode(v)
}

// Test GET /containers/{id} - Container retrieval with member listing
func TestContainerHandler_GetContainer(t *testing.T) {
	tests := []struct {
		name           string
		containerID    string
		setupMocks     func(*MockContainerService, *MockContainerStorageService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful container retrieval",
			containerID: "test-container-1",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				container := domain.NewContainer("test-container-1", "", domain.BasicContainer)
				container.AddMember("resource-1")
				container.AddMember("resource-2")

				cs.On("GetContainer", mock.Anything, "test-container-1").Return(container, nil)

				listing := &application.ContainerListing{
					ContainerID: "test-container-1",
					Members:     []string{"resource-1", "resource-2"},
					Pagination:  domain.GetDefaultPagination(),
				}
				cs.On("ListContainerMembers", mock.Anything, "test-container-1", mock.AnythingOfType("domain.PaginationOptions")).Return(listing, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"@id":"test-container-1"`,
		},
		{
			name:        "container not found",
			containerID: "nonexistent-container",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("GetContainer", mock.Anything, "nonexistent-container").Return(nil, domain.ErrResourceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"CONTAINER_NOT_FOUND"`,
		},
		{
			name:           "empty container ID",
			containerID:    "",
			setupMocks:     func(cs *MockContainerService, ss *MockContainerStorageService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockContainerService, mockStorageService := createTestContainerHandler()
			tt.setupMocks(mockContainerService, mockStorageService)

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("GET", "/containers/"+tt.containerID, nil, vars)

			err := handler.GetContainer(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)
			assert.Contains(t, ctx.(*mockHTTPContext).response.Body.String(), tt.expectedBody)

			mockContainerService.AssertExpectations(t)
			mockStorageService.AssertExpectations(t)
		})
	}
}

// Test POST /containers/{id} - Resource creation in containers
func TestContainerHandler_PostResource(t *testing.T) {
	tests := []struct {
		name           string
		containerID    string
		requestBody    []byte
		setupMocks     func(*MockContainerService, *MockContainerStorageService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful resource creation in container",
			containerID: "test-container-1",
			requestBody: []byte(`{"data": "test resource data"}`),
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				// Container exists
				cs.On("ContainerExists", mock.Anything, "test-container-1").Return(true, nil)

				// Resource creation
				resource := domain.NewResource("generated-id", "application/json", []byte(`{"data": "test resource data"}`))
				ss.On("StoreResource", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), "application/json").Return(resource, nil)

				// Add resource to container
				cs.On("AddResource", mock.Anything, "test-container-1", mock.AnythingOfType("string")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"message":"Resource created in container successfully"`,
		},
		{
			name:        "container not found",
			containerID: "nonexistent-container",
			requestBody: []byte(`{"data": "test resource data"}`),
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("ContainerExists", mock.Anything, "nonexistent-container").Return(false, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"CONTAINER_NOT_FOUND"`,
		},
		{
			name:           "empty request body",
			containerID:    "test-container-1",
			requestBody:    []byte{},
			setupMocks:     func(cs *MockContainerService, ss *MockContainerStorageService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"EMPTY_BODY"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockContainerService, mockStorageService := createTestContainerHandler()
			tt.setupMocks(mockContainerService, mockStorageService)

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("POST", "/containers/"+tt.containerID, tt.requestBody, vars)

			err := handler.PostResource(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)
			assert.Contains(t, ctx.(*mockHTTPContext).response.Body.String(), tt.expectedBody)

			mockContainerService.AssertExpectations(t)
			mockStorageService.AssertExpectations(t)
		})
	}
}

// Test PUT /containers/{id} - Container metadata updates
func TestContainerHandler_PutContainer(t *testing.T) {
	tests := []struct {
		name           string
		containerID    string
		requestBody    []byte
		setupMocks     func(*MockContainerService, *MockContainerStorageService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful container metadata update",
			containerID: "test-container-1",
			requestBody: []byte(`{"title": "Updated Title", "description": "Updated Description"}`),
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				container := domain.NewContainer("test-container-1", "", domain.BasicContainer)
				cs.On("GetContainer", mock.Anything, "test-container-1").Return(container, nil)
				cs.On("UpdateContainer", mock.Anything, mock.AnythingOfType("*domain.Container")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Container updated successfully"`,
		},
		{
			name:        "container not found",
			containerID: "nonexistent-container",
			requestBody: []byte(`{"title": "Updated Title"}`),
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("GetContainer", mock.Anything, "nonexistent-container").Return(nil, domain.ErrResourceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"CONTAINER_NOT_FOUND"`,
		},
		{
			name:           "invalid JSON body",
			containerID:    "test-container-1",
			requestBody:    []byte(`{invalid json}`),
			setupMocks:     func(cs *MockContainerService, ss *MockContainerStorageService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"INVALID_JSON"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockContainerService, mockStorageService := createTestContainerHandler()
			tt.setupMocks(mockContainerService, mockStorageService)

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("PUT", "/containers/"+tt.containerID, tt.requestBody, vars)

			err := handler.PutContainer(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)
			assert.Contains(t, ctx.(*mockHTTPContext).response.Body.String(), tt.expectedBody)

			mockContainerService.AssertExpectations(t)
			mockStorageService.AssertExpectations(t)
		})
	}
}

// Test DELETE /containers/{id} - Container deletion with empty validation
func TestContainerHandler_DeleteContainer(t *testing.T) {
	tests := []struct {
		name           string
		containerID    string
		setupMocks     func(*MockContainerService, *MockContainerStorageService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful container deletion",
			containerID: "test-container-1",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("DeleteContainer", mock.Anything, "test-container-1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"Container deleted successfully"`,
		},
		{
			name:        "container not found",
			containerID: "nonexistent-container",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("DeleteContainer", mock.Anything, "nonexistent-container").Return(domain.ErrResourceNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"CONTAINER_NOT_FOUND"`,
		},
		{
			name:        "container not empty",
			containerID: "test-container-1",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("DeleteContainer", mock.Anything, "test-container-1").Return(domain.ErrContainerNotEmpty)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `"code":"CONTAINER_NOT_EMPTY"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockContainerService, mockStorageService := createTestContainerHandler()
			tt.setupMocks(mockContainerService, mockStorageService)

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("DELETE", "/containers/"+tt.containerID, nil, vars)

			err := handler.DeleteContainer(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)
			assert.Contains(t, ctx.(*mockHTTPContext).response.Body.String(), tt.expectedBody)

			mockContainerService.AssertExpectations(t)
			mockStorageService.AssertExpectations(t)
		})
	}
}

// Test HEAD /containers/{id} - Container metadata headers
func TestContainerHandler_HeadContainer(t *testing.T) {
	tests := []struct {
		name            string
		containerID     string
		setupMocks      func(*MockContainerService, *MockContainerStorageService)
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:        "successful container head request",
			containerID: "test-container-1",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				container := domain.NewContainer("test-container-1", "", domain.BasicContainer)
				container.AddMember("resource-1")
				cs.On("GetContainer", mock.Anything, "test-container-1").Return(container, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Content-Type": "application/ld+json",
			},
		},
		{
			name:        "container not found",
			containerID: "nonexistent-container",
			setupMocks: func(cs *MockContainerService, ss *MockContainerStorageService) {
				cs.On("GetContainer", mock.Anything, "nonexistent-container").Return(nil, domain.ErrResourceNotFound)
			},
			expectedStatus:  http.StatusNotFound,
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockContainerService, mockStorageService := createTestContainerHandler()
			tt.setupMocks(mockContainerService, mockStorageService)

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("HEAD", "/containers/"+tt.containerID, nil, vars)

			err := handler.HeadContainer(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)

			for key, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, ctx.(*mockHTTPContext).response.Header().Get(key))
			}

			mockContainerService.AssertExpectations(t)
			mockStorageService.AssertExpectations(t)
		})
	}
}

// Test OPTIONS /containers/{id} - Container options support
func TestContainerHandler_OptionsContainer(t *testing.T) {
	tests := []struct {
		name            string
		containerID     string
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name:           "successful options request",
			containerID:    "test-container-1",
			expectedStatus: http.StatusOK,
			expectedBody:   `"methods"`,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, HEAD, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Accept, Authorization",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _ := createTestContainerHandler()

			vars := map[string][]string{}
			if tt.containerID != "" {
				vars["id"] = []string{tt.containerID}
			}

			ctx := createTestContext("OPTIONS", "/containers/"+tt.containerID, nil, vars)

			err := handler.OptionsContainer(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, ctx.(*mockHTTPContext).response.Code)
			assert.Contains(t, ctx.(*mockHTTPContext).response.Body.String(), tt.expectedBody)

			for key, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, ctx.(*mockHTTPContext).response.Header().Get(key))
			}
		})
	}
}

// Test LDP-compliant endpoint behavior
func TestContainerHandler_LDPCompliance(t *testing.T) {
	t.Run("container response includes LDP headers", func(t *testing.T) {
		handler, mockContainerService, _ := createTestContainerHandler()

		container := domain.NewContainer("test-container-1", "", domain.BasicContainer)
		mockContainerService.On("GetContainer", mock.Anything, "test-container-1").Return(container, nil)

		listing := &application.ContainerListing{
			ContainerID: "test-container-1",
			Members:     []string{},
			Pagination:  domain.GetDefaultPagination(),
		}
		mockContainerService.On("ListContainerMembers", mock.Anything, "test-container-1", mock.AnythingOfType("domain.PaginationOptions")).Return(listing, nil)

		vars := map[string][]string{"id": {"test-container-1"}}
		ctx := createTestContext("GET", "/containers/test-container-1", nil, vars)

		err := handler.GetContainer(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, ctx.(*mockHTTPContext).response.Code)

		// Check for LDP-specific headers
		response := ctx.(*mockHTTPContext).response
		assert.Equal(t, "application/ld+json", response.Header().Get("Content-Type"))

		mockContainerService.AssertExpectations(t)
	})

	t.Run("container creation returns Location header", func(t *testing.T) {
		handler, mockContainerService, mockStorageService := createTestContainerHandler()

		mockContainerService.On("ContainerExists", mock.Anything, "test-container-1").Return(true, nil)

		resource := domain.NewResource("new-resource-id", "application/json", []byte(`{"data": "test"}`))
		mockStorageService.On("StoreResource", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), "application/json").Return(resource, nil)
		mockContainerService.On("AddResource", mock.Anything, "test-container-1", mock.AnythingOfType("string")).Return(nil)

		vars := map[string][]string{"id": {"test-container-1"}}
		ctx := createTestContext("POST", "/containers/test-container-1", []byte(`{"data": "test"}`), vars)

		err := handler.PostResource(ctx)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, ctx.(*mockHTTPContext).response.Code)

		// Check for Location header
		location := ctx.(*mockHTTPContext).response.Header().Get("Location")
		assert.Contains(t, location, "/resources/")

		mockContainerService.AssertExpectations(t)
		mockStorageService.AssertExpectations(t)
	})
}
