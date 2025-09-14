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

// Mock implementations for testing security features

type MockTokenBlacklistRepository struct {
	blacklist map[string]*application.TokenBlacklist
}

func NewMockTokenBlacklistRepository() *MockTokenBlacklistRepository {
	return &MockTokenBlacklistRepository{
		blacklist: make(map[string]*application.TokenBlacklist),
	}
}

func (m *MockTokenBlacklistRepository) AddToBlacklist(ctx context.Context, blacklist *application.TokenBlacklist) error {
	m.blacklist[blacklist.TokenID] = blacklist
	return nil
}

func (m *MockTokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	entry, exists := m.blacklist[tokenID]
	if !exists {
		return false, nil
	}
	// Check if entry is still valid (not expired)
	return time.Now().Before(entry.ExpiresAt), nil
}

func (m *MockTokenBlacklistRepository) RemoveFromBlacklist(ctx context.Context, tokenID string) error {
	delete(m.blacklist, tokenID)
	return nil
}

func (m *MockTokenBlacklistRepository) CleanupExpired(ctx context.Context) error {
	now := time.Now()
	for tokenID, entry := range m.blacklist {
		if now.After(entry.ExpiresAt) {
			delete(m.blacklist, tokenID)
		}
	}
	return nil
}

type MockTokenAuditLogger struct {
	events []application.TokenAuditEvent
}

func NewMockTokenAuditLogger() *MockTokenAuditLogger {
	return &MockTokenAuditLogger{
		events: make([]application.TokenAuditEvent, 0),
	}
}

func (m *MockTokenAuditLogger) LogTokenEvent(ctx context.Context, event *application.TokenAuditEvent) error {
	m.events = append(m.events, *event)
	return nil
}

func (m *MockTokenAuditLogger) GetEvents() []application.TokenAuditEvent {
	return m.events
}

// Security feature tests

