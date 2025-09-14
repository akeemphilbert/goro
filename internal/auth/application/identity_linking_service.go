package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	userapplication "github.com/akeemphilbert/goro/internal/user/application"
)

// IdentityLinkingService handles external identity linking operations for existing users
type IdentityLinkingService struct {
	userService  userapplication.UserService
	identityRepo domain.ExternalIdentityRepository
}

// NewIdentityLinkingService creates a new identity linking service
func NewIdentityLinkingService(
	userService userapplication.UserService,
	identityRepo domain.ExternalIdentityRepository,
) *IdentityLinkingService {
	return &IdentityLinkingService{
		userService:  userService,
		identityRepo: identityRepo,
	}
}

// LinkIdentity links an external identity to an existing authenticated user
func (s *IdentityLinkingService) LinkIdentity(ctx context.Context, userID, provider string, profile domain.ExternalProfile) error {
	// Validate inputs
	if err := s.validateLinkingInputs(userID, provider, profile); err != nil {
		return err
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check for duplicate identity prevention
	if err := s.preventDuplicateIdentity(ctx, userID, provider, profile.ID); err != nil {
		return err
	}

	// Link the external identity
	err = s.identityRepo.LinkIdentity(ctx, userID, provider, profile.ID)
	if err != nil {
		return fmt.Errorf("failed to link external identity: %w", err)
	}

	return nil
}

// UnlinkIdentity removes the link between a user and an external identity
func (s *IdentityLinkingService) UnlinkIdentity(ctx context.Context, userID, provider, externalID string) error {
	// Validate inputs
	if err := s.validateUnlinkingInputs(userID, provider, externalID); err != nil {
		return err
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify the identity is actually linked to this user
	linkedUserID, err := s.identityRepo.FindByExternalID(ctx, provider, externalID)
	if err != nil {
		if err == domain.ErrExternalIdentityNotFound {
			return domain.ErrExternalIdentityNotFound
		}
		return fmt.Errorf("failed to verify identity ownership: %w", err)
	}

	if linkedUserID != userID {
		return fmt.Errorf("external identity is not linked to this user")
	}

	// Unlink the external identity
	err = s.identityRepo.UnlinkIdentity(ctx, userID, provider, externalID)
	if err != nil {
		return fmt.Errorf("failed to unlink external identity: %w", err)
	}

	return nil
}

// GetLinkedIdentities retrieves all external identities linked to a user
func (s *IdentityLinkingService) GetLinkedIdentities(ctx context.Context, userID string) ([]*domain.ExternalIdentity, error) {
	// Validate input
	if userID == "" {
		return nil, domain.ErrUserIDEmpty
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

// UnlinkAllIdentities removes all external identity links for a user
func (s *IdentityLinkingService) UnlinkAllIdentities(ctx context.Context, userID string) error {
	// Validate input
	if userID == "" {
		return domain.ErrUserIDEmpty
	}

	// Verify user exists
	_, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Unlink all identities
	err = s.identityRepo.UnlinkAllIdentities(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to unlink all external identities: %w", err)
	}

	return nil
}

// IsIdentityLinked checks if an external identity is linked to any user
func (s *IdentityLinkingService) IsIdentityLinked(ctx context.Context, provider, externalID string) (bool, string, error) {
	// Validate inputs
	if provider == "" {
		return false, "", domain.ErrProviderEmpty
	}

	if externalID == "" {
		return false, "", domain.ErrExternalIDEmpty
	}

	// Check if identity is linked
	userID, err := s.identityRepo.FindByExternalID(ctx, provider, externalID)
	if err != nil {
		if err == domain.ErrExternalIdentityNotFound {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check identity linking: %w", err)
	}

	return true, userID, nil
}

// GetIdentitiesByProvider retrieves all external identities for a specific provider
func (s *IdentityLinkingService) GetIdentitiesByProvider(ctx context.Context, provider string) ([]*domain.ExternalIdentity, error) {
	// Validate input
	if provider == "" {
		return nil, domain.ErrProviderEmpty
	}

	// Get identities by provider
	identities, err := s.identityRepo.GetByProvider(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get identities by provider: %w", err)
	}

	return identities, nil
}

// validateLinkingInputs validates inputs for identity linking
func (s *IdentityLinkingService) validateLinkingInputs(userID, provider string, profile domain.ExternalProfile) error {
	if userID == "" {
		return domain.ErrUserIDEmpty
	}

	if provider == "" {
		return domain.ErrProviderEmpty
	}

	if !profile.IsValid() {
		return domain.ErrExternalIdentityInvalid
	}

	return nil
}

// validateUnlinkingInputs validates inputs for identity unlinking
func (s *IdentityLinkingService) validateUnlinkingInputs(userID, provider, externalID string) error {
	if userID == "" {
		return domain.ErrUserIDEmpty
	}

	if provider == "" {
		return domain.ErrProviderEmpty
	}

	if externalID == "" {
		return domain.ErrExternalIDEmpty
	}

	return nil
}

// preventDuplicateIdentity checks and prevents duplicate identity linking
func (s *IdentityLinkingService) preventDuplicateIdentity(ctx context.Context, userID, provider, externalID string) error {
	// Check if external identity is already linked to any user
	existingUserID, err := s.identityRepo.FindByExternalID(ctx, provider, externalID)
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

	return nil
}
