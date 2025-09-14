package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for AccountService testing

// MockAccountRepository mocks the account repository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByOwner(ctx context.Context, ownerID string) ([]*domain.Account, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Account), args.Error(1)
}

// MockRoleRepository mocks the role repository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) GetSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Role), args.Error(1)
}

// MockInvitationRepository mocks the invitation repository
type MockInvitationRepository struct {
	mock.Mock
}

func (m *MockInvitationRepository) GetByID(ctx context.Context, id string) (*domain.Invitation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.Invitation, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invitation), args.Error(1)
}

func (m *MockInvitationRepository) ListByEmail(ctx context.Context, email string) ([]*domain.Invitation, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invitation), args.Error(1)
}

// MockAccountMemberRepository mocks the account member repository
type MockAccountMemberRepository struct {
	mock.Mock
}

func (m *MockAccountMemberRepository) GetByID(ctx context.Context, id string) (*domain.AccountMember, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AccountMember), args.Error(1)
}

func (m *MockAccountMemberRepository) GetByAccountAndUser(ctx context.Context, accountID, userID string) (*domain.AccountMember, error) {
	args := m.Called(ctx, accountID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AccountMember), args.Error(1)
}

func (m *MockAccountMemberRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.AccountMember, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AccountMember), args.Error(1)
}

func (m *MockAccountMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.AccountMember, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AccountMember), args.Error(1)
}

// MockInvitationGenerator mocks the invitation generator
type MockInvitationGenerator struct {
	mock.Mock
}

func (m *MockInvitationGenerator) GenerateToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInvitationGenerator) GenerateInvitationID() string {
	args := m.Called()
	return args.String(0)
}

// Test data helpers
func createTestAccount(id, ownerID, name string) *domain.Account {
	owner := createTestUser(ownerID, "owner@example.com", "Owner User")
	account, _ := domain.NewAccount(context.Background(), id, owner, name, "Test account description")
	account.MarkEventsAsCommitted() // Clear creation events for test setup
	return account
}

func createTestRole(id, name string) *domain.Role {
	permissions := []domain.Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "create", Scope: "own"},
	}
	role, _ := domain.NewRole(context.Background(), id, name, "Test role", permissions)
	role.MarkEventsAsCommitted() // Clear creation events for test setup
	return role
}

func createTestInvitation(id, accountID, email, roleID, invitedBy string) *domain.Invitation {
	account := createTestAccount(accountID, "owner-id", "Test Account")
	role := createTestRole(roleID, "Test Role")
	inviter := createTestUser(invitedBy, "inviter@example.com", "Inviter User")

	expiresAt := time.Now().Add(24 * time.Hour)
	invitation, _ := domain.NewInvitation(context.Background(), id, "test-token", account, email, role, inviter, expiresAt)
	invitation.MarkEventsAsCommitted() // Clear creation events for test setup
	return invitation
}

func createTestAccountMember(id, accountID, userID, roleID string) *domain.AccountMember {
	account := createTestAccount(accountID, "owner-id", "Test Account")
	user := createTestUser(userID, "member@example.com", "Member User")
	role := createTestRole(roleID, "Test Role")
	inviter := createTestUser("inviter-id", "inviter@example.com", "Inviter User")

	member, _ := domain.NewAccountMember(context.Background(), id, account, user, role, inviter, time.Now())
	member.MarkEventsAsCommitted() // Clear creation events for test setup
	return member
}

// Test CreateAccount method

func TestAccountService_CreateAccount_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	ownerID := "owner-user-id"
	accountName := "Test Account"
	owner := createTestUser(ownerID, "owner@example.com", "Owner User")

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	account, err := service.CreateAccount(ctx, ownerID, accountName)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, ownerID, account.OwnerID)
	assert.Equal(t, accountName, account.Name)

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestAccountService_CreateAccount_OwnerNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	ownerID := "non-existent-user"
	accountName := "Test Account"

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, ownerID).Return(nil, errors.New("user not found"))

	// Act
	account, err := service.CreateAccount(ctx, ownerID, accountName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "failed to get owner user")

	mockUserRepo.AssertExpectations(t)
}

