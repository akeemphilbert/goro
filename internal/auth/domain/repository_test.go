package domain

import (
	"context"
	"testing"
	"time"
)

// Mock implementations to test interface compliance

type mockSessionRepository struct {
	sessions map[string]*Session
}

func (m *mockSessionRepository) Save(ctx context.Context, session *Session) error {
	if m.sessions == nil {
		m.sessions = make(map[string]*Session)
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, id string) (*Session, error) {
	if session, exists := m.sessions[id]; exists {
		return session, nil
	}
	return nil, ErrSessionNotFound
}

func (m *mockSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*Session, error) {
	var sessions []*Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for id, session := range m.sessions {
		if session.ExpiresAt.Before(now) {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepository) UpdateActivity(ctx context.Context, sessionID string) error {
	if session, exists := m.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
		return nil
	}
	return ErrSessionNotFound
}

func (m *mockSessionRepository) FindByAccountID(ctx context.Context, accountID string) ([]*Session, error) {
	var sessions []*Session
	for _, session := range m.sessions {
		if session.AccountID == accountID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *mockSessionRepository) FindByUserIDAndAccountID(ctx context.Context, userID, accountID string) ([]*Session, error) {
	var sessions []*Session
	for _, session := range m.sessions {
		if session.UserID == userID && session.AccountID == accountID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func TestSessionRepository_Interface(t *testing.T) {
	// Test that our mock implements the interface
	var _ SessionRepository = &mockSessionRepository{}

	repo := &mockSessionRepository{}
	ctx := context.Background()

	// Test Save and FindByID
	session := &Session{
		ID:           "test-session",
		UserID:       "test-user",
		WebID:        "https://example.com/user#me",
		AccountID:    "test-account",
		RoleID:       "test-role",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	err := repo.Save(ctx, session)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	found, err := repo.FindByID(ctx, session.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}
	if found.ID != session.ID {
		t.Errorf("FindByID() returned wrong session")
	}

	// Test FindByUserID
	sessions, err := repo.FindByUserID(ctx, session.UserID)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("FindByUserID() returned %d sessions, want 1", len(sessions))
	}

	// Test UpdateActivity
	err = repo.UpdateActivity(ctx, session.ID)
	if err != nil {
		t.Errorf("UpdateActivity() error = %v", err)
	}

	// Test Delete
	err = repo.Delete(ctx, session.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(ctx, session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("FindByID() after delete should return ErrSessionNotFound")
	}
}

type mockPasswordRepository struct {
	credentials map[string]*PasswordCredential
}

func (m *mockPasswordRepository) Save(ctx context.Context, credential *PasswordCredential) error {
	if m.credentials == nil {
		m.credentials = make(map[string]*PasswordCredential)
	}
	m.credentials[credential.UserID] = credential
	return nil
}

func (m *mockPasswordRepository) FindByUserID(ctx context.Context, userID string) (*PasswordCredential, error) {
	if credential, exists := m.credentials[userID]; exists {
		return credential, nil
	}
	return nil, ErrInvalidCredentials
}

func (m *mockPasswordRepository) Update(ctx context.Context, credential *PasswordCredential) error {
	if _, exists := m.credentials[credential.UserID]; exists {
		m.credentials[credential.UserID] = credential
		return nil
	}
	return ErrInvalidCredentials
}

func (m *mockPasswordRepository) Delete(ctx context.Context, userID string) error {
	delete(m.credentials, userID)
	return nil
}

func (m *mockPasswordRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, exists := m.credentials[userID]
	return exists, nil
}

func TestPasswordRepository_Interface(t *testing.T) {
	// Test that our mock implements the interface
	var _ PasswordRepository = &mockPasswordRepository{}

	repo := &mockPasswordRepository{}
	ctx := context.Background()

	credential := &PasswordCredential{
		UserID:       "test-user",
		PasswordHash: "hash123",
		Salt:         "salt123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test Save
	err := repo.Save(ctx, credential)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Test Exists
	exists, err := repo.Exists(ctx, credential.UserID)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() should return true after Save()")
	}

	// Test FindByUserID
	found, err := repo.FindByUserID(ctx, credential.UserID)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}
	if found.UserID != credential.UserID {
		t.Error("FindByUserID() returned wrong credential")
	}

	// Test Update
	credential.PasswordHash = "newHash"
	err = repo.Update(ctx, credential)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Test Delete
	err = repo.Delete(ctx, credential.UserID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	exists, err = repo.Exists(ctx, credential.UserID)
	if err != nil {
		t.Errorf("Exists() after delete error = %v", err)
	}
	if exists {
		t.Error("Exists() should return false after Delete()")
	}
}

type mockPasswordResetRepository struct {
	tokens map[string]*PasswordResetToken
}

func (m *mockPasswordResetRepository) Save(ctx context.Context, token *PasswordResetToken) error {
	if m.tokens == nil {
		m.tokens = make(map[string]*PasswordResetToken)
	}
	m.tokens[token.Token] = token
	return nil
}

func (m *mockPasswordResetRepository) FindByToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	if resetToken, exists := m.tokens[token]; exists {
		return resetToken, nil
	}
	return nil, ErrPasswordResetNotFound
}

func (m *mockPasswordResetRepository) FindByUserID(ctx context.Context, userID string) ([]*PasswordResetToken, error) {
	var tokens []*PasswordResetToken
	for _, token := range m.tokens {
		if token.UserID == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (m *mockPasswordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	if resetToken, exists := m.tokens[token]; exists {
		resetToken.Used = true
		return nil
	}
	return ErrPasswordResetNotFound
}

func (m *mockPasswordResetRepository) Delete(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *mockPasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	for token, resetToken := range m.tokens {
		if resetToken.UserID == userID {
			delete(m.tokens, token)
		}
	}
	return nil
}

func (m *mockPasswordResetRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for token, resetToken := range m.tokens {
		if resetToken.ExpiresAt.Before(now) {
			delete(m.tokens, token)
		}
	}
	return nil
}

func TestPasswordResetRepository_Interface(t *testing.T) {
	// Test that our mock implements the interface
	var _ PasswordResetRepository = &mockPasswordResetRepository{}

	repo := &mockPasswordResetRepository{}
	ctx := context.Background()

	resetToken := &PasswordResetToken{
		Token:     "reset-token-123",
		UserID:    "test-user",
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		Used:      false,
	}

	// Test Save
	err := repo.Save(ctx, resetToken)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Test FindByToken
	found, err := repo.FindByToken(ctx, resetToken.Token)
	if err != nil {
		t.Errorf("FindByToken() error = %v", err)
	}
	if found.Token != resetToken.Token {
		t.Error("FindByToken() returned wrong token")
	}

	// Test FindByUserID
	tokens, err := repo.FindByUserID(ctx, resetToken.UserID)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("FindByUserID() returned %d tokens, want 1", len(tokens))
	}

	// Test MarkAsUsed
	err = repo.MarkAsUsed(ctx, resetToken.Token)
	if err != nil {
		t.Errorf("MarkAsUsed() error = %v", err)
	}

	// Verify token is marked as used
	found, err = repo.FindByToken(ctx, resetToken.Token)
	if err != nil {
		t.Errorf("FindByToken() after MarkAsUsed error = %v", err)
	}
	if !found.Used {
		t.Error("Token should be marked as used")
	}

	// Test Delete
	err = repo.Delete(ctx, resetToken.Token)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.FindByToken(ctx, resetToken.Token)
	if err != ErrPasswordResetNotFound {
		t.Error("FindByToken() after delete should return ErrPasswordResetNotFound")
	}
}

type mockExternalIdentityRepository struct {
	identities map[string]*ExternalIdentity
}

func (m *mockExternalIdentityRepository) LinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	if m.identities == nil {
		m.identities = make(map[string]*ExternalIdentity)
	}
	key := provider + ":" + externalID
	m.identities[key] = &ExternalIdentity{
		UserID:     userID,
		Provider:   provider,
		ExternalID: externalID,
		CreatedAt:  time.Now(),
	}
	return nil
}

func (m *mockExternalIdentityRepository) FindByExternalID(ctx context.Context, provider, externalID string) (string, error) {
	key := provider + ":" + externalID
	if identity, exists := m.identities[key]; exists {
		return identity.UserID, nil
	}
	return "", ErrExternalIdentityNotFound
}

func (m *mockExternalIdentityRepository) GetLinkedIdentities(ctx context.Context, userID string) ([]*ExternalIdentity, error) {
	var identities []*ExternalIdentity
	for _, identity := range m.identities {
		if identity.UserID == userID {
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func (m *mockExternalIdentityRepository) UnlinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	key := provider + ":" + externalID
	if identity, exists := m.identities[key]; exists && identity.UserID == userID {
		delete(m.identities, key)
		return nil
	}
	return ErrExternalIdentityNotFound
}

func (m *mockExternalIdentityRepository) UnlinkAllIdentities(ctx context.Context, userID string) error {
	for key, identity := range m.identities {
		if identity.UserID == userID {
			delete(m.identities, key)
		}
	}
	return nil
}

func (m *mockExternalIdentityRepository) IsLinked(ctx context.Context, provider, externalID string) (bool, error) {
	key := provider + ":" + externalID
	_, exists := m.identities[key]
	return exists, nil
}

func (m *mockExternalIdentityRepository) GetByProvider(ctx context.Context, provider string) ([]*ExternalIdentity, error) {
	var identities []*ExternalIdentity
	for _, identity := range m.identities {
		if identity.Provider == provider {
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func TestExternalIdentityRepository_Interface(t *testing.T) {
	// Test that our mock implements the interface
	var _ ExternalIdentityRepository = &mockExternalIdentityRepository{}

	repo := &mockExternalIdentityRepository{}
	ctx := context.Background()

	userID := "test-user"
	provider := "google"
	externalID := "google-123"

	// Test LinkIdentity
	err := repo.LinkIdentity(ctx, userID, provider, externalID)
	if err != nil {
		t.Errorf("LinkIdentity() error = %v", err)
	}

	// Test IsLinked
	linked, err := repo.IsLinked(ctx, provider, externalID)
	if err != nil {
		t.Errorf("IsLinked() error = %v", err)
	}
	if !linked {
		t.Error("IsLinked() should return true after LinkIdentity()")
	}

	// Test FindByExternalID
	foundUserID, err := repo.FindByExternalID(ctx, provider, externalID)
	if err != nil {
		t.Errorf("FindByExternalID() error = %v", err)
	}
	if foundUserID != userID {
		t.Errorf("FindByExternalID() returned %s, want %s", foundUserID, userID)
	}

	// Test GetLinkedIdentities
	identities, err := repo.GetLinkedIdentities(ctx, userID)
	if err != nil {
		t.Errorf("GetLinkedIdentities() error = %v", err)
	}
	if len(identities) != 1 {
		t.Errorf("GetLinkedIdentities() returned %d identities, want 1", len(identities))
	}

	// Test GetByProvider
	providerIdentities, err := repo.GetByProvider(ctx, provider)
	if err != nil {
		t.Errorf("GetByProvider() error = %v", err)
	}
	if len(providerIdentities) != 1 {
		t.Errorf("GetByProvider() returned %d identities, want 1", len(providerIdentities))
	}

	// Test UnlinkIdentity
	err = repo.UnlinkIdentity(ctx, userID, provider, externalID)
	if err != nil {
		t.Errorf("UnlinkIdentity() error = %v", err)
	}

	// Verify unlinking
	linked, err = repo.IsLinked(ctx, provider, externalID)
	if err != nil {
		t.Errorf("IsLinked() after unlink error = %v", err)
	}
	if linked {
		t.Error("IsLinked() should return false after UnlinkIdentity()")
	}
}
