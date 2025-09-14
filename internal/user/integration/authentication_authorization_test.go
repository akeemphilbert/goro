package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAuthenticationRequiredEndpoints tests that protected endpoints require authentication
func TestAuthenticationRequiredEndpoints(t *testing.T) {
	// Setup test server with authentication middleware
	server, cleanup := setupAuthenticatedHTTPServer(t)
	defer cleanup()

	// Test cases for endpoints that should require authentication
	testCases := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
	}{
		{
			name:   "UpdateProfile",
			method: "PUT",
			path:   "/api/users/test-user-id/profile",
			body: map[string]interface{}{
				"profile": map[string]interface{}{
					"name": "Updated Name",
				},
			},
		},
		{
			name:   "DeleteAccount",
			method: "DELETE",
			path:   "/api/users/test-user-id",
			body: map[string]interface{}{
				"confirmation": "DELETE",
			},
		},
		{
			name:   "CreateAccount",
			method: "POST",
			path:   "/api/accounts/owner/test-owner-id",
			body: map[string]interface{}{
				"name": "Test Account",
			},
		},
		{
			name:   "InviteUser",
			method: "POST",
			path:   "/api/accounts/test-account-id/invitations/inviter/test-inviter-id",
			body: map[string]interface{}{
				"email":   "test@example.com",
				"role_id": "member",
			},
		},
		{
			name:   "UpdateMemberRole",
			method: "PUT",
			path:   "/api/accounts/test-account-id/members/test-user-id/role",
			body: map[string]interface{}{
				"role_id": "admin",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			if tc.body != nil {
				jsonBody, err := json.Marshal(tc.body)
				require.NoError(t, err)
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			// Don't set authentication header

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			// Should return 401 Unauthorized
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "UNAUTHORIZED", response["error"])
		})
	}
}

