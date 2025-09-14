package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccountCreationWorkflow tests the complete account creation workflow
func TestAccountCreationWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database
	db := setupTestDatabase(t)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Setup WebID generator and invitation generator
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")
	inviteGen := &testInvitationGenerator{}

	// Setup event dispatcher and unit of work
	eventDispatcher := setupAccountTestEventDispatcher(t, userWriteRepo, accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup services
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)
	accountService := application.NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo)

	// Create owner user first
	ownerProfile := domain.UserProfile{
		Name:        "Account Owner",
		Bio:         "Owner of the account",
		Avatar:      "https://example.com/owner.jpg",
		Preferences: map[string]interface{}{"role": "owner"},
	}

	ownerReq := application.RegisterUserRequest{
		Email:   "owner@example.com",
		Profile: ownerProfile,
	}

	owner, err := userService.RegisterUser(ctx, ownerReq)
	require.NoError(t, err)

	// Create account
	accountName := "Test Account"
	account, err := accountService.CreateAccount(ctx, owner.ID(), accountName)
	require.NoError(t, err)
	require.NotNil(t, account)

	// Verify account entity
	assert.NotEmpty(t, account.ID())
	assert.Equal(t, owner.ID(), account.OwnerID)
	assert.Equal(t, accountName, account.Name)
	assert.True(t, account.Settings.AllowInvitations)
	assert.Equal(t, "member", account.Settings.DefaultRoleID)
	assert.Equal(t, 100, account.Settings.MaxMembers)

	// Verify database persistence
	dbAccount, err := accountRepo.GetByID(ctx, account.ID())
	require.NoError(t, err)
	assert.Equal(t, account.ID(), dbAccount.ID())
	assert.Equal(t, account.OwnerID, dbAccount.OwnerID)
	assert.Equal(t, account.Name, dbAccount.Name)
}

// TestInvitationWorkflow tests the complete invitation workflow
func TestInvitationWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database and repositories
	db := setupTestDatabase(t)

	// Seed system roles
	err := infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup generators
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")
	inviteGen := &testInvitationGenerator{}

	// Setup event dispatcher and unit of work
	eventDispatcher := setupAccountTestEventDispatcher(t, userWriteRepo, accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup services
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)
	accountService := application.NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo)

	// Create owner and account
	owner, account := createTestOwnerAndAccount(t, ctx, userService, accountService)

	// Create invitee user
	inviteeProfile := domain.UserProfile{
		Name:        "Invitee User",
		Bio:         "User to be invited",
		Avatar:      "https://example.com/invitee.jpg",
		Preferences: map[string]interface{}{},
	}

	inviteeReq := application.RegisterUserRequest{
		Email:   "invitee@example.com",
		Profile: inviteeProfile,
	}

	invitee, err := userService.RegisterUser(ctx, inviteeReq)
	require.NoError(t, err)

	// Send invitation
	invitation, err := accountService.InviteUser(ctx, account.ID(), owner.ID(), invitee.Email, "member")
	require.NoError(t, err)
	require.NotNil(t, invitation)

	// Verify invitation entity
	assert.NotEmpty(t, invitation.ID())
	assert.Equal(t, account.ID(), invitation.AccountID)
	assert.Equal(t, invitee.Email, invitation.Email)
	assert.Equal(t, "member", invitation.RoleID)
	assert.Equal(t, owner.ID(), invitation.InvitedBy)
	assert.Equal(t, domain.InvitationStatusPending, invitation.Status)
	assert.NotEmpty(t, invitation.Token)

	// Verify database persistence
	dbInvitation, err := invitationRepo.GetByToken(ctx, invitation.Token)
	require.NoError(t, err)
	assert.Equal(t, invitation.ID(), dbInvitation.ID())
	assert.Equal(t, invitation.Token, dbInvitation.Token)
	assert.Equal(t, invitation.Status, dbInvitation.Status)

	// Accept invitation
	err = accountService.AcceptInvitation(ctx, invitation.Token, invitee.ID())
	require.NoError(t, err)

	// Verify invitation is accepted
	dbInvitation, err = invitationRepo.GetByToken(ctx, invitation.Token)
	require.NoError(t, err)
	assert.Equal(t, domain.InvitationStatusAccepted, dbInvitation.Status)

	// Verify account membership is created
	member, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), invitee.ID())
	require.NoError(t, err)
	assert.Equal(t, account.ID(), member.AccountID)
	assert.Equal(t, invitee.ID(), member.UserID)
	assert.Equal(t, "member", member.RoleID)
	assert.Equal(t, owner.ID(), member.InvitedBy)
}

