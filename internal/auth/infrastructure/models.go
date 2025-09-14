package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// SessionModel represents the GORM model for sessions table
type SessionModel struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)"`
	UserID       string    `gorm:"not null;type:varchar(255);index:idx_session_user"`
	WebID        string    `gorm:"not null;type:varchar(500)"`
	AccountID    string    `gorm:"type:varchar(255);index:idx_session_account"` // Optional - user might not be in account context
	RoleID       string    `gorm:"type:varchar(255);index:idx_session_role"`    // Optional - user's role in the account
	TokenHash    string    `gorm:"not null;type:varchar(255)"`
	ExpiresAt    time.Time `gorm:"not null;index:idx_session_expires"`
	CreatedAt    time.Time `gorm:"not null"`
	LastActivity time.Time `gorm:"not null;index:idx_session_activity"`
}

// TableName specifies the table name for SessionModel
func (SessionModel) TableName() string {
	return "session_models"
}

// UpdateActivity updates the last activity timestamp with logging
func (s *SessionModel) UpdateActivity(ctx context.Context) error {
	log.Context(ctx).Debugf("[SessionModel.UpdateActivity] Updating session activity: sessionID=%s", s.ID)

	s.LastActivity = time.Now()

	log.Context(ctx).Debugf("[SessionModel.UpdateActivity] Session activity updated: sessionID=%s, lastActivity=%s", s.ID, s.LastActivity)
	return nil
}

// IsExpired checks if the session has expired
func (s *SessionModel) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session model has required fields
func (s *SessionModel) IsValid() bool {
	return s.ID != "" && s.UserID != "" && s.WebID != "" && s.TokenHash != ""
}

// PasswordCredentialModel represents the GORM model for password credentials table
type PasswordCredentialModel struct {
	UserID       string    `gorm:"primaryKey;type:varchar(255)"`
	PasswordHash string    `gorm:"not null;type:varchar(255)"`
	Salt         string    `gorm:"not null;type:varchar(255)"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// TableName specifies the table name for PasswordCredentialModel
func (PasswordCredentialModel) TableName() string {
	return "password_credential_models"
}

// UpdatePassword updates the password hash and salt with logging
func (pc *PasswordCredentialModel) UpdatePassword(ctx context.Context, hash, salt string) error {
	log.Context(ctx).Debugf("[PasswordCredentialModel.UpdatePassword] Updating password: userID=%s", pc.UserID)

	if hash == "" || salt == "" {
		err := fmt.Errorf("password hash and salt cannot be empty")
		log.Context(ctx).Debug("[PasswordCredentialModel.UpdatePassword] Validation failed: hash or salt empty")
		return err
	}

	pc.PasswordHash = hash
	pc.Salt = salt
	pc.UpdatedAt = time.Now()

	log.Context(ctx).Debugf("[PasswordCredentialModel.UpdatePassword] Password updated successfully: userID=%s", pc.UserID)
	return nil
}

// IsValid checks if the password credential model has required fields
func (pc *PasswordCredentialModel) IsValid() bool {
	return pc.UserID != "" && pc.PasswordHash != "" && pc.Salt != ""
}

// PasswordResetTokenModel represents the GORM model for password reset tokens table
type PasswordResetTokenModel struct {
	Token     string    `gorm:"primaryKey;type:varchar(255)"`
	UserID    string    `gorm:"not null;type:varchar(255);index:idx_reset_token_user"`
	Email     string    `gorm:"not null;type:varchar(255);index:idx_reset_token_email"`
	ExpiresAt time.Time `gorm:"not null;index:idx_reset_token_expires"`
	CreatedAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false;index:idx_reset_token_used"`
}

// TableName specifies the table name for PasswordResetTokenModel
func (PasswordResetTokenModel) TableName() string {
	return "password_reset_token_models"
}

// MarkAsUsed marks the token as used with logging
func (prt *PasswordResetTokenModel) MarkAsUsed(ctx context.Context) error {
	log.Context(ctx).Debugf("[PasswordResetTokenModel.MarkAsUsed] Marking token as used: token=%s", prt.Token)

	if prt.Used {
		err := fmt.Errorf("token is already used")
		log.Context(ctx).Debug("[PasswordResetTokenModel.MarkAsUsed] Validation failed: token already used")
		return err
	}

	prt.Used = true

	log.Context(ctx).Debugf("[PasswordResetTokenModel.MarkAsUsed] Token marked as used: token=%s", prt.Token)
	return nil
}

// IsExpired checks if the reset token has expired
func (prt *PasswordResetTokenModel) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt)
}

// IsValid checks if the reset token is valid (not expired, not used, has required fields)
func (prt *PasswordResetTokenModel) IsValid() bool {
	if prt.Token == "" || prt.UserID == "" || prt.Email == "" {
		return false
	}
	return !prt.IsExpired() && !prt.Used
}

// ExternalIdentityModel represents the GORM model for external identity linking table
type ExternalIdentityModel struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	UserID     string    `gorm:"not null;type:varchar(255);index:idx_external_identity_user"`
	Provider   string    `gorm:"not null;type:varchar(100);index:idx_external_identity_provider"`
	ExternalID string    `gorm:"not null;type:varchar(255)"`
	CreatedAt  time.Time `gorm:"not null"`
}

// TableName specifies the table name for ExternalIdentityModel
func (ExternalIdentityModel) TableName() string {
	return "external_identity_models"
}

// BeforeCreate hook to add unique constraint validation
func (eim *ExternalIdentityModel) BeforeCreate(tx *gorm.DB) error {
	// Check for existing provider + external_id combination
	var count int64
	err := tx.Model(&ExternalIdentityModel{}).
		Where("provider = ? AND external_id = ?", eim.Provider, eim.ExternalID).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check for existing external identity: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("external identity already exists for provider %s and external ID %s", eim.Provider, eim.ExternalID)
	}

	return nil
}

// IsValid checks if the external identity model has required fields
func (eim *ExternalIdentityModel) IsValid() bool {
	return eim.UserID != "" && eim.Provider != "" && eim.ExternalID != ""
}

// UpdateProvider updates the provider information with logging
func (eim *ExternalIdentityModel) UpdateProvider(ctx context.Context, provider string) error {
	log.Context(ctx).Debugf("[ExternalIdentityModel.UpdateProvider] Updating provider: identityID=%d, currentProvider=%s, newProvider=%s",
		eim.ID, eim.Provider, provider)

	if provider == "" {
		err := fmt.Errorf("provider cannot be empty")
		log.Context(ctx).Debug("[ExternalIdentityModel.UpdateProvider] Validation failed: provider cannot be empty")
		return err
	}

	eim.Provider = provider

	log.Context(ctx).Debugf("[ExternalIdentityModel.UpdateProvider] Provider updated successfully: identityID=%d, provider=%s", eim.ID, provider)
	return nil
}