func TestJWTTokenManager_RevokeToken_Success(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	blacklistRepo := NewMockTokenBlacklistRepository()
	auditLogger := NewMockTokenAuditLogger()

	tokenManager := application.NewJWTTokenManagerWithSecurity(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
		blacklistRepo,
		auditLogger,
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

	// Revoke token
	err = tokenManager.RevokeToken(ctx, token, "security_breach")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify token is revoked
	isRevoked, err := tokenManager.IsTokenRevoked(ctx, token)
	if err != nil {
		t.Fatalf("Failed to check token revocation: %v", err)
	}

	if !isRevoked {
		t.Error("Expected token to be revoked")
	}

	// Verify audit log
	events := auditLogger.GetEvents()
	found := false
	for _, event := range events {
		if event.EventType == "token_revoked" && event.Success {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected token revocation to be logged")
	}
}

func TestJWTTokenManager_ValidateToken_RevokedToken(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	blacklistRepo := NewMockTokenBlacklistRepository()
	auditLogger := NewMockTokenAuditLogger()

	tokenManager := application.NewJWTTokenManagerWithSecurity(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
		blacklistRepo,
		auditLogger,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate and revoke token
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	err = tokenManager.RevokeToken(ctx, token, "test_revocation")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Try to validate revoked token
	claims, err := tokenManager.ValidateToken(ctx, token)

	// Assertions
	if err == nil {
		t.Error("Expected validation to fail for revoked token")
	}

	if claims != nil {
		t.Error("Expected no claims to be returned for revoked token")
	}

	if err.Error() != "token has been revoked" {
		t.Errorf("Expected 'token has been revoked' error, got: %v", err)
	}

	// Verify audit log contains blacklisted token usage attempt
	events := auditLogger.GetEvents()
	found := false
	for _, event := range events {
		if event.EventType == "blacklisted_token_used" && !event.Success {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected blacklisted token usage to be logged")
	}
}

func TestJWTTokenManager_CleanupExpiredBlacklist(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	blacklistRepo := NewMockTokenBlacklistRepository()
	auditLogger := NewMockTokenAuditLogger()

	tokenManager := application.NewJWTTokenManagerWithSecurity(
		privateKey,
		"test-issuer",
		time.Hour, // Normal expiry
		30*time.Minute,
		blacklistRepo,
		auditLogger,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate and revoke token
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	err = tokenManager.RevokeToken(ctx, token, "test_revocation")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify token is revoked
	isRevoked, err := tokenManager.IsTokenRevoked(ctx, token)
	if err != nil {
		t.Fatalf("Failed to check token revocation: %v", err)
	}

	if !isRevoked {
		t.Error("Expected token to be revoked")
	}

	// Manually add an expired entry to the blacklist for testing cleanup
	expiredBlacklist := &application.TokenBlacklist{
		TokenID:   "expired-token-id",
		UserID:    "test-user-id",
		Reason:    "test_expired",
		RevokedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour), // Already expired
	}
	err = blacklistRepo.AddToBlacklist(ctx, expiredBlacklist)
	if err != nil {
		t.Fatalf("Failed to add expired blacklist entry: %v", err)
	}

	// Cleanup expired blacklist entries
	err = tokenManager.CleanupExpiredBlacklist(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup expired blacklist: %v", err)
	}

	// Verify expired entry was removed
	isExpiredRevoked, err := blacklistRepo.IsBlacklisted(ctx, "expired-token-id")
	if err != nil {
		t.Fatalf("Failed to check expired token blacklist status: %v", err)
	}

	if isExpiredRevoked {
		t.Error("Expected expired token to be removed from blacklist")
	}

	// Verify non-expired entry is still there
	isStillRevoked, err := tokenManager.IsTokenRevoked(ctx, token)
	if err != nil {
		t.Fatalf("Failed to check token revocation after cleanup: %v", err)
	}

	if !isStillRevoked {
		t.Error("Expected non-expired revoked token to still be in blacklist")
	}
}

func TestJWTTokenManager_RevokeAllUserTokens(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	auditLogger := NewMockTokenAuditLogger()

	tokenManager := application.NewJWTTokenManagerWithSecurity(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
		nil, // No blacklist repo for this test
		auditLogger,
	)

	ctx := context.Background()

	// Revoke all tokens for user
	err = tokenManager.RevokeAllUserTokens(ctx, "test-user-id", "account_compromise")
	if err != nil {
		t.Fatalf("Failed to revoke all user tokens: %v", err)
	}

	// Verify audit log
	events := auditLogger.GetEvents()
	found := false
	for _, event := range events {
		if event.EventType == "all_user_tokens_revoked" && event.UserID == "test-user-id" && event.Success {
			found = true
			if event.Reason != "account_compromise" {
				t.Errorf("Expected reason 'account_compromise', got: %s", event.Reason)
			}
			break
		}
	}
	if !found {
		t.Error("Expected all user tokens revocation to be logged")
	}
}

func TestJWTTokenManager_SecurityAuditLogging(t *testing.T) {
	// Setup
	privateKey, err := generateTestKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate test key pair: %v", err)
	}

	auditLogger := NewMockTokenAuditLogger()

	tokenManager := application.NewJWTTokenManagerWithSecurity(
		privateKey,
		"test-issuer",
		time.Hour,
		30*time.Minute,
		nil,
		auditLogger,
	)

	session := &domain.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	ctx := context.Background()

	// Generate token (should log generation)
	token, err := tokenManager.GenerateToken(ctx, session)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token (should log validation)
	_, err = tokenManager.ValidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Validate invalid token (should log failure)
	_, err = tokenManager.ValidateToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected validation to fail for invalid token")
	}

	// Check audit events
	events := auditLogger.GetEvents()

	// Should have at least: token_validated (success), token_validation_failed
	if len(events) < 2 {
		t.Errorf("Expected at least 2 audit events, got %d", len(events))
	}

	// Verify successful validation event
	foundSuccess := false
	foundFailure := false
	for _, event := range events {
		if event.EventType == "token_validated" && event.Success {
			foundSuccess = true
			if event.UserID != "test-user-id" {
				t.Errorf("Expected user ID 'test-user-id', got: %s", event.UserID)
			}
		}
		if event.EventType == "token_validation_failed" && !event.Success {
			foundFailure = true
		}
	}

	if !foundSuccess {
		t.Error("Expected successful token validation to be logged")
	}
	if !foundFailure {
		t.Error("Expected failed token validation to be logged")
	}
}
