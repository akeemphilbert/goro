package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
)

// Mock token manager for session manager tests
type mockTokenManagerForSession struct {
	tokens map[string]*application.TokenClaims
}

func (m *mockTokenManagerForSession) GenerateToken(ctx context.Context, session *domain.Session) (string, error) {
	token := "token_" + session.ID
	m.tokens[token] = &application.TokenClaims{
		SessionID: session.ID,
		UserID:    session.UserID,
		WebID:     session.WebID,
		AccountID: session.AccountID,
		RoleID:    session.RoleID,
		ExpiresAt: session.ExpiresAt,
		IssuedAt:  time.Now(),
	}
	return token, nil
}

func (m *mockTokenManagerForSession) ValidateToken(ctx context.Context, token string) (*application.TokenClaims, error) {
	if claims, exists := m.tokens[token]; exists {
		if time.Now().Before(claims.ExpiresAt) {
			return claims, nil
		}
	}
	return nil, domain.ErrSessionExpired
}

func (m *mockTokenManagerForSession) RefreshToken(ctx context.Context, token string) (string, error) {
	return "refreshed_" + token, nil
}

// Mock session repository for session manager tests
type mockSessionRepositoryForSession struct {
	sessions map[string]*domain.Session
}

func (m *mockSessionRepositoryForSession) Save(ctx context.Context, session *domain.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepositoryForSession) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	if session, exists := m.sessions[id]; exists {
		return session, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (m *mockSessionRepositoryForSession) FindByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}
	return result, nil
}

func (m *mockSessionRepositoryForSession) Delete(ctx context.Context, id string) error {
	if _, exists := m.sessions[id]; !exists {
		return domain.ErrSessionNotFound
	}
	delete(m.sessions, id)
	return nil
}

func (m *mockSessionRepositoryForSession) DeleteByUserID(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepositoryForSession) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for id, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepositoryForSession) UpdateActivity(ctx context.Context, sessionID string) error {
	if session, exists := m.sessions[sessionID]; exists {
		session.UpdateActivity()
		return nil
	}
	return domain.ErrSessionNotFound
}

func (m *mockSessionRepositoryForSession) FindByAccountID(ctx context.Context, accountID string) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, session := range m.sessions {
		if session.AccountID == accountID {
			result = append(result, session)
		}
	}
	return result, nil
}

func (m *mockSessionRepositoryForSession) FindByUserIDAndAccountID(ctx context.Context, userID, accountID string) ([]*domain.Session, error) {
	var result []*domain.Session
	for _, session := range m.sessions {
		if session.UserID == userID && session.AccountID == accountID {
			result = append(result, session)
		}
	}
	return result, nil
}

func setupSessionManager() (*application.SessionManager, *mockSessionRepositoryForSession, *mockTokenManagerForSession) {
	sessionRepo := &mockSessionRepositoryForSession{
		sessions: make(map[string]*domain.Session),
	}
	tokenManager := &mockTokenManagerForSession{
		tokens: make(map[string]*application.TokenClaims),
	}

	sessionManager := application.NewSessionManager(
		sessionRepo,
		tokenManager,
		24*time.Hour,   // session expiry
		time.Hour,      // refresh threshold
		30*time.Minute, // cleanup interval
	)

	return sessionManager, sessionRepo, tokenManager
}

func TestSessionManager_CreateSession_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Test session creation
	session, jwtToken, err := sessionManager.CreateSession(ctx, "user1", "https://example.com/profile#me")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful session creation, got error: %v", err)
	}

	if session == nil {
		t.Fatal("Expected session to be created")
	}

	if jwtToken == "" {
		t.Fatal("Expected JWT token to be generated")
	}

	if session.UserID != "user1" {
		t.Errorf("Expected session UserID to be 'user1', got '%s'", session.UserID)
	}

	if session.WebID != "https://example.com/profile#me" {
		t.Errorf("Expected session WebID to be 'https://example.com/profile#me', got '%s'", session.WebID)
	}

	// Verify session was saved
	if _, exists := sessionRepo.sessions[session.ID]; !exists {
		t.Error("Expected session to be saved in repository")
	}
}

