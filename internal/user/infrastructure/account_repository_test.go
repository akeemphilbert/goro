package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// createTestAccount creates a test account in the database
func createTestAccount(t *testing.T, db *gorm.DB, id, ownerID, name, description string) {
	settingsJSON := `{"allow_invitations":true,"default_role_id":"member","max_members":100}`

	account := AccountModel{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		Settings:    settingsJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := db.Create(&account).Error
	require.NoError(t, err)
}

func TestGormAccountRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountRepository(db)
	ctx := context.Background()

	// Create test owner user first
	ownerID := "owner-user-1"
	createTestUser(t, db, ownerID, "https://example.com/users/owner-1", "owner1@example.com", "Owner 1", string(domain.UserStatusActive))

	// Create test account
	accountID := "test-account-1"
	name := "Test Account 1"
	description := "Test account description"

	createTestAccount(t, db, accountID, ownerID, name, description)

	t.Run("should return account when found", func(t *testing.T) {
		account, err := repo.GetByID(ctx, accountID)
		require.NoError(t, err)
		assert.NotNil(t, account)
		assert.Equal(t, accountID, account.ID())
		assert.Equal(t, ownerID, account.OwnerID)
		assert.Equal(t, name, account.Name)
		assert.Equal(t, description, account.Description)
		assert.True(t, account.Settings.AllowInvitations)
		assert.Equal(t, "member", account.Settings.DefaultRoleID)
		assert.Equal(t, 100, account.Settings.MaxMembers)
	})

	t.Run("should return error when account not found", func(t *testing.T) {
		account, err := repo.GetByID(ctx, "non-existent-account")
		assert.Error(t, err)
		assert.Nil(t, account)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		account, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, account)
	})
}

func TestGormAccountRepository_GetByOwner(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountRepository(db)
	ctx := context.Background()

	// Create test owner users
	owner1ID := "owner-user-2"
	owner2ID := "owner-user-3"
	createTestUser(t, db, owner1ID, "https://example.com/users/owner-2", "owner2@example.com", "Owner 2", string(domain.UserStatusActive))
	createTestUser(t, db, owner2ID, "https://example.com/users/owner-3", "owner3@example.com", "Owner 3", string(domain.UserStatusActive))

	// Create test accounts
	accounts := []struct {
		id, ownerID, name, description string
	}{
		{"account-1", owner1ID, "Account 1", "First account"},
		{"account-2", owner1ID, "Account 2", "Second account"},
		{"account-3", owner2ID, "Account 3", "Third account"},
	}

	for _, acc := range accounts {
		createTestAccount(t, db, acc.id, acc.ownerID, acc.name, acc.description)
	}

	t.Run("should return accounts for owner with multiple accounts", func(t *testing.T) {
		accountList, err := repo.GetByOwner(ctx, owner1ID)
		require.NoError(t, err)
		assert.Len(t, accountList, 2)

		// Verify all accounts belong to the correct owner
		for _, account := range accountList {
			assert.Equal(t, owner1ID, account.OwnerID)
		}
	})

	t.Run("should return single account for owner with one account", func(t *testing.T) {
		accountList, err := repo.GetByOwner(ctx, owner2ID)
		require.NoError(t, err)
		assert.Len(t, accountList, 1)
		assert.Equal(t, owner2ID, accountList[0].OwnerID)
		assert.Equal(t, "Account 3", accountList[0].Name)
	})

	t.Run("should return empty list for owner with no accounts", func(t *testing.T) {
		// Create owner without accounts
		noAccountsOwnerID := "owner-no-accounts"
		createTestUser(t, db, noAccountsOwnerID, "https://example.com/users/no-accounts", "noaccounts@example.com", "No Accounts", string(domain.UserStatusActive))

		accountList, err := repo.GetByOwner(ctx, noAccountsOwnerID)
		require.NoError(t, err)
		assert.Len(t, accountList, 0)
	})

	t.Run("should return error for empty owner ID", func(t *testing.T) {
		accountList, err := repo.GetByOwner(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, accountList)
	})

	t.Run("should return empty list for non-existent owner", func(t *testing.T) {
		accountList, err := repo.GetByOwner(ctx, "non-existent-owner")
		require.NoError(t, err)
		assert.Len(t, accountList, 0)
	})
}
