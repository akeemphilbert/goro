package integration

import (
	"context"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestGormRepositoryImplementations tests GORM repository implementations with temporary database
func TestGormRepositoryImplementations(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)

	// Seed system roles
	err := infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	t.Run("UserRepository", func(t *testing.T) {
		testUserRepository(t, db)
	})

	t.Run("AccountRepository", func(t *testing.T) {
		testAccountRepository(t, db)
	})

	t.Run("RoleRepository", func(t *testing.T) {
		testRoleRepository(t, db)
	})

	t.Run("InvitationRepository", func(t *testing.T) {
		testInvitationRepository(t, db)
	})

	t.Run("AccountMemberRepository", func(t *testing.T) {
		testAccountMemberRepository(t, db)
	})
}

func testUserRepository(t *testing.T, db *gorm.DB) {
	ctx := context.Background()

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)

	// Create test user
	profile := domain.UserProfile{
		Name:        "Test User",
		Bio:         "Test bio",
		Avatar:      "https://example.com/avatar.jpg",
		Preferences: map[string]interface{}{"theme": "dark"},
	}

	user, err := domain.NewUser(ctx, "test-user-1", "https://example.com/users/test-user-1", "test@example.com", profile)
	require.NoError(t, err)

	// Test Create
	err = userWriteRepo.Create(ctx, user)
	require.NoError(t, err)

	// Test GetByID
	retrievedUser, err := userRepo.GetByID(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, user.ID(), retrievedUser.ID())
	assert.Equal(t, user.WebID, retrievedUser.WebID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.Profile.Name, retrievedUser.Profile.Name)

	// Test GetByWebID
	retrievedUser, err = userRepo.GetByWebID(ctx, user.WebID)
	require.NoError(t, err)
	assert.Equal(t, user.ID(), retrievedUser.ID())

	// Test GetByEmail
	retrievedUser, err = userRepo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID(), retrievedUser.ID())

	// Test Exists
	exists, err := userRepo.Exists(ctx, user.ID())
	require.NoError(t, err)
	assert.True(t, exists)

	// Test List
	filter := domain.UserFilter{
		Status: domain.UserStatusActive,
		Limit:  10,
		Offset: 0,
	}
	users, err := userRepo.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, user.ID(), users[0].ID())

	// Test Update
	user.Profile.Name = "Updated Name"
	user.UpdatedAt = time.Now()
	err = userWriteRepo.Update(ctx, user)
	require.NoError(t, err)

	retrievedUser, err = userRepo.GetByID(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrievedUser.Profile.Name)

	// Test Delete
	err = userWriteRepo.Delete(ctx, user.ID())
	require.NoError(t, err)

	exists, err = userRepo.Exists(ctx, user.ID())
	require.NoError(t, err)
	assert.False(t, exists)
}

func testAccountRepository(t *testing.T, db *gorm.DB) {
	ctx := context.Background()

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)

	// Create owner user first
	ownerProfile := domain.UserProfile{
		Name:        "Owner User",
		Bio:         "Account owner",
		Avatar:      "https://example.com/owner.jpg",
		Preferences: map[string]interface{}{},
	}

	owner, err := domain.NewUser(ctx, "owner-1", "https://example.com/users/owner-1", "owner@example.com", ownerProfile)
	require.NoError(t, err)

	err = userWriteRepo.Create(ctx, owner)
	require.NoError(t, err)

	// Create test account
	account, err := domain.NewAccount(ctx, "test-account-1", owner, "Test Account", "Test description")
	require.NoError(t, err)

	// Test Create
	err = accountWriteRepo.Create(ctx, account)
	require.NoError(t, err)

	// Test GetByID
	retrievedAccount, err := accountRepo.GetByID(ctx, account.ID())
	require.NoError(t, err)
	assert.Equal(t, account.ID(), retrievedAccount.ID())
	assert.Equal(t, account.OwnerID, retrievedAccount.OwnerID)
	assert.Equal(t, account.Name, retrievedAccount.Name)
	assert.Equal(t, account.Description, retrievedAccount.Description)

	// Test GetByOwner
	ownerAccounts, err := accountRepo.GetByOwner(ctx, owner.ID())
	require.NoError(t, err)
	assert.Len(t, ownerAccounts, 1)
	assert.Equal(t, account.ID(), ownerAccounts[0].ID())

	// Test Update
	account.Name = "Updated Account Name"
	account.UpdatedAt = time.Now()
	err = accountWriteRepo.Update(ctx, account)
	require.NoError(t, err)

	retrievedAccount, err = accountRepo.GetByID(ctx, account.ID())
	require.NoError(t, err)
	assert.Equal(t, "Updated Account Name", retrievedAccount.Name)

	// Test Delete
	err = accountWriteRepo.Delete(ctx, account.ID())
	require.NoError(t, err)

	_, err = accountRepo.GetByID(ctx, account.ID())
	assert.Error(t, err) // Should not be found
}

