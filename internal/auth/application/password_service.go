package application

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/akeemphilbert/goro/internal/infrastructure/email"
	userDomain "github.com/akeemphilbert/goro/internal/user/domain"
)

// PasswordService handles password management operations
type PasswordService struct {
	userRepo       userDomain.UserRepository
	passwordRepo   domain.PasswordRepository
	resetRepo      domain.PasswordResetRepository
	emailService   email.Service
	tokenGenerator domain.SecureTokenGenerator
	hasher         domain.PasswordHasher
	validator      domain.PasswordValidator
	resetExpiry    time.Duration
	baseURL        string
}

// NewPasswordService creates a new password service
func NewPasswordService(
	userRepo userDomain.UserRepository,
	passwordRepo domain.PasswordRepository,
	resetRepo domain.PasswordResetRepository,
	emailService email.Service,
	tokenGenerator domain.SecureTokenGenerator,
	hasher domain.PasswordHasher,
	validator domain.PasswordValidator,
	resetExpiry time.Duration,
	baseURL string,
) *PasswordService {
	return &PasswordService{
		userRepo:       userRepo,
		passwordRepo:   passwordRepo,
		resetRepo:      resetRepo,
		emailService:   emailService,
		tokenGenerator: tokenGenerator,
		hasher:         hasher,
		validator:      validator,
		resetExpiry:    resetExpiry,
		baseURL:        baseURL,
	}
}

// SetPassword sets a new password for a user
func (s *PasswordService) SetPassword(ctx context.Context, userID, password string) error {
	// Validate the password strength
	if err := s.validator.Validate(password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Hash the password
	hash, salt, err := s.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create or update password credential
	credential := &domain.PasswordCredential{
		UserID:       userID,
		PasswordHash: hash,
		Salt:         salt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Check if password already exists
	existing, err := s.passwordRepo.FindByUserID(ctx, userID)
	if err != nil && err != domain.ErrPasswordCredentialNotFound {
		return fmt.Errorf("failed to check existing password: %w", err)
	}

	if existing != nil {
		// Update existing password
		if err := s.passwordRepo.Update(ctx, credential); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}
	} else {
		// Save new password
		if err := s.passwordRepo.Save(ctx, credential); err != nil {
			return fmt.Errorf("failed to save password: %w", err)
		}
	}

	return nil
}

// ChangePassword changes a user's password after verifying the current password
func (s *PasswordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Verify current password
	if err := s.ValidatePassword(ctx, userID, currentPassword); err != nil {
		return fmt.Errorf("current password verification failed: %w", err)
	}

	// Set the new password
	if err := s.SetPassword(ctx, userID, newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	return nil
}

// InitiatePasswordReset initiates a password reset process by sending a reset email
func (s *PasswordService) InitiatePasswordReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate reset token
	token, err := s.tokenGenerator.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create reset token record
	resetToken := &domain.PasswordResetToken{
		Token:     token,
		UserID:    user.ID(),
		Email:     email,
		ExpiresAt: time.Now().Add(s.resetExpiry),
		CreatedAt: time.Now(),
		Used:      false,
	}

	// Save reset token
	if err := s.resetRepo.Save(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	// Send reset email
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.baseURL, token)
	templateData := map[string]interface{}{
		"UserName":   user.Profile.Name,
		"ResetURL":   resetURL,
		"ExpiryTime": resetToken.ExpiresAt,
		"SupportURL": fmt.Sprintf("%s/support", s.baseURL),
	}

	if err := s.emailService.SendTemplatedEmail(ctx, "password_reset", templateData, email); err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

// CompletePasswordReset completes a password reset using a valid token
func (s *PasswordService) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	// Find and validate reset token
	resetToken, err := s.resetRepo.FindByToken(ctx, token)
	if err != nil {
		if err == domain.ErrPasswordResetTokenNotFound {
			return domain.ErrPasswordResetInvalid
		}
		return fmt.Errorf("failed to find reset token: %w", err)
	}

	// Validate token
	if !resetToken.IsValid() {
		if resetToken.Used {
			return domain.ErrPasswordResetUsed
		}
		if resetToken.IsExpired() {
			return domain.ErrPasswordResetExpired
		}
		return domain.ErrPasswordResetInvalid
	}

	// Set new password
	if err := s.SetPassword(ctx, resetToken.UserID, newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	// Mark token as used
	if err := s.resetRepo.MarkAsUsed(ctx, token); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// ValidatePassword validates a password against stored credentials
func (s *PasswordService) ValidatePassword(ctx context.Context, userID, password string) error {
	// Find password credential
	credential, err := s.passwordRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrPasswordCredentialNotFound {
			return domain.ErrInvalidCredentials
		}
		return fmt.Errorf("failed to find password credential: %w", err)
	}

	// Verify password
	if !s.hasher.Verify(password, credential.PasswordHash, credential.Salt) {
		return domain.ErrInvalidCredentials
	}

	return nil
}

// CleanupExpiredTokens removes expired password reset tokens
func (s *PasswordService) CleanupExpiredTokens(ctx context.Context) error {
	if err := s.resetRepo.DeleteExpired(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

// HasPassword checks if a user has a password set
func (s *PasswordService) HasPassword(ctx context.Context, userID string) (bool, error) {
	_, err := s.passwordRepo.FindByUserID(ctx, userID)
	if err != nil {
		if err == domain.ErrPasswordCredentialNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check password existence: %w", err)
	}
	return true, nil
}
