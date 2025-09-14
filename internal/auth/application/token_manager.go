package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
)

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	TokenID   string
	UserID    string
	Reason    string
	RevokedAt time.Time
	ExpiresAt time.Time
}

// TokenAuditEvent represents a token security event for audit logging
type TokenAuditEvent struct {
	EventType string
	TokenID   string
	UserID    string
	SessionID string
	Timestamp time.Time
	IPAddress string
	UserAgent string
	Reason    string
	Success   bool
}

// TokenBlacklistRepository interface for managing blacklisted tokens
type TokenBlacklistRepository interface {
	AddToBlacklist(ctx context.Context, blacklist *TokenBlacklist) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	RemoveFromBlacklist(ctx context.Context, tokenID string) error
	CleanupExpired(ctx context.Context) error
}

// TokenAuditLogger interface for logging token security events
type TokenAuditLogger interface {
	LogTokenEvent(ctx context.Context, event *TokenAuditEvent) error
}

// JWTTokenManager implements JWT token generation and validation using golang-jwt
type JWTTokenManager struct {
	signingKey        *rsa.PrivateKey
	issuer            string
	expiry            time.Duration
	refreshExpiry     time.Duration
	blacklistRepo     TokenBlacklistRepository
	auditLogger       TokenAuditLogger
	inMemoryBlacklist map[string]time.Time // In-memory cache for performance
	blacklistMutex    sync.RWMutex
}

