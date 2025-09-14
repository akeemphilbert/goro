package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/google/wire"
)

// Service Providers
func ProvideUserService(
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	webidGen infrastructure.WebIDGenerator,
	userRepo domain.UserRepository,
) (UserService, error) {
	if unitOfWorkFactory == nil {
		return nil, fmt.Errorf("unit of work factory cannot be nil")
	}
	if webidGen == nil {
		return nil, fmt.Errorf("WebID generator cannot be nil")
	}
	if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	}

	return NewUserService(unitOfWorkFactory, webidGen, userRepo), nil
}

func ProvideAccountService(
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	inviteGen InvitationGenerator,
	accountRepo domain.AccountRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	invitationRepo domain.InvitationRepository,
	memberRepo domain.AccountMemberRepository,
) (AccountService, error) {
	if unitOfWorkFactory == nil {
		return nil, fmt.Errorf("unit of work factory cannot be nil")
	}
	if inviteGen == nil {
		return nil, fmt.Errorf("invitation generator cannot be nil")
	}
	if accountRepo == nil {
		return nil, fmt.Errorf("account repository cannot be nil")
	}
	if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	}
	if roleRepo == nil {
		return nil, fmt.Errorf("role repository cannot be nil")
	}
	if invitationRepo == nil {
		return nil, fmt.Errorf("invitation repository cannot be nil")
	}
	if memberRepo == nil {
		return nil, fmt.Errorf("member repository cannot be nil")
	}

	return NewAccountService(unitOfWorkFactory, inviteGen, accountRepo, userRepo, roleRepo, invitationRepo, memberRepo), nil
}

// Event Handler Providers
func ProvideUserEventHandler(
	userWriteRepo domain.UserWriteRepository,
	fileStorage FileStorage,
) (*UserEventHandler, error) {
	if userWriteRepo == nil {
		return nil, fmt.Errorf("user write repository cannot be nil")
	}
	if fileStorage == nil {
		return nil, fmt.Errorf("file storage cannot be nil")
	}

	return NewUserEventHandler(userWriteRepo, fileStorage), nil
}

func ProvideAccountEventHandler(
	accountWriteRepo domain.AccountWriteRepository,
	accountMemberWriteRepo domain.AccountMemberWriteRepository,
	invitationWriteRepo domain.InvitationWriteRepository,
	fileStorage FileStorage,
) (*AccountEventHandler, error) {
	if accountWriteRepo == nil {
		return nil, fmt.Errorf("account write repository cannot be nil")
	}
	if accountMemberWriteRepo == nil {
		return nil, fmt.Errorf("account member write repository cannot be nil")
	}
	if invitationWriteRepo == nil {
		return nil, fmt.Errorf("invitation write repository cannot be nil")
	}
	if fileStorage == nil {
		return nil, fmt.Errorf("file storage cannot be nil")
	}

	return NewAccountEventHandler(accountWriteRepo, accountMemberWriteRepo, invitationWriteRepo, fileStorage), nil
}

// Event Handler Registration Provider
func ProvideEventHandlerRegistrar(
	eventDispatcher pericarpdomain.EventDispatcher,
	userEventHandler *UserEventHandler,
	accountEventHandler *AccountEventHandler,
) (*EventHandlerRegistrar, error) {
	if eventDispatcher == nil {
		return nil, fmt.Errorf("event dispatcher cannot be nil")
	}
	if userEventHandler == nil {
		return nil, fmt.Errorf("user event handler cannot be nil")
	}
	if accountEventHandler == nil {
		return nil, fmt.Errorf("account event handler cannot be nil")
	}

	registrar := NewEventHandlerRegistrar(eventDispatcher)

	// Register user event handlers
	if err := registrar.RegisterUserEventHandlers(userEventHandler); err != nil {
		return nil, fmt.Errorf("failed to register user event handlers: %w", err)
	}

	// Register account event handlers
	if err := registrar.RegisterAccountEventHandlers(accountEventHandler); err != nil {
		return nil, fmt.Errorf("failed to register account event handlers: %w", err)
	}

	return registrar, nil
}