func TestAccountService_CreateAccount_EmptyName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	ownerID := "owner-user-id"
	accountName := ""
	owner := createTestUser(ownerID, "owner@example.com", "Owner User")

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, ownerID).Return(owner, nil)

	// Act
	account, err := service.CreateAccount(ctx, ownerID, accountName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, account)
	assert.Contains(t, err.Error(), "account name is required")

	mockUserRepo.AssertExpectations(t)
}

// Test InviteUser method

func TestAccountService_InviteUser_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "test-account-id"
	inviterID := "inviter-user-id"
	email := "invitee@example.com"
	roleID := "member"

	account := createTestAccount(accountID, "owner-id", "Test Account")
	inviter := createTestUser(inviterID, "inviter@example.com", "Inviter User")
	role := createTestRole(roleID, "Member")

	// Mock expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockUserRepo.On("GetByID", ctx, inviterID).Return(inviter, nil)
	mockRoleRepo.On("GetByID", ctx, roleID).Return(role, nil)
	mockInviteGen.On("GenerateInvitationID").Return("invitation-id")
	mockInviteGen.On("GenerateToken").Return("invitation-token")
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	invitation, err := service.InviteUser(ctx, accountID, inviterID, email, roleID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, invitation)
	assert.Equal(t, accountID, invitation.AccountID)
	assert.Equal(t, email, invitation.Email)
	assert.Equal(t, roleID, invitation.RoleID)
	assert.Equal(t, inviterID, invitation.InvitedBy)
	assert.Equal(t, domain.InvitationStatusPending, invitation.Status)

	mockAccountRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockInviteGen.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestAccountService_InviteUser_AccountNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "non-existent-account"
	inviterID := "inviter-user-id"
	email := "invitee@example.com"
	roleID := "member"

	// Mock expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(nil, errors.New("account not found"))

	// Act
	invitation, err := service.InviteUser(ctx, accountID, inviterID, email, roleID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, invitation)
	assert.Contains(t, err.Error(), "failed to get account")

	mockAccountRepo.AssertExpectations(t)
}

