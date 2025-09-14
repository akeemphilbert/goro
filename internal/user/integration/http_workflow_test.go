package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUserRegistrationHTTPWorkflow tests the complete user registration workflow via HTTP
func TestUserRegistrationHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// Test data
	requestBody := map[string]interface{}{
		"email": "john.doe@example.com",
		"profile": map[string]interface{}{
			"name":        "John Doe",
			"bio":         "Test user",
			"avatar":      "https://example.com/avatar.jpg",
			"preferences": map[string]interface{}{"theme": "dark"},
		},
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Make HTTP request
	req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response["id"])
	assert.NotEmpty(t, response["webid"])
	assert.Equal(t, "john.doe@example.com", response["email"])
	assert.Equal(t, "active", response["status"])

	profile, ok := response["profile"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "John Doe", profile["name"])
	assert.Equal(t, "Test user", profile["bio"])

	userID := response["id"].(string)

	// Test duplicate registration (should fail)
	req2 := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	// Should return conflict or bad request
	assert.True(t, w2.Code == http.StatusConflict || w2.Code == http.StatusBadRequest)

	// Test getting the created user
	req3 := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%s", userID), nil)
	w3 := httptest.NewRecorder()
	server.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)

	var getUserResponse map[string]interface{}
	err = json.Unmarshal(w3.Body.Bytes(), &getUserResponse)
	require.NoError(t, err)

	assert.Equal(t, userID, getUserResponse["id"])
	assert.Equal(t, "john.doe@example.com", getUserResponse["email"])
}

// TestUserProfileUpdateHTTPWorkflow tests the complete user profile update workflow via HTTP
func TestUserProfileUpdateHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// First, create a user
	userID := createTestUserViaHTTP(t, server, "jane.doe@example.com", "Jane Doe")

	// Update profile
	updateBody := map[string]interface{}{
		"profile": map[string]interface{}{
			"name":        "Jane Smith",
			"bio":         "Updated bio",
			"avatar":      "https://example.com/new-avatar.jpg",
			"preferences": map[string]interface{}{"theme": "light", "language": "en"},
		},
	}

	jsonBody, err := json.Marshal(updateBody)
	require.NoError(t, err)

	// Make update request
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/users/%s/profile", userID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Profile updated successfully", response["message"])

	// Verify the update by getting the user
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%s", userID), nil)
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var getUserResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getUserResponse)
	require.NoError(t, err)

	profile, ok := getUserResponse["profile"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Jane Smith", profile["name"])
	assert.Equal(t, "Updated bio", profile["bio"])
	assert.Equal(t, "https://example.com/new-avatar.jpg", profile["avatar"])

	preferences, ok := profile["preferences"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "light", preferences["theme"])
	assert.Equal(t, "en", preferences["language"])
}

// TestUserDeletionHTTPWorkflow tests the complete user deletion workflow via HTTP
func TestUserDeletionHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// First, create a user
	userID := createTestUserViaHTTP(t, server, "delete.me@example.com", "Delete Me")

	// Delete user
	deleteBody := map[string]interface{}{
		"confirmation": "DELETE",
	}

	jsonBody, err := json.Marshal(deleteBody)
	require.NoError(t, err)

	// Make delete request
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/users/%s", userID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Try to get the deleted user (should fail or return deleted status)
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%s", userID), nil)
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	// Should return not found or show deleted status
	if w2.Code == http.StatusOK {
		var getUserResponse map[string]interface{}
		err = json.Unmarshal(w2.Body.Bytes(), &getUserResponse)
		require.NoError(t, err)
		assert.Equal(t, "deleted", getUserResponse["status"])
	} else {
		assert.Equal(t, http.StatusNotFound, w2.Code)
	}

	// Test invalid confirmation
	invalidDeleteBody := map[string]interface{}{
		"confirmation": "INVALID",
	}

	invalidJsonBody, err := json.Marshal(invalidDeleteBody)
	require.NoError(t, err)

	// Create another user for invalid deletion test
	userID2 := createTestUserViaHTTP(t, server, "invalid.delete@example.com", "Invalid Delete")

	req3 := httptest.NewRequest("DELETE", fmt.Sprintf("/api/users/%s", userID2), bytes.NewBuffer(invalidJsonBody))
	req3.Header.Set("Content-Type", "application/json")

	w3 := httptest.NewRecorder()
	server.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

