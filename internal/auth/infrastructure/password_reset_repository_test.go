package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGormPasswordResetRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormPasswordResetRepository(db)
	ctx := context.Background()

	t.Run("Save valid password reset token", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := repo.Save(ctx, token)
		assert.NoError(t, err)

		// Verify it was saved
		found, err := repo.FindByToken(ctx, token.Token)
		assert.NoError(t, err)
		assert.Equal(t, token.Token, found.Token)
		assert.Equal(t, token.UserID, found.UserID)
		assert.Equal(t, token.Email, found.Email)
		assert.False(t, found.Used)
	})

	t.Run("Save nil token", func(t *testing.T) {
		err := repo.Save(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password reset token cannot be nil")
	})

	t.Run("Save invalid token - missing fields", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:  "token123",
			UserID: "user-456",
			// Missing Email
		}

		err := repo.Save(ctx, token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password reset token is invalid")
	})

	t.Run("FindByToken existing token", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:     "token-find-123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := repo.Save(ctx, token)
		require.NoError(t, err)

		found, err := repo.FindByToken(ctx, token.Token)
		assert.NoError(t, err)
		assert.Equal(t, token.Token, found.Token)
		assert.Equal(t, token.UserID, found.UserID)
		assert.Equal(t, token.Email, found.Email)
	})

	t.Run("FindByToken non-existent token", func(t *testing.T) {
		_, err := repo.FindByToken(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)
	})

	t.Run("FindByToken empty token", func(t *testing.T) {
		_, err := repo.FindByToken(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("FindByUserID", func(t *testing.T) {
		userID := "user-multi-tokens"

		// Create multiple tokens for the same user
		tokens := []*domain.PasswordResetToken{
			{
				Token:     "token-1",
				UserID:    userID,
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
				Used:      false,
			},
			{
				Token:     "token-2",
				UserID:    userID,
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
				Used:      false,
			},
		}

		for _, token := range tokens {
			err := repo.Save(ctx, token)
			require.NoError(t, err)
		}

		found, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, found, 2)
	})

	t.Run("FindByUserID empty user ID", func(t *testing.T) {
		_, err := repo.FindByUserID(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("MarkAsUsed existing token", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:     "token-mark-used",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := repo.Save(ctx, token)
		require.NoError(t, err)

		err = repo.MarkAsUsed(ctx, token.Token)
		assert.NoError(t, err)

		// Verify it was marked as used
		found, err := repo.FindByToken(ctx, token.Token)
		assert.NoError(t, err)
		assert.True(t, found.Used)
	})

	t.Run("MarkAsUsed non-existent token", func(t *testing.T) {
		err := repo.MarkAsUsed(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)
	})

	t.Run("MarkAsUsed already used token", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:     "token-already-used",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      true, // Already used
		}

		err := repo.Save(ctx, token)
		require.NoError(t, err)

		err = repo.MarkAsUsed(ctx, token.Token)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)
	})

	t.Run("MarkAsUsed empty token", func(t *testing.T) {
		err := repo.MarkAsUsed(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("Delete existing token", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			Token:     "token-delete-123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := repo.Save(ctx, token)
		require.NoError(t, err)

		err = repo.Delete(ctx, token.Token)
		assert.NoError(t, err)

		// Verify it was deleted
		_, err = repo.FindByToken(ctx, token.Token)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)
	})

	t.Run("Delete non-existent token", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)
	})

	t.Run("Delete empty token", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("DeleteByUserID", func(t *testing.T) {
		userID := "user-delete-all-tokens"

		// Create multiple tokens for the user
		tokens := []*domain.PasswordResetToken{
			{
				Token:     "token-del-1",
				UserID:    userID,
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
				Used:      false,
			},
			{
				Token:     "token-del-2",
				UserID:    userID,
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
				Used:      false,
			},
		}

		for _, token := range tokens {
			err := repo.Save(ctx, token)
			require.NoError(t, err)
		}

		err := repo.DeleteByUserID(ctx, userID)
		assert.NoError(t, err)

		// Verify all tokens were deleted
		found, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, found, 0)
	})

	t.Run("DeleteByUserID empty user ID", func(t *testing.T) {
		err := repo.DeleteByUserID(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")
	})

	t.Run("DeleteExpired", func(t *testing.T) {
		// Create expired token
		expiredToken := &domain.PasswordResetToken{
			Token:     "token-expired",
			UserID:    "user-expired",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt: time.Now().Add(-2 * time.Hour),
			Used:      false,
		}

		// Create valid token
		validToken := &domain.PasswordResetToken{
			Token:     "token-valid",
			UserID:    "user-valid",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := repo.Save(ctx, expiredToken)
		require.NoError(t, err)
		err = repo.Save(ctx, validToken)
		require.NoError(t, err)

		err = repo.DeleteExpired(ctx)
		assert.NoError(t, err)

		// Verify expired token was deleted
		_, err = repo.FindByToken(ctx, expiredToken.Token)
		assert.Equal(t, domain.ErrPasswordResetTokenNotFound, err)

		// Verify valid token still exists
		_, err = repo.FindByToken(ctx, validToken.Token)
		assert.NoError(t, err)
	})
}
