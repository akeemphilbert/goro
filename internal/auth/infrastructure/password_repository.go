package infrastructure

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"gorm.io/gorm"
)

// GormPasswordRepository implements PasswordRepository using GORM
type GormPasswordRepository struct {
	db *gorm.DB
}

// NewGormPasswordRepository creates a new GORM password repository
func NewGormPasswordRepository(db *gorm.DB) domain.PasswordRepository {
	return &GormPasswordRepository{db: db}
}

// Save creates or updates password credentials for a user
func (r *GormPasswordRepository) Save(ctx context.Context, credential *domain.PasswordCredential) error {
	if credential == nil {
		return fmt.Errorf("password credential cannot be nil")
	}

	if !credential.IsValid() {
		return fmt.Errorf("password credential is invalid")
	}

	model := &PasswordCredentialModel{
		UserID:       credential.UserID,
		PasswordHash: credential.PasswordHash,
		Salt:         credential.Salt,
		CreatedAt:    credential.CreatedAt,
		UpdatedAt:    credential.UpdatedAt,
	}

	err := r.db.WithContext(ctx).Save(model).Error
	if err != nil {
		return fmt.Errorf("failed to save password credential: %w", err)
	}

	return nil
}

// FindByUserID retrieves password credentials for a user
func (r *GormPasswordRepository) FindByUserID(ctx context.Context, userID string) (*domain.PasswordCredential, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var model PasswordCredentialModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPasswordCredentialNotFound
		}
		return nil, fmt.Errorf("failed to find password credential by user ID: %w", err)
	}

	return &domain.PasswordCredential{
		UserID:       model.UserID,
		PasswordHash: model.PasswordHash,
		Salt:         model.Salt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

// Update modifies existing password credentials
func (r *GormPasswordRepository) Update(ctx context.Context, credential *domain.PasswordCredential) error {
	if credential == nil {
		return fmt.Errorf("password credential cannot be nil")
	}

	if !credential.IsValid() {
		return fmt.Errorf("password credential is invalid")
	}

	result := r.db.WithContext(ctx).Model(&PasswordCredentialModel{}).
		Where("user_id = ?", credential.UserID).
		Updates(map[string]interface{}{
			"password_hash": credential.PasswordHash,
			"salt":          credential.Salt,
			"updated_at":    credential.UpdatedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update password credential: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPasswordCredentialNotFound
	}

	return nil
}

// Delete removes password credentials for a user
func (r *GormPasswordRepository) Delete(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&PasswordCredentialModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete password credential: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPasswordCredentialNotFound
	}

	return nil
}

// Exists checks if password credentials exist for a user
func (r *GormPasswordRepository) Exists(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&PasswordCredentialModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if password credential exists: %w", err)
	}

	return count > 0, nil
}
