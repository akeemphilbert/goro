package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/google/uuid"
)

// InvitationGenerator defines the interface for generating invitations
type InvitationGenerator interface {
	GenerateToken() string
	GenerateInvitationID() string
}

// AccountService defines the interface for account management operations
type AccountService interface {
	CreateAccount(ctx context.Context, ownerID string, name string) (*domain.Account, error)
	InviteUser(ctx context.Context, accountID, inviterID, email string, roleID string) (*domain.Invitation, error)
	AcceptInvitation(ctx context.Context, token string, userID string) error
	UpdateMemberRole(ctx context.Context, accountID, userID string, roleID string) error
}

// accountService implements the AccountService interface
type accountService struct {
	unitOfWorkFactory func() pericarpdomain.UnitOfWork
	inviteGen         InvitationGenerator
	accountRepo       domain.AccountRepository
	userRepo          domain.UserRepository
	roleRepo          domain.RoleRepository
	invitationRepo    domain.InvitationRepository
	memberRepo        domain.AccountMemberRepository
}

// NewAccountService creates a new AccountService instance
func NewAccountService(
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	inviteGen InvitationGenerator,
	accountRepo domain.AccountRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	invitationRepo domain.InvitationRepository,
	memberRepo domain.AccountMemberRepository,
) AccountService {
	return &accountService{
		unitOfWorkFactory: unitOfWorkFactory,
		inviteGen:         inviteGen,
		accountRepo:       accountRepo,
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		invitationRepo:    invitationRepo,
		memberRepo:        memberRepo,
	}
}

// CreateAccount creates a new account with owner assignment
func (s *accountService) CreateAccount(ctx context.Context, ownerID string, name string) (*domain.Account, error) {
	// Get owner user first
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner user: %w", err)
	}

	// Validate input
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("account name is required")
	}

	// Generate account ID
	accountID := generateAccountID()

	// Create account entity
	account, err := domain.NewAccount(ctx, accountID, owner, name, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := account.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work - this persists events and dispatches them
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback on failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			// Log rollback error but return original error
		}
		return nil, fmt.Errorf("failed to commit account creation: %w", err)
	}

	// Mark events as committed
	account.MarkEventsAsCommitted()

	return account, nil
}

// InviteUser invites a user to join an account with a specific role
func (s *accountService) InviteUser(ctx context.Context, accountID, inviterID, email string, roleID string) (*domain.Invitation, error) {
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Get inviter user
	inviter, err := s.userRepo.GetByID(ctx, inviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inviter user: %w", err)
	}

	// Get role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Generate invitation details
	invitationID := s.inviteGen.GenerateInvitationID()
	token := s.inviteGen.GenerateToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days from now

	// Create invitation entity
	invitation, err := domain.NewInvitation(ctx, invitationID, token, account, email, role, inviter, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := invitation.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work - this persists events and dispatches them
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback on failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			// Log rollback error but return original error
		}
		return nil, fmt.Errorf("failed to commit invitation creation: %w", err)
	}

	// Mark events as committed
	invitation.MarkEventsAsCommitted()

	return invitation, nil
}

// AcceptInvitation accepts an invitation and creates account membership
func (s *accountService) AcceptInvitation(ctx context.Context, token string, userID string) error {
	// Get invitation by token
	invitation, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get account
	account, err := s.accountRepo.GetByID(ctx, invitation.AccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get role
	role, err := s.roleRepo.GetByID(ctx, invitation.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Create account member
	memberID := s.inviteGen.GenerateInvitationID() // Reuse the ID generator
	member, err := domain.NewAccountMember(ctx, memberID, account, user, role, user, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create account member: %w", err)
	}

	// Accept invitation
	if err := invitation.Accept(ctx, user, account, role, member); err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events from both invitation and member
	invitationEvents := invitation.UncommittedEvents()
	memberEvents := member.UncommittedEvents()

	allEvents := make([]domain.Event, 0, len(invitationEvents)+len(memberEvents))
	allEvents = append(allEvents, invitationEvents...)
	allEvents = append(allEvents, memberEvents...)

	if len(allEvents) > 0 {
		unitOfWork.RegisterEvents(allEvents)
	}

	// Commit unit of work - this persists events and dispatches them
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback on failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			// Log rollback error but return original error
		}
		return fmt.Errorf("failed to commit invitation acceptance: %w", err)
	}

	// Mark events as committed
	invitation.MarkEventsAsCommitted()
	member.MarkEventsAsCommitted()

	return nil
}

// UpdateMemberRole updates a member's role in an account
func (s *accountService) UpdateMemberRole(ctx context.Context, accountID, userID string, roleID string) error {
	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Get account member
	member, err := s.memberRepo.GetByAccountAndUser(ctx, accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to get account member: %w", err)
	}

	// Get old role
	oldRole, err := s.roleRepo.GetByID(ctx, member.RoleID)
	if err != nil {
		return fmt.Errorf("failed to get old role: %w", err)
	}

	// Get new role
	newRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get new role: %w", err)
	}

	// Update member role
	if err := member.UpdateRole(ctx, account, user, oldRole, newRole); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := member.UncommittedEvents()
	if len(events) > 0 {
		unitOfWork.RegisterEvents(events)
	}

	// Commit unit of work - this persists events and dispatches them
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		// Rollback on failure
		if rollbackErr := unitOfWork.Rollback(); rollbackErr != nil {
			// Log rollback error but return original error
		}
		return fmt.Errorf("failed to commit member role update: %w", err)
	}

	// Mark events as committed
	member.MarkEventsAsCommitted()

	return nil
}

// Helper functions
func generateAccountID() string {
	return uuid.New().String()
}

func generateInvitationToken() string {
	return uuid.New().String()
}
