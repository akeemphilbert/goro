package infrastructure_test

import (
	"context"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// createTestAccountMember creates a test account member in the database
func createTestAccountMember(t *testing.T, db *gorm.DB, id, accountID, userID, roleID, invitedBy string, joinedAt time.Time) {
	member := infrastructure.AccountMemberModel{
		ID:        id,
		AccountID: accountID,
		UserID:    userID,
		RoleID:    roleID,
		InvitedBy: invitedBy,
		JoinedAt:  joinedAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := db.Create(&member).Error
	require.NoError(t, err)
}

func TestGormAccountMemberRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormAccountMemberRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-member-1"
	userID := "member-user-1"
	accountID := "account-member-1"
	memberID := "member-1"
	roleID := "member"
	joinedAt := time.Now().Add(-24 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-member-1", "owner-member1@example.com", "Owner Member 1", string(domain.UserStatusActive))
	createTestUser(t, db, userID, "https://example.com/users/member-1", "member1@example.com", "Member 1", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Member Account 1", "Account for member testing")
	createTestAccountMember(t, db, memberID, accountID, userID, roleID, ownerID, joinedAt)

	t.Run("should return account member when found", func(t *testing.T) {
		member, err := repo.GetByID(ctx, memberID)
		require.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, memberID, member.ID())
		assert.Equal(t, accountID, member.AccountID)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, roleID, member.RoleID)
		assert.Equal(t, ownerID, member.InvitedBy)
		assert.WithinDuration(t, joinedAt, member.JoinedAt, time.Second)
	})

	t.Run("should return error when member not found", func(t *testing.T) {
		member, err := repo.GetByID(ctx, "non-existent-member")
		assert.Error(t, err)
		assert.Nil(t, member)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		member, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, member)
	})
}

func TestGormAccountMemberRepository_GetByAccountAndUser(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormAccountMemberRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-member-2"
	userID := "member-user-2"
	accountID := "account-member-2"
	memberID := "member-2"
	roleID := "admin"
	joinedAt := time.Now().Add(-48 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-member-2", "owner-member2@example.com", "Owner Member 2", string(domain.UserStatusActive))
	createTestUser(t, db, userID, "https://example.com/users/member-2", "member2@example.com", "Member 2", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Member Account 2", "Account for member testing 2")
	createTestAccountMember(t, db, memberID, accountID, userID, roleID, ownerID, joinedAt)

	t.Run("should return member when found by account and user", func(t *testing.T) {
		member, err := repo.GetByAccountAndUser(ctx, accountID, userID)
		require.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, memberID, member.ID())
		assert.Equal(t, accountID, member.AccountID)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, roleID, member.RoleID)
	})

	t.Run("should return error when member not found", func(t *testing.T) {
		member, err := repo.GetByAccountAndUser(ctx, accountID, "non-existent-user")
		assert.Error(t, err)
		assert.Nil(t, member)
	})

	t.Run("should return error for empty account ID", func(t *testing.T) {
		member, err := repo.GetByAccountAndUser(ctx, "", userID)
		assert.Error(t, err)
		assert.Nil(t, member)
	})

	t.Run("should return error for empty user ID", func(t *testing.T) {
		member, err := repo.GetByAccountAndUser(ctx, accountID, "")
		assert.Error(t, err)
		assert.Nil(t, member)
	})
}

