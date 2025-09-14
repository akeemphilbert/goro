package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGormPasswordRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordRepository(db)
	ctx := context.Background()

	t.Run("Save valid password credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID:       "user-123",
			PasswordHash: "hash123",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential)
		assert.NoError(t, err)

		// Verify it was saved
		found, err := repo.FindByUserID(ctx, credential.UserID)
		assert.NoError(t, err)
		assert.Equal(t, credential.UserID, found.UserID)
		assert.Equal(t, credential.PasswordHash, found.PasswordHash)
		assert.Equal(t, credential.Salt, found.Salt)
	})

	t.Run("Save nil credential", func(t *testing.T) {
		err := repo.Save(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password credential cannot be nil")
	})

	t.Run("Save invalid credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID: "user-123",
			// Missing PasswordHash and Salt
		}

		err := repo.Save(ctx, credential)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password credential is invalid")
	})

	t.Run("Save updates existing credential", func(t *testing.T) {
		userID := "user-update"

		// Save initial credential
		credential1 := &domain.PasswordCredential{
			UserID:       userID,
			PasswordHash: "hash1",
			Salt:         "salt1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential1)
		require.NoError(t, err)

		// Save updated credential (same user ID)
		credential2 := &domain.PasswordCredential{
			UserID:       userID,
			PasswordHash: "hash2",
			Salt:         "salt2",
			CreatedAt:    credential1.CreatedAt, // Keep original created time
			UpdatedAt:    time.Now(),
		}

		err = repo.Save(ctx, credential2)
		assert.NoError(t, err)

		// Verify it was updated, not duplicated
		found, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "hash2", found.PasswordHash)
		assert.Equal(t, "salt2", found.Salt)
	})

	t.Run("FindByUserID existing credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID:       "user-find-123",
			PasswordHash: "hash123",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential)
		require.NoError(t, err)

		found, err := repo.FindByUserID(ctx, credential.UserID)
		assert.NoError(t, err)
		assert.Equal(t, credential.UserID, found.UserID)
		assert.Equal(t, credential.PasswordHash, found.PasswordHash)
		assert.Equal(t, credential.Salt, found.Salt)
	})

	t.Run("FindByUserID non-existent credential", func(t *testing.T) {
		_, err := repo.FindByUserID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordCredentialNotFound, err)
	})

	t.Run("FindByUserID empty user ID", func(t *testing.T) {
		_, err := repo.FindByUserID(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("Update existing credential", func(t *testing.T) {
		userID := "user-update-test"

		// Save initial credential
		credential := &domain.PasswordCredential{
			UserID:       userID,
			PasswordHash: "oldHash",
			Salt:         "oldSalt",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential)
		require.NoError(t, err)

		// Update credential
		credential.PasswordHash = "newHash"
		credential.Salt = "newSalt"
		credential.UpdatedAt = time.Now()

		err = repo.Update(ctx, credential)
		assert.NoError(t, err)

		// Verify it was updated
		found, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "newHash", found.PasswordHash)
		assert.Equal(t, "newSalt", found.Salt)
	})

	t.Run("Update non-existent credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID:       "non-existent",
			PasswordHash: "hash",
			Salt:         "salt",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Update(ctx, credential)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordCredentialNotFound, err)
	})

	t.Run("Update nil credential", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password credential cannot be nil")
	})

	t.Run("Update invalid credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID: "user-123",
			// Missing PasswordHash and Salt
		}

		err := repo.Update(ctx, credential)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password credential is invalid")
	})

	t.Run("Delete existing credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID:       "user-delete-123",
			PasswordHash: "hash123",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential)
		require.NoError(t, err)

		err = repo.Delete(ctx, credential.UserID)
		assert.NoError(t, err)

		// Verify it was deleted
		_, err = repo.FindByUserID(ctx, credential.UserID)
		assert.Equal(t, domain.ErrPasswordCredentialNotFound, err)
	})

	t.Run("Delete non-existent credential", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordCredentialNotFound, err)
	})

	t.Run("Delete empty user ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("Exists with existing credential", func(t *testing.T) {
		credential := &domain.PasswordCredential{
			UserID:       "user-exists-123",
			PasswordHash: "hash123",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Save(ctx, credential)
		require.NoError(t, err)

		exists, err := repo.Exists(ctx, credential.UserID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Exists with non-existent credential", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "non-existent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Exists empty user ID", func(t *testing.T) {
		_, err := repo.Exists(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})
}
