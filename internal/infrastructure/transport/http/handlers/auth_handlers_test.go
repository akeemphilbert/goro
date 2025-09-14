package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Mock implementations for testing

type mockAuthService struct {
	sessions map[string]*domain.Session
	users    map[string]string // userID -> webID mapping
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		sessions: make(map[string]*domain.Session),
		users: map[string]string{
			"user1": "https://example.com/user1#me",
			"user2": "https://example.com/user2#me",
		},
	}
}

func (m *mockAuthService) AuthenticateWithPassword(ctx context.Context, username, password string) (*domain.Session, error) {
	if username == "test@example.com" && password == "password123" {
		session := &domain.Session{
			ID:           "session1",
			UserID:       "user1",
			WebID:        "https://example.com/user1#me",
			TokenHash:    "hash1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		m.sessions[session.ID] = session
		return session, nil
	}
	return nil, domain.ErrInvalidCredentials
}

func (m *mockAuthService) AuthenticateWithWebID(ctx context.Context, webID string, oidcToken string) (*domain.Session, error) {
	if webID == "https://example.com/user1#me" && oidcToken == "valid_oidc_token" {
		session := &domain.Session{
			ID:           "session2",
			UserID:       "user1",
			WebID:        webID,
			TokenHash:    "hash2",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		m.sessions[session.ID] = session
		return session, nil
	}
	return nil, domain.ErrWebIDValidationFailed
}

func (m *mockAuthService) AuthenticateWithOAuth(ctx context.Context, provider string, oauthCode string) (*domain.Session, error) {
	if provider == "google" && oauthCode == "valid_oauth_code" {
		session := &domain.Session{
			ID:           "session3",
			UserID:       "user2",
			WebID:        "https://example.com/user2#me",
			TokenHash:    "hash3",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		m.sessions[session.ID] = session
		return session, nil
	}
	return nil, domain.ErrExternalAuthFailed
}

func (m *mockAuthService) ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, domain.ErrSessionNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		delete(m.sessions, sessionID)
		return nil, domain.ErrSessionExpired
	}
	return session, nil
}

func (m *mockAuthService) RefreshSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, err := m.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return session, nil
}

