package infrastructure_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
)

func TestGormRoleRepository_GetByID(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormRoleRepository(db)
	ctx := context.Background()

	t.Run("should return system role when found", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "owner")
		require.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "owner", role.ID())
		assert.Equal(t, "Owner", role.Name)
		assert.Equal(t, "Full access to account and all resources", role.Description)
		assert.Len(t, role.Permissions, 1)
		assert.Equal(t, "*", role.Permissions[0].Resource)
		assert.Equal(t, "*", role.Permissions[0].Action)
		assert.Equal(t, "account", role.Permissions[0].Scope)
	})

	t.Run("should return admin role with correct permissions", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "admin")
		require.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "admin", role.ID())
		assert.Equal(t, "Administrator", role.Name)
		assert.Equal(t, "Administrative access to account management", role.Description)
		assert.Len(t, role.Permissions, 4)

		// Check specific permissions
		hasUserPermission := false
		hasAccountReadPermission := false
		hasAccountUpdatePermission := false
		hasResourcePermission := false

		for _, perm := range role.Permissions {
			if perm.Resource == "user" && perm.Action == "*" && perm.Scope == "account" {
				hasUserPermission = true
			}
			if perm.Resource == "account" && perm.Action == "read" && perm.Scope == "account" {
				hasAccountReadPermission = true
			}
			if perm.Resource == "account" && perm.Action == "update" && perm.Scope == "account" {
				hasAccountUpdatePermission = true
			}
			if perm.Resource == "resource" && perm.Action == "*" && perm.Scope == "account" {
				hasResourcePermission = true
			}
		}

		assert.True(t, hasUserPermission)
		assert.True(t, hasAccountReadPermission)
		assert.True(t, hasAccountUpdatePermission)
		assert.True(t, hasResourcePermission)
	})

	t.Run("should return member role with correct permissions", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "member")
		require.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "member", role.ID())
		assert.Equal(t, "Member", role.Name)
		assert.Equal(t, "Standard member access", role.Description)
		assert.Len(t, role.Permissions, 4)

		// Check specific permissions
		hasUserReadPermission := false
		hasResourceCreatePermission := false
		hasResourceReadPermission := false
		hasResourceUpdatePermission := false

		for _, perm := range role.Permissions {
			if perm.Resource == "user" && perm.Action == "read" && perm.Scope == "account" {
				hasUserReadPermission = true
			}
			if perm.Resource == "resource" && perm.Action == "create" && perm.Scope == "own" {
				hasResourceCreatePermission = true
			}
			if perm.Resource == "resource" && perm.Action == "read" && perm.Scope == "own" {
				hasResourceReadPermission = true
			}
			if perm.Resource == "resource" && perm.Action == "update" && perm.Scope == "own" {
				hasResourceUpdatePermission = true
			}
		}

		assert.True(t, hasUserReadPermission)
		assert.True(t, hasResourceCreatePermission)
		assert.True(t, hasResourceReadPermission)
		assert.True(t, hasResourceUpdatePermission)
	})

	t.Run("should return viewer role with correct permissions", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "viewer")
		require.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "viewer", role.ID())
		assert.Equal(t, "Viewer", role.Name)
		assert.Equal(t, "Read-only access", role.Description)
		assert.Len(t, role.Permissions, 2)

		// Check specific permissions
		hasUserReadPermission := false
		hasResourceReadPermission := false

		for _, perm := range role.Permissions {
			if perm.Resource == "user" && perm.Action == "read" && perm.Scope == "account" {
				hasUserReadPermission = true
			}
			if perm.Resource == "resource" && perm.Action == "read" && perm.Scope == "account" {
				hasResourceReadPermission = true
			}
		}

		assert.True(t, hasUserReadPermission)
		assert.True(t, hasResourceReadPermission)
	})

	t.Run("should return error when role not found", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "non-existent-role")
		assert.Error(t, err)
		assert.Nil(t, role)
	})

	t.Run("should return error for empty ID", func(t *testing.T) {
		role, err := repo.GetByID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, role)
	})
}

func TestGormRoleRepository_List(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormRoleRepository(db)
	ctx := context.Background()

	t.Run("should return all roles", func(t *testing.T) {
		roles, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, roles, 4)

		// Check that all system roles are present
		roleIDs := make(map[string]bool)
		for _, role := range roles {
			roleIDs[role.ID()] = true
		}

		assert.True(t, roleIDs["owner"])
		assert.True(t, roleIDs["admin"])
		assert.True(t, roleIDs["member"])
		assert.True(t, roleIDs["viewer"])
	})

	t.Run("should return roles with correct structure", func(t *testing.T) {
		roles, err := repo.List(ctx)
		require.NoError(t, err)

		for _, role := range roles {
			assert.NotEmpty(t, role.ID())
			assert.NotEmpty(t, role.Name)
			assert.NotEmpty(t, role.Description)
			assert.NotEmpty(t, role.Permissions)

			// Validate permissions structure
			for _, perm := range role.Permissions {
				assert.NotEmpty(t, perm.Resource)
				assert.NotEmpty(t, perm.Action)
				assert.NotEmpty(t, perm.Scope)
			}
		}
	})
}

func TestGormRoleRepository_GetSystemRoles(t *testing.T) {
	db := setupTestDBWithMigration(t)
	repo := infrastructure.NewGormRoleRepository(db)
	ctx := context.Background()

	t.Run("should return all system roles", func(t *testing.T) {
		roles, err := repo.GetSystemRoles(ctx)
		require.NoError(t, err)
		assert.Len(t, roles, 4)

		// Check that all system roles are present
		roleIDs := make(map[string]bool)
		for _, role := range roles {
			roleIDs[role.ID()] = true
		}

		assert.True(t, roleIDs["owner"])
		assert.True(t, roleIDs["admin"])
		assert.True(t, roleIDs["member"])
		assert.True(t, roleIDs["viewer"])
	})

	t.Run("should return roles in consistent order", func(t *testing.T) {
		roles1, err := repo.GetSystemRoles(ctx)
		require.NoError(t, err)

		roles2, err := repo.GetSystemRoles(ctx)
		require.NoError(t, err)

		assert.Len(t, roles1, len(roles2))

		// Create maps for comparison
		roles1Map := make(map[string]*domain.Role)
		roles2Map := make(map[string]*domain.Role)

		for _, role := range roles1 {
			roles1Map[role.ID()] = role
		}
		for _, role := range roles2 {
			roles2Map[role.ID()] = role
		}

		// Compare each role
		for id, role1 := range roles1Map {
			role2, exists := roles2Map[id]
			require.True(t, exists)
			assert.Equal(t, role1.ID(), role2.ID())
			assert.Equal(t, role1.Name, role2.Name)
			assert.Equal(t, role1.Description, role2.Description)
			assert.Len(t, role1.Permissions, len(role2.Permissions))
		}
	})

	t.Run("should validate role permissions", func(t *testing.T) {
		roles, err := repo.GetSystemRoles(ctx)
		require.NoError(t, err)

		for _, role := range roles {
			for _, perm := range role.Permissions {
				err := perm.Validate()
				assert.NoError(t, err, "Permission validation failed for role %s: %+v", role.ID(), perm)
			}
		}
	})
}
