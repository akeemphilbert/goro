package infrastructure

import (
	"database/sql"
	"fmt"
	"time"
)

// createSQLiteContainerSchema creates the container and membership tables with indexes for SQLite
func createSQLiteContainerSchema(db *sql.DB) error {
	// Create containers table
	containersSQL := `
	CREATE TABLE IF NOT EXISTS containers (
		id TEXT PRIMARY KEY,
		parent_id TEXT,
		type TEXT NOT NULL DEFAULT 'BasicContainer',
		title TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (parent_id) REFERENCES containers(id)
	);`

	if _, err := db.Exec(containersSQL); err != nil {
		return fmt.Errorf("failed to create containers table: %w", err)
	}

	// Create memberships table
	membershipsSQL := `
	CREATE TABLE IF NOT EXISTS memberships (
		container_id TEXT NOT NULL,
		member_id TEXT NOT NULL,
		member_type TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (container_id, member_id),
		FOREIGN KEY (container_id) REFERENCES containers(id)
	);`

	if _, err := db.Exec(membershipsSQL); err != nil {
		return fmt.Errorf("failed to create memberships table: %w", err)
	}

	// Create indexes for efficient queries
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_containers_parent ON containers(parent_id);",
		"CREATE INDEX IF NOT EXISTS idx_memberships_container ON memberships(container_id);",
		"CREATE INDEX IF NOT EXISTS idx_memberships_member ON memberships(member_id);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// createSQLiteSchemaMigrationsTable creates the schema migrations tracking table for SQLite
func createSQLiteSchemaMigrationsTable(db *sql.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(sql); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// recordMigration records a migration in the schema_migrations table
func recordMigration(db *sql.DB, version int, description string) error {
	sql := `
	INSERT OR IGNORE INTO schema_migrations (version, description, applied_at) 
	VALUES (?, ?, ?);`

	_, err := db.Exec(sql, version, description, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", version, err)
	}

	return nil
}

// getCurrentSchemaVersion returns the current schema version
func getCurrentSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current schema version: %w", err)
	}
	return version, nil
}

// migrateDatabase applies all pending migrations (legacy function for backward compatibility)
func migrateDatabase(db *sql.DB) error {
	provider := &SQLiteSchemaProvider{}
	return MigrateDatabaseWithProvider(db, provider)
}

// createContainerSchema creates the container schema (legacy function for backward compatibility)
func createContainerSchema(db *sql.DB) error {
	return createSQLiteContainerSchema(db)
}

// createSchemaMigrationsTable creates schema migrations table (legacy function for backward compatibility)
func createSchemaMigrationsTable(db *sql.DB) error {
	return createSQLiteSchemaMigrationsTable(db)
}

// validateSchema validates the schema (legacy function for backward compatibility)
func validateSchema(db *sql.DB) error {
	return validateSQLiteSchema(db)
}

// validateSQLiteSchema validates that the SQLite schema is correctly applied
func validateSQLiteSchema(db *sql.DB) error {
	// Check that required tables exist
	requiredTables := []string{"containers", "memberships", "schema_migrations"}

	for _, table := range requiredTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			return fmt.Errorf("required table %s not found: %w", table, err)
		}
	}

	// Check that required indexes exist
	requiredIndexes := []string{
		"idx_containers_parent",
		"idx_memberships_container",
		"idx_memberships_member",
	}

	for _, index := range requiredIndexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
		if err != nil {
			return fmt.Errorf("required index %s not found: %w", index, err)
		}
	}

	return nil
}
