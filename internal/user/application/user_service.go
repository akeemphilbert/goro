package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/google/uuid"
)

// RegisterUserRequest represents a request to register a new user
type RegisterUserRequest struct {
	Email   string             `json:"email"`
	Profile domain.UserProfile `json:"profile"`
}

// UserService defines the interface for user management operations
type UserService interface {
	RegisterUser(ctx context.Context, req RegisterUserRequest) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, profile domain.UserProfile) error
	DeleteAccount(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	GetUserByWebID(ctx context.Context, webID string) (*domain.User, error)
}

// userService implements the UserService interface
type userService struct {
	unitOfWorkFactory func() pericarpdomain.UnitOfWork
	webidGen          infrastructure.WebIDGenerator
	userRepo          domain.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(
	unitOfWorkFactory func() pericarpdomain.UnitOfWork,
	webidGen infrastructure.WebIDGenerator,
	userRepo domain.UserRepository,
) UserService {
	return &userService{
		unitOfWorkFactory: unitOfWorkFactory,
		webidGen:          webidGen,
		userRepo:          userRepo,
	}
}

// RegisterUser registers a new user with WebID generation
func (s *userService) RegisterUser(ctx context.Context, req RegisterUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	// Generate user ID
	userID := generateUserID()

	// Create user entity first to validate email format
	user, err := domain.NewUser(ctx, userID, "temp-webid", req.Email, req.Profile)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate WebID
	webID, err := s.webidGen.GenerateWebID(ctx, userID, req.Email, req.Profile.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate WebID: %w", err)
	}

	// Update user with real WebID
	user.WebID = webID

	// Validate WebID format
	if err := s.webidGen.ValidateWebID(ctx, webID); err != nil {
		return nil, fmt.Errorf("invalid WebID format: %w", err)
	}

	// Check WebID uniqueness
	isUnique, err := s.webidGen.IsUniqueWebID(ctx, webID)
	if err != nil {
		return nil, fmt.Errorf("failed to check WebID uniqueness: %w", err)
	}

	// Generate alternative WebID if not unique
	if !isUnique {
		webID, err = s.webidGen.GenerateAlternativeWebID(ctx, webID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate alternative WebID: %w", err)
		}

		// Validate alternative WebID
		if err := s.webidGen.ValidateWebID(ctx, webID); err != nil {
			return nil, fmt.Errorf("invalid alternative WebID format: %w", err)
		}

		// Check alternative WebID uniqueness
		isUnique, err = s.webidGen.IsUniqueWebID(ctx, webID)
		if err != nil {
			return nil, fmt.Errorf("failed to check alternative WebID uniqueness: %w", err)
		}

		if !isUnique {
			return nil, fmt.Errorf("failed to generate unique WebID")
		}

		// Update user with alternative WebID
		user.WebID = webID
	}

	// Clear any events from temporary user creation and add final events
	user.MarkEventsAsCommitted()

	// Add the final user created event with correct WebID
	event := domain.NewUserCreatedEvent(user, webID, string(domain.UserStatusActive))
	user.AddEvent(event)

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := user.UncommittedEvents()
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
		return nil, fmt.Errorf("failed to commit user registration: %w", err)
	}

	// Mark events as committed
	user.MarkEventsAsCommitted()

	return user, nil
}

// UpdateProfile updates a user's profile
func (s *userService) UpdateProfile(ctx context.Context, userID string, profile domain.UserProfile) error {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update profile
	if err := user.UpdateProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := user.UncommittedEvents()
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
		return fmt.Errorf("failed to commit profile update: %w", err)
	}

	// Mark events as committed
	user.MarkEventsAsCommitted()

	return nil
}

// DeleteAccount marks a user account as deleted
func (s *userService) DeleteAccount(ctx context.Context, userID string) error {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user
	if err := user.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Create unit of work for event processing
	unitOfWork := s.unitOfWorkFactory()

	// Register events with unit of work
	events := user.UncommittedEvents()
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
		return fmt.Errorf("failed to commit user deletion: %w", err)
	}

	// Mark events as committed
	user.MarkEventsAsCommitted()

	return nil
}

// GetUserByID retrieves a user by their ID
func (s *userService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByWebID retrieves a user by their WebID
func (s *userService) GetUserByWebID(ctx context.Context, webID string) (*domain.User, error) {
	return s.userRepo.GetByWebID(ctx, webID)
}

// generateUserID generates a new unique user ID
func generateUserID() string {
	return uuid.New().String()
}
