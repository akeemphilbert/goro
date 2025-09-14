package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// GormAccountMemberWriteRepository implements domain.AccountMemberWriteRepository using GORM
type GormAccountMemberWriteRepository struct {
	db *gorm.DB
}

// NewGormAccountMemberWriteRepository creates a new GORM-based account member write repository
func NewGormAccountMemberWriteRepository(db *gorm.DB) domain.AccountMemberWriteRepository {
	return &GormAccountMemberWriteRepository{db: db}
}

// Create creates a new account member in the database
func (r *GormAccountMemberWriteRepository) Create(ctx context.Context, member *domain.AccountMember) error {
	if member == nil {
		return fmt.Errorf("account member cannot be nil")
	}

	if strings.TrimSpace(member.ID()) == "" {
		return fmt.Errorf("member ID cannot be empty")
	}

	if strings.TrimSpace(member.AccountID) == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if strings.TrimSpace(member.UserID) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if strings.TrimSpace(member.RoleID) == "" {
		return fmt.Errorf("role ID cannot be empty")
	}

	// Convert domain member to GORM model
	memberModel := &AccountMemberModel{
		ID:        member.ID(),
		AccountID: member.AccountID,
		UserID:    member.UserID,
		RoleID:    member.RoleID,
		InvitedBy: member.InvitedBy,
		JoinedAt:  member.JoinedAt,
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}

	err := r.db.WithContext(ctx).Create(memberModel).Error
	if err != nil {
		return fmt.Errorf("failed to create account member: %w", err)
	}

	return nil
}

// Update updates an existing account member in the database
func (r *GormAccountMemberWriteRepository) Update(ctx context.Context, member *domain.AccountMember) error {
	if member == nil {
		return fmt.Errorf("account member cannot be nil")
	}

	if strings.TrimSpace(member.ID()) == "" {
		return fmt.Errorf("member ID cannot be empty")
	}

	// Convert domain member to GORM model
	memberModel := &AccountMemberModel{
		ID:        member.ID(),
		AccountID: member.AccountID,
		UserID:    member.UserID,
		RoleID:    member.RoleID,
		InvitedBy: member.InvitedBy,
		JoinedAt:  member.JoinedAt,
		CreatedAt: member.CreatedAt,
		UpdatedAt: member.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&AccountMemberModel{}).Where("id = ?", member.ID()).Updates(memberModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update account member: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account member not found: %s", member.ID())
	}

	return nil
}

// Delete deletes an account member from the database
func (r *GormAccountMemberWriteRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("member ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&AccountMemberModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete account member: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account member not found: %s", id)
	}

	return nil
}
