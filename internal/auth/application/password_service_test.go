package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/application"
	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/akeemphilbert/goro/internal/infrastructure/email"
	userDomain "github.com/akeemphilbert/goro/internal/user/domain"
)

// Mock implementations for testing

type mockUserRepository struct {
	users map[string]*userDomain.User
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*userDomain.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByWebID(ctx context.Context, webID string) (*userDomain.User, error) {
	for _, user := range m.users {
		if user.WebID == webID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) List(ctx context.Context, filter userDomain.UserFilter) ([]*userDomain.User, error) {
	var users []*userDomain.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, exists := m.users[id]
	return exists, nil
}

type mockPasswordRepository struct {
	credentials map[string]*domain.PasswordCredential
}

func (m *mockPasswordRepository) Save(ctx context.Context, credential *domain.PasswordCredential) error {
	m.credentials[credential.UserID] = credential
	return nil
}

func (m *mockPasswordRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	if cred, exists := m.credentials[userID]; exists {
		return cred, nil
	}
	return nil, domain.ErrPasswordCredentialNotFound
}

func (m *mockPasswordRepository) Update(ctx context.Context, credential *domain.PasswordCredential) error {
	if _, exists := m.credentials[credential.UserID]; !exists {
		return domain.ErrPasswordCredentialNotFound
	}
	m.credentials[credential.UserID] = credential
	return nil
}

func (m *mockPasswordRepository) Delete(ctx context.Context, userID string) error {
	delete(m.credentials, userID)
	return nil
}

func (m *mockPasswordRepository) Exists(ctx context.Context, userID string) (bool, error) {
	_, exists := m.credentials[userID]
	return exists, nil
}

type mockPasswordResetRepository struct {
	tokens map[string]*domain.PasswordResetToken
}

func (m *mockPasswordResetRepository) Save(ctx context.Context, token *domain.PasswordResetToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *mockPasswordResetRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	if resetToken, exists := m.tokens[token]; exists {
		return resetToken, nil
	}
	return nil, domain.ErrPasswordResetTokenNotFound
}

func (m *mockPasswordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	if resetToken, exists := m.tokens[token]; exists {
		resetToken.MarkAsUsed()
		return nil
	}
	return domain.ErrPasswordResetTokenNotFound
}

func (m *mockPasswordResetRepository) DeleteExpired(ctx context.Context) error {
	for token, resetToken := range m.tokens {
		if resetToken.IsExpired() {
			delete(m.tokens, token)
		}
	}
	return nil
}

func (m *mockPasswordResetRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.PasswordResetToken, error) {
	var tokens []*domain.PasswordResetToken
	for _, token := range m.tokens {
		if token.UserID == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
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

type mockEmailService struct {
	sentEmails []email.Email
}

func (m *mockEmailService) SendEmail(ctx context.Context, email *email.Email) error {
	m.sentEmails = append(m.sentEmails, *email)
	return nil
}

func (m *mockEmailService) SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients ...string) error {
	emailMsg := email.Email{
		To:      recipients,
		Subject: "Password Reset",
	}
	m.sentEmails = append(m.sentEmails, emailMsg)
	return nil
}

type mockTokenGenerator struct {
	tokens []string
	index  int
}

func (m *mockTokenGenerator) GenerateToken() (string, error) {
	if m.index >= len(m.tokens) {
		return "random-token", nil
	}
	token := m.tokens[m.index]
	m.index++
	return token, nil
}

type mockPasswordHasher struct{}

func (m *mockPasswordHasher) Hash(password string) (hash, salt string, err error) {
	return "hashed-" + password, "salt-123", nil
}

func (m *mockPasswordHasher) Verify(password, hash, salt string) bool {
	return hash == "hashed-"+password && salt == "salt-123"
}

type mockPasswordValidator struct {
	shouldFail bool
}

func (m *mockPasswordValidator) Validate(password string) error {
	if m.shouldFail {
		return domain.ErrPasswordTooWeak
	}
	return nil
}

func setupPasswordService() (*application.PasswordService, *mockUserRepository, *mockPasswordRepository, *mockPasswordResetRepository, *mockEmailService) {
	userRepo := &mockUserRepository{
		users: make(map[string]*userDomain.User),
	}
	passwordRepo := &mockPasswordRepository{
		credentials: make(map[string]*domain.PasswordCredential),
	}
	resetRepo := &mockPasswordResetRepository{
		tokens: make(map[string]*domain.PasswordResetToken),
	}
	emailService := &mockEmailService{}
	tokenGenerator := &mockTokenGenerator{
		tokens: []string{"test-token-1", "test-token-2"},
	}
	hasher := &mockPasswordHasher{}
	validator := &mockPasswordValidator{}

	service := application.NewPasswordService(
		userRepo,
		passwordRepo,
		resetRepo,
		emailService,
		tokenGenerator,
		hasher,
		validator,
		time.Hour,
		"https://example.com",
	)

	return service, userRepo, passwordRepo, resetRepo, emailService
}

func TestPasswordService_SetPassword(t *testing.T) {
	service, _, passwordRepo, _, _ := setupPasswordService()
	ctx := context.Background()

	tests := []struct {
		name     string
		userID   string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			userID:   "user1",
			password: "validPassword123!",
			wantErr:  false,
		},
		{
			name:     "empty user ID",
			userID:   "",
			password: "validPassword123!",
			wantErr:  false, // Service doesn't validate userID
		},
		{
			name:     "empty password",
			userID:   "user1",
			password: "",
			wantErr:  false, // Validator mock doesn't fail by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetPassword(ctx, tt.userID, tt.password)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordService.SetPassword() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordService.SetPassword() unexpected error: %v", err)
			}

			if !tt.wantErr && tt.userID != "" {
				// Verify password was saved
				cred, err := passwordRepo.FindByUserID(ctx, tt.userID)
				if err != nil {
					t.Errorf("Password was not saved: %v", err)
				} else if cred.UserID != tt.userID {
					t.Errorf("Saved credential has wrong user ID: got %s, want %s", cred.UserID, tt.userID)
				}
			}
		})
	}
}

