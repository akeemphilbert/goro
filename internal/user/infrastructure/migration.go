package infrastructure

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"github.com/akeemphilbert/goro/internal/user/domain"
)

// MigrateUserModels performs auto-migration for all user management models
func MigrateUserModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&UserModel{},
		&RoleModel{},
		&AccountModel{},
		&AccountMemberModel{},
		&InvitationModel{},
	)
}

// SeedSystemRoles creates the predefined system roles with proper permissions
func SeedSystemRoles(db *gorm.DB) error {
	systemRoles := []struct {
		ID          string
		Name        string
		Description string
		Permissions []domain.Permission
	}{
		{
			ID:          "owner",
			Name:        "Owner",
			Description: "Full access to account and all resources",
			Permissions: []domain.Permission{
				{Resource: "*", Action: "*", Scope: "account"},
			},
		},
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Administrative access to account management",
			Permissions: []domain.Permission{
				{Resource: "user", Action: "*", Scope: "account"},
				{Resource: "account", Action: "read", Scope: "account"},
				{Resource: "account", Action: "update", Scope: "account"},
				{Resource: "resource", Action: "*", Scope: "account"},
			},
		},
		{
			ID:          "member",
			Name:        "Member",
			Description: "Standard member access",
			Permissions: []domain.Permission{
				{Resource: "user", Action: "read", Scope: "account"},
				{Resource: "resource", Action: "create", Scope: "own"},
				{Resource: "resource", Action: "read", Scope: "own"},
				{Resource: "resource", Action: "update", Scope: "own"},
			},
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "Read-only access",
			Permissions: []domain.Permission{
				{Resource: "user", Action: "read", Scope: "account"},
				{Resource: "resource", Action: "read", Scope: "account"},
			},
		},
	}

	for _, roleData := range systemRoles {
		// Check if role already exists
		var existingRole RoleModel
		err := db.First(&existingRole, "id = ?", roleData.ID).Error
		if err == nil {
			// Role already exists, skip
			continue
		}
		if err != gorm.ErrRecordNotFound {
			// Some other error occurred
			return err
		}

		// Serialize permissions to JSON
		permissionsJSON, err := json.Marshal(roleData.Permissions)
		if err != nil {
			return err
		}

		// Create the role
		role := RoleModel{
			ID:          roleData.ID,
			Name:        roleData.Name,
			Description: roleData.Description,
			Permissions: string(permissionsJSON),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	return nil
}