func TestSessionManager_ValidateSession_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a valid session
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	sessionRepo.sessions["session1"] = session

	// Test validation
	validatedSession, err := sessionManager.ValidateSession(ctx, "session1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful validation, got error: %v", err)
	}

	if validatedSession == nil {
		t.Fatal("Expected session to be returned")
	}

	if validatedSession.ID != "session1" {
		t.Errorf("Expected session ID to be 'session1', got '%s'", validatedSession.ID)
	}
}

func TestSessionManager_ValidateSession_Expired(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create an expired session
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(-time.Hour), // Expired
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		LastActivity: time.Now().Add(-time.Hour),
	}
	sessionRepo.sessions["session1"] = session

	// Test validation
	validatedSession, err := sessionManager.ValidateSession(ctx, "session1")

	// Assertions
	if err != domain.ErrSessionExpired {
		t.Errorf("Expected ErrSessionExpired, got: %v", err)
	}

	if validatedSession != nil {
		t.Error("Expected no session to be returned for expired session")
	}

	// Verify session was deleted
	if _, exists := sessionRepo.sessions["session1"]; exists {
		t.Error("Expected expired session to be deleted")
	}
}

func TestSessionManager_ValidateSession_NotFound(t *testing.T) {
	sessionManager, _, _ := setupSessionManager()
	ctx := context.Background()

	// Test with non-existent session
	validatedSession, err := sessionManager.ValidateSession(ctx, "nonexistent")

	// Assertions
	if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}

	if validatedSession != nil {
		t.Error("Expected no session to be returned for non-existent session")
	}
}

func TestSessionManager_ValidateJWTToken_Success(t *testing.T) {
	sessionManager, sessionRepo, tokenManager := setupSessionManager()
	ctx := context.Background()

	// Create a valid session
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	sessionRepo.sessions["session1"] = session

	// Create a valid token
	token := "token_session1"
	tokenManager.tokens[token] = &application.TokenClaims{
		SessionID: "session1",
		UserID:    "user1",
		WebID:     "https://example.com/profile#me",
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
	}

	// Test JWT validation
	validatedSession, err := sessionManager.ValidateJWTToken(ctx, token)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful JWT validation, got error: %v", err)
	}

	if validatedSession == nil {
		t.Fatal("Expected session to be returned")
	}

	if validatedSession.ID != "session1" {
		t.Errorf("Expected session ID to be 'session1', got '%s'", validatedSession.ID)
	}
}

func TestSessionManager_RefreshSession_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a session that needs refresh (expires within refresh threshold)
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(30 * time.Minute), // Within refresh threshold
		CreatedAt:    time.Now().Add(-time.Hour),
		LastActivity: time.Now().Add(-10 * time.Minute),
	}
	sessionRepo.sessions["session1"] = session
	originalExpiry := session.ExpiresAt

	// Test refresh
	refreshedSession, newToken, err := sessionManager.RefreshSession(ctx, "session1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful refresh, got error: %v", err)
	}

	if refreshedSession == nil {
		t.Fatal("Expected session to be returned")
	}

	if newToken == "" {
		t.Fatal("Expected new JWT token to be generated")
	}

	if !refreshedSession.ExpiresAt.After(originalExpiry) {
		t.Error("Expected session expiry to be extended")
	}
}

func TestSessionManager_RefreshSession_NoRefreshNeeded(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a session that doesn't need refresh
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(2 * time.Hour), // Beyond refresh threshold
		CreatedAt:    time.Now().Add(-time.Hour),
		LastActivity: time.Now().Add(-10 * time.Minute),
	}
	sessionRepo.sessions["session1"] = session
	originalExpiry := session.ExpiresAt

	// Test refresh
	refreshedSession, newToken, err := sessionManager.RefreshSession(ctx, "session1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful refresh, got error: %v", err)
	}

	if refreshedSession == nil {
		t.Fatal("Expected session to be returned")
	}

	if newToken == "" {
		t.Fatal("Expected JWT token to be generated")
	}

	// Expiry should not have changed significantly (allowing for small time differences)
	timeDiff := refreshedSession.ExpiresAt.Sub(originalExpiry)
	if timeDiff > time.Second {
		t.Error("Expected session expiry to remain unchanged when refresh not needed")
	}
}