func TestPasswordService_SetPassword_WithValidation(t *testing.T) {
	// Create a service with a validator that fails
	userRepo := &mockUserRepository{
		users: make(map[string]*userDomain.User),
	}
	passwordRepo := &mockPasswordRepository{
		credentials: make(map[string]*domain.PasswordCredential),
	}
	resetRepo := &mockPasswordResetRepository{
		tokens: make(map[string]*domain.PasswordResetToken),
	}
	emailService := &mockEmailService{}
	tokenGenerator := &mockTokenGenerator{
		tokens: []string{"test-token-1", "test-token-2"},
	}
	hasher := &mockPasswordHasher{}
	validator := &mockPasswordValidator{shouldFail: true} // This one fails

	service := application.NewPasswordService(
		userRepo,
		passwordRepo,
		resetRepo,
		emailService,
		tokenGenerator,
		hasher,
		validator,
		time.Hour,
		"https://example.com",
	)

	ctx := context.Background()

	err := service.SetPassword(ctx, "user1", "weakpassword")
	if err == nil {
		t.Error("PasswordService.SetPassword() expected validation error, got nil")
	}
}

func TestPasswordService_ChangePassword(t *testing.T) {
	service, _, _, _, _ := setupPasswordService()
	ctx := context.Background()

	// Set up existing password
	userID := "user1"
	currentPassword := "currentPass123!"
	newPassword := "newPass456!"

	// First set a password
	err := service.SetPassword(ctx, userID, currentPassword)
	if err != nil {
		t.Fatalf("Failed to set initial password: %v", err)
	}

	tests := []struct {
		name            string
		userID          string
		currentPassword string
		newPassword     string
		wantErr         bool
	}{
		{
			name:            "valid password change",
			userID:          userID,
			currentPassword: currentPassword,
			newPassword:     newPassword,
			wantErr:         false,
		},
		{
			name:            "wrong current password",
			userID:          userID,
			currentPassword: "wrongPassword",
			newPassword:     newPassword,
			wantErr:         true,
		},
		{
			name:            "non-existent user",
			userID:          "nonexistent",
			currentPassword: currentPassword,
			newPassword:     newPassword,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ChangePassword(ctx, tt.userID, tt.currentPassword, tt.newPassword)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordService.ChangePassword() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordService.ChangePassword() unexpected error: %v", err)
			}

			if !tt.wantErr {
				// Verify new password works
				err := service.ValidatePassword(ctx, tt.userID, tt.newPassword)
				if err != nil {
					t.Errorf("New password validation failed: %v", err)
				}

				// Verify old password no longer works
				err = service.ValidatePassword(ctx, tt.userID, tt.currentPassword)
				if err == nil {
					t.Error("Old password should no longer work")
				}
			}
		})
	}
}