// Provider Sets
var UserApplicationProviderSet = wire.NewSet(
	ProvideUserService,
	ProvideAccountService,
	ProvideUserEventHandler,
	ProvideAccountEventHandler,
	ProvideEventHandlerRegistrar,
	ProvideInvitationGenerator,
	ProvideFileStorageAdapter,
	ProvideUnitOfWorkFactory,
)

// Additional providers for missing dependencies

// ProvideInvitationGenerator provides a simple invitation generator
func ProvideInvitationGenerator() InvitationGenerator {
	return &simpleInvitationGenerator{}
}

// ProvideFileStorageAdapter provides a file storage adapter for the application layer
func ProvideFileStorageAdapter(domainFileStorage domain.FileStorage) FileStorage {
	return &fileStorageAdapter{
		domainStorage: domainFileStorage,
	}
}

// ProvideUnitOfWorkFactory provides a unit of work factory
func ProvideUnitOfWorkFactory(
	eventStore pericarpdomain.EventStore,
	eventDispatcher pericarpdomain.EventDispatcher,
) func() pericarpdomain.UnitOfWork {
	return func() pericarpdomain.UnitOfWork {
		// This would typically come from the pericarp infrastructure
		// For now, return a mock implementation
		return &mockUnitOfWork{}
	}
}

// Simple implementations for testing

type simpleInvitationGenerator struct{}

func (g *simpleInvitationGenerator) GenerateToken() string {
	return "test-token-123"
}

func (g *simpleInvitationGenerator) GenerateInvitationID() string {
	return "test-invitation-id-123"
}

type fileStorageAdapter struct {
	domainStorage domain.FileStorage
}

func (f *fileStorageAdapter) WriteUserProfile(ctx context.Context, userID string, profile domain.UserProfile) error {
	// Create a minimal user for storage
	user, err := domain.NewUser(ctx, userID, "", "", profile)
	if err != nil {
		return err
	}
	return f.domainStorage.StoreUserProfile(ctx, userID, user)
}

func (f *fileStorageAdapter) WriteWebIDDocument(ctx context.Context, userID, webID, document string) error {
	return f.domainStorage.StoreWebIDDocument(ctx, userID, document)
}

func (f *fileStorageAdapter) DeleteUserFiles(ctx context.Context, userID string) error {
	if err := f.domainStorage.DeleteUserProfile(ctx, userID); err != nil {
		return err
	}
	return f.domainStorage.DeleteWebIDDocument(ctx, userID)
}

func (f *fileStorageAdapter) ReadUserProfile(ctx context.Context, userID string) (domain.UserProfile, error) {
	user, err := f.domainStorage.LoadUserProfile(ctx, userID)
	if err != nil {
		return domain.UserProfile{}, err
	}
	return user.Profile, nil
}

func (f *fileStorageAdapter) ReadWebIDDocument(ctx context.Context, userID string) (string, error) {
	return f.domainStorage.LoadWebIDDocument(ctx, userID)
}

func (f *fileStorageAdapter) UserExists(ctx context.Context, userID string) (bool, error) {
	_, err := f.domainStorage.LoadUserProfile(ctx, userID)
	if err != nil {
		return false, nil // Assume user doesn't exist if we can't load profile
	}
	return true, nil
}

type mockUnitOfWork struct{}

func (m *mockUnitOfWork) RegisterEvents(events []domain.Event) {
	// Mock implementation - do nothing
}

func (m *mockUnitOfWork) Commit(ctx context.Context) ([]pericarpdomain.Envelope, error) {
	// Mock implementation - return empty envelopes
	return []pericarpdomain.Envelope{}, nil
}

func (m *mockUnitOfWork) Rollback() error {
	// Mock implementation - do nothing
	return nil
}