func TestGormAccountMemberRepository_ListByAccount(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormAccountMemberRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-member-3"
	accountID := "account-member-3"
	joinedAt := time.Now().Add(-72 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-member-3", "owner-member3@example.com", "Owner Member 3", string(domain.UserStatusActive))
	createTestAccount(t, db, accountID, ownerID, "Member Account 3", "Account for member testing 3")

	// Create multiple members
	members := []struct {
		id, userID, roleID string
	}{
		{"member-3-1", "user-3-1", "member"},
		{"member-3-2", "user-3-2", "admin"},
		{"member-3-3", "user-3-3", "viewer"},
	}

	for _, m := range members {
		createTestUser(t, db, m.userID, "https://example.com/users/"+m.userID, m.userID+"@example.com", "User "+m.userID, string(domain.UserStatusActive))
		createTestAccountMember(t, db, m.id, accountID, m.userID, m.roleID, ownerID, joinedAt)
	}

	t.Run("should return all members for account", func(t *testing.T) {
		memberList, err := repo.ListByAccount(ctx, accountID)
		require.NoError(t, err)
		assert.Len(t, memberList, 3)

		// Verify all members belong to the correct account
		for _, member := range memberList {
			assert.Equal(t, accountID, member.AccountID)
		}
	})

	t.Run("should return empty list for account with no members", func(t *testing.T) {
		// Create account without members
		emptyAccountID := "empty-account"
		createTestAccount(t, db, emptyAccountID, ownerID, "Empty Account", "Account with no members")

		memberList, err := repo.ListByAccount(ctx, emptyAccountID)
		require.NoError(t, err)
		assert.Len(t, memberList, 0)
	})

	t.Run("should return error for empty account ID", func(t *testing.T) {
		memberList, err := repo.ListByAccount(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, memberList)
	})

	t.Run("should return empty list for non-existent account", func(t *testing.T) {
		memberList, err := repo.ListByAccount(ctx, "non-existent-account")
		require.NoError(t, err)
		assert.Len(t, memberList, 0)
	})
}

func TestGormAccountMemberRepository_ListByUser(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormAccountMemberRepository(db)
	ctx := context.Background()

	// Create test data
	ownerID := "owner-user-member-4"
	userID := "multi-member-user"
	joinedAt := time.Now().Add(-96 * time.Hour)

	createTestUser(t, db, ownerID, "https://example.com/users/owner-member-4", "owner-member4@example.com", "Owner Member 4", string(domain.UserStatusActive))
	createTestUser(t, db, userID, "https://example.com/users/multi-member", "multimember@example.com", "Multi Member User", string(domain.UserStatusActive))

	// Create multiple accounts and memberships for the same user
	accounts := []struct {
		id, name, roleID string
	}{
		{"account-multi-1", "Multi Account 1", "member"},
		{"account-multi-2", "Multi Account 2", "admin"},
		{"account-multi-3", "Multi Account 3", "viewer"},
	}

	for i, acc := range accounts {
		createTestAccount(t, db, acc.id, ownerID, acc.name, "Multi account "+acc.name)
		memberID := "multi-member-" + acc.id
		createTestAccountMember(t, db, memberID, acc.id, userID, acc.roleID, ownerID, joinedAt.Add(time.Duration(i)*time.Hour))
	}

	t.Run("should return all memberships for user", func(t *testing.T) {
		memberList, err := repo.ListByUser(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, memberList, 3)

		// Verify all memberships belong to the correct user
		for _, member := range memberList {
			assert.Equal(t, userID, member.UserID)
		}

		// Verify different roles
		roles := make(map[string]bool)
		for _, member := range memberList {
			roles[member.RoleID] = true
		}
		assert.True(t, roles["member"])
		assert.True(t, roles["admin"])
		assert.True(t, roles["viewer"])
	})

	t.Run("should return empty list for user with no memberships", func(t *testing.T) {
		// Create user without memberships
		noMemberUserID := "no-member-user"
		createTestUser(t, db, noMemberUserID, "https://example.com/users/no-member", "nomember@example.com", "No Member User", string(domain.UserStatusActive))

		memberList, err := repo.ListByUser(ctx, noMemberUserID)
		require.NoError(t, err)
		assert.Len(t, memberList, 0)
	})

	t.Run("should return error for empty user ID", func(t *testing.T) {
		memberList, err := repo.ListByUser(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, memberList)
	})

	t.Run("should return empty list for non-existent user", func(t *testing.T) {
		memberList, err := repo.ListByUser(ctx, "non-existent-user")
		require.NoError(t, err)
		assert.Len(t, memberList, 0)
	})
}
