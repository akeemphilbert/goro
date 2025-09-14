package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a temporary in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

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
