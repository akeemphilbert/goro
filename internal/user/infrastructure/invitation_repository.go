package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// GormInvitationRepository implements domain.InvitationRepository using GORM
type GormInvitationRepository struct {
	db *gorm.DB
}

// NewGormInvitationRepository creates a new GORM-based invitation repository
func NewGormInvitationRepository(db *gorm.DB) domain.InvitationRepository {
	return &GormInvitationRepository{db: db}
}

// GetByID retrieves an invitation by ID
func (r *GormInvitationRepository) GetByID(ctx context.Context, id string) (*domain.Invitation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("invitation ID cannot be empty")
	}

	var invitationModel InvitationModel
	err := r.db.WithContext(ctx).First(&invitationModel, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invitation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get invitation by ID %s: %w", id, err)
	}

	return r.modelToDomain(&invitationModel), nil
}

// GetByToken retrieves an invitation by token
func (r *GormInvitationRepository) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("invitation token cannot be empty")
	}

	var invitationModel InvitationModel
	err := r.db.WithContext(ctx).First(&invitationModel, "token = ?", token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invitation not found with token: %s", token)
		}
		return nil, fmt.Errorf("failed to get invitation by token %s: %w", token, err)
	}

	return r.modelToDomain(&invitationModel), nil
}

// ListByAccount retrieves all invitations for an account
func (r *GormInvitationRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.Invitation, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var invitationModels []InvitationModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&invitationModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations for account %s: %w", accountID, err)
	}

	invitations := make([]*domain.Invitation, len(invitationModels))
	for i, model := range invitationModels {
		invitations[i] = r.modelToDomain(&model)
	}

	return invitations, nil
}

// ListByEmail retrieves all invitations for an email address
func (r *GormInvitationRepository) ListByEmail(ctx context.Context, email string) ([]*domain.Invitation, error) {
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var invitationModels []InvitationModel
	err := r.db.WithContext(ctx).Where("email = ?", email).Find(&invitationModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations for email %s: %w", email, err)
	}

	invitations := make([]*domain.Invitation, len(invitationModels))
	for i, model := range invitationModels {
		invitations[i] = r.modelToDomain(&model)
	}

	return invitations, nil
}

// modelToDomain converts an InvitationModel to a domain.Invitation
func (r *GormInvitationRepository) modelToDomain(model *InvitationModel) *domain.Invitation {
	invitation := &domain.Invitation{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		AccountID:   model.AccountID,
		Email:       model.Email,
		RoleID:      model.RoleID,
		Token:       model.Token,
		InvitedBy:   model.InvitedBy,
		Status:      domain.InvitationStatus(model.Status),
		ExpiresAt:   model.ExpiresAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return invitation
}
