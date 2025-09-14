package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// GormAccountWriteRepository implements domain.AccountWriteRepository using GORM
type GormAccountWriteRepository struct {
	db *gorm.DB
}

// NewGormAccountWriteRepository creates a new GORM-based account write repository
func NewGormAccountWriteRepository(db *gorm.DB) domain.AccountWriteRepository {
	return &GormAccountWriteRepository{db: db}
}

// Create creates a new account in the database
func (r *GormAccountWriteRepository) Create(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}

	if strings.TrimSpace(account.ID()) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if strings.TrimSpace(account.OwnerID) == "" {
		return fmt.Errorf("owner ID cannot be empty")
	}

	if strings.TrimSpace(account.Name) == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(account.Settings)
	if err != nil {
		return fmt.Errorf("failed to serialize account settings: %w", err)
	}

	// Convert domain account to GORM model
	accountModel := &AccountModel{
		ID:          account.ID(),
		OwnerID:     account.OwnerID,
		Name:        account.Name,
		Description: account.Description,
		Settings:    string(settingsJSON),
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}

	err = r.db.WithContext(ctx).Create(accountModel).Error
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// Update updates an existing account in the database
func (r *GormAccountWriteRepository) Update(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}

	if strings.TrimSpace(account.ID()) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(account.Settings)
	if err != nil {
		return fmt.Errorf("failed to serialize account settings: %w", err)
	}

	// Convert domain account to GORM model
	accountModel := &AccountModel{
		ID:          account.ID(),
		OwnerID:     account.OwnerID,
		Name:        account.Name,
		Description: account.Description,
		Settings:    string(settingsJSON),
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&AccountModel{}).Where("id = ?", account.ID()).Updates(accountModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found: %s", account.ID())
	}

	return nil
}

// Delete deletes an account from the database (with cascade delete for related data)
func (r *GormAccountWriteRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	// Use a transaction to ensure all related data is deleted
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete related account members
		err := tx.Delete(&AccountMemberModel{}, "account_id = ?", id).Error
		if err != nil {
			return fmt.Errorf("failed to delete account members: %w", err)
		}

		// Delete related invitations
		err = tx.Delete(&InvitationModel{}, "account_id = ?", id).Error
		if err != nil {
			return fmt.Errorf("failed to delete invitations: %w", err)
		}

		// Delete the account
		result := tx.Delete(&AccountModel{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("failed to delete account: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("account not found: %s", id)
		}

		return nil
	})
}