// TestMembershipManagementWorkflow tests the complete membership management workflow
func TestMembershipManagementWorkflow(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database and repositories
	db := setupTestDatabase(t)

	// Seed system roles
	err := infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup generators
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")
	inviteGen := &testInvitationGenerator{}

	// Setup event dispatcher and unit of work
	eventDispatcher := setupAccountTestEventDispatcher(t, userWriteRepo, accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup services
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)
	accountService := application.NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo)

	// Create owner and account
	owner, account := createTestOwnerAndAccount(t, ctx, userService, accountService)

	// Create and invite member
	member, _ := createTestMemberAndInvite(t, ctx, userService, accountService, account, owner, "member@example.com", "member")

	// Update member role from member to admin
	err = accountService.UpdateMemberRole(ctx, account.ID(), member.ID(), "admin")
	require.NoError(t, err)

	// Verify role update
	dbMember, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), member.ID())
	require.NoError(t, err)
	assert.Equal(t, "admin", dbMember.RoleID)

	// Verify member can be found by account
	accountMembers, err := memberRepo.ListByAccount(ctx, account.ID())
	require.NoError(t, err)
	assert.Len(t, accountMembers, 1)
	assert.Equal(t, member.ID(), accountMembers[0].UserID)
	assert.Equal(t, "admin", accountMembers[0].RoleID)

	// Verify member can be found by user
	userMemberships, err := memberRepo.ListByUser(ctx, member.ID())
	require.NoError(t, err)
	assert.Len(t, userMemberships, 1)
	assert.Equal(t, account.ID(), userMemberships[0].AccountID)
	assert.Equal(t, "admin", userMemberships[0].RoleID)
}

// TestAtomicEntityRelationships tests atomic entity relationships through event projections
func TestAtomicEntityRelationships(t *testing.T) {
	// Setup test environment
	ctx := context.Background()
	tempDir := t.TempDir()

	// Setup database and repositories
	db := setupTestDatabase(t)

	// Seed system roles
	err := infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Setup repositories
	userRepo := infrastructure.NewGormUserRepository(db)
	userWriteRepo := infrastructure.NewGormUserWriteRepository(db)
	accountRepo := infrastructure.NewGormAccountRepository(db)
	accountWriteRepo := infrastructure.NewGormAccountWriteRepository(db)
	roleRepo := infrastructure.NewGormRoleRepository(db)
	invitationRepo := infrastructure.NewGormInvitationRepository(db)
	memberRepo := infrastructure.NewGormAccountMemberRepository(db)
	memberWriteRepo := infrastructure.NewGormAccountMemberWriteRepository(db)
	invitationWriteRepo := infrastructure.NewGormInvitationWriteRepository(db)

	// Setup file storage
	fileStorage := setupTestFileStorage(t, tempDir)

	// Setup generators
	webidGen := infrastructure.NewWebIDGenerator("https://example.com")
	inviteGen := &testInvitationGenerator{}

	// Setup event dispatcher and unit of work
	eventDispatcher := setupAccountTestEventDispatcher(t, userWriteRepo, accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)
	unitOfWorkFactory := setupTestUnitOfWorkFactory(t, eventDispatcher)

	// Setup services
	userService := application.NewUserService(unitOfWorkFactory, webidGen, userRepo)
	accountService := application.NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo)

	// Create owner and account
	owner, account := createTestOwnerAndAccount(t, ctx, userService, accountService)

	// Create multiple members with different roles
	member1, _ := createTestMemberAndInvite(t, ctx, userService, accountService, account, owner, "member1@example.com", "member")
	member2, _ := createTestMemberAndInvite(t, ctx, userService, accountService, account, owner, "member2@example.com", "admin")
	member3, _ := createTestMemberAndInvite(t, ctx, userService, accountService, account, owner, "member3@example.com", "viewer")

	// Verify atomic relationships

	// 1. Account -> Members relationship
	accountMembers, err := memberRepo.ListByAccount(ctx, account.ID())
	require.NoError(t, err)
	assert.Len(t, accountMembers, 3)

	memberIDs := make([]string, len(accountMembers))
	for i, m := range accountMembers {
		memberIDs[i] = m.UserID
	}
	assert.Contains(t, memberIDs, member1.ID())
	assert.Contains(t, memberIDs, member2.ID())
	assert.Contains(t, memberIDs, member3.ID())

	// 2. User -> Memberships relationship
	user1Memberships, err := memberRepo.ListByUser(ctx, member1.ID())
	require.NoError(t, err)
	assert.Len(t, user1Memberships, 1)
	assert.Equal(t, account.ID(), user1Memberships[0].AccountID)
	assert.Equal(t, "member", user1Memberships[0].RoleID)

	// 3. Role consistency across relationships
	member1Data, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), member1.ID())
	require.NoError(t, err)
	assert.Equal(t, "member", member1Data.RoleID)

	member2Data, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), member2.ID())
	require.NoError(t, err)
	assert.Equal(t, "admin", member2Data.RoleID)

	member3Data, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), member3.ID())
	require.NoError(t, err)
	assert.Equal(t, "viewer", member3Data.RoleID)

	// 4. Invitation -> Member relationship consistency
	// All invitations should be accepted and linked to members
	accountInvitations, err := invitationRepo.ListByAccount(ctx, account.ID())
	require.NoError(t, err)
	assert.Len(t, accountInvitations, 3)

	for _, invitation := range accountInvitations {
		assert.Equal(t, domain.InvitationStatusAccepted, invitation.Status)

		// Find corresponding member
		member, err := memberRepo.GetByAccountAndUser(ctx, account.ID(), getMemberIDByEmail(t, ctx, userRepo, invitation.Email))
		require.NoError(t, err)
		assert.Equal(t, invitation.RoleID, member.RoleID)
		assert.Equal(t, invitation.InvitedBy, member.InvitedBy)
	}
}