func TestAccountService_InviteUser_InvalidRole(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "test-account-id"
	inviterID := "inviter-user-id"
	email := "invitee@example.com"
	roleID := "invalid-role"

	account := createTestAccount(accountID, "owner-id", "Test Account")
	inviter := createTestUser(inviterID, "inviter@example.com", "Inviter User")

	// Mock expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockUserRepo.On("GetByID", ctx, inviterID).Return(inviter, nil)
	mockRoleRepo.On("GetByID", ctx, roleID).Return(nil, errors.New("role not found"))

	// Act
	invitation, err := service.InviteUser(ctx, accountID, inviterID, email, roleID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, invitation)
	assert.Contains(t, err.Error(), "failed to get role")

	mockAccountRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

// Test AcceptInvitation method

func TestAccountService_AcceptInvitation_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	token := "invitation-token"
	userID := "accepting-user-id"

	invitation := createTestInvitation("invitation-id", "account-id", "user@example.com", "member", "inviter-id")
	user := createTestUser(userID, "user@example.com", "Accepting User")
	account := createTestAccount("account-id", "owner-id", "Test Account")
	role := createTestRole("member", "Member")

	// Mock expectations
	mockInvitationRepo.On("GetByToken", ctx, token).Return(invitation, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockAccountRepo.On("GetByID", ctx, invitation.AccountID).Return(account, nil)
	mockRoleRepo.On("GetByID", ctx, invitation.RoleID).Return(role, nil)
	mockInviteGen.On("GenerateInvitationID").Return("member-id")
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.AcceptInvitation(ctx, token, userID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.InvitationStatusAccepted, invitation.Status)

	mockInvitationRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockInviteGen.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestAccountService_AcceptInvitation_InvalidToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	token := "invalid-token"
	userID := "accepting-user-id"

	// Mock expectations
	mockInvitationRepo.On("GetByToken", ctx, token).Return(nil, errors.New("invitation not found"))

	// Act
	err := service.AcceptInvitation(ctx, token, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get invitation")

	mockInvitationRepo.AssertExpectations(t)
}

// Test UpdateMemberRole method

func TestAccountService_UpdateMemberRole_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "test-account-id"
	userID := "member-user-id"
	newRoleID := "admin"

	account := createTestAccount(accountID, "owner-id", "Test Account")
	user := createTestUser(userID, "member@example.com", "Member User")
	member := createTestAccountMember("member-id", accountID, userID, "member")
	oldRole := createTestRole("member", "Member")
	newRole := createTestRole(newRoleID, "Admin")

	// Mock expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockMemberRepo.On("GetByAccountAndUser", ctx, accountID, userID).Return(member, nil)
	mockRoleRepo.On("GetByID", ctx, "member").Return(oldRole, nil)
	mockRoleRepo.On("GetByID", ctx, newRoleID).Return(newRole, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	err := service.UpdateMemberRole(ctx, accountID, userID, newRoleID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, newRoleID, member.RoleID)

	mockAccountRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestAccountService_UpdateMemberRole_MemberNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "test-account-id"
	userID := "non-existent-user"
	newRoleID := "admin"

	account := createTestAccount(accountID, "owner-id", "Test Account")
	user := createTestUser(userID, "member@example.com", "Member User")

	// Mock expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockMemberRepo.On("GetByAccountAndUser", ctx, accountID, userID).Return(nil, errors.New("member not found"))

	// Act
	err := service.UpdateMemberRole(ctx, accountID, userID, newRoleID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get account member")

	mockAccountRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

// Test event emission validation

func TestAccountService_CreateAccount_EmitsCorrectEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	ownerID := "owner-user-id"
	accountName := "Test Account"
	owner := createTestUser(ownerID, "owner@example.com", "Owner User")

	var capturedEvents []domain.Event

	// Mock expectations with event capture
	mockUserRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).
		Run(func(args mock.Arguments) {
			capturedEvents = args.Get(0).([]domain.Event)
		}).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	account, err := service.CreateAccount(ctx, ownerID, accountName)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, account)
	assert.NotEmpty(t, capturedEvents)

	entityEvent := capturedEvents[0].(*domain.EntityEvent)
	assert.Equal(t, "account.created", entityEvent.EventType())
	assert.Equal(t, account.ID(), entityEvent.AggregateID())

	mockUserRepo.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}

func TestAccountService_InviteUser_EmitsCorrectEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUnitOfWork := &MockUnitOfWork{}
	mockAccountRepo := &MockAccountRepository{}
	mockUserRepo := &MockUserRepository{}
	mockRoleRepo := &MockRoleRepository{}
	mockInvitationRepo := &MockInvitationRepository{}
	mockMemberRepo := &MockAccountMemberRepository{}
	mockInviteGen := &MockInvitationGenerator{}

	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return mockUnitOfWork
	}

	service := NewAccountService(unitOfWorkFactory, mockInviteGen, mockAccountRepo, mockUserRepo, mockRoleRepo, mockInvitationRepo, mockMemberRepo)

	accountID := "test-account-id"
	inviterID := "inviter-user-id"
	email := "invitee@example.com"
	roleID := "member"

	account := createTestAccount(accountID, "owner-id", "Test Account")
	inviter := createTestUser(inviterID, "inviter@example.com", "Inviter User")
	role := createTestRole(roleID, "Member")

	var capturedEvents []domain.Event

	// Mock expectations with event capture
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockUserRepo.On("GetByID", ctx, inviterID).Return(inviter, nil)
	mockRoleRepo.On("GetByID", ctx, roleID).Return(role, nil)
	mockInviteGen.On("GenerateInvitationID").Return("invitation-id")
	mockInviteGen.On("GenerateToken").Return("invitation-token")
	mockUnitOfWork.On("RegisterEvents", mock.AnythingOfType("[]domain.Event")).
		Run(func(args mock.Arguments) {
			capturedEvents = args.Get(0).([]domain.Event)
		}).Return()
	mockUnitOfWork.On("Commit", ctx).Return([]pericarpdomain.Envelope{}, nil)

	// Act
	invitation, err := service.InviteUser(ctx, accountID, inviterID, email, roleID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, invitation)
	assert.NotEmpty(t, capturedEvents)

	entityEvent := capturedEvents[0].(*domain.EntityEvent)
	assert.Equal(t, "invitation.created", entityEvent.EventType())
	assert.Equal(t, invitation.ID(), entityEvent.AggregateID())

	mockAccountRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockInviteGen.AssertExpectations(t)
	mockUnitOfWork.AssertExpectations(t)
}
