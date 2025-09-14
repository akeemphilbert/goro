package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"gorm.io/gorm"
)

// GormExternalIdentityRepository implements ExternalIdentityRepository using GORM
type GormExternalIdentityRepository struct {
	db *gorm.DB
}

// NewGormExternalIdentityRepository creates a new GORM external identity repository
func NewGormExternalIdentityRepository(db *gorm.DB) domain.ExternalIdentityRepository {
	return &GormExternalIdentityRepository{db: db}
}

// LinkIdentity creates a link between a user and an external identity
func (r *GormExternalIdentityRepository) LinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	if userID == "" || provider == "" || externalID == "" {
		return fmt.Errorf("userID, provider, and externalID cannot be empty")
	}

	// Check if the external identity is already linked to any user
	isLinked, err := r.IsLinked(ctx, provider, externalID)
	if err != nil {
		return fmt.Errorf("failed to check if external identity is linked: %w", err)
	}

	if isLinked {
		return domain.ErrExternalIdentityAlreadyLinked
	}

	model := &ExternalIdentityModel{
		UserID:     userID,
		Provider:   provider,
		ExternalID: externalID,
		CreatedAt:  time.Now(),
	}

	err = r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to link external identity: %w", err)
	}

	return nil
}

// FindByExternalID finds a user ID by external provider and ID
func (r *GormExternalIdentityRepository) FindByExternalID(ctx context.Context, provider, externalID string) (string, error) {
	if provider == "" || externalID == "" {
		return "", fmt.Errorf("provider and externalID cannot be empty")
	}

	var model ExternalIdentityModel
	err := r.db.WithContext(ctx).
		Where("provider = ? AND external_id = ?", provider, externalID).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", domain.ErrExternalIdentityNotFound
		}
		return "", fmt.Errorf("failed to find external identity: %w", err)
	}

	return model.UserID, nil
}

// GetLinkedIdentities retrieves all external identities linked to a user
func (r *GormExternalIdentityRepository) GetLinkedIdentities(ctx context.Context, userID string) ([]*domain.ExternalIdentity, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var models []ExternalIdentityModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get linked identities: %w", err)
	}

	identities := make([]*domain.ExternalIdentity, len(models))
	for i, model := range models {
		identities[i] = &domain.ExternalIdentity{
			ID:         model.ID,
			UserID:     model.UserID,
			Provider:   model.Provider,
			ExternalID: model.ExternalID,
			CreatedAt:  model.CreatedAt,
		}
	}

	return identities, nil
}

// UnlinkIdentity removes the link between a user and an external identity
func (r *GormExternalIdentityRepository) UnlinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	if userID == "" || provider == "" || externalID == "" {
		return fmt.Errorf("userID, provider, and externalID cannot be empty")
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ? AND external_id = ?", userID, provider, externalID).
		Delete(&ExternalIdentityModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to unlink external identity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrExternalIdentityNotFound
	}

	return nil
}

// UnlinkAllIdentities removes all external identity links for a user
func (r *GormExternalIdentityRepository) UnlinkAllIdentities(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&ExternalIdentityModel{}).Error
	if err != nil {
		return fmt.Errorf("failed to unlink all external identities: %w", err)
	}

	return nil
}

// IsLinked checks if an external identity is already linked to any user
func (r *GormExternalIdentityRepository) IsLinked(ctx context.Context, provider, externalID string) (bool, error) {
	if provider == "" || externalID == "" {
		return false, fmt.Errorf("provider and externalID cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&ExternalIdentityModel{}).
		Where("provider = ? AND external_id = ?", provider, externalID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if external identity is linked: %w", err)
	}

	return count > 0, nil
}

// GetByProvider retrieves all external identities for a specific provider
func (r *GormExternalIdentityRepository) GetByProvider(ctx context.Context, provider string) ([]*domain.ExternalIdentity, error) {
	if provider == "" {
		return nil, fmt.Errorf("provider cannot be empty")
	}

	var models []ExternalIdentityModel
	err := r.db.WithContext(ctx).Where("provider = ?", provider).Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get external identities by provider: %w", err)
	}

	identities := make([]*domain.ExternalIdentity, len(models))
	for i, model := range models {
		identities[i] = &domain.ExternalIdentity{
			ID:         model.ID,
			UserID:     model.UserID,
			Provider:   model.Provider,
			ExternalID: model.ExternalID,
			CreatedAt:  model.CreatedAt,
		}
	}

	return identities, nil
}