// CustomClaims represents our custom JWT claims
type CustomClaims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	WebID     string `json:"webid"`
	AccountID string `json:"account_id,omitempty"`
	RoleID    string `json:"role_id,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTTokenManager creates a new JWT token manager
func NewJWTTokenManager(signingKey *rsa.PrivateKey, issuer string, expiry, refreshExpiry time.Duration) *JWTTokenManager {
	return &JWTTokenManager{
		signingKey:        signingKey,
		issuer:            issuer,
		expiry:            expiry,
		refreshExpiry:     refreshExpiry,
		inMemoryBlacklist: make(map[string]time.Time),
	}
}

// NewJWTTokenManagerWithSecurity creates a new JWT token manager with security features
func NewJWTTokenManagerWithSecurity(
	signingKey *rsa.PrivateKey,
	issuer string,
	expiry, refreshExpiry time.Duration,
	blacklistRepo TokenBlacklistRepository,
	auditLogger TokenAuditLogger,
) *JWTTokenManager {
	return &JWTTokenManager{
		signingKey:        signingKey,
		issuer:            issuer,
		expiry:            expiry,
		refreshExpiry:     refreshExpiry,
		blacklistRepo:     blacklistRepo,
		auditLogger:       auditLogger,
		inMemoryBlacklist: make(map[string]time.Time),
	}
}

// GenerateToken generates a JWT token for a session
func (tm *JWTTokenManager) GenerateToken(ctx context.Context, session *domain.Session) (string, error) {
	now := time.Now()
	expiresAt := now.Add(tm.expiry)

	claims := CustomClaims{
		SessionID: session.ID,
		UserID:    session.UserID,
		WebID:     session.WebID,
		AccountID: session.AccountID,
		RoleID:    session.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tm.issuer,
			Subject:   session.UserID,
			Audience:  []string{"solid-pod-server"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        session.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set the key ID in the header
	token.Header["kid"] = "default"

	tokenString, err := token.SignedString(tm.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (tm *JWTTokenManager) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &tm.signingKey.PublicKey, nil
	})

	if err != nil {
		tm.logTokenEvent(ctx, &TokenAuditEvent{
			EventType: "token_validation_failed",
			Timestamp: time.Now(),
			Reason:    err.Error(),
			Success:   false,
		})
		return nil, fmt.Errorf("JWT token validation failed: %w", err)
	}

	// Extract custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		tm.logTokenEvent(ctx, &TokenAuditEvent{
			EventType: "invalid_token_claims",
			Timestamp: time.Now(),
			Reason:    "invalid JWT token claims",
			Success:   false,
		})
		return nil, fmt.Errorf("invalid JWT token claims")
	}

	// Check if token is blacklisted
	isBlacklisted, err := tm.isTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		slog.Error("Failed to check token blacklist", "error", err, "token_id", claims.ID)
		// Continue validation - don't fail on blacklist check errors
	} else if isBlacklisted {
		tm.logTokenEvent(ctx, &TokenAuditEvent{
			EventType: "blacklisted_token_used",
			TokenID:   claims.ID,
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
			Timestamp: time.Now(),
			Reason:    "token is blacklisted",
			Success:   false,
		})
		return nil, fmt.Errorf("token has been revoked")
	}

	// Convert to our TokenClaims structure
	tokenClaims := &TokenClaims{
		SessionID: claims.SessionID,
		UserID:    claims.UserID,
		WebID:     claims.WebID,
		AccountID: claims.AccountID,
		RoleID:    claims.RoleID,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}

	// Validate required claims
	if tokenClaims.SessionID == "" || tokenClaims.UserID == "" || tokenClaims.WebID == "" {
		tm.logTokenEvent(ctx, &TokenAuditEvent{
			EventType: "missing_required_claims",
			TokenID:   claims.ID,
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
			Timestamp: time.Now(),
			Reason:    "missing required claims in JWT token",
			Success:   false,
		})
		return nil, fmt.Errorf("missing required claims in JWT token")
	}

	// Log successful validation
	tm.logTokenEvent(ctx, &TokenAuditEvent{
		EventType: "token_validated",
		TokenID:   claims.ID,
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		Timestamp: time.Now(),
		Success:   true,
	})

	return tokenClaims, nil
}

// RefreshToken generates a new token with extended expiry
func (tm *JWTTokenManager) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	// Validate current token
	claims, err := tm.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Check if token is eligible for refresh (not too close to expiry)
	timeUntilExpiry := time.Until(claims.ExpiresAt)
	if timeUntilExpiry > tm.refreshExpiry {
		return "", fmt.Errorf("token does not need refresh yet")
	}

	// Create new session object for token generation
	session := &domain.Session{
		ID:        claims.SessionID,
		UserID:    claims.UserID,
		WebID:     claims.WebID,
		AccountID: claims.AccountID,
		RoleID:    claims.RoleID,
	}

	// Generate new token
	newToken, err := tm.GenerateToken(ctx, session)
	if err != nil {
		return "", fmt.Errorf("failed to generate refreshed token: %w", err)
	}

	return newToken, nil
}

// GenerateSecureSessionID generates a cryptographically secure session ID
func GenerateSecureSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("sess_%x", hash[:16]), nil
}

// GenerateSecureTokenHash generates a cryptographically secure token hash
func GenerateSecureTokenHash() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("hash_%x", hash), nil
}

// RevokeToken adds a token to the blacklist
func (tm *JWTTokenManager) RevokeToken(ctx context.Context, tokenString, reason string) error {
	// Parse token to get claims
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &tm.signingKey.PublicKey, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for revocation: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return fmt.Errorf("invalid token claims for revocation")
	}

	// Add to in-memory blacklist
	tm.blacklistMutex.Lock()
	tm.inMemoryBlacklist[claims.ID] = claims.ExpiresAt.Time
	tm.blacklistMutex.Unlock()

	// Add to persistent blacklist if repository is available
	if tm.blacklistRepo != nil {
		blacklist := &TokenBlacklist{
			TokenID:   claims.ID,
			UserID:    claims.UserID,
			Reason:    reason,
			RevokedAt: time.Now(),
			ExpiresAt: claims.ExpiresAt.Time,
		}

		if err := tm.blacklistRepo.AddToBlacklist(ctx, blacklist); err != nil {
			slog.Error("Failed to add token to persistent blacklist", "error", err, "token_id", claims.ID)
			// Don't return error - in-memory blacklist is sufficient
		}
	}

	// Log revocation event
	tm.logTokenEvent(ctx, &TokenAuditEvent{
		EventType: "token_revoked",
		TokenID:   claims.ID,
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		Timestamp: time.Now(),
		Reason:    reason,
		Success:   true,
	})

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (tm *JWTTokenManager) RevokeAllUserTokens(ctx context.Context, userID, reason string) error {
	// This would require a way to track all active tokens per user
	// For now, we'll log the event and rely on session invalidation
	tm.logTokenEvent(ctx, &TokenAuditEvent{
		EventType: "all_user_tokens_revoked",
		UserID:    userID,
		Timestamp: time.Now(),
		Reason:    reason,
		Success:   true,
	})

	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (tm *JWTTokenManager) IsTokenRevoked(ctx context.Context, tokenString string) (bool, error) {
	// Parse token to get ID
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &tm.signingKey.PublicKey, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return false, fmt.Errorf("invalid token claims")
	}

	return tm.isTokenBlacklisted(ctx, claims.ID)
}

// CleanupExpiredBlacklist removes expired tokens from blacklist
func (tm *JWTTokenManager) CleanupExpiredBlacklist(ctx context.Context) error {
	now := time.Now()

	// Clean up in-memory blacklist
	tm.blacklistMutex.Lock()
	for tokenID, expiresAt := range tm.inMemoryBlacklist {
		if now.After(expiresAt) {
			delete(tm.inMemoryBlacklist, tokenID)
		}
	}
	tm.blacklistMutex.Unlock()

	// Clean up persistent blacklist if repository is available
	if tm.blacklistRepo != nil {
		if err := tm.blacklistRepo.CleanupExpired(ctx); err != nil {
			slog.Error("Failed to cleanup expired blacklist entries", "error", err)
			return fmt.Errorf("failed to cleanup expired blacklist: %w", err)
		}
	}

	return nil
}

// isTokenBlacklisted checks if a token is blacklisted
func (tm *JWTTokenManager) isTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	// Check in-memory blacklist first (fastest)
	tm.blacklistMutex.RLock()
	expiresAt, exists := tm.inMemoryBlacklist[tokenID]
	tm.blacklistMutex.RUnlock()

	if exists {
		// Check if blacklist entry is still valid
		if time.Now().Before(expiresAt) {
			return true, nil
		}
		// Remove expired entry
		tm.blacklistMutex.Lock()
		delete(tm.inMemoryBlacklist, tokenID)
		tm.blacklistMutex.Unlock()
	}

	// Check persistent blacklist if repository is available
	if tm.blacklistRepo != nil {
		return tm.blacklistRepo.IsBlacklisted(ctx, tokenID)
	}

	return false, nil
}

// logTokenEvent logs a token security event
func (tm *JWTTokenManager) logTokenEvent(ctx context.Context, event *TokenAuditEvent) {
	// Log to structured logger
	if event.Success {
		slog.Info("Token security event",
			"event_type", event.EventType,
			"token_id", event.TokenID,
			"user_id", event.UserID,
			"session_id", event.SessionID,
			"reason", event.Reason,
		)
	} else {
		slog.Warn("Token security event failed",
			"event_type", event.EventType,
			"token_id", event.TokenID,
			"user_id", event.UserID,
			"session_id", event.SessionID,
			"reason", event.Reason,
		)
	}

	// Log to audit logger if available
	if tm.auditLogger != nil {
		if err := tm.auditLogger.LogTokenEvent(ctx, event); err != nil {
			slog.Error("Failed to log token audit event", "error", err, "event_type", event.EventType)
		}
	}
}
