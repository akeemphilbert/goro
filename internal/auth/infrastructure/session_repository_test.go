package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGormSessionRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormSessionRepository(db)
	ctx := context.Background()

	t.Run("Save valid session", func(t *testing.T) {
		session := &domain.Session{
			ID:           "session-123",
			UserID:       "user-456",
			WebID:        "https://example.com/user/456",
			AccountID:    "account-789",
			RoleID:       "role-admin",
			TokenHash:    "hash123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, session)
		assert.NoError(t, err)

		// Verify it was saved
		found, err := repo.FindByID(ctx, session.ID)
		assert.NoError(t, err)
		assert.Equal(t, session.ID, found.ID)
		assert.Equal(t, session.UserID, found.UserID)
		assert.Equal(t, session.WebID, found.WebID)
		assert.Equal(t, session.TokenHash, found.TokenHash)
	})

	t.Run("Save nil session", func(t *testing.T) {
		err := repo.Save(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session cannot be nil")
	})

	t.Run("Save invalid session", func(t *testing.T) {
		session := &domain.Session{
			ID:     "session-123",
			UserID: "user-456",
			// Missing required fields
		}

		err := repo.Save(ctx, session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session is invalid")
	})

	t.Run("FindByID existing session", func(t *testing.T) {
		session := &domain.Session{
			ID:           "session-find-123",
			UserID:       "user-456",
			WebID:        "https://example.com/user/456",
			AccountID:    "account-789",
			RoleID:       "role-member",
			TokenHash:    "hash123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, session)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, session.ID)
		assert.NoError(t, err)
		assert.Equal(t, session.ID, found.ID)
		assert.Equal(t, session.UserID, found.UserID)
	})

	t.Run("FindByID non-existent session", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("FindByID empty ID", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")
	})

	t.Run("FindByUserID", func(t *testing.T) {
		userID := "user-multi-sessions"

		// Create multiple sessions for the same user
		sessions := []*domain.Session{
			{
				ID:           "session-1",
				UserID:       userID,
				WebID:        "https://example.com/user/multi",
				AccountID:    "account-multi-1",
				RoleID:       "role-owner",
				TokenHash:    "hash1",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
			{
				ID:           "session-2",
				UserID:       userID,
				WebID:        "https://example.com/user/multi",
				AccountID:    "account-multi-2",
				RoleID:       "role-member",
				TokenHash:    "hash2",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
		}

		for _, session := range sessions {
			err := repo.Save(ctx, session)
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

	t.Run("Delete existing session", func(t *testing.T) {
		session := &domain.Session{
			ID:           "session-delete-123",
			UserID:       "user-456",
			WebID:        "https://example.com/user/456",
			AccountID:    "account-delete",
			RoleID:       "role-admin",
			TokenHash:    "hash123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, session)
		require.NoError(t, err)

		err = repo.Delete(ctx, session.ID)
		assert.NoError(t, err)

		// Verify it was deleted
		_, err = repo.FindByID(ctx, session.ID)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("Delete non-existent session", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("Delete empty session ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")
	})

	t.Run("DeleteByUserID", func(t *testing.T) {
		userID := "user-delete-all"

		// Create multiple sessions for the user
		sessions := []*domain.Session{
			{
				ID:           "session-del-1",
				UserID:       userID,
				WebID:        "https://example.com/user/del",
				AccountID:    "account-del-1",
				RoleID:       "role-member",
				TokenHash:    "hash1",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
			{
				ID:           "session-del-2",
				UserID:       userID,
				WebID:        "https://example.com/user/del",
				AccountID:    "account-del-2",
				RoleID:       "role-admin",
				TokenHash:    "hash2",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
		}

		for _, session := range sessions {
			err := repo.Save(ctx, session)
			require.NoError(t, err)
		}

		err := repo.DeleteByUserID(ctx, userID)
		assert.NoError(t, err)

		// Verify all sessions were deleted
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
		// Create session that will be expired (save it as valid first, then manually update expiry)
		expiredSession := &domain.Session{
			ID:           "session-expired",
			UserID:       "user-expired",
			WebID:        "https://example.com/user/expired",
			AccountID:    "account-expired",
			RoleID:       "role-member",
			TokenHash:    "hash-expired",
			ExpiresAt:    time.Now().Add(24 * time.Hour), // Valid initially
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			LastActivity: time.Now().Add(-1 * time.Hour),
		}

		// Create valid session
		validSession := &domain.Session{
			ID:           "session-valid",
			UserID:       "user-valid",
			WebID:        "https://example.com/user/valid",
			AccountID:    "account-valid",
			RoleID:       "role-owner",
			TokenHash:    "hash-valid",
			ExpiresAt:    time.Now().Add(24 * time.Hour), // Valid
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, expiredSession)
		require.NoError(t, err)
		err = repo.Save(ctx, validSession)
		require.NoError(t, err)

		// Manually update the expired session to be expired in the database
		db.Model(&SessionModel{}).Where("id = ?", expiredSession.ID).Update("expires_at", time.Now().Add(-1*time.Hour))

		err = repo.DeleteExpired(ctx)
		assert.NoError(t, err)

		// Verify expired session was deleted
		_, err = repo.FindByID(ctx, expiredSession.ID)
		assert.Equal(t, domain.ErrSessionNotFound, err)

		// Verify valid session still exists
		_, err = repo.FindByID(ctx, validSession.ID)
		assert.NoError(t, err)
	})

	t.Run("UpdateActivity", func(t *testing.T) {
		session := &domain.Session{
			ID:           "session-activity",
			UserID:       "user-activity",
			WebID:        "https://example.com/user/activity",
			AccountID:    "account-activity",
			RoleID:       "role-viewer",
			TokenHash:    "hash-activity",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now().Add(-1 * time.Hour),
		}

		err := repo.Save(ctx, session)
		require.NoError(t, err)

		oldActivity := session.LastActivity
		time.Sleep(1 * time.Millisecond) // Ensure time difference

		err = repo.UpdateActivity(ctx, session.ID)
		assert.NoError(t, err)

		// Verify activity was updated
		updated, err := repo.FindByID(ctx, session.ID)
		assert.NoError(t, err)
		assert.True(t, updated.LastActivity.After(oldActivity))
	})

	t.Run("UpdateActivity non-existent session", func(t *testing.T) {
		err := repo.UpdateActivity(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("UpdateActivity empty session ID", func(t *testing.T) {
		err := repo.UpdateActivity(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session ID cannot be empty")
	})

	t.Run("FindByAccountID", func(t *testing.T) {
		accountID := "account-find-test"

		// Create sessions for the same account with different users
		sessions := []*domain.Session{
			{
				ID:           "session-acc-1",
				UserID:       "user-1",
				WebID:        "https://example.com/user/1",
				AccountID:    accountID,
				RoleID:       "role-admin",
				TokenHash:    "hash1",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
			{
				ID:           "session-acc-2",
				UserID:       "user-2",
				WebID:        "https://example.com/user/2",
				AccountID:    accountID,
				RoleID:       "role-member",
				TokenHash:    "hash2",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				CreatedAt:    time.Now(),
				LastActivity: time.Now(),
			},
		}

		for _, session := range sessions {
			err := repo.Save(ctx, session)
			require.NoError(t, err)
		}

		found, err := repo.FindByAccountID(ctx, accountID)
		assert.NoError(t, err)
		assert.Len(t, found, 2)

		// Verify all sessions belong to the correct account
		for _, session := range found {
			assert.Equal(t, accountID, session.AccountID)
		}
	})

	t.Run("FindByAccountID empty account ID", func(t *testing.T) {
		_, err := repo.FindByAccountID(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("FindByUserIDAndAccountID", func(t *testing.T) {
		userID := "user-specific"
		accountID := "account-specific"

		// Create session for specific user and account
		session := &domain.Session{
			ID:           "session-user-acc",
			UserID:       userID,
			WebID:        "https://example.com/user/specific",
			AccountID:    accountID,
			RoleID:       "role-owner",
			TokenHash:    "hash-specific",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, session)
		require.NoError(t, err)

		found, err := repo.FindByUserIDAndAccountID(ctx, userID, accountID)
		assert.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, session.ID, found[0].ID)
		assert.Equal(t, userID, found[0].UserID)
		assert.Equal(t, accountID, found[0].AccountID)
	})

	t.Run("FindByUserIDAndAccountID empty parameters", func(t *testing.T) {
		_, err := repo.FindByUserIDAndAccountID(ctx, "", "account-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID cannot be empty")

		_, err = repo.FindByUserIDAndAccountID(ctx, "user-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "account ID cannot be empty")
	})

	t.Run("Session with optional account context", func(t *testing.T) {
		// Test session without account context (global session)
		globalSession := &domain.Session{
			ID:     "session-global",
			UserID: "user-global",
			WebID:  "https://example.com/user/global",
			// No AccountID or RoleID
			TokenHash:    "hash-global",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		err := repo.Save(ctx, globalSession)
		assert.NoError(t, err)

		// Verify it was saved correctly
		found, err := repo.FindByID(ctx, globalSession.ID)
		assert.NoError(t, err)
		assert.Equal(t, "", found.AccountID)
		assert.Equal(t, "", found.RoleID)
		assert.False(t, found.HasAccountContext())
	})
}
