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

// createTestInvitation creates a test invitation in the database
func createTestInvitation(t *testing.T, db *gorm.DB, id, accountID, email, roleID, token, invitedBy, status string, expiresAt time.Time) {
	invitation := InvitationModel{
		ID:        id,
		AccountID: accountID,
		Email:     email,
		RoleID:    roleID,
		Token:     token,
		InvitedBy: invitedBy,
		Status:    status,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&invitation).Error
	require.NoError(t, err)
}

func TestGormInvitationRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-invite-1"
	accountID := "account-invite-1"
	invitationID := "invitation-1"
	email := "invite1@example.com"
	roleID := "member"
	token := "test-token-1"
	status := string(domain.InvitationStatusPending)
	expiresAt := time.Now().Add(24 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-invite-1", "owner-invite1@example.com", "Owner Invite 1", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Invite Account 1", "Account for invitation testing")
	createTestInvitation(t, db, invitationID, accountID, email, roleID, token, ownerID, status, expiresAt)

	t.Run("should return invitation when found", func(t *testing.T) {
		invitation, err := repo.GetByID(ctx, invitationID)
		require.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, invitationID, invitation.ID())
		assert.Equal(t, accountID, invitation.AccountID)
		assert.Equal(t, email, invitation.Email)
		assert.Equal(t, roleID, invitation.RoleID)
		assert.Equal(t, token, invitation.Token)
		assert.Equal(t, ownerID, invitation.InvitedBy)
		assert.Equal(t, domain.InvitationStatus(status), invitation.Status)
		assert.WithinDuration(t, expiresAt, invitation.ExpiresAt, time.Second)
	})

	t.Run("should return error when invitation not found", func(t *testing.T) {
		invitation, err := repo.GetByID(ctx, "non-existent-invitation")
		assert.Error(t, err)
		assert.Nil(t, invitation)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		invitation, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, invitation)
	})
}

func TestGormInvitationRepository_GetByToken(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-invite-2"
	accountID := "account-invite-2"
	invitationID := "invitation-2"
	email := "invite2@example.com"
	roleID := "admin"
	token := "unique-test-token-2"
	status := string(domain.InvitationStatusPending)
	expiresAt := time.Now().Add(48 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-invite-2", "owner-invite2@example.com", "Owner Invite 2", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Invite Account 2", "Account for invitation testing 2")
	createTestInvitation(t, db, invitationID, accountID, email, roleID, token, ownerID, status, expiresAt)

	t.Run("should return invitation when found by token", func(t *testing.T) {
		invitation, err := repo.GetByToken(ctx, token)
		require.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, invitationID, invitation.ID())
		assert.Equal(t, accountID, invitation.AccountID)
		assert.Equal(t, email, invitation.Email)
		assert.Equal(t, roleID, invitation.RoleID)
		assert.Equal(t, token, invitation.Token)
	})

	t.Run("should return error when token not found", func(t *testing.T) {
		invitation, err := repo.GetByToken(ctx, "non-existent-token")
		assert.Error(t, err)
		assert.Nil(t, invitation)
	})

	t.Run("should return error for empty token", func(t *testing.T) {
		invitation, err := repo.GetByToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, invitation)
	})
}

