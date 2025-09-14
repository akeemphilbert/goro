package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Context keys for authentication information
type authContextKey string

const (
	userIDKey          authContextKey = "user_id"
	sessionIDKey       authContextKey = "session_id"
	webIDKey           authContextKey = "webid"
	accountIDKey       authContextKey = "account_id"
	roleIDKey          authContextKey = "role_id"
	isAuthenticatedKey authContextKey = "is_authenticated"
)

// AuthInfo holds authentication information
type AuthInfo struct {
	UserID          string
	SessionID       string
	WebID           string
	AccountID       string
	RoleID          string
	IsAuthenticated bool
}

// WithAuthInfo adds authentication information to the context
func WithAuthInfo(ctx context.Context, info AuthInfo) context.Context {
	ctx = context.WithValue(ctx, userIDKey, info.UserID)
	ctx = context.WithValue(ctx, sessionIDKey, info.SessionID)
	ctx = context.WithValue(ctx, webIDKey, info.WebID)
	ctx = context.WithValue(ctx, accountIDKey, info.AccountID)
	ctx = context.WithValue(ctx, roleIDKey, info.RoleID)
	ctx = context.WithValue(ctx, isAuthenticatedKey, info.IsAuthenticated)
	return ctx
}

// GetAuthInfo retrieves authentication information from the context
func GetAuthInfo(ctx context.Context) AuthInfo {
	return AuthInfo{
		UserID:          GetUserID(ctx),
		SessionID:       GetSessionID(ctx),
		WebID:           GetWebID(ctx),
		AccountID:       GetAccountID(ctx),
		RoleID:          GetRoleID(ctx),
		IsAuthenticated: IsAuthenticated(ctx),
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetSessionID retrieves the session ID from the context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// GetWebID retrieves the WebID from the context
func GetWebID(ctx context.Context) string {
	if webID, ok := ctx.Value(webIDKey).(string); ok {
		return webID
	}
	return ""
}

// GetAccountID retrieves the account ID from the context
func GetAccountID(ctx context.Context) string {
	if accountID, ok := ctx.Value(accountIDKey).(string); ok {
		return accountID
	}
	return ""
}

// GetRoleID retrieves the role ID from the context
func GetRoleID(ctx context.Context) string {
	if roleID, ok := ctx.Value(roleIDKey).(string); ok {
		return roleID
	}
	return ""
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	if isAuth, ok := ctx.Value(isAuthenticatedKey).(bool); ok {
		return isAuth
	}
	return false
}

// SessionValidation returns a filter that validates sessions and injects user context
func SessionValidation(
	authService *application.AuthenticationService,
	tokenManager application.TokenManager,
	logger log.Logger,
) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from request
			token := extractToken(r)

			if token != "" {
				// Validate token and session
				authInfo, err := validateTokenAndSession(r.Context(), token, authService, tokenManager, logger)
				if err != nil {
					// Log validation error but don't fail the request
					// Some endpoints might be optional authentication
					logger.Log(log.LevelWarn, "msg", "Token validation failed", "error", err.Error())
				} else {
					// Inject authentication info into context
					ctx := WithAuthInfo(r.Context(), authInfo)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthentication returns a filter that requires valid authentication
func RequireAuthentication(
	authService *application.AuthenticationService,
	tokenManager application.TokenManager,
	logger log.Logger,
) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from request
			token := extractToken(r)
			if token == "" {
				writeAuthError(w, &AuthenticationError{
					Code:    "MISSING_TOKEN",
					Message: "Authentication token is required",
					Status:  http.StatusUnauthorized,
				})
				return
			}

			// Validate token and session
			authInfo, err := validateTokenAndSession(r.Context(), token, authService, tokenManager, logger)
			if err != nil {
				logger.Log(log.LevelWarn, "msg", "Authentication failed", "error", err.Error())
				writeAuthError(w, mapAuthError(err))
				return
			}

			// Inject authentication info into context
			ctx := WithAuthInfo(r.Context(), authInfo)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// JWTTokenExtraction returns a filter that extracts and validates JWT tokens
func JWTTokenExtraction(
	tokenManager application.TokenManager,
	logger log.Logger,
) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from request
			token := extractToken(r)

			if token != "" {
				// Validate JWT token
				claims, err := tokenManager.ValidateToken(r.Context(), token)
				if err != nil {
					// Log validation error but don't fail the request
					logger.Log(log.LevelWarn, "msg", "JWT validation failed", "error", err.Error())
				} else {
					// Inject token claims into context
					authInfo := AuthInfo{
						UserID:          claims.UserID,
						SessionID:       claims.SessionID,
						WebID:           claims.WebID,
						AccountID:       claims.AccountID,
						RoleID:          claims.RoleID,
						IsAuthenticated: true,
					}
					ctx := WithAuthInfo(r.Context(), authInfo)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns a filter that requires a specific role
func RequireRole(requiredRole string, logger log.Logger) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			if !IsAuthenticated(r.Context()) {
				writeAuthError(w, &AuthenticationError{
					Code:    "AUTHENTICATION_REQUIRED",
					Message: "Authentication is required",
					Status:  http.StatusUnauthorized,
				})
				return
			}

			// Check role
			userRole := GetRoleID(r.Context())
			if userRole != requiredRole {
				logger.Log(log.LevelWarn,
					"msg", "Insufficient permissions",
					"user_id", GetUserID(r.Context()),
					"required_role", requiredRole,
					"user_role", userRole,
				)
				writeAuthError(w, &AuthenticationError{
					Code:    "INSUFFICIENT_PERMISSIONS",
					Message: "Insufficient permissions for this operation",
					Status:  http.StatusForbidden,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuthentication returns a filter that provides optional authentication
func OptionalAuthentication(
	authService *application.AuthenticationService,
	tokenManager application.TokenManager,
	logger log.Logger,
) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from request
			token := extractToken(r)

			if token != "" {
				// Try to validate token and session
				authInfo, err := validateTokenAndSession(r.Context(), token, authService, tokenManager, logger)
				if err == nil {
					// Inject authentication info into context
					ctx := WithAuthInfo(r.Context(), authInfo)
					r = r.WithContext(ctx)
				}
				// Don't fail on authentication errors for optional auth
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	Code    string
	Message string
	Status  int
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// Helper functions

// extractToken extracts the authentication token from the HTTP request
func extractToken(req *http.Request) string {
	// Try Authorization header first (Bearer token)
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	// Try query parameter as fallback
	token := req.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Try session_id query parameter for session-based auth
	sessionID := req.URL.Query().Get("session_id")
	if sessionID != "" {
		return sessionID
	}

	return ""
}

// validateTokenAndSession validates both JWT token and session
func validateTokenAndSession(
	ctx context.Context,
	token string,
	authService *application.AuthenticationService,
	tokenManager application.TokenManager,
	logger log.Logger,
) (AuthInfo, error) {
	// First try to validate as JWT token
	claims, err := tokenManager.ValidateToken(ctx, token)
	if err == nil {
		// JWT is valid, now validate the session
		session, err := authService.ValidateSession(ctx, claims.SessionID)
		if err != nil {
			return AuthInfo{}, err
		}

		return AuthInfo{
			UserID:          session.UserID,
			SessionID:       session.ID,
			WebID:           session.WebID,
			AccountID:       claims.AccountID,
			RoleID:          claims.RoleID,
			IsAuthenticated: true,
		}, nil
	}

	// If JWT validation fails, try to validate as session ID directly
	session, err := authService.ValidateSession(ctx, token)
	if err != nil {
		return AuthInfo{}, err
	}

	return AuthInfo{
		UserID:          session.UserID,
		SessionID:       session.ID,
		WebID:           session.WebID,
		IsAuthenticated: true,
	}, nil
}

// mapAuthError maps domain authentication errors to HTTP errors
func mapAuthError(err error) *AuthenticationError {
	switch err {
	case domain.ErrSessionNotFound:
		return &AuthenticationError{
			Code:    "SESSION_NOT_FOUND",
			Message: "Session not found",
			Status:  http.StatusUnauthorized,
		}
	case domain.ErrSessionExpired:
		return &AuthenticationError{
			Code:    "SESSION_EXPIRED",
			Message: "Session has expired",
			Status:  http.StatusUnauthorized,
		}
	case domain.ErrInvalidCredentials:
		return &AuthenticationError{
			Code:    "INVALID_CREDENTIALS",
			Message: "Invalid credentials",
			Status:  http.StatusUnauthorized,
		}
	case domain.ErrWebIDValidationFailed:
		return &AuthenticationError{
			Code:    "WEBID_VALIDATION_FAILED",
			Message: "WebID validation failed",
			Status:  http.StatusUnauthorized,
		}
	default:
		return &AuthenticationError{
			Code:    "AUTHENTICATION_FAILED",
			Message: "Authentication failed",
			Status:  http.StatusUnauthorized,
		}
	}
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, err *AuthenticationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	response := map[string]interface{}{
		"error":   err.Code,
		"message": err.Message,
	}

	json.NewEncoder(w).Encode(response)
}

// AuthenticationConfig holds configuration for authentication middleware
type AuthenticationConfig struct {
	AuthService   *application.AuthenticationService
	TokenManager  application.TokenManager
	Logger        log.Logger
	SkipPaths     []string
	RequiredRoles map[string]string // path -> required role
}

// ConfigurableAuthentication returns a filter with configurable authentication rules
func ConfigurableAuthentication(config AuthenticationConfig) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath := r.URL.Path

			// Check if path should be skipped
			for _, skipPath := range config.SkipPaths {
				if requestPath == skipPath {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract and validate token
			token := extractToken(r)
			if token != "" {
				authInfo, err := validateTokenAndSession(r.Context(), token, config.AuthService, config.TokenManager, config.Logger)
				if err == nil {
					ctx := WithAuthInfo(r.Context(), authInfo)
					r = r.WithContext(ctx)

					// Check role requirements
					if requiredRole, exists := config.RequiredRoles[requestPath]; exists {
						if authInfo.RoleID != requiredRole {
							writeAuthError(w, &AuthenticationError{
								Code:    "INSUFFICIENT_PERMISSIONS",
								Message: "Insufficient permissions for this operation",
								Status:  http.StatusForbidden,
							})
							return
						}
					}
				} else {
					config.Logger.Log(log.LevelWarn, "msg", "Authentication failed", "path", requestPath, "error", err.Error())
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
