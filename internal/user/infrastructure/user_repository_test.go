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

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *gorm.DB, id, webID, email, name, status string) {
	user := UserModel{
		ID:        id,
		WebID:     webID,
		Email:     email,
		Name:      name,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&user).Error
	require.NoError(t, err)
}

func TestGormUserRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	// Create test user
	userID := "test-user-1"
	webID := "https://example.com/users/test-user-1"
	email := "test1@example.com"
	name := "Test User 1"
	status := string(domain.UserStatusActive)

	createTestUser(t, db, userID, webID, email, name, status)

	t.Run("should return user when found", func(t *testing.T) {
		user, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID())
		assert.Equal(t, webID, user.WebID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, name, user.Profile.Name)
		assert.Equal(t, domain.UserStatus(status), user.Status)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "non-existent-user")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGormUserRepository_GetByWebID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	// Create test user
	userID := "test-user-2"
	webID := "https://example.com/users/test-user-2"
	email := "test2@example.com"
	name := "Test User 2"
	status := string(domain.UserStatusActive)

	createTestUser(t, db, userID, webID, email, name, status)

	t.Run("should return user when found by WebID", func(t *testing.T) {
		user, err := repo.GetByWebID(ctx, webID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID())
		assert.Equal(t, webID, user.WebID)
		assert.Equal(t, email, user.Email)
	})

	t.Run("should return error when WebID not found", func(t *testing.T) {
		user, err := repo.GetByWebID(ctx, "https://example.com/users/non-existent")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("should return error for empty WebID", func(t *testing.T) {
		user, err := repo.GetByWebID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGormUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	// Create test user
	userID := "test-user-3"
	webID := "https://example.com/users/test-user-3"
	email := "test3@example.com"
	name := "Test User 3"
	status := string(domain.UserStatusActive)

	createTestUser(t, db, userID, webID, email, name, status)

	t.Run("should return user when found by email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID())
		assert.Equal(t, webID, user.WebID)
		assert.Equal(t, email, user.Email)
	})

	t.Run("should return error when email not found", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("should return error for empty email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestGormUserRepository_List(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	// Create multiple test users
	users := []struct {
		id, webID, email, name, status string
	}{
		{"user-1", "https://example.com/users/user-1", "user1@example.com", "User 1", string(domain.UserStatusActive)},
		{"user-2", "https://example.com/users/user-2", "user2@example.com", "User 2", string(domain.UserStatusActive)},
		{"user-3", "https://example.com/users/user-3", "user3@example.com", "User 3", string(domain.UserStatusSuspended)},
		{"user-4", "https://example.com/users/user-4", "user4@example.com", "User 4", string(domain.UserStatusDeleted)},
	}

	for _, u := range users {
		createTestUser(t, db, u.id, u.webID, u.email, u.name, u.status)
	}

	t.Run("should return all users with no filter", func(t *testing.T) {
		filter := domain.UserFilter{}
		userList, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, userList, 4)
	})

	t.Run("should filter by status", func(t *testing.T) {
		filter := domain.UserFilter{Status: domain.UserStatusActive}
		userList, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, userList, 2)
		for _, user := range userList {
			assert.Equal(t, domain.UserStatusActive, user.Status)
		}
	})

	t.Run("should filter by email pattern", func(t *testing.T) {
		filter := domain.UserFilter{EmailPattern: "user1@"}
		userList, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, userList, 1)
		assert.Equal(t, "user1@example.com", userList[0].Email)
	})

	t.Run("should apply limit", func(t *testing.T) {
		filter := domain.UserFilter{Limit: 2}
		userList, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, userList, 2)
	})

	t.Run("should apply offset", func(t *testing.T) {
		filter := domain.UserFilter{Offset: 2}
		userList, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, userList, 2)
	})

	t.Run("should return error for invalid filter", func(t *testing.T) {
		filter := domain.UserFilter{Limit: -1}
		userList, err := repo.List(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, userList)
	})
}

func TestGormUserRepository_Exists(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormUserRepository(db)
	ctx := context.Background()

	// Create test user
	userID := "test-user-exists"
	webID := "https://example.com/users/test-user-exists"
	email := "exists@example.com"
	name := "Exists User"
	status := string(domain.UserStatusActive)

	createTestUser(t, db, userID, webID, email, name, status)

	t.Run("should return true when user exists", func(t *testing.T) {
		exists, err := repo.Exists(ctx, userID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false when user does not exist", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "non-existent-user")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "")
		assert.Error(t, err)
		assert.False(t, exists)
	})
}