func (m *mockAuthService) Logout(ctx context.Context, sessionID string) error {
	if _, exists := m.sessions[sessionID]; !exists {
		return domain.ErrSessionNotFound
	}
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockAuthService) LogoutAllSessions(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockAuthService) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	var sessions []*domain.Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *mockAuthService) CleanupExpiredSessions(ctx context.Context) error {
	for id, session := range m.sessions {
		if time.Now().After(session.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
	return nil
}

type mockPasswordService struct {
	resetTokens map[string]string // token -> userID
}

func newMockPasswordService() *mockPasswordService {
	return &mockPasswordService{
		resetTokens: make(map[string]string),
	}
}

func (m *mockPasswordService) SetPassword(ctx context.Context, userID, password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password validation failed: password too short")
	}
	return nil
}

func (m *mockPasswordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if currentPassword != "oldpassword" {
		return domain.ErrCurrentPasswordInvalid
	}
	return m.SetPassword(ctx, userID, newPassword)
}

func (m *mockPasswordService) InitiatePasswordReset(ctx context.Context, email string) error {
	if email == "test@example.com" {
		m.resetTokens["reset_token_123"] = "user1"
	}
	return nil
}

func (m *mockPasswordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	if token == "reset_token_123" {
		delete(m.resetTokens, token)
		return m.SetPassword(ctx, "user1", newPassword)
	}
	if token == "expired_token" {
		return domain.ErrPasswordResetExpired
	}
	if token == "used_token" {
		return domain.ErrPasswordResetUsed
	}
	return domain.ErrPasswordResetInvalid
}

func (m *mockPasswordService) ValidatePassword(ctx context.Context, userID, password string) error {
	if password == "password123" {
		return nil
	}
	return domain.ErrInvalidCredentials
}

func (m *mockPasswordService) CleanupExpiredTokens(ctx context.Context) error {
	return nil
}

func (m *mockPasswordService) HasPassword(ctx context.Context, userID string) (bool, error) {
	return true, nil
}

type mockRegistrationService struct{}

func newMockRegistrationService() *mockRegistrationService {
	return &mockRegistrationService{}
}

type mockTokenManager struct{}

func newMockTokenManager() *mockTokenManager {
	return &mockTokenManager{}
}

func (m *mockTokenManager) GenerateToken(ctx context.Context, session *domain.Session) (string, error) {
	return fmt.Sprintf("token_%s", session.ID), nil
}

func (m *mockTokenManager) ValidateToken(ctx context.Context, token string) (*application.TokenClaims, error) {
	if strings.HasPrefix(token, "token_") {
		sessionID := strings.TrimPrefix(token, "token_")
		return &application.TokenClaims{
			SessionID: sessionID,
			UserID:    "user1",
			WebID:     "https://example.com/user1#me",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IssuedAt:  time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (m *mockTokenManager) RefreshToken(ctx context.Context, token string) (string, error) {
	claims, err := m.ValidateToken(ctx, token)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("refreshed_%s", token), nil
}

type mockOAuthProvider struct {
	name string
}

func newMockOAuthProvider(name string) *mockOAuthProvider {
	return &mockOAuthProvider{name: name}
}

func (m *mockOAuthProvider) GetProviderName() string {
	return m.name
}

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://%s.com/oauth/authorize?state=%s", m.name, state)
}

func (m *mockOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthToken, error) {
	if code == "valid_oauth_code" {
		return &domain.OAuthToken{
			AccessToken:  "access_token_123",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			RefreshToken: "refresh_token_123",
		}, nil
	}
	return nil, fmt.Errorf("invalid code")
}

func (m *mockOAuthProvider) GetUserProfile(ctx context.Context, token *domain.OAuthToken) (*domain.ExternalProfile, error) {
	if token.AccessToken == "access_token_123" {
		return &domain.ExternalProfile{
			ID:    "external_user_123",
			Name:  "Test User",
			Email: "test@example.com",
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// Helper function to create test context
func createTestContext(method, path string, body interface{}) (khttp.Context, *httptest.ResponseRecorder) {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// Create a mock kratos context
	ctx := &mockKratosContext{
		request:  req,
		recorder: recorder,
		vars:     make(map[string][]string),
	}

	return ctx, recorder
}

type mockKratosContext struct {
	request  *http.Request
	recorder *httptest.ResponseRecorder
	vars     map[string][]string
}

func (m *mockKratosContext) Request() *http.Request {
	return m.request
}

func (m *mockKratosContext) JSON(code int, v interface{}) error {
	m.recorder.WriteHeader(code)
	return json.NewEncoder(m.recorder).Encode(v)
}

func (m *mockKratosContext) Vars() map[string][]string {
	return m.vars
}

func (m *mockKratosContext) SetVar(key string, value string) {
	m.vars[key] = []string{value}
}

// Test cases

func TestAuthHandler_Login_Password(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		request        LoginRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful password login",
			request: LoginRequest{
				Method:   "password",
				Username: "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			request: LoginRequest{
				Method:   "password",
				Username: "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
		{
			name: "missing username",
			request: LoginRequest{
				Method:   "password",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing password",
			request: LoginRequest{
				Method:   "password",
				Username: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing method",
			request: LoginRequest{
				Username: "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("POST", "/auth/login", tt.request)

			err := handler.Login(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			} else {
				var response LoginResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.SessionID == "" {
					t.Error("Expected session ID in response")
				}
				if response.AccessToken == "" {
					t.Error("Expected access token in response")
				}
				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type Bearer, got %s", response.TokenType)
				}
			}
		})
	}
}

func TestAuthHandler_Login_WebIDOIDC(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		request        LoginRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful WebID-OIDC login",
			request: LoginRequest{
				Method: "webid-oidc",
				WebID:  "https://example.com/user1#me",
				Token:  "valid_oidc_token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid WebID",
			request: LoginRequest{
				Method: "webid-oidc",
				WebID:  "https://example.com/unknown#me",
				Token:  "valid_oidc_token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "WEBID_VALIDATION_FAILED",
		},
		{
			name: "invalid token",
			request: LoginRequest{
				Method: "webid-oidc",
				WebID:  "https://example.com/user1#me",
				Token:  "invalid_token",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "WEBID_VALIDATION_FAILED",
		},
		{
			name: "missing WebID",
			request: LoginRequest{
				Method: "webid-oidc",
				Token:  "valid_oidc_token",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "invalid WebID format",
			request: LoginRequest{
				Method: "webid-oidc",
				WebID:  "not-a-url",
				Token:  "valid_oidc_token",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("POST", "/auth/login", tt.request)

			err := handler.Login(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	// Create a session first
	session := &domain.Session{
		ID:           "test_session",
		UserID:       "user1",
		WebID:        "https://example.com/user1#me",
		TokenHash:    "hash1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	authService.sessions[session.ID] = session

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful logout",
			sessionID:      "test_session",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "session not found",
			sessionID:      "nonexistent_session",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "SESSION_NOT_FOUND",
		},
		{
			name:           "missing session ID",
			sessionID:      "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_SESSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("POST", "/auth/logout", nil)

			// Set Authorization header if session ID provided
			if tt.sessionID != "" {
				ctx.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.sessionID))
			}

			err := handler.Logout(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestAuthHandler_InitiatePasswordReset(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		request        PasswordResetRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful password reset initiation",
			request: PasswordResetRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "nonexistent email (should still return success for security)",
			request: PasswordResetRequest{
				Email: "nonexistent@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing email",
			request: PasswordResetRequest{
				Email: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "invalid email format",
			request: PasswordResetRequest{
				Email: "not-an-email",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("POST", "/auth/password-reset", tt.request)

			err := handler.InitiatePasswordReset(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestAuthHandler_CompletePasswordReset(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		request        PasswordResetCompleteRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful password reset completion",
			request: PasswordResetCompleteRequest{
				Token:       "reset_token_123",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			request: PasswordResetCompleteRequest{
				Token:       "invalid_token",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_RESET_TOKEN",
		},
		{
			name: "expired token",
			request: PasswordResetCompleteRequest{
				Token:       "expired_token",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "RESET_TOKEN_EXPIRED",
		},
		{
			name: "used token",
			request: PasswordResetCompleteRequest{
				Token:       "used_token",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "RESET_TOKEN_USED",
		},
		{
			name: "weak password",
			request: PasswordResetCompleteRequest{
				Token:       "reset_token_123",
				NewPassword: "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "WEAK_PASSWORD",
		},
		{
			name: "missing token",
			request: PasswordResetCompleteRequest{
				Token:       "",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing password",
			request: PasswordResetCompleteRequest{
				Token:       "reset_token_123",
				NewPassword: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("POST", "/auth/password-reset/complete", tt.request)

			err := handler.CompletePasswordReset(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestAuthHandler_GetOAuthAuthURL(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := map[string]application.OAuthProvider{
		"google": newMockOAuthProvider("google"),
		"github": newMockOAuthProvider("github"),
	}
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		provider       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful Google OAuth URL",
			provider:       "google",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful GitHub OAuth URL",
			provider:       "github",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unsupported provider",
			provider:       "facebook",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "UNSUPPORTED_PROVIDER",
		},
		{
			name:           "empty provider",
			provider:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_PROVIDER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/auth/oauth/%s/url", tt.provider)
			ctx, recorder := createTestContext("GET", path, nil)

			// Set provider in vars
			if tt.provider != "" {
				ctx.(*mockKratosContext).SetVar("provider", tt.provider)
			}

			err := handler.GetOAuthAuthURL(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			} else {
				var response OAuthAuthURLResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.AuthURL == "" {
					t.Error("Expected auth URL in response")
				}
				if response.State == "" {
					t.Error("Expected state in response")
				}

				expectedURL := fmt.Sprintf("https://%s.com/oauth/authorize", tt.provider)
				if !strings.Contains(response.AuthURL, expectedURL) {
					t.Errorf("Expected auth URL to contain %s, got %s", expectedURL, response.AuthURL)
				}
			}
		})
	}
}

func TestAuthHandler_HandleOAuthCallback(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := map[string]application.OAuthProvider{
		"google": newMockOAuthProvider("google"),
	}
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	tests := []struct {
		name           string
		provider       string
		code           string
		state          string
		errorParam     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful OAuth callback",
			provider:       "google",
			code:           "valid_oauth_code",
			state:          "state123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OAuth error",
			provider:       "google",
			errorParam:     "access_denied",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "OAUTH_ERROR",
		},
		{
			name:           "missing code",
			provider:       "google",
			state:          "state123",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_CODE",
		},
		{
			name:           "missing state",
			provider:       "google",
			code:           "valid_oauth_code",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_STATE",
		},
		{
			name:           "invalid code",
			provider:       "google",
			code:           "invalid_code",
			state:          "state123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "EXTERNAL_AUTH_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build URL with query parameters
			u, _ := url.Parse(fmt.Sprintf("/auth/oauth/%s/callback", tt.provider))
			q := u.Query()
			if tt.code != "" {
				q.Set("code", tt.code)
			}
			if tt.state != "" {
				q.Set("state", tt.state)
			}
			if tt.errorParam != "" {
				q.Set("error", tt.errorParam)
			}
			u.RawQuery = q.Encode()

			ctx, recorder := createTestContext("GET", u.String(), nil)

			// Set provider in vars
			ctx.(*mockKratosContext).SetVar("provider", tt.provider)

			err := handler.HandleOAuthCallback(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			} else {
				var response LoginResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.SessionID == "" {
					t.Error("Expected session ID in response")
				}
				if response.AccessToken == "" {
					t.Error("Expected access token in response")
				}
			}
		})
	}
}

func TestAuthHandler_ValidateSession(t *testing.T) {
	// Setup
	authService := newMockAuthService()
	passwordService := newMockPasswordService()
	registrationService := newMockRegistrationService()
	tokenManager := newMockTokenManager()
	oauthProviders := make(map[string]application.OAuthProvider)
	logger := log.NewStdLogger(log.NewFilter(log.NewHelper(log.DefaultLogger)))

	handler := NewAuthHandler(
		authService,
		passwordService,
		registrationService,
		tokenManager,
		oauthProviders,
		logger,
	)

	// Create a valid session
	session := &domain.Session{
		ID:           "valid_session",
		UserID:       "user1",
		WebID:        "https://example.com/user1#me",
		TokenHash:    "hash1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	authService.sessions[session.ID] = session

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid session",
			sessionID:      "valid_session",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid session",
			sessionID:      "invalid_session",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "SESSION_NOT_FOUND",
		},
		{
			name:           "missing session ID",
			sessionID:      "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_SESSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, recorder := createTestContext("GET", "/auth/validate", nil)

			// Set Authorization header if session ID provided
			if tt.sessionID != "" {
				ctx.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.sessionID))
			}

			err := handler.ValidateSession(ctx)
			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Expected error %s, got %v", tt.expectedError, response["error"])
				}
			} else {
				var response map[string]interface{}
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if valid, ok := response["valid"].(bool); !ok || !valid {
					t.Error("Expected valid session response")
				}
				if sessionID, ok := response["session_id"].(string); !ok || sessionID != tt.sessionID {
					t.Errorf("Expected session ID %s, got %v", tt.sessionID, response["session_id"])
				}
			}
		})
	}
}
