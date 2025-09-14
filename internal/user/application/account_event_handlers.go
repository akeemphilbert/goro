package application

import (
	"context"
	"fmt"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// AccountEventHandler handles account-related domain events for persistence operations
type AccountEventHandler struct {
	accountRepo    domain.AccountWriteRepository
	memberRepo     domain.AccountMemberWriteRepository
	invitationRepo domain.InvitationWriteRepository
	fileStorage    FileStorage
}

// NewAccountEventHandler creates a new account event handler
func NewAccountEventHandler(
	accountRepo domain.AccountWriteRepository,
	memberRepo domain.AccountMemberWriteRepository,
	invitationRepo domain.InvitationWriteRepository,
	fileStorage FileStorage,
) *AccountEventHandler {
	return &AccountEventHandler{
		accountRepo:    accountRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
		fileStorage:    fileStorage,
	}
}

// HandleAccountCreated handles account creation events by persisting to database and file storage
func (h *AccountEventHandler) HandleAccountCreated(ctx context.Context, event *domain.AccountCreatedEventData) error {
	// First persist to database
	if err := h.accountRepo.Create(ctx, event.Account); err != nil {
		return fmt.Errorf("failed to persist account creation: %w", err)
	}

	return nil
}

// HandleAccountUpdated handles account update events by updating database and file storage
func (h *AccountEventHandler) HandleAccountUpdated(ctx context.Context, event *domain.AccountUpdatedEventData) error {
	// Update database
	if err := h.accountRepo.Update(ctx, event.Account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// HandleAccountMemberAdded handles member addition events by creating membership projection
func (h *AccountEventHandler) HandleAccountMemberAdded(ctx context.Context, event *domain.AccountMemberAddedEventData) error {
	// Create membership projection in database
	if err := h.memberRepo.Create(ctx, event.AccountMember); err != nil {
		return fmt.Errorf("failed to create account member: %w", err)
	}

	return nil
}

// HandleAccountMemberRemoved handles member removal events by deleting membership projection
func (h *AccountEventHandler) HandleAccountMemberRemoved(ctx context.Context, event *domain.AccountMemberRemovedEventData) error {
	// Remove membership projection from database
	if err := h.memberRepo.Delete(ctx, event.AccountMember.ID()); err != nil {
		return fmt.Errorf("failed to remove account member: %w", err)
	}

	return nil
}

// HandleAccountMemberRoleUpdated handles member role update events by updating membership projection
func (h *AccountEventHandler) HandleAccountMemberRoleUpdated(ctx context.Context, event *domain.AccountMemberRoleUpdatedEventData) error {
	// Update membership projection in database
	if err := h.memberRepo.Update(ctx, event.AccountMember); err != nil {
		return fmt.Errorf("failed to update account member role: %w", err)
	}

	return nil
}

// HandleInvitationCreated handles invitation creation events by persisting invitation
func (h *AccountEventHandler) HandleInvitationCreated(ctx context.Context, event *domain.InvitationCreatedEventData) error {
	// Persist invitation to database
	if err := h.invitationRepo.Create(ctx, event.Invitation); err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// HandleInvitationAccepted handles invitation acceptance events by updating invitation and creating membership
func (h *AccountEventHandler) HandleInvitationAccepted(ctx context.Context, event *domain.InvitationAcceptedEventData) error {
	// Update invitation status in database
	if err := h.invitationRepo.Update(ctx, event.Invitation); err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	// Create membership projection
	if err := h.memberRepo.Create(ctx, event.AccountMember); err != nil {
		return fmt.Errorf("failed to create account member from invitation: %w", err)
	}

	return nil
}

// HandleInvitationRevoked handles invitation revocation events by updating invitation status
func (h *AccountEventHandler) HandleInvitationRevoked(ctx context.Context, event *domain.InvitationRevokedEventData) error {
	// Update invitation status in database
	if err := h.invitationRepo.Update(ctx, event.Invitation); err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	return nil
}

// HandleInvitationExpired handles invitation expiration events by updating invitation status
func (h *AccountEventHandler) HandleInvitationExpired(ctx context.Context, event *domain.InvitationExpiredEventData) error {
	// Update invitation status in database
	if err := h.invitationRepo.Update(ctx, event.Invitation); err != nil {
		return fmt.Errorf("failed to expire invitation: %w", err)
	}

	return nil
}
