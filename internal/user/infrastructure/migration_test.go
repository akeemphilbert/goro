package infrastructure_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/akeemphilbert/goro/internal/user/infrastructure"
)

func TestMigrateUserModels(t *testing.T) {
	db := setupTestDB(t)

	// Test that MigrateUserModels function exists and works
	err := infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	// Verify that all tables were created
	tables := []string{"user_models", "role_models", "account_models", "account_member_models", "invitation_models"}

	for _, table := range tables {
		var count int64
		err := db.Table(table).Count(&count).Error
		assert.NoError(t, err, "Table %s should exist after migration", table)
	}

	// Test that migration is idempotent (can be run multiple times)
	err = infrastructure.MigrateUserModels(db)
	assert.NoError(t, err, "Migration should be idempotent")
}

func TestSeedSystemRoles(t *testing.T) {
	db := setupTestDB(t)

	// Migrate models first
	err := infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	// Test that SeedSystemRoles function exists and works
	err = infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Verify that system roles were created
	expectedRoles := []struct {
		ID          string
		Name        string
		Description string
	}{
		{"owner", "Owner", "Full access to account and all resources"},
		{"admin", "Administrator", "Administrative access to account management"},
		{"member", "Member", "Standard member access"},
		{"viewer", "Viewer", "Read-only access"},
	}

	for _, expectedRole := range expectedRoles {
		var role infrastructure.RoleModel
		err := db.First(&role, "id = ?", expectedRole.ID).Error
		require.NoError(t, err, "System role %s should exist", expectedRole.ID)
		assert.Equal(t, expectedRole.Name, role.Name)
		assert.Equal(t, expectedRole.Description, role.Description)
		assert.NotEmpty(t, role.Permissions, "Role should have permissions")
	}

	// Test that seeding is idempotent (can be run multiple times)
	err = infrastructure.SeedSystemRoles(db)
	assert.NoError(t, err, "Seeding should be idempotent")

	// Verify no duplicate roles were created
	var roleCount int64
	err = db.Model(&infrastructure.RoleModel{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(4), roleCount, "Should have exactly 4 system roles")
}

func TestSystemRolePermissions(t *testing.T) {
	db := setupTestDB(t)

	// Migrate and seed
	err := infrastructure.MigrateUserModels(db)
	require.NoError(t, err)
	err = infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Test Owner role permissions
	var ownerRole infrastructure.RoleModel
	err = db.First(&ownerRole, "id = ?", "owner").Error
	require.NoError(t, err)

	ownerPermissions := parsePermissions(t, ownerRole.Permissions)
	assert.Contains(t, ownerPermissions, domain.Permission{Resource: "*", Action: "*", Scope: "account"})

	// Test Admin role permissions
	var adminRole infrastructure.RoleModel
	err = db.First(&adminRole, "id = ?", "admin").Error
	require.NoError(t, err)

	adminPermissions := parsePermissions(t, adminRole.Permissions)
	expectedAdminPerms := []domain.Permission{
		{Resource: "user", Action: "*", Scope: "account"},
		{Resource: "account", Action: "read", Scope: "account"},
		{Resource: "account", Action: "update", Scope: "account"},
		{Resource: "resource", Action: "*", Scope: "account"},
	}
	for _, perm := range expectedAdminPerms {
		assert.Contains(t, adminPermissions, perm)
	}

	// Test Member role permissions
	var memberRole infrastructure.RoleModel
	err = db.First(&memberRole, "id = ?", "member").Error
	require.NoError(t, err)

	memberPermissions := parsePermissions(t, memberRole.Permissions)
	expectedMemberPerms := []domain.Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "create", Scope: "own"},
		{Resource: "resource", Action: "read", Scope: "own"},
		{Resource: "resource", Action: "update", Scope: "own"},
	}
	for _, perm := range expectedMemberPerms {
		assert.Contains(t, memberPermissions, perm)
	}

	// Test Viewer role permissions
	var viewerRole infrastructure.RoleModel
	err = db.First(&viewerRole, "id = ?", "viewer").Error
	require.NoError(t, err)

	viewerPermissions := parsePermissions(t, viewerRole.Permissions)
	expectedViewerPerms := []domain.Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "read", Scope: "account"},
	}
	for _, perm := range expectedViewerPerms {
		assert.Contains(t, viewerPermissions, perm)
	}
}

func TestWireProviderIntegration(t *testing.T) {
	// Test that Wire provider can be created with existing GORM instance
	db := setupTestDB(t)

	// Test that we can create a provider function that accepts a GORM DB
	provider := func(db *gorm.DB) error {
		// Ensure models are migrated
		if err := infrastructure.MigrateUserModels(db); err != nil {
			return err
		}

		// Seed system roles
		if err := infrastructure.SeedSystemRoles(db); err != nil {
			return err
		}

		// This would return an actual repository implementation
		// For now, we just test that the provider pattern works
		return nil
	}

	// Test that provider can be called with database
	err := provider(db)
	assert.NoError(t, err, "Wire provider should work with existing GORM instance")
}

func TestDatabaseIntegrationWithExistingSetup(t *testing.T) {
	// Test integration with existing database provider from ldp infrastructure
	// This simulates how the user management system would integrate with existing setup

	// Create a database similar to how it's done in ldp/infrastructure
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Test that user models can be migrated alongside existing models
	// First migrate user models
	err = infrastructure.MigrateUserModels(db)
	require.NoError(t, err)

	// Then seed system roles
	err = infrastructure.SeedSystemRoles(db)
	require.NoError(t, err)

	// Verify everything works together
	var roleCount int64
	err = db.Model(&infrastructure.RoleModel{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(4), roleCount)

	// Test that we can create users, accounts, etc.
	user := &infrastructure.UserModel{
		ID:     "test-user",
		WebID:  "https://example.com/user/test",
		Email:  "test@example.com",
		Name:   "Test User",
		Status: string(domain.UserStatusActive),
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	account := &infrastructure.AccountModel{
		ID:       "test-account",
		OwnerID:  "test-user",
		Name:     "Test Account",
		Settings: `{"allow_invitations":true,"default_role_id":"member","max_members":100}`,
	}
	err = db.Create(account).Error
	require.NoError(t, err)
}

// Helper function to parse permissions JSON for testing
func parsePermissions(t *testing.T, permissionsJSON string) []domain.Permission {
	var permissions []domain.Permission
	err := json.Unmarshal([]byte(permissionsJSON), &permissions)
	require.NoError(t, err)
	return permissions
}