func TestGormInvitationRepository_ListByAccount(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-invite-3"
	accountID := "account-invite-3"
	baseTime := time.Now()

	createTestUser(t, db, ownerID, "https://example.com/users/owner-invite-3", "owner-invite3@example.com", "Owner Invite 3", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Invite Account 3", "Account for invitation testing 3")

	// Create multiple invitations
	invitations := []struct {
		id, email, roleID, token, status string
		expiresAt                        time.Time
	}{
		{"invite-3-1", "invite3-1@example.com", "member", "token-3-1", string(domain.InvitationStatusPending), baseTime.Add(24 * time.Hour)},
		{"invite-3-2", "invite3-2@example.com", "admin", "token-3-2", string(domain.InvitationStatusAccepted), baseTime.Add(48 * time.Hour)},
		{"invite-3-3", "invite3-3@example.com", "viewer", "token-3-3", string(domain.InvitationStatusExpired), baseTime.Add(-24 * time.Hour)},
	}

	for _, inv := range invitations {
		createTestInvitation(t, db, inv.id, accountID, inv.email, inv.roleID, inv.token, ownerID, inv.status, inv.expiresAt)
	}

	t.Run("should return all invitations for account", func(t *testing.T) {
		invitationList, err := repo.ListByAccount(ctx, accountID)
		require.NoError(t, err)
		assert.Len(t, invitationList, 3)

		// Verify all invitations belong to the correct account
		for _, invitation := range invitationList {
			assert.Equal(t, accountID, invitation.AccountID)
		}

		// Verify different statuses are present
		statuses := make(map[domain.InvitationStatus]bool)
		for _, invitation := range invitationList {
			statuses[invitation.Status] = true
		}
		assert.True(t, statuses[domain.InvitationStatusPending])
		assert.True(t, statuses[domain.InvitationStatusAccepted])
		assert.True(t, statuses[domain.InvitationStatusExpired])
	})

	t.Run("should return empty list for account with no invitations", func(t *testing.T) {
		// Create account without invitations
		emptyAccountID := "empty-invite-account"
		createTestAccount(t, db, emptyAccountID, ownerID, "Empty Invite Account", "Account with no invitations")

		invitationList, err := repo.ListByAccount(ctx, emptyAccountID)
		require.NoError(t, err)
		assert.Len(t, invitationList, 0)
	})

	t.Run("should return error for empty account ID", func(t *testing.T) {
		invitationList, err := repo.ListByAccount(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, invitationList)
	})

	t.Run("should return empty list for non-existent account", func(t *testing.T) {
		invitationList, err := repo.ListByAccount(ctx, "non-existent-account")
		require.NoError(t, err)
		assert.Len(t, invitationList, 0)
	})
}

func TestGormInvitationRepository_ListByEmail(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := NewGormInvitationRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-invite-4"
	email := "multi-invite@example.com"
	baseTime := time.Now()

	createTestUser(t, db, ownerID, "https://example.com/users/owner-invite-4", "owner-invite4@example.com", "Owner Invite 4", string(domain.UserStatusActive))

	// Create multiple accounts and invitations for the same email
	accounts := []struct {
		id, name, inviteID, roleID, token, status string
		expiresAt                                 time.Time
	}{
		{"account-multi-invite-1", "Multi Invite Account 1", "multi-invite-1", "member", "multi-token-1", string(domain.InvitationStatusPending), baseTime.Add(24 * time.Hour)},
		{"account-multi-invite-2", "Multi Invite Account 2", "multi-invite-2", "admin", "multi-token-2", string(domain.InvitationStatusAccepted), baseTime.Add(48 * time.Hour)},
		{"account-multi-invite-3", "Multi Invite Account 3", "multi-invite-3", "viewer", "multi-token-3", string(domain.InvitationStatusRevoked), baseTime.Add(72 * time.Hour)},
	}

	for _, acc := range accounts {
		createTestAccount(t, db, acc.id, ownerID, acc.name, "Multi invite account "+acc.name)
		createTestInvitation(t, db, acc.inviteID, acc.id, email, acc.roleID, acc.token, ownerID, acc.status, acc.expiresAt)
	}

	t.Run("should return all invitations for email", func(t *testing.T) {
		invitationList, err := repo.ListByEmail(ctx, email)
		require.NoError(t, err)
		assert.Len(t, invitationList, 3)

		// Verify all invitations belong to the correct email
		for _, invitation := range invitationList {
			assert.Equal(t, email, invitation.Email)
		}

		// Verify different accounts and statuses
		accountIDs := make(map[string]bool)
		statuses := make(map[domain.InvitationStatus]bool)
		for _, invitation := range invitationList {
			accountIDs[invitation.AccountID] = true
			statuses[invitation.Status] = true
		}
		assert.Len(t, accountIDs, 3)
		assert.True(t, statuses[domain.InvitationStatusPending])
		assert.True(t, statuses[domain.InvitationStatusAccepted])
		assert.True(t, statuses[domain.InvitationStatusRevoked])
	})

	t.Run("should return empty list for email with no invitations", func(t *testing.T) {
		invitationList, err := repo.ListByEmail(ctx, "no-invites@example.com")
		require.NoError(t, err)
		assert.Len(t, invitationList, 0)
	})

	t.Run("should return error for empty email", func(t *testing.T) {
		invitationList, err := repo.ListByEmail(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, invitationList)
	})

	t.Run("should handle case sensitivity correctly", func(t *testing.T) {
		// Test with different case
		invitationList, err := repo.ListByEmail(ctx, "MULTI-INVITE@EXAMPLE.COM")
		require.NoError(t, err)
		// Should return empty list as email matching should be case-sensitive
		assert.Len(t, invitationList, 0)
	})
}