// Helper functions

func setupAccountTestEventDispatcher(t *testing.T, userWriteRepo domain.UserWriteRepository, accountWriteRepo domain.AccountWriteRepository, memberWriteRepo domain.AccountMemberWriteRepository, invitationWriteRepo domain.InvitationWriteRepository, fileStorage application.FileStorage) pericarpdomain.EventDispatcher {
	dispatcher := pericarpdomain.NewInMemoryEventDispatcher()

	// Register user event handlers
	userEventHandler := application.NewUserEventHandler(userWriteRepo, fileStorage)
	accountEventHandler := application.NewAccountEventHandler(accountWriteRepo, memberWriteRepo, invitationWriteRepo, fileStorage)

	// Register user event handlers
	dispatcher.RegisterHandler("UserRegisteredEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserRegisteredEventData); ok {
			return userEventHandler.HandleUserRegistered(ctx, userEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("UserProfileUpdatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserProfileUpdatedEventData); ok {
			return userEventHandler.HandleUserProfileUpdated(ctx, userEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("UserDeletedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if userEvent, ok := event.Data().(*domain.UserDeletedEventData); ok {
			return userEventHandler.HandleUserDeleted(ctx, userEvent)
		}
		return nil
	})

	// Register account event handlers
	dispatcher.RegisterHandler("AccountCreatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if accountEvent, ok := event.Data().(*domain.AccountCreatedEventData); ok {
			return accountEventHandler.HandleAccountCreated(ctx, accountEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("InvitationCreatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if invitationEvent, ok := event.Data().(*domain.InvitationCreatedEventData); ok {
			return accountEventHandler.HandleInvitationCreated(ctx, invitationEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("InvitationAcceptedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if invitationEvent, ok := event.Data().(*domain.InvitationAcceptedEventData); ok {
			return accountEventHandler.HandleInvitationAccepted(ctx, invitationEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("AccountMemberCreatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if memberEvent, ok := event.Data().(*domain.AccountMemberCreatedEventData); ok {
			return accountEventHandler.HandleAccountMemberCreated(ctx, memberEvent)
		}
		return nil
	})

	dispatcher.RegisterHandler("AccountMemberRoleUpdatedEvent", func(ctx context.Context, event pericarpdomain.Event) error {
		if memberEvent, ok := event.Data().(*domain.AccountMemberRoleUpdatedEventData); ok {
			return accountEventHandler.HandleAccountMemberRoleUpdated(ctx, memberEvent)
		}
		return nil
	})

	return dispatcher
}

func createTestOwnerAndAccount(t *testing.T, ctx context.Context, userService application.UserService, accountService application.AccountService) (*domain.User, *domain.Account) {
	// Create owner user
	ownerProfile := domain.UserProfile{
		Name:        "Account Owner",
		Bio:         "Owner of the account",
		Avatar:      "https://example.com/owner.jpg",
		Preferences: map[string]interface{}{"role": "owner"},
	}

	ownerReq := application.RegisterUserRequest{
		Email:   "owner@example.com",
		Profile: ownerProfile,
	}

	owner, err := userService.RegisterUser(ctx, ownerReq)
	require.NoError(t, err)

	// Create account
	account, err := accountService.CreateAccount(ctx, owner.ID(), "Test Account")
	require.NoError(t, err)

	return owner, account
}

func createTestMemberAndInvite(t *testing.T, ctx context.Context, userService application.UserService, accountService application.AccountService, account *domain.Account, owner *domain.User, email, roleID string) (*domain.User, *domain.Invitation) {
	// Create member user
	memberProfile := domain.UserProfile{
		Name:        "Member User",
		Bio:         "Member of the account",
		Avatar:      "https://example.com/member.jpg",
		Preferences: map[string]interface{}{},
	}

	memberReq := application.RegisterUserRequest{
		Email:   email,
		Profile: memberProfile,
	}

	member, err := userService.RegisterUser(ctx, memberReq)
	require.NoError(t, err)

	// Send invitation
	invitation, err := accountService.InviteUser(ctx, account.ID(), owner.ID(), member.Email, roleID)
	require.NoError(t, err)

	// Accept invitation
	err = accountService.AcceptInvitation(ctx, invitation.Token, member.ID())
	require.NoError(t, err)

	return member, invitation
}

func getMemberIDByEmail(t *testing.T, ctx context.Context, userRepo domain.UserRepository, email string) string {
	user, err := userRepo.GetByEmail(ctx, email)
	require.NoError(t, err)
	return user.ID()
}

// testInvitationGenerator implements InvitationGenerator for testing
type testInvitationGenerator struct {
	counter int
}

func (g *testInvitationGenerator) GenerateToken() string {
	g.counter++
	return fmt.Sprintf("test-token-%d-%d", g.counter, time.Now().Unix())
}

func (g *testInvitationGenerator) GenerateInvitationID() string {
	g.counter++
	return fmt.Sprintf("test-invitation-%d", g.counter)
}
