package application

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/akeemphilbert/goro/internal/auth/infrastructure"
	userDomain "github.com/akeemphilbert/goro/internal/user/domain"
)

// AuthenticationService handles multi-method authentication
type AuthenticationService struct {
	userRepo         userDomain.UserRepository
	sessionRepo      domain.SessionRepository
	passwordRepo     domain.PasswordRepository
	identityRepo     domain.ExternalIdentityRepository
	passwordService  *PasswordService
	tokenManager     TokenManager
	webidProvider    *infrastructure.WebIDOIDCProvider
	oauthProviders   map[string]OAuthProvider
	sessionExpiry    time.Duration
	refreshThreshold time.Duration
}

// TokenManager interface for JWT token operations
type TokenManager interface {
	GenerateToken(ctx context.Context, session *domain.Session) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	RefreshToken(ctx context.Context, token string) (string, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	SessionID string
	UserID    string
	WebID     string
	AccountID string
	RoleID    string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// OAuthProvider interface for OAuth authentication providers
type OAuthProvider interface {
	GetProviderName() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*domain.OAuthToken, error)
	GetUserProfile(ctx context.Context, token *domain.OAuthToken) (*domain.ExternalProfile, error)
}

// NewAuthenticationService creates a new authentication service
func NewAuthenticationService(
	userRepo userDomain.UserRepository,
	sessionRepo domain.SessionRepository,
	passwordRepo domain.PasswordRepository,
	identityRepo domain.ExternalIdentityRepository,
	passwordService *PasswordService,
	tokenManager TokenManager,
	webidProvider *infrastructure.WebIDOIDCProvider,
	oauthProviders map[string]OAuthProvider,
	sessionExpiry time.Duration,
	refreshThreshold time.Duration,
) *AuthenticationService {
	return &AuthenticationService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		passwordRepo:     passwordRepo,
		identityRepo:     identityRepo,
		passwordService:  passwordService,
		tokenManager:     tokenManager,
		webidProvider:    webidProvider,
		oauthProviders:   oauthProviders,
		sessionExpiry:    sessionExpiry,
		refreshThreshold: refreshThreshold,
	}
}

// AuthenticateWithPassword authenticates a user with username/email and password
func (s *AuthenticationService) AuthenticateWithPassword(ctx context.Context, username, password string) (*domain.Session, error) {
	// Find user by email or username (assuming username is email for now)
	user, err := s.userRepo.GetByEmail(ctx, username)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Validate password
	if err := s.passwordService.ValidatePassword(ctx, user.ID(), password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Create session
	session, err := s.createSession(ctx, user.ID(), user.WebID, domain.MethodPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// AuthenticateWithWebID authenticates a user with WebID-OIDC token
func (s *AuthenticationService) AuthenticateWithWebID(ctx context.Context, webID string, oidcToken string) (*domain.Session, error) {
	// Validate WebID-OIDC token
	claims, err := s.webidProvider.ValidateWebIDToken(ctx, oidcToken)
	if err != nil {
		return nil, fmt.Errorf("WebID-OIDC validation failed: %w", err)
	}

	// Verify WebID matches
	if claims.WebID != webID {
		return nil, domain.ErrWebIDValidationFailed
	}

	// Find user by WebID
	user, err := s.userRepo.GetByWebID(ctx, webID)
	if err != nil {
		return nil, fmt.Errorf("user not found for WebID: %w", err)
	}

	// Create session
	session, err := s.createSession(ctx, user.ID(), webID, domain.MethodWebIDOIDC)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// AuthenticateWithOAuth authenticates a user with OAuth provider
func (s *AuthenticationService) AuthenticateWithOAuth(ctx context.Context, provider string, oauthCode string) (*domain.Session, error) {
	// Get OAuth provider
	oauthProvider, exists := s.oauthProviders[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Exchange code for token
	token, err := oauthProvider.ExchangeCode(ctx, oauthCode)
	if err != nil {
		return nil, fmt.Errorf("OAuth code exchange failed: %w", err)
	}

	// Get user profile from provider
	profile, err := oauthProvider.GetUserProfile(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Find user by external identity
	userID, err := s.identityRepo.FindByExternalID(ctx, provider, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("external identity not linked to any user: %w", err)
	}

	// Get user details
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create session
	session, err := s.createSession(ctx, user.ID(), user.WebID, domain.MethodOAuth)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates an existing session
func (s *AuthenticationService) ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	// Find session
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	// Check if session is valid
	if !session.IsValid() {
		// Delete expired session
		_ = s.sessionRepo.Delete(ctx, sessionID)
		return nil, domain.ErrSessionExpired
	}

	// Update activity timestamp
	session.UpdateActivity()
	if err := s.sessionRepo.Save(ctx, session); err != nil {
		// Log error but don't fail validation
		// The session is still valid even if we can't update activity
	}

	return session, nil
}

// RefreshSession refreshes a session if it's close to expiry
func (s *AuthenticationService) RefreshSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	// Validate current session
	session, err := s.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Check if refresh is needed
	timeUntilExpiry := time.Until(session.ExpiresAt)
	if timeUntilExpiry > s.refreshThreshold {
		// No refresh needed
		return session, nil
	}

	// Extend session expiry
	session.ExpiresAt = time.Now().Add(s.sessionExpiry)
	session.UpdateActivity()

	// Save updated session
	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	return session, nil
}

// Logout invalidates a session
func (s *AuthenticationService) Logout(ctx context.Context, sessionID string) error {
	// Delete session
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		if err == domain.ErrSessionNotFound {
			// Session already doesn't exist, consider it logged out
			return nil
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// LogoutAllSessions invalidates all sessions for a user
func (s *AuthenticationService) LogoutAllSessions(ctx context.Context, userID string) error {
	// Delete all sessions for user
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (s *AuthenticationService) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	sessions, err := s.sessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Filter out expired sessions
	var activeSessions []*domain.Session
	for _, session := range sessions {
		if session.IsValid() {
			activeSessions = append(activeSessions, session)
		} else {
			// Clean up expired session
			_ = s.sessionRepo.Delete(ctx, session.ID)
		}
	}

	return activeSessions, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *AuthenticationService) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.sessionRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// createSession creates a new session for a user
func (s *AuthenticationService) createSession(ctx context.Context, userID, webID string, method domain.AuthenticationMethod) (*domain.Session, error) {
	// Generate session ID
	sessionID, err := GenerateSecureSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate token hash
	tokenHash, err := GenerateSecureTokenHash()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token hash: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:           sessionID,
		UserID:       userID,
		WebID:        webID,
		TokenHash:    tokenHash,
		ExpiresAt:    time.Now().Add(s.sessionExpiry),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Save session
	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}
