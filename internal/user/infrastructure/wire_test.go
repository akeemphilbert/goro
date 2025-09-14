package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideUserDatabase(t *testing.T) {
	// Create a test database
	db := setupTestDB(t)

	// Test the provider function
	userDB, err := ProvideUserDatabase(db)
	require.NoError(t, err)
	assert.NotNil(t, userDB)

	// Verify that models were migrated
	tables := []string{"user_models", "role_models", "account_models", "account_member_models", "invitation_models"}
	for _, table := range tables {
		var count int64
		err := userDB.Table(table).Count(&count).Error
		assert.NoError(t, err, "Table %s should exist after migration", table)
	}

	// Verify that system roles were seeded
	var roleCount int64
	err = userDB.Model(&RoleModel{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(4), roleCount, "Should have 4 system roles")

	// Test that provider is idempotent
	userDB2, err := ProvideUserDatabase(db)
	require.NoError(t, err)
	assert.NotNil(t, userDB2)

	// Verify no duplicate roles were created
	err = userDB2.Model(&RoleModel{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(4), roleCount, "Should still have exactly 4 system roles")
}
