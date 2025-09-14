package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/akeemphilbert/goro/internal/user/application"
	userdomain "github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
)

// RegistrationService handles user registration with external identities
type RegistrationService struct {
	userService    application.UserService
	identityRepo   domain.ExternalIdentityRepository
	webidGenerator infrastructure.WebIDGenerator
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	userService application.UserService,
	identityRepo domain.ExternalIdentityRepository,
	webidGenerator infrastructure.WebIDGenerator,
) *RegistrationService {
	return &RegistrationService{
		userService:    userService,
		identityRepo:   identityRepo,
		webidGenerator: webidGenerator,
	}
}

// RegisterWithExternalIdentity registers a new user using external identity provider
func (s *RegistrationService) RegisterWithExternalIdentity(ctx context.Context, provider string, profile domain.ExternalProfile) (*userdomain.User, error) {
	// Validate external profile
	if !profile.IsValid() {
		return nil, fmt.Errorf("invalid external profile: missing required fields")
	}

	// Check if external identity is already linked to another user
	existingUserID, err := s.identityRepo.FindByExternalID(ctx, provider, profile.ID)
	if err != nil && err != domain.ErrExternalIdentityNotFound {
		return nil, fmt.Errorf("failed to check existing external identity: %w", err)
	}

	if existingUserID != "" {
		return nil, domain.ErrExternalIdentityAlreadyLinked
	}

	// Create user profile from external profile
	userProfile := userdomain.UserProfile{
		Name: profile.Name,
		Bio:  "", // External providers typically don't provide bio
		Preferences: map[string]interface{}{
			"registration_provider": provider,
		},
	}

	// Register new user
	registerReq := application.RegisterUserRequest{
		Email:   profile.Email,
		Profile: userProfile,
	}

	user, err := s.userService.RegisterUser(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Link external identity to the new user
	err = s.identityRepo.LinkIdentity(ctx, user.ID(), provider, profile.ID)
	if err != nil {
		// If linking fails, we should ideally rollback user creation
		// For now, we'll return the error and let the caller handle it
		return nil, fmt.Errorf("failed to link external identity: %w", err)
	}

	return user, nil
}

// LinkExternalIdentity links an external identity to an existing authenticated user
func (s *RegistrationService) LinkExternalIdentity(ctx context.Context, userID, provider string, profile domain.ExternalProfile) error {
	// Validate inputs
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if provider == "" {
		return fmt.Errorf("provider is required")
	}

	if !profile.IsValid() {
		return fmt.Errorf("invalid external profile: missing required fields")
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if external identity is already linked to any user (including this one)
	existingUserID, err := s.identityRepo.FindByExternalID(ctx, provider, profile.ID)
	if err != nil && err != domain.ErrExternalIdentityNotFound {
		return fmt.Errorf("failed to check existing external identity: %w", err)
	}

	if existingUserID != "" {
		if existingUserID == userID {
			// Identity is already linked to this user
			return domain.ErrExternalIdentityAlreadyLinked
		}
		// Identity is linked to a different user
		return domain.ErrExternalIdentityAlreadyLinked
	}

	// Link the external identity
	err = s.identityRepo.LinkIdentity(ctx, userID, provider, profile.ID)
	if err != nil {
		return fmt.Errorf("failed to link external identity: %w", err)
	}

	return nil
}

// UnlinkExternalIdentity removes the link between a user and an external identity
func (s *RegistrationService) UnlinkExternalIdentity(ctx context.Context, userID, provider, externalID string) error {
	// Validate inputs
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if provider == "" {
		return fmt.Errorf("provider is required")
	}

	if externalID == "" {
		return fmt.Errorf("external ID is required")
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Unlink the external identity
	err = s.identityRepo.UnlinkIdentity(ctx, userID, provider, externalID)
	if err != nil {
		return fmt.Errorf("failed to unlink external identity: %w", err)
	}

	return nil
}

// GetLinkedIdentities retrieves all external identities linked to a user
func (s *RegistrationService) GetLinkedIdentities(ctx context.Context, userID string) ([]*domain.ExternalIdentity, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get linked identities
	identities, err := s.identityRepo.GetLinkedIdentities(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked identities: %w", err)
	}

	return identities, nil
}
