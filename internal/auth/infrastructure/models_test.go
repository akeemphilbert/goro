package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = MigrateAuthTables(db)
	require.NoError(t, err)

	return db
}

func TestSessionModel(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid session model", func(t *testing.T) {
		session := &SessionModel{
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

		assert.True(t, session.IsValid())
		assert.False(t, session.IsExpired())
	})

	t.Run("Invalid session model - missing fields", func(t *testing.T) {
		session := &SessionModel{
			ID:     "session-123",
			UserID: "user-456",
			// Missing WebID and TokenHash
		}

		assert.False(t, session.IsValid())
	})

	t.Run("Expired session", func(t *testing.T) {
		session := &SessionModel{
			ID:           "session-123",
			UserID:       "user-456",
			WebID:        "https://example.com/user/456",
			AccountID:    "account-789",
			RoleID:       "role-member",
			TokenHash:    "hash123",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		assert.True(t, session.IsValid())   // Valid fields
		assert.True(t, session.IsExpired()) // But expired
	})

	t.Run("Update activity", func(t *testing.T) {
		session := &SessionModel{
			ID:           "session-123",
			UserID:       "user-456",
			WebID:        "https://example.com/user/456",
			AccountID:    "account-789",
			RoleID:       "role-viewer",
			TokenHash:    "hash123",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			LastActivity: time.Now().Add(-1 * time.Hour),
		}

		oldActivity := session.LastActivity
		time.Sleep(1 * time.Millisecond) // Ensure time difference

		err := session.UpdateActivity(ctx)
		assert.NoError(t, err)
		assert.True(t, session.LastActivity.After(oldActivity))
	})

	t.Run("TableName", func(t *testing.T) {
		session := &SessionModel{}
		assert.Equal(t, "session_models", session.TableName())
	})
}

func TestPasswordCredentialModel(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid password credential model", func(t *testing.T) {
		pc := &PasswordCredentialModel{
			UserID:       "user-123",
			PasswordHash: "hash123",
			Salt:         "salt123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.True(t, pc.IsValid())
	})

	t.Run("Invalid password credential model - missing fields", func(t *testing.T) {
		pc := &PasswordCredentialModel{
			UserID: "user-123",
			// Missing PasswordHash and Salt
		}

		assert.False(t, pc.IsValid())
	})

	t.Run("Update password", func(t *testing.T) {
		pc := &PasswordCredentialModel{
			UserID:       "user-123",
			PasswordHash: "oldHash",
			Salt:         "oldSalt",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now().Add(-1 * time.Hour),
		}

		oldUpdatedAt := pc.UpdatedAt
		time.Sleep(1 * time.Millisecond) // Ensure time difference

		err := pc.UpdatePassword(ctx, "newHash", "newSalt")
		assert.NoError(t, err)
		assert.Equal(t, "newHash", pc.PasswordHash)
		assert.Equal(t, "newSalt", pc.Salt)
		assert.True(t, pc.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("Update password with empty values", func(t *testing.T) {
		pc := &PasswordCredentialModel{
			UserID:       "user-123",
			PasswordHash: "oldHash",
			Salt:         "oldSalt",
		}

		err := pc.UpdatePassword(ctx, "", "newSalt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password hash and salt cannot be empty")

		err = pc.UpdatePassword(ctx, "newHash", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password hash and salt cannot be empty")
	})

	t.Run("TableName", func(t *testing.T) {
		pc := &PasswordCredentialModel{}
		assert.Equal(t, "password_credential_models", pc.TableName())
	})
}

func TestPasswordResetTokenModel(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid password reset token model", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		assert.True(t, prt.IsValid())
		assert.False(t, prt.IsExpired())
	})

	t.Run("Invalid password reset token model - missing fields", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:  "token123",
			UserID: "user-456",
			// Missing Email
		}

		assert.False(t, prt.IsValid())
	})

	t.Run("Expired token", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			CreatedAt: time.Now(),
			Used:      false,
		}

		assert.False(t, prt.IsValid()) // Invalid because expired
		assert.True(t, prt.IsExpired())
	})

	t.Run("Used token", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      true, // Already used
		}

		assert.False(t, prt.IsValid()) // Invalid because used
		assert.False(t, prt.IsExpired())
	})

	t.Run("Mark as used", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      false,
		}

		err := prt.MarkAsUsed(ctx)
		assert.NoError(t, err)
		assert.True(t, prt.Used)
	})

	t.Run("Mark already used token as used", func(t *testing.T) {
		prt := &PasswordResetTokenModel{
			Token:     "token123",
			UserID:    "user-456",
			Email:     "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			CreatedAt: time.Now(),
			Used:      true,
		}

		err := prt.MarkAsUsed(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is already used")
	})

	t.Run("TableName", func(t *testing.T) {
		prt := &PasswordResetTokenModel{}
		assert.Equal(t, "password_reset_token_models", prt.TableName())
	})
}

