package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

func TestGormAccountWriteRepository_Create(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountWriteRepository(db)
	ctx := context.Background()

	t.Run("should create account successfully", func(t *testing.T) {
		settings := domain.AccountSettings{
			AllowInvitations: true,
			DefaultRoleID:    "member",
			MaxMembers:       50,
		}

		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity("account-123"),
			OwnerID:     "owner-123",
			Name:        "Test Account",
			Description: "A test account for testing",
			Settings:    settings,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, account)
		require.NoError(t, err)

		// Verify account was created in database
		var accountModel AccountModel
		err = db.First(&accountModel, "id = ?", "account-123").Error
		require.NoError(t, err)
		assert.Equal(t, "account-123", accountModel.ID)
		assert.Equal(t, "owner-123", accountModel.OwnerID)
		assert.Equal(t, "Test Account", accountModel.Name)
		assert.Equal(t, "A test account for testing", accountModel.Description)
		assert.Contains(t, accountModel.Settings, "allow_invitations")
	})

	t.Run("should return error for nil account", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account cannot be nil")
	})

	t.Run("should return error for account with empty ID", func(t *testing.T) {
		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity(""),
			OwnerID:     "owner-456",
			Name:        "No ID Account",
		}

		err := repo.Create(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("should return error for account with empty owner ID", func(t *testing.T) {
		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity("account-no-owner"),
			Name:        "No Owner Account",
		}

		err := repo.Create(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID cannot be empty")
	})

	t.Run("should return error for account with empty name", func(t *testing.T) {
		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity("account-no-name"),
			OwnerID:     "owner-789",
		}

		err := repo.Create(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account name cannot be empty")
	})
}

func TestGormAccountWriteRepository_Update(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountWriteRepository(db)
	ctx := context.Background()

	t.Run("should update account successfully", func(t *testing.T) {
		// Create initial account
		createTestAccount(t, db, "account-update-1", "owner-update-1", "Original Account", "Original description")

		// Update account
		settings := domain.AccountSettings{
			AllowInvitations: false,
			DefaultRoleID:    "viewer",
			MaxMembers:       25,
		}

		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity("account-update-1"),
			OwnerID:     "owner-update-1",
			Name:        "Updated Account",
			Description: "Updated description",
			Settings:    settings,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		}

		err := repo.Update(ctx, account)
		require.NoError(t, err)

		// Verify account was updated in database
		var accountModel AccountModel
		err = db.First(&accountModel, "id = ?", "account-update-1").Error
		require.NoError(t, err)
		assert.Equal(t, "Updated Account", accountModel.Name)
		assert.Equal(t, "Updated description", accountModel.Description)
		assert.Contains(t, accountModel.Settings, "allow_invitations")
		assert.Contains(t, accountModel.Settings, "false")
	})

	t.Run("should return error for nil account", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account cannot be nil")
	})

	t.Run("should return error for account with empty ID", func(t *testing.T) {
		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity(""),
			OwnerID:     "owner-update-2",
			Name:        "No ID Update",
		}

		err := repo.Update(ctx, account)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("should return error for non-existent account", func(t *testing.T) {
		account := &domain.Account{
			BasicEntity: pericarpdomain.NewEntity("non-existent-account"),
			OwnerID:     "owner-non-existent",
			Name:        "Non-existent Account",
		}

		err := repo.Update(ctx, account)
		assert.Error(t, err)
	})
}

func TestGormAccountWriteRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountWriteRepository(db)
	ctx := context.Background()

	t.Run("should delete account successfully", func(t *testing.T) {
		// Create account to delete
		createTestAccount(t, db, "account-delete-1", "owner-delete-1", "Delete Account", "Account to be deleted")

		err := repo.Delete(ctx, "account-delete-1")
		require.NoError(t, err)

		// Verify account was deleted from database
		var count int64
		err = db.Model(&AccountModel{}).Where("id = ?", "account-delete-1").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("should return error for non-existent account", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-account")
		assert.Error(t, err)
	})

	t.Run("should cascade delete related members and invitations", func(t *testing.T) {
		// Create account with members and invitations
		createTestAccount(t, db, "account-cascade-1", "owner-cascade-1", "Cascade Account", "Account for cascade testing")

		// Create related data
		createTestAccountMember(t, db, "member-cascade-1", "account-cascade-1", "user-cascade-1", "member", "owner-cascade-1", time.Now())
		createTestInvitation(t, db, "invitation-cascade-1", "account-cascade-1", "invite@example.com", "member", "token-cascade-1", "owner-cascade-1", string(domain.InvitationStatusPending), time.Now().Add(24*time.Hour))

		err := repo.Delete(ctx, "account-cascade-1")
		require.NoError(t, err)

		// Verify account was deleted
		var accountCount int64
		err = db.Model(&AccountModel{}).Where("id = ?", "account-cascade-1").Count(&accountCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), accountCount)

		// Verify related members were deleted
		var memberCount int64
		err = db.Model(&AccountMemberModel{}).Where("account_id = ?", "account-cascade-1").Count(&memberCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), memberCount)

		// Verify related invitations were deleted
		var invitationCount int64
		err = db.Model(&InvitationModel{}).Where("account_id = ?", "account-cascade-1").Count(&invitationCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), invitationCount)
	})
}
