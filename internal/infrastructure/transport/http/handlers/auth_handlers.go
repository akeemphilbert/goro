package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService         *application.AuthenticationService
	passwordService     *application.PasswordService
	registrationService *application.RegistrationService
	tokenManager        application.TokenManager
	oauthProviders      map[string]application.OAuthProvider
	logger              log.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	authService *application.AuthenticationService,
	passwordService *application.PasswordService,
	registrationService *application.RegistrationService,
	tokenManager application.TokenManager,
	oauthProviders map[string]application.OAuthProvider,
	logger log.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		passwordService:     passwordService,
		registrationService: registrationService,
		tokenManager:        tokenManager,
		oauthProviders:      oauthProviders,
		logger:              logger,
	}
}

// Request/Response types

// LoginRequest represents the HTTP request for login
type LoginRequest struct {
	Method   string `json:"method"`   // "password", "webid-oidc"
	Username string `json:"username"` // For password method
	Password string `json:"password"` // For password method
	WebID    string `json:"webid"`    // For WebID-OIDC method
	Token    string `json:"token"`    // For WebID-OIDC method
}

// LoginResponse represents the HTTP response for successful login
type LoginResponse struct {
	SessionID   string    `json:"session_id"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	UserID      string    `json:"user_id"`
	WebID       string    `json:"webid"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// PasswordResetRequest represents the HTTP request for password reset
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetCompleteRequest represents the HTTP request for completing password reset
type PasswordResetCompleteRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordRequest represents the HTTP request for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// OAuthAuthURLResponse represents the response for OAuth authorization URL
type OAuthAuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// Login handles user login requests
func (h *AuthHandler) Login(ctx khttp.Context) error {
	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate request
	if err := h.validateLoginRequest(req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	var session *domain.Session
	var err error

	// Authenticate based on method
	switch req.Method {
	case "password":
		session, err = h.authService.AuthenticateWithPassword(ctx.Request().Context(), req.Username, req.Password)
	case "webid-oidc":
		session, err = h.authService.AuthenticateWithWebID(ctx.Request().Context(), req.WebID, req.Token)
	default:
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_METHOD", "Unsupported authentication method")
	}

	if err != nil {
		return h.handleAuthError(ctx, err)
	}

	// Generate access token
	accessToken, err := h.tokenManager.GenerateToken(ctx.Request().Context(), session)
	if err != nil {
		h.logger.Log(log.LevelError, "msg", "Failed to generate access token", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate access token")
	}

	// Build response
	response := LoginResponse{
		SessionID:   session.ID,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(session.ExpiresAt).Seconds()),
		UserID:      session.UserID,
		WebID:       session.WebID,
		ExpiresAt:   session.ExpiresAt,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Logout handles user logout requests
func (h *AuthHandler) Logout(ctx khttp.Context) error {
	// Extract session ID from Authorization header or query parameter
	sessionID := h.extractSessionID(ctx)
	if sessionID == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_SESSION", "Session ID is required")
	}

	// Logout
	if err := h.authService.Logout(ctx.Request().Context(), sessionID); err != nil {
		return h.handleAuthError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// InitiatePasswordReset handles password reset initiation requests
func (h *AuthHandler) InitiatePasswordReset(ctx khttp.Context) error {
	// Parse request body
	var req PasswordResetRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required")
	}

	if !isValidEmail(req.Email) {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid email format")
	}

	// Initiate password reset
	if err := h.passwordService.InitiatePasswordReset(ctx.Request().Context(), req.Email); err != nil {
		h.logger.Log(log.LevelError, "msg", "Failed to initiate password reset", "error", err.Error())
		// Don't reveal if email exists or not for security
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "If the email exists, a password reset link has been sent"})
}

// CompletePasswordReset handles password reset completion requests
func (h *AuthHandler) CompletePasswordReset(ctx khttp.Context) error {
	// Parse request body
	var req PasswordResetCompleteRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate request
	if strings.TrimSpace(req.Token) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "Reset token is required")
	}

	if strings.TrimSpace(req.NewPassword) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "New password is required")
	}

	// Complete password reset
	if err := h.passwordService.CompletePasswordReset(ctx.Request().Context(), req.Token, req.NewPassword); err != nil {
		return h.handlePasswordError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(ctx khttp.Context) error {
	// Extract user ID from authenticated context (this will be set by middleware)
	userID := h.extractUserID(ctx)
	if userID == "" {
		return h.handleError(ctx, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	}

	// Parse request body
	var req ChangePasswordRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate request
	if strings.TrimSpace(req.CurrentPassword) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "Current password is required")
	}

	if strings.TrimSpace(req.NewPassword) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", "New password is required")
	}

	// Change password
	if err := h.passwordService.ChangePassword(ctx.Request().Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		return h.handlePasswordError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// GetOAuthAuthURL handles OAuth authorization URL requests
func (h *AuthHandler) GetOAuthAuthURL(ctx khttp.Context) error {
	// Extract provider from path
	vars := ctx.Vars()
	providerSlice, exists := vars["provider"]
	if !exists || len(providerSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_PROVIDER", "OAuth provider is required")
	}

	provider := providerSlice[0]
	if strings.TrimSpace(provider) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_PROVIDER", "OAuth provider is required")
	}

	// Get OAuth provider
	oauthProvider, exists := h.oauthProviders[provider]
	if !exists {
		return h.handleError(ctx, http.StatusBadRequest, "UNSUPPORTED_PROVIDER", fmt.Sprintf("Unsupported OAuth provider: %s", provider))
	}

	// Generate state parameter for CSRF protection
	state, err := generateSecureState()
	if err != nil {
		h.logger.Log(log.LevelError, "msg", "Failed to generate OAuth state", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "STATE_GENERATION_FAILED", "Failed to generate OAuth state")
	}

	// Get authorization URL
	authURL := oauthProvider.GetAuthURL(state)

	// Store state in session/cookie for validation (simplified for now)
	// In production, you'd want to store this securely

	response := OAuthAuthURLResponse{
		AuthURL: authURL,
		State:   state,
	}

	return ctx.JSON(http.StatusOK, response)
}