// TestAccountCreationHTTPWorkflow tests the complete account creation workflow via HTTP
func TestAccountCreationHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// First, create an owner user
	ownerID := createTestUserViaHTTP(t, server, "owner@example.com", "Account Owner")

	// Create account
	accountBody := map[string]interface{}{
		"name":        "Test Account",
		"description": "Test account description",
	}

	jsonBody, err := json.Marshal(accountBody)
	require.NoError(t, err)

	// Make account creation request
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/owner/%s", ownerID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response["id"])
	assert.Equal(t, ownerID, response["owner_id"])
	assert.Equal(t, "Test Account", response["name"])
	assert.Equal(t, "Test account description", response["description"])

	settings, ok := response["settings"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, settings["allow_invitations"])
	assert.Equal(t, "member", settings["default_role_id"])
	assert.Equal(t, float64(100), settings["max_members"]) // JSON numbers are float64
}

// TestInvitationHTTPWorkflow tests the complete invitation workflow via HTTP
func TestInvitationHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// Create owner and account
	ownerID := createTestUserViaHTTP(t, server, "owner@example.com", "Account Owner")
	accountID := createTestAccountViaHTTP(t, server, ownerID, "Test Account")

	// Create invitee user
	inviteeID := createTestUserViaHTTP(t, server, "invitee@example.com", "Invitee User")

	// Send invitation
	inviteBody := map[string]interface{}{
		"email":   "invitee@example.com",
		"role_id": "member",
	}

	jsonBody, err := json.Marshal(inviteBody)
	require.NoError(t, err)

	// Make invitation request
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, ownerID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var inviteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &inviteResponse)
	require.NoError(t, err)

	// Verify invitation structure
	assert.NotEmpty(t, inviteResponse["id"])
	assert.Equal(t, accountID, inviteResponse["account_id"])
	assert.Equal(t, "invitee@example.com", inviteResponse["email"])
	assert.Equal(t, "member", inviteResponse["role_id"])
	assert.Equal(t, "pending", inviteResponse["status"])
	assert.Equal(t, ownerID, inviteResponse["invited_by"])

	token := inviteResponse["id"].(string) // Using ID as token for simplicity

	// Accept invitation
	acceptBody := map[string]interface{}{
		"token":   token,
		"user_id": inviteeID,
	}

	acceptJsonBody, err := json.Marshal(acceptBody)
	require.NoError(t, err)

	req2 := httptest.NewRequest("POST", "/api/invitations/accept", bytes.NewBuffer(acceptJsonBody))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	// Verify acceptance response
	assert.Equal(t, http.StatusOK, w2.Code)

	var acceptResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &acceptResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invitation accepted successfully", acceptResponse["message"])
}

// TestMemberRoleUpdateHTTPWorkflow tests the complete member role update workflow via HTTP
func TestMemberRoleUpdateHTTPWorkflow(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	// Create owner and account
	ownerID := createTestUserViaHTTP(t, server, "owner@example.com", "Account Owner")
	accountID := createTestAccountViaHTTP(t, server, ownerID, "Test Account")

	// Create member user and invite them
	memberID := createTestUserViaHTTP(t, server, "member@example.com", "Member User")
	inviteAndAcceptUser(t, server, accountID, ownerID, memberID, "member@example.com", "member")

	// Update member role
	updateRoleBody := map[string]interface{}{
		"role_id": "admin",
	}

	jsonBody, err := json.Marshal(updateRoleBody)
	require.NoError(t, err)

	// Make role update request
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/accounts/%s/members/%s/role", accountID, memberID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Member role updated successfully", response["message"])
}

// TestHTTPErrorScenarios tests various error scenarios across the system
func TestHTTPErrorScenarios(t *testing.T) {
	// Setup test server
	server, cleanup := setupTestHTTPServer(t)
	defer cleanup()

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_JSON", response["error"])
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"email": "", // Empty email
			"profile": map[string]interface{}{
				"name": "",
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response["error"])
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/nonexistent-user-id", nil)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", response["error"])
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"email": "invalid-email",
			"profile": map[string]interface{}{
				"name": "Test User",
			},
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response["error"])
	})
}

// Helper functions