func TestPasswordService_InitiatePasswordReset(t *testing.T) {
	service, userRepo, _, resetRepo, emailService := setupPasswordService()
	ctx := context.Background()

	// Set up test user
	user, err := userDomain.NewUser(ctx, "user1", "https://example.com/user1", "test@example.com", userDomain.UserProfile{
		Name: "Test User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userRepo.users[user.ID()] = user

	tests := []struct {
		name      string
		email     string
		wantErr   bool
		wantEmail bool
	}{
		{
			name:      "valid email",
			email:     "test@example.com",
			wantErr:   false,
			wantEmail: true,
		},
		{
			name:      "non-existent email",
			email:     "nonexistent@example.com",
			wantErr:   false, // Service doesn't reveal if email exists
			wantEmail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialEmailCount := len(emailService.sentEmails)
			initialTokenCount := len(resetRepo.tokens)

			err := service.InitiatePasswordReset(ctx, tt.email)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordService.InitiatePasswordReset() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordService.InitiatePasswordReset() unexpected error: %v", err)
			}

			if tt.wantEmail {
				if len(emailService.sentEmails) != initialEmailCount+1 {
					t.Errorf("Expected email to be sent, but email count didn't increase")
				}
				if len(resetRepo.tokens) != initialTokenCount+1 {
					t.Errorf("Expected reset token to be created, but token count didn't increase")
				}
			} else {
				if len(emailService.sentEmails) != initialEmailCount {
					t.Errorf("Expected no email to be sent, but email count increased")
				}
			}
		})
	}
}

func TestPasswordService_CompletePasswordReset(t *testing.T) {
	service, _, _, resetRepo, _ := setupPasswordService()
	ctx := context.Background()

	// Set up valid reset token
	validToken := "valid-token"
	expiredToken := "expired-token"
	usedToken := "used-token"

	resetRepo.tokens[validToken] = &domain.PasswordResetToken{
		Token:     validToken,
		UserID:    "user1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		Used:      false,
	}

	resetRepo.tokens[expiredToken] = &domain.PasswordResetToken{
		Token:     expiredToken,
		UserID:    "user1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		CreatedAt: time.Now(),
		Used:      false,
	}

	resetRepo.tokens[usedToken] = &domain.PasswordResetToken{
		Token:     usedToken,
		UserID:    "user1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		Used:      true, // Already used
	}

	tests := []struct {
		name        string
		token       string
		newPassword string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "valid token and password",
			token:       validToken,
			newPassword: "newPassword123!",
			wantErr:     false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			newPassword: "newPassword123!",
			wantErr:     true,
			expectedErr: domain.ErrPasswordResetInvalid,
		},
		{
			name:        "expired token",
			token:       expiredToken,
			newPassword: "newPassword123!",
			wantErr:     true,
			expectedErr: domain.ErrPasswordResetExpired,
		},
		{
			name:        "used token",
			token:       usedToken,
			newPassword: "newPassword123!",
			wantErr:     true,
			expectedErr: domain.ErrPasswordResetUsed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CompletePasswordReset(ctx, tt.token, tt.newPassword)

			if tt.wantErr {
				if err == nil {
					t.Errorf("PasswordService.CompletePasswordReset() expected error, got nil")
				} else if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("PasswordService.CompletePasswordReset() expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("PasswordService.CompletePasswordReset() unexpected error: %v", err)
				}

				// Verify token is marked as used
				token, _ := resetRepo.FindByToken(ctx, tt.token)
				if !token.Used {
					t.Error("Token should be marked as used")
				}
			}
		})
	}
}

