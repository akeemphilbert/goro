package application

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
)

// SessionManager handles session lifecycle management
type SessionManager struct {
	sessionRepo      domain.SessionRepository
	tokenManager     TokenManager
	sessionExpiry    time.Duration
	refreshThreshold time.Duration
	cleanupInterval  time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(
	sessionRepo domain.SessionRepository,
	tokenManager TokenManager,
	sessionExpiry time.Duration,
	refreshThreshold time.Duration,
	cleanupInterval time.Duration,
) *SessionManager {
	return &SessionManager{
		sessionRepo:      sessionRepo,
		tokenManager:     tokenManager,
		sessionExpiry:    sessionExpiry,
		refreshThreshold: refreshThreshold,
		cleanupInterval:  cleanupInterval,
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(ctx context.Context, userID, webID string) (*domain.Session, string, error) {
	// Generate secure session ID
	sessionID, err := GenerateSecureSessionID()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate secure token hash
	tokenHash, err := GenerateSecureTokenHash()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token hash: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:           sessionID,
		UserID:       userID,
		WebID:        webID,
		TokenHash:    tokenHash,
		ExpiresAt:    time.Now().Add(sm.sessionExpiry),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// Save session
	if err := sm.sessionRepo.Save(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to save session: %w", err)
	}

	// Generate JWT token
	jwtToken, err := sm.tokenManager.GenerateToken(ctx, session)
	if err != nil {
		// Clean up session if token generation fails
		_ = sm.sessionRepo.Delete(ctx, sessionID)
		return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return session, jwtToken, nil
}

// ValidateSession validates a session by ID
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	// Find session
	session, err := sm.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	// Check if session is valid
	if !session.IsValid() {
		// Delete expired session
		_ = sm.sessionRepo.Delete(ctx, sessionID)
		return nil, domain.ErrSessionExpired
	}

	// Update activity timestamp
	session.UpdateActivity()
	if err := sm.sessionRepo.Save(ctx, session); err != nil {
		// Log error but don't fail validation
		// The session is still valid even if we can't update activity
	}

	return session, nil
}

// ValidateJWTToken validates a JWT token and returns the session
func (sm *SessionManager) ValidateJWTToken(ctx context.Context, token string) (*domain.Session, error) {
	// Validate JWT token
	claims, err := sm.tokenManager.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("JWT token validation failed: %w", err)
	}

	// Validate the session
	session, err := sm.ValidateSession(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}

	// Verify token claims match session
	if session.UserID != claims.UserID || session.WebID != claims.WebID {
		return nil, fmt.Errorf("JWT claims do not match session")
	}

	return session, nil
}

// RefreshSession refreshes a session if it's close to expiry
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) (*domain.Session, string, error) {
	// Validate current session
	session, err := sm.ValidateSession(ctx, sessionID)
	if err != nil {
		return nil, "", err
	}

	// Check if refresh is needed
	timeUntilExpiry := time.Until(session.ExpiresAt)
	if timeUntilExpiry > sm.refreshThreshold {
		// No refresh needed, return current session and generate new token
		jwtToken, err := sm.tokenManager.GenerateToken(ctx, session)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
		}
		return session, jwtToken, nil
	}

	// Extend session expiry
	session.ExpiresAt = time.Now().Add(sm.sessionExpiry)
	session.UpdateActivity()

	// Save updated session
	if err := sm.sessionRepo.Save(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to refresh session: %w", err)
	}

	// Generate new JWT token
	jwtToken, err := sm.tokenManager.GenerateToken(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return session, jwtToken, nil
}

// RefreshJWTToken refreshes a JWT token
func (sm *SessionManager) RefreshJWTToken(ctx context.Context, token string) (string, error) {
	// Validate current token and get session
	session, err := sm.ValidateJWTToken(ctx, token)
	if err != nil {
		return "", err
	}

	// Refresh session and get new token
	_, newToken, err := sm.RefreshSession(ctx, session.ID)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	// Delete session
	if err := sm.sessionRepo.Delete(ctx, sessionID); err != nil {
		if err == domain.ErrSessionNotFound {
			// Session already doesn't exist, consider it invalidated
			return nil
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	// Delete all sessions for user
	if err := sm.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	sessions, err := sm.sessionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Filter out expired sessions and clean them up
	var activeSessions []*domain.Session
	for _, session := range sessions {
		if session.IsValid() {
			activeSessions = append(activeSessions, session)
		} else {
			// Clean up expired session
			_ = sm.sessionRepo.Delete(ctx, session.ID)
		}
	}

	return activeSessions, nil
}

// SetAccountContext sets the account context for a session
func (sm *SessionManager) SetAccountContext(ctx context.Context, sessionID, accountID, roleID string) error {
	// Find session
	session, err := sm.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Validate session
	if !session.IsValid() {
		return domain.ErrSessionExpired
	}

	// Set account context
	session.SetAccountContext(accountID, roleID)
	session.UpdateActivity()

	// Save updated session
	if err := sm.sessionRepo.Save(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// ClearAccountContext clears the account context for a session
func (sm *SessionManager) ClearAccountContext(ctx context.Context, sessionID string) error {
	// Find session
	session, err := sm.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Validate session
	if !session.IsValid() {
		return domain.ErrSessionExpired
	}

	// Clear account context
	session.ClearAccountContext()
	session.UpdateActivity()

	// Save updated session
	if err := sm.sessionRepo.Save(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	if err := sm.sessionRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// StartCleanupRoutine starts a background routine to clean up expired sessions
func (sm *SessionManager) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sm.CleanupExpiredSessions(ctx); err != nil {
				// Log error but continue cleanup routine
				// In a real implementation, you'd use a proper logger
			}
		}
	}
}
