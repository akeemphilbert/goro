package application_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
)

// Test helper to generate RSA key pair
func generateTestKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func TestJWTTokenManager_GenerateToken_Success(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		AccountID: "test-account-id",
		RoleID:    "test-role-id",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Test
	token, err := tokenManager.GenerateToken(ctx, session)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful token generation, got error: %v", err)
	}

	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	// Verify token can be validated
	claims, err := tokenManager.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Generated token should be valid, got error: %v", err)
	}

	if claims.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, claims.SessionID)
	}

	if claims.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, claims.UserID)
	}

	if claims.WebID != session.WebID {
		t.Errorf("Expected WebID %s, got %s", session.WebID, claims.WebID)
	}

	if claims.AccountID != session.AccountID {
		t.Errorf("Expected account ID %s, got %s", session.AccountID, claims.AccountID)
	}

	if claims.RoleID != session.RoleID {
		t.Errorf("Expected role ID %s, got %s", session.RoleID, claims.RoleID)
	}
}

func TestJWTTokenManager_GenerateToken_WithoutAccountContext(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		// No AccountID or RoleID
	}

	ctx := context.Background()

	// Test
	token, err := tokenManager.GenerateToken(ctx, session)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful token generation, got error: %v", err)
	}

	claims, err := tokenManager.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Generated token should be valid, got error: %v", err)
	}

	if claims.AccountID != "" {
		t.Errorf("Expected empty account ID, got %s", claims.AccountID)
	}

	if claims.RoleID != "" {
		t.Errorf("Expected empty role ID, got %s", claims.RoleID)
	}
}

func TestJWTTokenManager_ValidateToken_Success(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate token
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test validation
	claims, err := tokenManager.ValidateToken(ctx, token)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful token validation, got error: %v", err)
	}

	if claims == nil {
		t.Fatal("Expected claims to be returned")
	}

	if claims.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, claims.SessionID)
	}

	if claims.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, claims.UserID)
	}

	if claims.WebID != session.WebID {
		t.Errorf("Expected WebID %s, got %s", session.WebID, claims.WebID)
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Expected token to not be expired")
	}
}

func TestJWTTokenManager_ValidateToken_InvalidSignature(t *testing.T) {
	// Setup with different keys
	privateKey1, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair 1: %v", err)
	}

	privateKey2, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair 2: %v", err)
	}

	tokenManager1 := application.NewJWTTokenManager(
		privateKey1,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	tokenManager2 := application.NewJWTTokenManager(
		privateKey2,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate token with first key
	token, err := tokenManager1.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with second key
	claims, err := tokenManager2.ValidateToken(ctx, token)

	// Assertions
	if err == nil {
		t.Error("Expected validation to fail with different key")
	}

	if claims != nil {
		t.Error("Expected no claims to be returned for invalid token")
	}
}

func TestJWTTokenManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	// Use very short expiry
	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Millisecond, // Very short expiry
		30*time.Minute,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate token
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	claims, err := tokenManager.ValidateToken(ctx, token)

	// Assertions
	if err == nil {
		t.Error("Expected validation to fail for expired token")
	}

	if claims != nil {
		t.Error("Expected no claims to be returned for expired token")
	}
}

func TestJWTTokenManager_RefreshToken_Success(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		2*time.Hour, // Refresh threshold longer than token expiry - always refresh
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		AccountID: "test-account-id",
		RoleID:    "test-role-id",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate original token
	originalToken, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate original token: %v", err)
	}

	// Test refresh
	refreshedToken, err := tokenManager.RefreshToken(ctx, originalToken)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful token refresh, got error: %v", err)
	}

	if refreshedToken == "" {
		t.Fatal("Expected non-empty refreshed token")
	}

	// Note: Tokens might be the same if generated at the same time with same claims
	// The important thing is that refresh succeeded and returned a valid token

	// Validate refreshed token
	claims, err := tokenManager.ValidateToken(ctx, refreshedToken)
	if err != nil {
		t.Fatalf("Refreshed token should be valid, got error: %v", err)
	}

	if claims.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, claims.SessionID)
	}

	if claims.UserID != session.UserID {
		t.Errorf("Expected user ID %s, got %s", session.UserID, claims.UserID)
	}
}

func TestJWTTokenManager_RefreshToken_NotEligible(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		10*time.Minute, // Low refresh threshold
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate token (will have 1 hour expiry, more than 10 minute threshold)
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to refresh token that doesn't need refresh
	refreshedToken, err := tokenManager.RefreshToken(ctx, token)

	// Assertions
	if err == nil {
		t.Error("Expected refresh to fail for token that doesn't need refresh")
	}

	if refreshedToken != "" {
		t.Error("Expected no refreshed token for ineligible token")
	}
}

func TestJWTTokenManager_RefreshToken_InvalidToken(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	tokenManager := application.NewJWTTokenManager(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
	)

	ctx := context.Background()

	// Try to refresh invalid token
	refreshedToken, err := tokenManager.RefreshToken(ctx, "invalid-token")

	// Assertions
	if err == nil {
		t.Error("Expected refresh to fail for invalid token")
	}

	if refreshedToken != "" {
		t.Error("Expected no refreshed token for invalid token")
	}
}

func TestGenerateSecureSessionID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id, err := application.GenerateSecureSessionID()
		if err != nil {
			t.Fatalf("Failed to generate session ID: %v", err)
		}

		if id == "" {
			t.Error("Expected non-empty session ID")
		}

		if ids[id] {
			t.Errorf("Generated duplicate session ID: %s", id)
		}
		ids[id] = true

		// Check format
		if len(id) < 10 {
			t.Errorf("Expected session ID to be reasonably long, got: %s", id)
		}

		if id[:5] != "sess_" {
			t.Errorf("Expected session ID to start with 'sess_', got: %s", id)
		}
	}
}

func TestGenerateSecureTokenHash(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	hashes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		hash, err := application.GenerateSecureTokenHash()
		if err != nil {
			t.Fatalf("Failed to generate token hash: %v", err)
		}

		if hash == "" {
			t.Error("Expected non-empty token hash")
		}

		if hashes[hash] {
			t.Errorf("Generated duplicate token hash: %s", hash)
		}
		hashes[hash] = true

		// Check format
		if len(hash) < 10 {
			t.Errorf("Expected token hash to be reasonably long, got: %s", hash)
		}

		if hash[:5] != "hash_" {
			t.Errorf("Expected token hash to start with 'hash_', got: %s", hash)
		}
	}
}
