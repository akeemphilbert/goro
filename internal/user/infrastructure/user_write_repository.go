package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// GormUserWriteRepository implements domain.UserWriteRepository using GORM
type GormUserWriteRepository struct {
	db *gorm.DB
}

// NewGormUserWriteRepository creates a new GORM-based user write repository
func NewGormUserWriteRepository(db *gorm.DB) domain.UserWriteRepository {
	return &GormUserWriteRepository{db: db}
}

// Create creates a new user in the database
func (r *GormUserWriteRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if strings.TrimSpace(user.ID()) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Convert domain user to GORM model
	userModel := &UserModel{
		ID:        user.ID(),
		WebID:     user.WebID,
		Email:     user.Email,
		Name:      user.Profile.Name,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	err := r.db.WithContext(ctx).Create(userModel).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates an existing user in the database
func (r *GormUserWriteRepository) Update(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if strings.TrimSpace(user.ID()) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Convert domain user to GORM model
	userModel := &UserModel{
		ID:        user.ID(),
		WebID:     user.WebID,
		Email:     user.Email,
		Name:      user.Profile.Name,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	result := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", user.ID()).Updates(userModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID())
	}

	return nil
}

// Delete deletes a user from the database
func (r *GormUserWriteRepository) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}
