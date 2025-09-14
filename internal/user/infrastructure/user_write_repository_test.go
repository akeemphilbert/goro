package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

func TestGormUserWriteRepository_Create(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormUserWriteRepository(db)
	ctx := context.Background()

	t.Run("should create user successfully", func(t *testing.T) {
		profile := domain.UserProfile{
			Name:        "John Doe",
			Bio:         "Software developer",
			Avatar:      "https://example.com/avatar.jpg",
			Preferences: map[string]interface{}{"theme": "dark"},
		}

		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-123"),
			WebID:       "https://example.com/users/john-doe",
			Email:       "john.doe@example.com",
			Profile:     profile,
			Status:      domain.UserStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify user was created in database
		var userModel infrastructure.UserModel
		err = db.First(&userModel, "id = ?", "user-123").Error
		require.NoError(t, err)
		assert.Equal(t, "user-123", userModel.ID)
		assert.Equal(t, "https://example.com/users/john-doe", userModel.WebID)
		assert.Equal(t, "john.doe@example.com", userModel.Email)
		assert.Equal(t, "John Doe", userModel.Name)
		assert.Equal(t, string(domain.UserStatusActive), userModel.Status)
	})

	t.Run("should return error for nil user", func(t *testing.T) {
		err := repo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user cannot be nil")
	})

	t.Run("should return error for user with empty ID", func(t *testing.T) {
		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity(""),
			WebID:       "https://example.com/users/no-id",
			Email:       "noid@example.com",
			Status:      domain.UserStatusActive,
		}

		err := repo.Create(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("should return error for duplicate WebID", func(t *testing.T) {
		// Create first user
		user1 := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-1"),
			WebID:       "https://example.com/users/duplicate",
			Email:       "user1@example.com",
			Status:      domain.UserStatusActive,
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create second user with same WebID
		user2 := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-2"),
			WebID:       "https://example.com/users/duplicate",
			Email:       "user2@example.com",
			Status:      domain.UserStatusActive,
		}

		err = repo.Create(ctx, user2)
		assert.Error(t, err)
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		// Create first user
		user1 := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-dup-1"),
			WebID:       "https://example.com/users/user1",
			Email:       "duplicate@example.com",
			Status:      domain.UserStatusActive,
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create second user with same email
		user2 := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-dup-2"),
			WebID:       "https://example.com/users/user2",
			Email:       "duplicate@example.com",
			Status:      domain.UserStatusActive,
		}

		err = repo.Create(ctx, user2)
		assert.Error(t, err)
	})
}

func TestGormUserWriteRepository_Update(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormUserWriteRepository(db)
	ctx := context.Background()

	t.Run("should update user successfully", func(t *testing.T) {
		// Create initial user
		createTestUser(t, db, "user-update-1", "https://example.com/users/update-1", "update1@example.com", "Original Name", string(domain.UserStatusActive))

		// Update user
		profile := domain.UserProfile{
			Name:        "Updated Name",
			Bio:         "Updated bio",
			Avatar:      "https://example.com/new-avatar.jpg",
			Preferences: map[string]interface{}{"theme": "light"},
		}

		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("user-update-1"),
			WebID:       "https://example.com/users/update-1",
			Email:       "update1@example.com",
			Profile:     profile,
			Status:      domain.UserStatusSuspended,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now(),
		}

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		// Verify user was updated in database
		var userModel infrastructure.UserModel
		err = db.First(&userModel, "id = ?", "user-update-1").Error
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", userModel.Name)
		assert.Equal(t, string(domain.UserStatusSuspended), userModel.Status)
	})

	t.Run("should return error for nil user", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user cannot be nil")
	})

	t.Run("should return error for user with empty ID", func(t *testing.T) {
		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity(""),
			WebID:       "https://example.com/users/no-id-update",
			Email:       "noidupdate@example.com",
			Status:      domain.UserStatusActive,
		}

		err := repo.Update(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		user := &domain.User{
			BasicEntity: pericarpdomain.NewEntity("non-existent-user"),
			WebID:       "https://example.com/users/non-existent",
			Email:       "nonexistent@example.com",
			Status:      domain.UserStatusActive,
		}

		err := repo.Update(ctx, user)
		assert.Error(t, err)
	})
}

func TestGormUserWriteRepository_Delete(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormUserWriteRepository(db)
	ctx := context.Background()

	t.Run("should delete user successfully", func(t *testing.T) {
		// Create user to delete
		createTestUser(t, db, "user-delete-1", "https://example.com/users/delete-1", "delete1@example.com", "Delete User", string(domain.UserStatusActive))

		err := repo.Delete(ctx, "user-delete-1")
		require.NoError(t, err)

		// Verify user was deleted from database
		var count int64
		err = db.Model(&infrastructure.UserModel{}).Where("id = ?", "user-delete-1").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-user")
		assert.Error(t, err)
	})
}
