package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"gorm.io/gorm"
)

// GormPasswordResetRepository implements PasswordResetRepository using GORM
type GormPasswordResetRepository struct {
	db *gorm.DB
}

// NewGormPasswordResetRepository creates a new GORM password reset repository
func NewGormPasswordResetRepository(db *gorm.DB) domain.PasswordResetRepository {
	return &GormPasswordResetRepository{db: db}
}

// Save creates a new password reset token
func (r *GormPasswordResetRepository) Save(ctx context.Context, token *domain.PasswordResetToken) error {
	if token == nil {
		return fmt.Errorf("password reset token cannot be nil")
	}

	if token.Token == "" || token.UserID == "" || token.Email == "" {
		return fmt.Errorf("password reset token is invalid: missing required fields")
	}

	model := &PasswordResetTokenModel{
		Token:     token.Token,
		UserID:    token.UserID,
		Email:     token.Email,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		Used:      token.Used,
	}

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to save password reset token: %w", err)
	}

	return nil
}

// FindByToken retrieves a password reset token by its token value
func (r *GormPasswordResetRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	var model PasswordResetTokenModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPasswordResetTokenNotFound
		}
		return nil, fmt.Errorf("failed to find password reset token: %w", err)
	}

	return &domain.PasswordResetToken{
		Token:     model.Token,
		UserID:    model.UserID,
		Email:     model.Email,
		ExpiresAt: model.ExpiresAt,
		CreatedAt: model.CreatedAt,
		Used:      model.Used,
	}, nil
}

// FindByUserID retrieves all password reset tokens for a user
func (r *GormPasswordResetRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.PasswordResetToken, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var models []PasswordResetTokenModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find password reset tokens by user ID: %w", err)
	}

	tokens := make([]*domain.PasswordResetToken, len(models))
	for i, model := range models {
		tokens[i] = &domain.PasswordResetToken{
			Token:     model.Token,
			UserID:    model.UserID,
			Email:     model.Email,
			ExpiresAt: model.ExpiresAt,
			CreatedAt: model.CreatedAt,
			Used:      model.Used,
		}
	}

	return tokens, nil
}

// MarkAsUsed marks a password reset token as used
func (r *GormPasswordResetRepository) MarkAsUsed(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	result := r.db.WithContext(ctx).Model(&PasswordResetTokenModel{}).
		Where("token = ? AND used = ?", token, false).
		Update("used", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPasswordResetTokenNotFound
	}

	return nil
}

// Delete removes a password reset token
func (r *GormPasswordResetRepository) Delete(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&PasswordResetTokenModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete password reset token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPasswordResetTokenNotFound
	}

	return nil
}

// DeleteByUserID removes all password reset tokens for a user
func (r *GormPasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&PasswordResetTokenModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete password reset tokens by user ID: %w", err)
	}

	return nil
}

// DeleteExpired removes all expired password reset tokens
func (r *GormPasswordResetRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&PasswordResetTokenModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete expired password reset tokens: %w", err)
	}

	return nil
}