func testRoleRepository(t *testing.T, db *gorm.DB) {
	ctx := context.Background()

	// Setup repository
	roleRepo := infrastructure.NewGormRoleRepository(db)

	// Test GetSystemRoles
	systemRoles, err := roleRepo.GetSystemRoles(ctx)
	require.NoError(t, err)
	assert.Len(t, systemRoles, 4) // owner, admin, member, viewer

	roleNames := make([]string, len(systemRoles))
	for i, role := range systemRoles {
		roleNames[i] = role.Name
	}
	assert.Contains(t, roleNames, "Owner")
	assert.Contains(t, roleNames, "Administrator")
	assert.Contains(t, roleNames, "Member")
	assert.Contains(t, roleNames, "Viewer")

	// Test GetByID
	ownerRole, err := roleRepo.GetByID(ctx, "owner")
	require.NoError(t, err)
	assert.Equal(t, "owner", ownerRole.ID())
	assert.Equal(t, "Owner", ownerRole.Name)
	assert.NotEmpty(t, ownerRole.Permissions)

	// Test List
	allRoles, err := roleRepo.List(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allRoles), 4) // At least the system roles
}

func testInvitationRepository(t *testing.T, db *gorm.DB) {
	ctx := context.Background()

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Create test data
	owner, account := createTestUserAndAccount(t, ctx, userWriteRepo, accountWriteRepo)

	// Get member role
	memberRole, err := roleRepo.GetByID(ctx, "member")
	require.NoError(t, err)

	// Create test invitation
	invitation, err := domain.NewInvitation(ctx, "test-invitation-1", "test-token-123", account, "invitee@example.com", memberRole, owner, time.Now().Add(7*24*time.Hour))
	require.NoError(t, err)

	// Test Create
	err = invitationWriteRepo.Create(ctx, invitation)
	require.NoError(t, err)

	// Test GetByID
	retrievedInvitation, err := invitationRepo.GetByID(ctx, invitation.ID())
	require.NoError(t, err)
	assert.Equal(t, invitation.ID(), retrievedInvitation.ID())
	assert.Equal(t, invitation.Token, retrievedInvitation.Token)
	assert.Equal(t, invitation.Email, retrievedInvitation.Email)

	// Test GetByToken
	retrievedInvitation, err = invitationRepo.GetByToken(ctx, invitation.Token)
	require.NoError(t, err)
	assert.Equal(t, invitation.ID(), retrievedInvitation.ID())

	// Test ListByAccount
	accountInvitations, err := invitationRepo.ListByAccount(ctx, account.ID())
	require.NoError(t, err)
	assert.Len(t, accountInvitations, 1)
	assert.Equal(t, invitation.ID(), accountInvitations[0].ID())

	// Test ListByEmail
	emailInvitations, err := invitationRepo.ListByEmail(ctx, invitation.Email)
	require.NoError(t, err)
	assert.Len(t, emailInvitations, 1)
	assert.Equal(t, invitation.ID(), emailInvitations[0].ID())

	// Test Update
	invitation.Status = domain.InvitationStatusAccepted
	invitation.UpdatedAt = time.Now()
	err = invitationWriteRepo.Update(ctx, invitation)
	require.NoError(t, err)

	retrievedInvitation, err = invitationRepo.GetByID(ctx, invitation.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.InvitationStatusAccepted, retrievedInvitation.Status)
}