// HandleOAuthCallback handles OAuth provider callback requests
func (h *AuthHandler) HandleOAuthCallback(ctx khttp.Context) error {
	// Extract provider from path
	vars := ctx.Vars()
	providerSlice, exists := vars["provider"]
	if !exists || len(providerSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_PROVIDER", "OAuth provider is required")
	}

	provider := providerSlice[0]

	// Get query parameters
	query := ctx.Request().URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errorParam := query.Get("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDescription := query.Get("error_description")
		h.logger.Log(log.LevelWarn, "msg", "OAuth error", "provider", provider, "error", errorParam, "description", errorDescription)
		return h.handleError(ctx, http.StatusBadRequest, "OAUTH_ERROR", fmt.Sprintf("OAuth error: %s", errorParam))
	}

	// Validate required parameters
	if code == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_CODE", "Authorization code is required")
	}

	if state == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_STATE", "State parameter is required")
	}

	// TODO: Validate state parameter against stored value for CSRF protection

	// Authenticate with OAuth
	session, err := h.authService.AuthenticateWithOAuth(ctx.Request().Context(), provider, code)
	if err != nil {
		return h.handleAuthError(ctx, err)
	}

	// Generate access token
	accessToken, err := h.tokenManager.GenerateToken(ctx.Request().Context(), session)
	if err != nil {
		h.logger.Log(log.LevelError, "msg", "Failed to generate access token", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate access token")
	}

	// Build response
	response := LoginResponse{
		SessionID:   session.ID,
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(session.ExpiresAt).Seconds()),
		UserID:      session.UserID,
		WebID:       session.WebID,
		ExpiresAt:   session.ExpiresAt,
	}

	return ctx.JSON(http.StatusOK, response)
}

