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

func TestGormAccountMemberWriteRepository_Create(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountMemberWriteRepository(db)
	ctx := context.Background()

	t.Run("should create account member successfully", func(t *testing.T) {
		joinedAt := time.Now().Add(-1 * time.Hour)

		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-create-1"),
			AccountID:   "account-member-create-1",
			UserID:      "user-member-create-1",
			RoleID:      "member",
			InvitedBy:   "owner-member-create-1",
			JoinedAt:    joinedAt,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, member)
		require.NoError(t, err)

		// Verify member was created in database
		var memberModel AccountMemberModel
		err = db.First(&memberModel, "id = ?", "member-create-1").Error
		require.NoError(t, err)
		assert.Equal(t, "member-create-1", memberModel.ID)
		assert.Equal(t, "account-member-create-1", memberModel.AccountID)
		assert.Equal(t, "user-member-create-1", memberModel.UserID)
		assert.Equal(t, "member", memberModel.RoleID)
		assert.Equal(t, "owner-member-create-1", memberModel.InvitedBy)
		assert.WithinDuration(t, joinedAt, memberModel.JoinedAt, time.Second)
	})

	t.Run("should return error for nil member", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account member cannot be nil")
	})

	t.Run("should return error for member with empty ID", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity(""),
			AccountID:   "account-123",
			UserID:      "user-123",
			RoleID:      "member",
		}

		err := repo.Create(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member ID cannot be empty")
	})

	t.Run("should return error for member with empty account ID", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-no-account"),
			UserID:      "user-456",
			RoleID:      "member",
		}

		err := repo.Create(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("should return error for member with empty user ID", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-no-user"),
			AccountID:   "account-456",
			RoleID:      "member",
		}

		err := repo.Create(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("should return error for member with empty role ID", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-no-role"),
			AccountID:   "account-789",
			UserID:      "user-789",
		}

		err := repo.Create(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "role ID cannot be empty")
	})

	t.Run("should return error for duplicate account+user combination", func(t *testing.T) {
		// Create first member
		member1 := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-duplicate-1"),
			AccountID:   "account-duplicate",
			UserID:      "user-duplicate",
			RoleID:      "member",
			JoinedAt:    time.Now(),
		}

		err := repo.Create(ctx, member1)
		require.NoError(t, err)

		// Try to create second member with same account+user combination
		member2 := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-duplicate-2"),
			AccountID:   "account-duplicate",
			UserID:      "user-duplicate",
			RoleID:      "admin",
			JoinedAt:    time.Now(),
		}

		err = repo.Create(ctx, member2)
		assert.Error(t, err)
	})
}

func TestGormAccountMemberWriteRepository_Update(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountMemberWriteRepository(db)
	ctx := context.Background()

	t.Run("should update account member successfully", func(t *testing.T) {
		// Create initial member
		joinedAt := time.Now().Add(-2 * time.Hour)
		createTestAccountMember(t, db, "member-update-1", "account-update-1", "user-update-1", "member", "owner-update-1", joinedAt)

		// Update member
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("member-update-1"),
			AccountID:   "account-update-1",
			UserID:      "user-update-1",
			RoleID:      "admin", // Changed role
			InvitedBy:   "owner-update-1",
			JoinedAt:    joinedAt,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now(),
		}

		err := repo.Update(ctx, member)
		require.NoError(t, err)

		// Verify member was updated in database
		var memberModel AccountMemberModel
		err = db.First(&memberModel, "id = ?", "member-update-1").Error
		require.NoError(t, err)
		assert.Equal(t, "admin", memberModel.RoleID)
	})

	t.Run("should return error for nil member", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account member cannot be nil")
	})

	t.Run("should return error for member with empty ID", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity(""),
			AccountID:   "account-update-2",
			UserID:      "user-update-2",
			RoleID:      "member",
		}

		err := repo.Update(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member ID cannot be empty")
	})

	t.Run("should return error for non-existent member", func(t *testing.T) {
		member := &domain.AccountMember{
			BasicEntity: pericarpdomain.NewEntity("non-existent-member"),
			AccountID:   "account-non-existent",
			UserID:      "user-non-existent",
			RoleID:      "member",
		}

		err := repo.Update(ctx, member)
		assert.Error(t, err)
	})
}

func TestGormAccountMemberWriteRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormAccountMemberWriteRepository(db)
	ctx := context.Background()

	t.Run("should delete account member successfully", func(t *testing.T) {
		// Create member to delete
		joinedAt := time.Now().Add(-3 * time.Hour)
		createTestAccountMember(t, db, "member-delete-1", "account-delete-1", "user-delete-1", "member", "owner-delete-1", joinedAt)

		err := repo.Delete(ctx, "member-delete-1")
		require.NoError(t, err)

		// Verify member was deleted from database
		var count int64
		err = db.Model(&AccountMemberModel{}).Where("id = ?", "member-delete-1").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member ID cannot be empty")
	})

	t.Run("should return error for non-existent member", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-member")
		assert.Error(t, err)
	})
}
