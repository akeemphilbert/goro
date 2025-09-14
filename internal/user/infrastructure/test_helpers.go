package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDBWithMigration creates a temporary in-memory SQLite database with migrations for testing
func setupTestDBWithMigration(t *testing.T) *gorm.DB {
	db := setupTestDB(t)

	// Run migrations
	err := MigrateUserModels(db)
	require.NoError(t, err)

	// Seed system roles
	err = SeedSystemRoles(db)
	require.NoError(t, err)

	return db
}