func TestExternalIdentityModel(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("Valid external identity model", func(t *testing.T) {
		eim := &ExternalIdentityModel{
			UserID:     "user-123",
			Provider:   "google",
			ExternalID: "google-456",
			CreatedAt:  time.Now(),
		}

		assert.True(t, eim.IsValid())
	})

	t.Run("Invalid external identity model - missing fields", func(t *testing.T) {
		eim := &ExternalIdentityModel{
			UserID:   "user-123",
			Provider: "google",
			// Missing ExternalID
		}

		assert.False(t, eim.IsValid())
	})

	t.Run("Update provider", func(t *testing.T) {
		eim := &ExternalIdentityModel{
			ID:         1,
			UserID:     "user-123",
			Provider:   "google",
			ExternalID: "google-456",
			CreatedAt:  time.Now(),
		}

		err := eim.UpdateProvider(ctx, "github")
		assert.NoError(t, err)
		assert.Equal(t, "github", eim.Provider)
	})

	t.Run("Update provider with empty value", func(t *testing.T) {
		eim := &ExternalIdentityModel{
			ID:         1,
			UserID:     "user-123",
			Provider:   "google",
			ExternalID: "google-456",
		}

		err := eim.UpdateProvider(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be empty")
	})

	t.Run("Unique constraint validation", func(t *testing.T) {
		// Create first identity
		eim1 := &ExternalIdentityModel{
			UserID:     "user-123",
			Provider:   "google",
			ExternalID: "google-456",
			CreatedAt:  time.Now(),
		}

		err := db.Create(eim1).Error
		assert.NoError(t, err)

		// Try to create duplicate identity with same provider + external_id
		eim2 := &ExternalIdentityModel{
			UserID:     "user-789",   // Different user
			Provider:   "google",     // Same provider
			ExternalID: "google-456", // Same external ID
			CreatedAt:  time.Now(),
		}

		err = db.Create(eim2).Error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "external identity already exists")
	})

	t.Run("Different providers allow same external ID", func(t *testing.T) {
		// Create identity with GitHub
		eim1 := &ExternalIdentityModel{
			UserID:     "user-123",
			Provider:   "github",
			ExternalID: "same-id-456",
			CreatedAt:  time.Now(),
		}

		err := db.Create(eim1).Error
		assert.NoError(t, err)

		// Create identity with Google using same external ID (should be allowed)
		eim2 := &ExternalIdentityModel{
			UserID:     "user-789",
			Provider:   "google",      // Different provider
			ExternalID: "same-id-456", // Same external ID
			CreatedAt:  time.Now(),
		}

		err = db.Create(eim2).Error
		assert.NoError(t, err)
	})

	t.Run("TableName", func(t *testing.T) {
		eim := &ExternalIdentityModel{}
		assert.Equal(t, "external_identity_models", eim.TableName())
	})
}

func TestMigrateAuthTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = MigrateAuthTables(db)
	assert.NoError(t, err)

	// Verify tables were created
	assert.True(t, db.Migrator().HasTable(&SessionModel{}))
	assert.True(t, db.Migrator().HasTable(&PasswordCredentialModel{}))
	assert.True(t, db.Migrator().HasTable(&PasswordResetTokenModel{}))
	assert.True(t, db.Migrator().HasTable(&ExternalIdentityModel{}))
}

func TestDropAuthTables(t *testing.T) {
	db := setupTestDB(t)

	// Verify tables exist
	assert.True(t, db.Migrator().HasTable(&SessionModel{}))
	assert.True(t, db.Migrator().HasTable(&PasswordCredentialModel{}))
	assert.True(t, db.Migrator().HasTable(&PasswordResetTokenModel{}))
	assert.True(t, db.Migrator().HasTable(&ExternalIdentityModel{}))

	err := DropAuthTables(db)
	assert.NoError(t, err)

	// Verify tables were dropped
	assert.False(t, db.Migrator().HasTable(&SessionModel{}))
	assert.False(t, db.Migrator().HasTable(&PasswordCredentialModel{}))
	assert.False(t, db.Migrator().HasTable(&PasswordResetTokenModel{}))
	assert.False(t, db.Migrator().HasTable(&ExternalIdentityModel{}))
}
