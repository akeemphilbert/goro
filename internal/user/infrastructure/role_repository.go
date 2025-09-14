package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// GormRoleRepository implements domain.RoleRepository using GORM
type GormRoleRepository struct {
	db *gorm.DB
}

// NewGormRoleRepository creates a new GORM-based role repository
func NewGormRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &GormRoleRepository{db: db}
}

// GetByID retrieves a role by ID
func (r *GormRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("role ID cannot be empty")
	}

	var roleModel RoleModel
	err := r.db.WithContext(ctx).First(&roleModel, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("role not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get role by ID %s: %w", id, err)
	}

	return r.modelToDomain(&roleModel)
}

// List retrieves all roles
func (r *GormRoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	var roleModels []RoleModel
	err := r.db.WithContext(ctx).Find(&roleModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roles := make([]*domain.Role, len(roleModels))
	for i, model := range roleModels {
		role, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert role model to domain: %w", err)
		}
		roles[i] = role
	}

	return roles, nil
}

// GetSystemRoles retrieves all system roles (predefined roles)
func (r *GormRoleRepository) GetSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	systemRoleIDs := []string{"owner", "admin", "member", "viewer"}

	var roleModels []RoleModel
	err := r.db.WithContext(ctx).Where("id IN ?", systemRoleIDs).Find(&roleModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}

	roles := make([]*domain.Role, len(roleModels))
	for i, model := range roleModels {
		role, err := r.modelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert role model to domain: %w", err)
		}
		roles[i] = role
	}

	return roles, nil
}

// modelToDomain converts a RoleModel to a domain.Role
func (r *GormRoleRepository) modelToDomain(model *RoleModel) (*domain.Role, error) {
	// Deserialize permissions from JSON
	var permissions []domain.Permission
	if model.Permissions != "" {
		err := json.Unmarshal([]byte(model.Permissions), &permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize role permissions: %w", err)
		}
	}

	role := &domain.Role{
		BasicEntity: pericarpdomain.NewEntity(model.ID),
		Name:        model.Name,
		Description: model.Description,
		Permissions: permissions,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	return role, nil
}
