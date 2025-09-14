package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// GormAccountMemberRepository implements domain.AccountMemberRepository using GORM
type GormAccountMemberRepository struct {
	db *gorm.DB
}

// NewGormAccountMemberRepository creates a new GORM-based account member repository
func NewGormAccountMemberRepository(db *gorm.DB) domain.AccountMemberRepository {
	return &GormAccountMemberRepository{db: db}
}

// GetByID retrieves an account member by ID
func (r *GormAccountMemberRepository) GetByID(ctx context.Context, id string) (*domain.AccountMember, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("member ID cannot be empty")
	}

	var memberModel AccountMemberModel
	err := r.db.WithContext(ctx).First(&memberModel, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account member not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get account member by ID %s: %w", id, err)
	}

	return r.modelToDomain(&memberModel), nil
}

// GetByAccountAndUser retrieves an account member by account ID and user ID
func (r *GormAccountMemberRepository) GetByAccountAndUser(ctx context.Context, accountID, userID string) (*domain.AccountMember, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var memberModel AccountMemberModel
	err := r.db.WithContext(ctx).First(&memberModel, "account_id = ? AND user_id = ?", accountID, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account member not found for account %s and user %s", accountID, userID)
		}
		return nil, fmt.Errorf("failed to get account member by account %s and user %s: %w", accountID, userID, err)
	}

	return r.modelToDomain(&memberModel), nil
}

// ListByAccount retrieves all members of an account
func (r *GormAccountMemberRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.AccountMember, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("account ID cannot be empty")
	}

	var memberModels []AccountMemberModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&memberModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list account members for account %s: %w", accountID, err)
	}

	members := make([]*domain.AccountMember, len(memberModels))
	for i, model := range memberModels {
		members[i] = r.modelToDomain(&model)
	}

	return members, nil
}

// ListByUser retrieves all account memberships for a user
func (r *GormAccountMemberRepository) ListByUser(ctx context.Context, userID string) ([]*domain.AccountMember, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var memberModels []AccountMemberModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&memberModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list account memberships for user %s: %w", userID, err)
	}

	members := make([]*domain.AccountMember, len(memberModels))
	for i, model := range memberModels {
		members[i] = r.modelToDomain(&model)
	}

	return members, nil
}

// modelToDomain converts an AccountMemberModel to a domain.AccountMember
func (r *GormAccountMemberRepository) modelToDomain(model *AccountMemberModel) *domain.AccountMember {
	member := &domain.AccountMember{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		AccountID:   model.AccountID,
		UserID:      model.UserID,
		RoleID:      model.RoleID,
		InvitedBy:   model.InvitedBy,
		JoinedAt:    model.JoinedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return member
}