// TestRoleBasedAccessControl tests role-based access control for protected endpoints
func TestRoleBasedAccessControl(t *testing.T) {
	// Setup test server
	server, cleanup := setupAuthenticatedHTTPServer(t)
	defer cleanup()

	// Create test users with different roles
	ownerID := createAuthenticatedUserViaHTTP(t, server, "owner@example.com", "Owner User")
	adminID := createAuthenticatedUserViaHTTP(t, server, "admin@example.com", "Admin User")
	memberID := createAuthenticatedUserViaHTTP(t, server, "member@example.com", "Member User")
	viewerID := createAuthenticatedUserViaHTTP(t, server, "viewer@example.com", "Viewer User")

	// Create account with owner
	accountID := createAuthenticatedAccountViaHTTP(t, server, ownerID, "Test Account")

	// Add users to account with different roles
	inviteAndAcceptAuthenticatedUser(t, server, accountID, ownerID, adminID, "admin@example.com", "admin")
	inviteAndAcceptAuthenticatedUser(t, server, accountID, ownerID, memberID, "member@example.com", "member")
	inviteAndAcceptAuthenticatedUser(t, server, accountID, ownerID, viewerID, "viewer@example.com", "viewer")

	t.Run("OwnerCanInviteUsers", func(t *testing.T) {
		// Owner should be able to invite users
		inviteBody := map[string]interface{}{
			"email":   "newuser@example.com",
			"role_id": "member",
		}

		jsonBody, err := json.Marshal(inviteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, ownerID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(ownerID, "owner")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("AdminCanInviteUsers", func(t *testing.T) {
		// Admin should be able to invite users
		inviteBody := map[string]interface{}{
			"email":   "newuser2@example.com",
			"role_id": "member",
		}

		jsonBody, err := json.Marshal(inviteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, adminID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(adminID, "admin")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("MemberCannotInviteUsers", func(t *testing.T) {
		// Member should not be able to invite users
		inviteBody := map[string]interface{}{
			"email":   "newuser3@example.com",
			"role_id": "member",
		}

		jsonBody, err := json.Marshal(inviteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, memberID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(memberID, "member")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "FORBIDDEN", response["error"])
	})

	t.Run("ViewerCannotInviteUsers", func(t *testing.T) {
		// Viewer should not be able to invite users
		inviteBody := map[string]interface{}{
			"email":   "newuser4@example.com",
			"role_id": "member",
		}

		jsonBody, err := json.Marshal(inviteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, viewerID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(viewerID, "viewer")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("OwnerCanUpdateMemberRoles", func(t *testing.T) {
		// Owner should be able to update member roles
		updateBody := map[string]interface{}{
			"role_id": "admin",
		}

		jsonBody, err := json.Marshal(updateBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/accounts/%s/members/%s/role", accountID, memberID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(ownerID, "owner")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MemberCannotUpdateRoles", func(t *testing.T) {
		// Member should not be able to update roles
		updateBody := map[string]interface{}{
			"role_id": "admin",
		}

		jsonBody, err := json.Marshal(updateBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/accounts/%s/members/%s/role", accountID, viewerID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(memberID, "member")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestUserCanOnlyAccessOwnData tests that users can only access their own data
func TestUserCanOnlyAccessOwnData(t *testing.T) {
	// Setup test server
	server, cleanup := setupAuthenticatedHTTPServer(t)
	defer cleanup()

	// Create two users
	user1ID := createAuthenticatedUserViaHTTP(t, server, "user1@example.com", "User One")
	user2ID := createAuthenticatedUserViaHTTP(t, server, "user2@example.com", "User Two")

	t.Run("UserCanAccessOwnProfile", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%s", user1ID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(user1ID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user1ID, response["id"])
	})

	t.Run("UserCannotAccessOtherProfile", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/users/%s", user2ID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(user1ID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "FORBIDDEN", response["error"])
	})

	t.Run("UserCanUpdateOwnProfile", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "Updated Name",
				"bio":  "Updated bio",
			},
		}

		jsonBody, err := json.Marshal(updateBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/users/%s/profile", user1ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(user1ID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UserCannotUpdateOtherProfile", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "Malicious Update",
			},
		}

		jsonBody, err := json.Marshal(updateBody)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/users/%s/profile", user2ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(user1ID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UserCanDeleteOwnAccount", func(t *testing.T) {
		// Create a user specifically for deletion test
		deleteUserID := createAuthenticatedUserViaHTTP(t, server, "delete@example.com", "Delete User")

		deleteBody := map[string]interface{}{
			"confirmation": "DELETE",
		}

		jsonBody, err := json.Marshal(deleteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/users/%s", deleteUserID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(deleteUserID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("UserCannotDeleteOtherAccount", func(t *testing.T) {
		deleteBody := map[string]interface{}{
			"confirmation": "DELETE",
		}

		jsonBody, err := json.Marshal(deleteBody)
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/users/%s", user2ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(user1ID, "user")))

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestInvalidTokenHandling tests handling of invalid authentication tokens
func TestInvalidTokenHandling(t *testing.T) {
	// Setup test server
	server, cleanup := setupAuthenticatedHTTPServer(t)
	defer cleanup()

	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "InvalidToken",
			token: "invalid-token",
		},
		{
			name:  "ExpiredToken",
			token: generateExpiredTestToken("user-id", "user"),
		},
		{
			name:  "MalformedToken",
			token: "malformed.token.here",
		},
		{
			name:  "EmptyToken",
			token: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/users/test-user-id", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			}

			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "UNAUTHORIZED", response["error"])
		})
	}
}

// Helper functions

func setupAuthenticatedHTTPServer(t *testing.T) (*mux.Router, func()) {
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

	// Setup handlers with authentication
	userHandler := handlers.NewUserHandler(userService, logger)
	accountHandler := handlers.NewAccountHandler(accountService, userService, logger)

	// Setup router with authentication middleware
	router := mux.NewRouter()

	// Public routes (no authentication required)
	router.HandleFunc("/api/users/register", adaptHandler(userHandler.RegisterUser)).Methods("POST")

	// Protected routes (authentication required)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(authenticationMiddleware(userRepo))

	// User routes
	protected.HandleFunc("/users/{id}", adaptHandlerWithAuth(userHandler.GetUser)).Methods("GET")
	protected.HandleFunc("/users/{id}/profile", adaptHandlerWithAuth(userHandler.UpdateProfile)).Methods("PUT")
	protected.HandleFunc("/users/{id}", adaptHandlerWithAuth(userHandler.DeleteAccount)).Methods("DELETE")

	// Account routes
	protected.HandleFunc("/accounts/owner/{owner_id}", adaptHandlerWithAuth(accountHandler.CreateAccount)).Methods("POST")
	protected.HandleFunc("/accounts/{account_id}/invitations/inviter/{inviter_id}", adaptHandlerWithAuth(accountHandler.InviteUser)).Methods("POST")
	protected.HandleFunc("/invitations/accept", adaptHandlerWithAuth(accountHandler.AcceptInvitation)).Methods("POST")
	protected.HandleFunc("/accounts/{account_id}/members/{user_id}/role", adaptHandlerWithAuth(accountHandler.UpdateMemberRole)).Methods("PUT")

	cleanup := func() {
		// Cleanup resources if needed
	}

	return router, cleanup
}

// authenticationMiddleware provides basic authentication middleware for testing
func authenticationMiddleware(userRepo domain.UserRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"Authentication required"}`, http.StatusUnauthorized)
				return
			}

			// Simple Bearer token validation
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"Invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			token := authHeader[7:]

			// Validate token (simplified for testing)
			userID, role, valid := validateTestToken(token)
			if !valid {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"Invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Check if user exists
			ctx := context.Background()
			user, err := userRepo.GetByID(ctx, userID)
			if err != nil {
				http.Error(w, `{"error":"UNAUTHORIZED","message":"User not found"}`, http.StatusUnauthorized)
				return
			}

			// Add user info to request context
			ctx = context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "user_role", role)
			ctx = context.WithValue(ctx, "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// adaptHandlerWithAuth adapts kratos HTTP handler with authentication context
func adaptHandlerWithAuth(handler func(khttp.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check authorization for resource access
		userID := r.Context().Value("user_id").(string)
		userRole := r.Context().Value("user_role").(string)

		// Simple authorization check
		if !isAuthorized(r, userID, userRole) {
			http.Error(w, `{"error":"FORBIDDEN","message":"Insufficient permissions"}`, http.StatusForbidden)
			return
		}

		// Create kratos context adapter
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

// isAuthorized performs simple authorization checks
func isAuthorized(r *http.Request, userID, userRole string) bool {
	vars := mux.Vars(r)

	// Check if user is accessing their own resources
	if resourceUserID, exists := vars["id"]; exists {
		if resourceUserID != userID && userRole != "admin" && userRole != "owner" {
			return false
		}
	}

	// Check role-based permissions for account operations
	if r.URL.Path == "/api/invitations/accept" {
		return true // Anyone can accept invitations
	}

	if r.Method == "POST" && (r.URL.Path == "/api/accounts" || r.URL.Path == "/api/invitations") {
		return userRole == "owner" || userRole == "admin"
	}

	if r.Method == "PUT" && r.URL.Path == "/api/members" {
		return userRole == "owner" || userRole == "admin"
	}

	return true
}

// generateTestToken generates a test JWT-like token
func generateTestToken(userID, role string) string {
	return fmt.Sprintf("test-token-%s-%s", userID, role)
}

// generateExpiredTestToken generates an expired test token
func generateExpiredTestToken(userID, role string) string {
	return fmt.Sprintf("expired-test-token-%s-%s", userID, role)
}

// validateTestToken validates a test token and returns user info
func validateTestToken(token string) (userID, role string, valid bool) {
	if token == "" {
		return "", "", false
	}

	// Handle expired tokens
	if len(token) > 7 && token[:7] == "expired" {
		return "", "", false
	}

	// Simple token format: test-token-{userID}-{role}
	if len(token) < 11 || token[:11] != "test-token-" {
		return "", "", false
	}

	parts := token[11:] // Remove "test-token-" prefix

	// Find the last dash to separate userID and role
	lastDash := -1
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == '-' {
			lastDash = i
			break
		}
	}

	if lastDash == -1 {
		return "", "", false
	}

	userID = parts[:lastDash]
	role = parts[lastDash+1:]

	return userID, role, true
}

func createAuthenticatedUserViaHTTP(t *testing.T, server *mux.Router, email, name string) string {
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

func createAuthenticatedAccountViaHTTP(t *testing.T, server *mux.Router, ownerID, name string) string {
	accountBody := map[string]interface{}{
		"name":        name,
		"description": "Test account",
	}

	jsonBody, err := json.Marshal(accountBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/owner/%s", ownerID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(ownerID, "owner")))

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return response["id"].(string)
}

func inviteAndAcceptAuthenticatedUser(t *testing.T, server *mux.Router, accountID, inviterID, userID, email, roleID string) {
	// Send invitation
	inviteBody := map[string]interface{}{
		"email":   email,
		"role_id": roleID,
	}

	jsonBody, err := json.Marshal(inviteBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/accounts/%s/invitations/inviter/%s", accountID, inviterID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(inviterID, "owner")))

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
	req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generateTestToken(userID, "user")))

	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
}