// ValidateSession handles session validation requests
func (h *AuthHandler) ValidateSession(ctx khttp.Context) error {
	// Extract session ID
	sessionID := h.extractSessionID(ctx)
	if sessionID == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_SESSION", "Session ID is required")
	}

	// Validate session
	session, err := h.authService.ValidateSession(ctx.Request().Context(), sessionID)
	if err != nil {
		return h.handleAuthError(ctx, err)
	}

	// Return session info
	response := map[string]interface{}{
		"valid":      true,
		"session_id": session.ID,
		"user_id":    session.UserID,
		"webid":      session.WebID,
		"expires_at": session.ExpiresAt,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Helper methods

func (h *AuthHandler) validateLoginRequest(req LoginRequest) error {
	if strings.TrimSpace(req.Method) == "" {
		return fmt.Errorf("authentication method is required")
	}

	switch req.Method {
	case "password":
		if strings.TrimSpace(req.Username) == "" {
			return fmt.Errorf("username is required for password authentication")
		}
		if strings.TrimSpace(req.Password) == "" {
			return fmt.Errorf("password is required for password authentication")
		}
	case "webid-oidc":
		if strings.TrimSpace(req.WebID) == "" {
			return fmt.Errorf("WebID is required for WebID-OIDC authentication")
		}
		if strings.TrimSpace(req.Token) == "" {
			return fmt.Errorf("OIDC token is required for WebID-OIDC authentication")
		}
		// Validate WebID format
		if _, err := url.Parse(req.WebID); err != nil {
			return fmt.Errorf("invalid WebID format")
		}
	default:
		return fmt.Errorf("unsupported authentication method: %s", req.Method)
	}

	return nil
}

func (h *AuthHandler) extractSessionID(ctx khttp.Context) string {
	// Try Authorization header first
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader != "" {
		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try query parameter
	return ctx.Request().URL.Query().Get("session_id")
}

func (h *AuthHandler) extractUserID(ctx khttp.Context) string {
	// This will be set by authentication middleware
	if userID, ok := ctx.Request().Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func (h *AuthHandler) handleError(ctx khttp.Context, status int, code, message string) error {
	response := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	return ctx.JSON(status, response)
}

func (h *AuthHandler) handleAuthError(ctx khttp.Context, err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case err == domain.ErrInvalidCredentials:
		return h.handleError(ctx, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid credentials")
	case err == domain.ErrSessionExpired:
		return h.handleError(ctx, http.StatusUnauthorized, "SESSION_EXPIRED", "Session has expired")
	case err == domain.ErrSessionNotFound:
		return h.handleError(ctx, http.StatusUnauthorized, "SESSION_NOT_FOUND", "Session not found")
	case err == domain.ErrWebIDValidationFailed:
		return h.handleError(ctx, http.StatusUnauthorized, "WEBID_VALIDATION_FAILED", "WebID validation failed")
	case strings.Contains(errMsg, "external authentication failed"):
		return h.handleError(ctx, http.StatusUnauthorized, "EXTERNAL_AUTH_FAILED", "External authentication failed")
	case strings.Contains(errMsg, "external identity not linked"):
		return h.handleError(ctx, http.StatusUnauthorized, "IDENTITY_NOT_LINKED", "External identity not linked to any user")
	case strings.Contains(errMsg, "user not found"):
		return h.handleError(ctx, http.StatusUnauthorized, "USER_NOT_FOUND", "User not found")
	default:
		h.logger.Log(log.LevelError, "msg", "Authentication error", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func (h *AuthHandler) handlePasswordError(ctx khttp.Context, err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case err == domain.ErrPasswordResetInvalid:
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_RESET_TOKEN", "Invalid or expired reset token")
	case err == domain.ErrPasswordResetExpired:
		return h.handleError(ctx, http.StatusBadRequest, "RESET_TOKEN_EXPIRED", "Reset token has expired")
	case err == domain.ErrPasswordResetUsed:
		return h.handleError(ctx, http.StatusBadRequest, "RESET_TOKEN_USED", "Reset token has already been used")
	case err == domain.ErrCurrentPasswordInvalid:
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "Current password is incorrect")
	case strings.Contains(errMsg, "password validation failed"):
		return h.handleError(ctx, http.StatusBadRequest, "WEAK_PASSWORD", "Password does not meet security requirements")
	case strings.Contains(errMsg, "current password verification failed"):
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_CURRENT_PASSWORD", "Current password is incorrect")
	default:
		h.logger.Log(log.LevelError, "msg", "Password error", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

// generateSecureState generates a secure state parameter for OAuth CSRF protection
func generateSecureState() (string, error) {
	// This should use the same secure token generator as other parts of the system
	// For now, we'll use a simple implementation
	return fmt.Sprintf("state_%d", time.Now().UnixNano()), nil
}

// isValidEmail validates email format using a simple regex
func isValidEmail(email string) bool {
	// Simple email validation regex
	// In production, you might want to use a more sophisticated validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}