func TestPasswordService_ValidatePassword(t *testing.T) {
	service, _, _, _, _ := setupPasswordService()
	ctx := context.Background()

	// Set up test password
	userID := "user1"
	password := "testPassword123!"
	err := service.SetPassword(ctx, userID, password)
	if err != nil {
		t.Fatalf("Failed to set test password: %v", err)
	}

	tests := []struct {
		name     string
		userID   string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			userID:   userID,
			password: password,
			wantErr:  false,
		},
		{
			name:     "invalid password",
			userID:   userID,
			password: "wrongPassword",
			wantErr:  true,
		},
		{
			name:     "non-existent user",
			userID:   "nonexistent",
			password: password,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(ctx, tt.userID, tt.password)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordService.ValidatePassword() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordService.ValidatePassword() unexpected error: %v", err)
			}
		})
	}
}

func TestPasswordService_HasPassword(t *testing.T) {
	service, _, _, _, _ := setupPasswordService()
	ctx := context.Background()

	// Set up user with password
	userWithPassword := "user1"
	err := service.SetPassword(ctx, userWithPassword, "testPassword123!")
	if err != nil {
		t.Fatalf("Failed to set test password: %v", err)
	}

	tests := []struct {
		name    string
		userID  string
		wantHas bool
		wantErr bool
	}{
		{
			name:    "user with password",
			userID:  userWithPassword,
			wantHas: true,
			wantErr: false,
		},
		{
			name:    "user without password",
			userID:  "user2",
			wantHas: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, err := service.HasPassword(ctx, tt.userID)

			if tt.wantErr && err == nil {
				t.Errorf("PasswordService.HasPassword() expected error, got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("PasswordService.HasPassword() unexpected error: %v", err)
			}

			if has != tt.wantHas {
				t.Errorf("PasswordService.HasPassword() = %v, want %v", has, tt.wantHas)
			}
		})
	}
}

func TestPasswordService_CleanupExpiredTokens(t *testing.T) {
	service, _, _, resetRepo, _ := setupPasswordService()
	ctx := context.Background()

	// Set up tokens - some expired, some valid
	resetRepo.tokens["valid-token"] = &domain.PasswordResetToken{
		Token:     "valid-token",
		UserID:    "user1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
		Used:      false,
	}

	resetRepo.tokens["expired-token"] = &domain.PasswordResetToken{
		Token:     "expired-token",
		UserID:    "user2",
		Email:     "test2@example.com",
		ExpiresAt: time.Now().Add(-time.Hour), // Expired
		CreatedAt: time.Now(),
		Used:      false,
	}

	initialCount := len(resetRepo.tokens)

	err := service.CleanupExpiredTokens(ctx)
	if err != nil {
		t.Errorf("PasswordService.CleanupExpiredTokens() unexpected error: %v", err)
	}

	finalCount := len(resetRepo.tokens)
	if finalCount >= initialCount {
		t.Errorf("Expected token count to decrease after cleanup, got %d, want < %d", finalCount, initialCount)
	}

	// Verify valid token still exists
	if _, exists := resetRepo.tokens["valid-token"]; !exists {
		t.Error("Valid token should not be deleted during cleanup")
	}

	// Verify expired token was removed
	if _, exists := resetRepo.tokens["expired-token"]; exists {
		t.Error("Expired token should be deleted during cleanup")
	}
}
