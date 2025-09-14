package infrastructure

import (
	"context"
	"fmt"
	"strings"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// GormUserRepository implements domain.UserRepository using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM-based user repository
func NewGormUserRepository(db *gorm.DB) domain.UserRepository {
	return &GormUserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var userModel UserModel
	err := r.db.WithContext(ctx).First(&userModel, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user by ID %s: %w", id, err)
	}

	return r.modelToDomain(&userModel)
}

// GetByWebID retrieves a user by WebID
func (r *GormUserRepository) GetByWebID(ctx context.Context, webid string) (*domain.User, error) {
	if strings.TrimSpace(webid) == "" {
		return nil, fmt.Errorf("WebID cannot be empty")
	}

	var userModel UserModel
	err := r.db.WithContext(ctx).First(&userModel, "web_id = ?", webid).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with WebID: %s", webid)
		}
		return nil, fmt.Errorf("failed to get user by WebID %s: %w", webid, err)
	}

	return r.modelToDomain(&userModel)
}

// GetByEmail retrieves a user by email
func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var userModel UserModel
	err := r.db.WithContext(ctx).First(&userModel, "email = ?", email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	return r.modelToDomain(&userModel)
}

// List retrieves users with filtering
func (r *GormUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&UserModel{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	if filter.EmailPattern != "" {
		query = query.Where("email LIKE ?", "%"+filter.EmailPattern+"%")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var userModels []UserModel
	err := query.Find(&userModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*domain.User, len(userModels))
	for i, model := range userModels {
		user, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user model to domain: %w", err)
		}
		users[i] = user
	}

	return users, nil
}

// Exists checks if a user exists by ID
func (r *GormUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	if strings.TrimSpace(id) == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user existence for ID %s: %w", id, err)
	}

	return count > 0, nil
}

// modelToDomain converts a UserModel to a domain.User
func (r *GormUserRepository) modelToDomain(model *UserModel) (*domain.User, error) {
	// Create user profile from name (simplified for now)
	profile := domain.UserProfile{
		Name:        model.Name,
		Bio:         "",
		Avatar:      "",
		Preferences: make(map[string]interface{}),
	}

	user := &domain.User{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		WebID:       model.WebID,
		Email:       model.Email,
		Profile:     profile,
		Status:      domain.UserStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return user, nil
}