func testAccountMemberRepository(t *testing.T, db *gorm.DB) {
	ctx := context.Background()

	// Setup repositories
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)

	// Create test data
	owner, account := createTestUserAndAccount(t, ctx, userWriteRepo, accountWriteRepo)

	// Create member user
	memberProfile := domain.UserProfile{
		Name:        "Member User",
		Bio:         "Account member",
		Avatar:      "https://example.com/member.jpg",
		Preferences: map[string]interface{}{},
	}

	member, err := domain.NewUser(ctx, "member-1", "https://example.com/users/member-1", "member@example.com", memberProfile)
	require.NoError(t, err)

	err = userWriteRepo.Create(ctx, member)
	require.NoError(t, err)

	// Get member role
	memberRole, err := roleRepo.GetByID(ctx, "member")
	require.NoError(t, err)

	// Create account member
	accountMember, err := domain.NewAccountMember(ctx, "test-member-1", account, member, memberRole, owner, time.Now())
	require.NoError(t, err)

	// Test Create
	err = memberWriteRepo.Create(ctx, accountMember)
	require.NoError(t, err)

	// Test GetByID
	retrievedMember, err := memberRepo.GetByID(ctx, accountMember.ID())
	require.NoError(t, err)
	assert.Equal(t, accountMember.ID(), retrievedMember.ID())
	assert.Equal(t, accountMember.AccountID, retrievedMember.AccountID)
	assert.Equal(t, accountMember.UserID, retrievedMember.UserID)
	assert.Equal(t, accountMember.RoleID, retrievedMember.RoleID)

	// Test GetByAccountAndUser
	retrievedMember, err = memberRepo.GetByAccountAndUser(ctx, account.ID(), member.ID())
	require.NoError(t, err)
	assert.Equal(t, accountMember.ID(), retrievedMember.ID())

	// Test ListByAccount
	accountMembers, err := memberRepo.ListByAccount(ctx, account.ID())
	require.NoError(t, err)
	assert.Len(t, accountMembers, 1)
	assert.Equal(t, accountMember.ID(), accountMembers[0].ID())

	// Test ListByUser
	userMemberships, err := memberRepo.ListByUser(ctx, member.ID())
	require.NoError(t, err)
	assert.Len(t, userMemberships, 1)
	assert.Equal(t, accountMember.ID(), userMemberships[0].ID())

	// Test Update
	adminRole, err := roleRepo.GetByID(ctx, "admin")
	require.NoError(t, err)

	accountMember.RoleID = adminRole.ID()
	accountMember.UpdatedAt = time.Now()
	err = memberWriteRepo.Update(ctx, accountMember)
	require.NoError(t, err)

	retrievedMember, err = memberRepo.GetByID(ctx, accountMember.ID())
	require.NoError(t, err)
	assert.Equal(t, adminRole.ID(), retrievedMember.RoleID)

	// Test Delete
	err = memberWriteRepo.Delete(ctx, accountMember.ID())
	require.NoError(t, err)

	_, err = memberRepo.GetByID(ctx, accountMember.ID())
	assert.Error(t, err) // Should not be found
}

// Helper functions

func createTestUserAndAccount(t *testing.T, ctx context.Context, userWriteRepo domain.UserWriteRepository, accountWriteRepo domain.AccountWriteRepository) (*domain.User, *domain.Account) {
	// Create owner user
	ownerProfile := domain.UserProfile{
		Name:        "Owner User",
		Bio:         "Account owner",
		Avatar:      "https://example.com/owner.jpg",
		Preferences: map[string]interface{}{},
	}

	owner, err := domain.NewUser(ctx, "owner-1", "https://example.com/users/owner-1", "owner@example.com", ownerProfile)
	require.NoError(t, err)

	err = userWriteRepo.Create(ctx, owner)
	require.NoError(t, err)

	// Create account
	account, err := domain.NewAccount(ctx, "test-account-1", owner, "Test Account", "Test description")
	require.NoError(t, err)

	err = accountWriteRepo.Create(ctx, account)
	require.NoError(t, err)

	return owner, account
}
