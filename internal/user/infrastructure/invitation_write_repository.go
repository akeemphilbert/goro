package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// GormInvitationWriteRepository implements domain.InvitationWriteRepository using GORM
type GormInvitationWriteRepository struct {
	db *gorm.DB
}

// NewGormInvitationWriteRepository creates a new GORM-based invitation write repository
func NewGormInvitationWriteRepository(db *gorm.DB) domain.InvitationWriteRepository {
	return &GormInvitationWriteRepository{db: db}
}

// Create creates a new invitation in the database
func (r *GormInvitationWriteRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	if invitation == nil {
		return fmt.Errorf("invitation cannot be nil")
	}

	if strings.TrimSpace(invitation.ID()) == "" {
		return fmt.Errorf("invitation ID cannot be empty")
	}

	if strings.TrimSpace(invitation.AccountID) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if strings.TrimSpace(invitation.Email) == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if strings.TrimSpace(invitation.RoleID) == "" {
		return fmt.Errorf("role ID cannot be empty")
	}

	if strings.TrimSpace(invitation.Token) == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Convert domain invitation to GORM model
	invitationModel := &InvitationModel{
		ID:        invitation.ID(),
		AccountID: invitation.AccountID,
		Email:     invitation.Email,
		RoleID:    invitation.RoleID,
		Token:     invitation.Token,
		InvitedBy: invitation.InvitedBy,
		Status:    string(invitation.Status),
		ExpiresAt: invitation.ExpiresAt,
		CreatedAt: invitation.CreatedAt,
		UpdatedAt: invitation.UpdatedAt,
	}

	err := r.db.WithContext(ctx).Create(invitationModel).Error
	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// Update updates an existing invitation in the database
func (r *GormInvitationWriteRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	if invitation == nil {
		return fmt.Errorf("invitation cannot be nil")
	}

	if strings.TrimSpace(invitation.ID()) == "" {
		return fmt.Errorf("invitation ID cannot be empty")
	}

	// Convert domain invitation to GORM model
	invitationModel := &InvitationModel{
		ID:        invitation.ID(),
		AccountID: invitation.AccountID,
		Email:     invitation.Email,
		RoleID:    invitation.RoleID,
		Token:     invitation.Token,
		InvitedBy: invitation.InvitedBy,
		Status:    string(invitation.Status),
		ExpiresAt: invitation.ExpiresAt,
		CreatedAt: invitation.CreatedAt,
		UpdatedAt: invitation.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&InvitationModel{}).Where("id = ?", invitation.ID()).Updates(invitationModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update invitation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invitation not found: %s", invitation.ID())
	}

	return nil
}

// Delete deletes an invitation from the database
func (r *GormInvitationWriteRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("invitation ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&InvitationModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete invitation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invitation not found: %s", id)
	}

	return nil
}
