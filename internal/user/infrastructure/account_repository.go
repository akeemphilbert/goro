package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// GormAccountRepository implements domain.AccountRepository using GORM
type GormAccountRepository struct {
	db *gorm.DB
}

// NewGormAccountRepository creates a new GORM-based account repository
func NewGormAccountRepository(db *gorm.DB) domain.AccountRepository {
	return &GormAccountRepository{db: db}
}

// GetByID retrieves an account by ID
func (r *GormAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var accountModel AccountModel
	err := r.db.WithContext(ctx).First(&accountModel, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get account by ID %s: %w", id, err)
	}

	return r.modelToDomain(&accountModel)
}

// GetByOwner retrieves accounts by owner ID
func (r *GormAccountRepository) GetByOwner(ctx context.Context, ownerID string) ([]*domain.Account, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, fmt.Errorf("owner ID cannot be empty")
	}

	var accountModels []AccountModel
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&accountModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by owner %s: %w", ownerID, err)
	}

	accounts := make([]*domain.Account, len(accountModels))
	for i, model := range accountModels {
		account, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert account model to domain: %w", err)
		}
		accounts[i] = account
	}

	return accounts, nil
}

// modelToDomain converts an AccountModel to a domain.Account
func (r *GormAccountRepository) modelToDomain(model *AccountModel) (*domain.Account, error) {
	// Deserialize settings from JSON
	var settings domain.AccountSettings
	if model.Settings != "" {
		err := json.Unmarshal([]byte(model.Settings), &settings)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize account settings: %w", err)
		}
	} else {
		// Default settings if none provided
		settings = domain.AccountSettings{
			AllowInvitations: true,
			DefaultRoleID:    "member",
			MaxMembers:       100,
		}
	}

	account := &domain.Account{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		OwnerID:     model.OwnerID,
		Name:        model.Name,
		Description: model.Description,
		Settings:    settings,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return account, nil
}
