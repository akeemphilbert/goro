package application

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
)

// JWTTokenManager implements JWT token generation and validation using golang-jwt
type JWTTokenManager struct {
	signingKey    *rsa.PrivateKey
	issuer        string
	expiry        time.Duration
	refreshExpiry time.Duration
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
		signingKey:    signingKey,
		issuer:        issuer,
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
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
		return nil, fmt.Errorf("JWT token validation failed: %w", err)
	}

	// Extract custom claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT token claims")
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
		return nil, fmt.Errorf("missing required claims in JWT token")
	}

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