func TestSessionManager_InvalidateSession_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a session
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	sessionRepo.sessions["session1"] = session

	// Test invalidation
	err := sessionManager.InvalidateSession(ctx, "session1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful invalidation, got error: %v", err)
	}

	// Verify session was deleted
	if _, exists := sessionRepo.sessions["session1"]; exists {
		t.Error("Expected session to be deleted after invalidation")
	}
}

func TestSessionManager_InvalidateAllUserSessions_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create multiple sessions for the same user
	session1 := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	session2 := &domain.Session{
		ID:           "session2",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash456",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	session3 := &domain.Session{
		ID:           "session3",
		UserID:       "user2", // Different user
		WebID:        "https://example.com/other#me",
		TokenHash:    "hash789",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	sessionRepo.sessions["session1"] = session1
	sessionRepo.sessions["session2"] = session2
	sessionRepo.sessions["session3"] = session3

	// Test invalidate all sessions for user1
	err := sessionManager.InvalidateAllUserSessions(ctx, "user1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful invalidation, got error: %v", err)
	}

	// Verify user1 sessions were deleted
	if _, exists := sessionRepo.sessions["session1"]; exists {
		t.Error("Expected session1 to be deleted")
	}
	if _, exists := sessionRepo.sessions["session2"]; exists {
		t.Error("Expected session2 to be deleted")
	}

	// Verify user2 session was not deleted
	if _, exists := sessionRepo.sessions["session3"]; !exists {
		t.Error("Expected session3 to remain for different user")
	}
}

func TestSessionManager_SetAccountContext_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a session
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	sessionRepo.sessions["session1"] = session

	// Test setting account context
	err := sessionManager.SetAccountContext(ctx, "session1", "account1", "role1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful account context setting, got error: %v", err)
	}

	// Verify account context was set
	updatedSession := sessionRepo.sessions["session1"]
	if updatedSession.AccountID != "account1" {
		t.Errorf("Expected account ID to be 'account1', got '%s'", updatedSession.AccountID)
	}
	if updatedSession.RoleID != "role1" {
		t.Errorf("Expected role ID to be 'role1', got '%s'", updatedSession.RoleID)
	}
}

func TestSessionManager_ClearAccountContext_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create a session with account context
	session := &domain.Session{
		ID:           "session1",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		AccountID:    "account1",
		RoleID:       "role1",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	sessionRepo.sessions["session1"] = session

	// Test clearing account context
	err := sessionManager.ClearAccountContext(ctx, "session1")

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful account context clearing, got error: %v", err)
	}

	// Verify account context was cleared
	updatedSession := sessionRepo.sessions["session1"]
	if updatedSession.AccountID != "" {
		t.Errorf("Expected account ID to be empty, got '%s'", updatedSession.AccountID)
	}
	if updatedSession.RoleID != "" {
		t.Errorf("Expected role ID to be empty, got '%s'", updatedSession.RoleID)
	}
}

func TestSessionManager_CleanupExpiredSessions_Success(t *testing.T) {
	sessionManager, sessionRepo, _ := setupSessionManager()
	ctx := context.Background()

	// Create sessions with different expiry times
	activeSession := &domain.Session{
		ID:           "active",
		UserID:       "user1",
		WebID:        "https://example.com/profile#me",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}
	expiredSession := &domain.Session{
		ID:           "expired",
		UserID:       "user2",
		WebID:        "https://example.com/other#me",
		TokenHash:    "hash456",
		ExpiresAt:    time.Now().Add(-time.Hour),
		CreatedAt:    time.Now().Add(-2 * time.Hour),
		LastActivity: time.Now().Add(-time.Hour),
	}

	sessionRepo.sessions["active"] = activeSession
	sessionRepo.sessions["expired"] = expiredSession

	// Test cleanup
	err := sessionManager.CleanupExpiredSessions(ctx)

	// Assertions
	if err != nil {
		t.Fatalf("Expected successful cleanup, got error: %v", err)
	}

	// Verify active session remains
	if _, exists := sessionRepo.sessions["active"]; !exists {
		t.Error("Expected active session to remain")
	}

	// Verify expired session was removed
	if _, exists := sessionRepo.sessions["expired"]; exists {
		t.Error("Expected expired session to be removed")
	}
}