func setupTestHTTPServer(t *testing.T) (*mux.Router, func()) {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	// Seed system roles
	err = infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Setup file storage
	tempDir := t.TempDir()
	fileStorage := infrastructure.NewFileStorageAdapter(tempDir)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Setup WebID generator and invitation generator
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")
	inviteGen := &testInvitationGenerator{}

	// Setup event dispatcher and unit of work
	eventDispatcher := setupAccountTestEventDispatcher(t, userWriteRepo, accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup services
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)
	accountService := application.NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo)

	// Setup logger
	logger := log.NewStdLogger(log.NewFilter(log.NewStdLogger(nil), log.FilterLevel(log.LevelDebug)))

	// Setup handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	accountHandler := handlers.NewAccountHandler(accountService, userService, logger)

	// Setup router
	router := mux.NewRouter()

	// User routes
	router.HandleFunc("/api/users/register", adaptHandler(userHandler.RegisterUser)).Methods("POST")
	router.HandleFunc("/api/users/{id}", adaptHandler(userHandler.GetUser)).Methods("GET")
	router.HandleFunc("/api/users/{id}/profile", adaptHandler(userHandler.UpdateProfile)).Methods("PUT")
	router.HandleFunc("/api/users/{id}", adaptHandler(userHandler.DeleteAccount)).Methods("DELETE")

	// Account routes
	router.HandleFunc("/api/accounts/owner/{owner_id}", adaptHandler(accountHandler.CreateAccount)).Methods("POST")
	router.HandleFunc("/api/accounts/{account_id}/invitations/inviter/{inviter_id}", adaptHandler(accountHandler.InviteUser)).Methods("POST")
	router.HandleFunc("/api/invitations/accept", adaptHandler(accountHandler.AcceptInvitation)).Methods("POST")
	router.HandleFunc("/api/accounts/{account_id}/members/{user_id}/role", adaptHandler(accountHandler.UpdateMemberRole)).Methods("PUT")

	cleanup := func() {
		// Cleanup resources if needed
	}

	return router, cleanup
}

// adaptHandler adapts kratos HTTP handler to standard HTTP handler
func adaptHandler(handler func(khttp.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a simple kratos context adapter
		ctx := &kratosContextAdapter{
			request:  r,
			response: w,
			vars:     mux.Vars(r),
		}

		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// kratosContextAdapter adapts standard HTTP to kratos HTTP context
type kratosContextAdapter struct {
	request  *http.Request
	response http.ResponseWriter
	vars     map[string]string
}

func (c *kratosContextAdapter) Request() *http.Request {
	return c.request
}

func (c *kratosContextAdapter) Response() http.ResponseWriter {
	return c.response
}

func (c *kratosContextAdapter) Vars() map[string][]string {
	result := make(map[string][]string)
	for k, v := range c.vars {
		result[k] = []string{v}
	}
	return result
}

func (c *kratosContextAdapter) JSON(code int, v interface{}) error {
	c.response.Header().Set("Content-Type", "application/json")
	c.response.WriteHeader(code)
	return json.NewEncoder(c.response).Encode(v)
}

func createTestUserViaHTTP(t *testing.T, server *mux.Router, email, name string) string {
	requestBody := map[string]interface{}{
		"email": email,
		"profile": map[string]interface{}{
			"name":        name,
			"bio":         "Test user",
			"avatar":      "https://example.com/avatar.jpg",
			"preferences": map[string]interface{}{"theme": "dark"},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/users/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return response["id"].(string)
}

func createTestAccountViaHTTP(t *testing.T, server *mux.Router, ownerID, name string) string {
	accountBody := map[string]interface{}{
		"name":        name,
		"description": "Test account",
	}

	jsonBody, err := json.Marshal(accountBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/owner/%s", ownerID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return response["id"].(string)
}

func inviteAndAcceptUser(t *testing.T, server *mux.Router, accountID, inviterID, userID, email, roleID string) {
	// Send invitation
	inviteBody := map[string]interface{}{
		"email":   email,
		"role_id": roleID,
	}

	jsonBody, err := json.Marshal(inviteBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, inviterID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var inviteResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &inviteResponse)
	require.NoError(t, err)

	token := inviteResponse["id"].(string)

	// Accept invitation
	acceptBody := map[string]interface{}{
		"token":   token,
		"user_id": userID,
	}

	acceptJsonBody, err := json.Marshal(acceptBody)
	require.NoError(t, err)

	req2 := httptest.NewRequest("POST", "/api/invitations/accept", bytes.NewBuffer(acceptJsonBody))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
}
