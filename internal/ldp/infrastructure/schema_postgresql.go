package infrastructure

import (
	"database/sql"
	"fmt"
)

// createPostgreSQLContainerSchema creates the container and membership tables for PostgreSQL
func createPostgreSQLContainerSchema(db *sql.DB) error {
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

// createPostgreSQLSchemaMigrationsTable creates the schema migrations tracking table for PostgreSQL
func createPostgreSQLSchemaMigrationsTable(db *sql.DB) error {
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

// validatePostgreSQLSchema validates that the PostgreSQL schema is correctly applied
func validatePostgreSQLSchema(db *sql.DB) error {
	// Check that required tables exist
	requiredTables := []string{"containers", "memberships", "schema_migrations"}

	for _, table := range requiredTables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("required table %s not found", table)
		}
	}

	// Check that required indexes exist
	requiredIndexes := []string{
		"idx_containers_parent",
		"idx_memberships_container",
		"idx_memberships_member",
	}

	for _, index := range requiredIndexes {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_indexes 
				WHERE schemaname = 'public' 
				AND indexname = $1
			)`, index).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", index, err)
		}
		if !exists {
			return fmt.Errorf("required index %s not found", index)
		}
	}

	return nil
}
