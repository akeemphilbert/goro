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

func TestGormInvitationWriteRepository_Create(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationWriteRepository(db)
	ctx := context.Background()

	t.Run("should create invitation successfully", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)

		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-create-1"),
			AccountID:   "account-invite-create-1",
			Email:       "invite.create@example.com",
			RoleID:      "member",
			Token:       "unique-token-create-1",
			InvitedBy:   "owner-invite-create-1",
			Status:      domain.InvitationStatusPending,
			ExpiresAt:   expiresAt,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		// Verify invitation was created in database
		var invitationModel InvitationModel
		err = db.First(&invitationModel, "id = ?", "invitation-create-1").Error
		require.NoError(t, err)
		assert.Equal(t, "invitation-create-1", invitationModel.ID)
		assert.Equal(t, "account-invite-create-1", invitationModel.AccountID)
		assert.Equal(t, "invite.create@example.com", invitationModel.Email)
		assert.Equal(t, "member", invitationModel.RoleID)
		assert.Equal(t, "unique-token-create-1", invitationModel.Token)
		assert.Equal(t, "owner-invite-create-1", invitationModel.InvitedBy)
		assert.Equal(t, string(domain.InvitationStatusPending), invitationModel.Status)
		assert.WithinDuration(t, expiresAt, invitationModel.ExpiresAt, time.Second)
	})

	t.Run("should return error for nil invitation", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation cannot be nil")
	})

	t.Run("should return error for invitation with empty ID", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity(""),
			AccountID:   "account-123",
			Email:       "test@example.com",
			RoleID:      "member",
			Token:       "token-123",
		}

		err := repo.Create(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation ID cannot be empty")
	})

	t.Run("should return error for invitation with empty account ID", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-no-account"),
			Email:       "test@example.com",
			RoleID:      "member",
			Token:       "token-456",
		}

		err := repo.Create(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("should return error for invitation with empty email", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-no-email"),
			AccountID:   "account-456",
			RoleID:      "member",
			Token:       "token-789",
		}

		err := repo.Create(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot be empty")
	})

	t.Run("should return error for invitation with empty role ID", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-no-role"),
			AccountID:   "account-789",
			Email:       "test@example.com",
			Token:       "token-abc",
		}

		err := repo.Create(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role ID cannot be empty")
	})

	t.Run("should return error for invitation with empty token", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-no-token"),
			AccountID:   "account-abc",
			Email:       "test@example.com",
			RoleID:      "member",
		}

		err := repo.Create(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("should return error for duplicate token", func(t *testing.T) {
		// Create first invitation
		invitation1 := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-duplicate-1"),
			AccountID:   "account-duplicate-token-1",
			Email:       "invite1@example.com",
			RoleID:      "member",
			Token:       "duplicate-token",
			Status:      domain.InvitationStatusPending,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}

		err := repo.Create(ctx, invitation1)
		require.NoError(t, err)

		// Try to create second invitation with same token
		invitation2 := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-duplicate-2"),
			AccountID:   "account-duplicate-token-2",
			Email:       "invite2@example.com",
			RoleID:      "member",
			Token:       "duplicate-token",
			Status:      domain.InvitationStatusPending,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}

		err = repo.Create(ctx, invitation2)
		assert.Error(t, err)
	})
}

func TestGormInvitationWriteRepository_Update(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationWriteRepository(db)
	ctx := context.Background()

	t.Run("should update invitation successfully", func(t *testing.T) {
		// Create initial invitation
		expiresAt := time.Now().Add(24 * time.Hour)
		createTestInvitation(t, db, "invitation-update-1", "account-update-1", "update@example.com", "member", "token-update-1", "owner-update-1", string(domain.InvitationStatusPending), expiresAt)

		// Update invitation
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("invitation-update-1"),
			AccountID:   "account-update-1",
			Email:       "update@example.com",
			RoleID:      "admin", // Changed role
			Token:       "token-update-1",
			InvitedBy:   "owner-update-1",
			Status:      domain.InvitationStatusAccepted, // Changed status
			ExpiresAt:   expiresAt,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now(),
		}

		err := repo.Update(ctx, invitation)
		require.NoError(t, err)

		// Verify invitation was updated in database
		var invitationModel InvitationModel
		err = db.First(&invitationModel, "id = ?", "invitation-update-1").Error
		require.NoError(t, err)
		assert.Equal(t, "admin", invitationModel.RoleID)
		assert.Equal(t, string(domain.InvitationStatusAccepted), invitationModel.Status)
	})

	t.Run("should return error for nil invitation", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation cannot be nil")
	})

	t.Run("should return error for invitation with empty ID", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity(""),
			AccountID:   "account-update-2",
			Email:       "update2@example.com",
			RoleID:      "member",
			Token:       "token-update-2",
		}

		err := repo.Update(ctx, invitation)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation ID cannot be empty")
	})

	t.Run("should return error for non-existent invitation", func(t *testing.T) {
		invitation := &domain.Invitation{
			BasicEntity: pericarpdomain.NewEntity("non-existent-invitation"),
			AccountID:   "account-non-existent",
			Email:       "nonexistent@example.com",
			RoleID:      "member",
			Token:       "token-non-existent",
		}

		err := repo.Update(ctx, invitation)
		assert.Error(t, err)
	})
}

func TestGormInvitationWriteRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationWriteRepository(db)
	ctx := context.Background()

	t.Run("should delete invitation successfully", func(t *testing.T) {
		// Create invitation to delete
		expiresAt := time.Now().Add(24 * time.Hour)
		createTestInvitation(t, db, "invitation-delete-1", "account-delete-1", "delete@example.com", "member", "token-delete-1", "owner-delete-1", string(domain.InvitationStatusPending), expiresAt)

		err := repo.Delete(ctx, "invitation-delete-1")
		require.NoError(t, err)

		// Verify invitation was deleted from database
		var count int64
		err = db.Model(&InvitationModel{}).Where("id = ?", "invitation-delete-1").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation ID cannot be empty")
	})

	t.Run("should return error for non-existent invitation", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-invitation")
		assert.Error(t, err)
	})
}
