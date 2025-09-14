package infrastructure

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateContainerSchema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test schema creation
	err := createContainerSchema(db)
	if err != nil {
		t.Fatalf("Failed to create container schema: %v", err)
	}

	// Verify containers table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='containers'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Containers table not found: %v", err)
	}
	if tableName != "containers" {
		t.Errorf("Expected table name 'containers', got '%s'", tableName)
	}

	// Verify memberships table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memberships'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Memberships table not found: %v", err)
	}
	if tableName != "memberships" {
		t.Errorf("Expected table name 'memberships', got '%s'", tableName)
	}
}

func TestCreateContainerSchemaIndexes(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create schema
	err := createContainerSchema(db)
	if err != nil {
		t.Fatalf("Failed to create container schema: %v", err)
	}

	// Verify indexes exist
	expectedIndexes := []string{
		"idx_containers_parent",
		"idx_memberships_container",
		"idx_memberships_member",
	}

	for _, indexName := range expectedIndexes {
		var name string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&name)
		if err != nil {
			t.Errorf("Index %s not found: %v", indexName, err)
		}
	}
}

func TestContainerSchemaConstraints(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create schema
	err := createContainerSchema(db)
	if err != nil {
		t.Fatalf("Failed to create container schema: %v", err)
	}

	// Test containers table constraints
	// Insert valid container
	_, err = db.Exec(`
		INSERT INTO containers (id, type, title, description) 
		VALUES ('container-1', 'BasicContainer', 'Test Container', 'Test Description')
	`)
	if err != nil {
		t.Fatalf("Failed to insert valid container: %v", err)
	}

	// Test primary key constraint
	_, err = db.Exec(`
		INSERT INTO containers (id, type, title, description) 
		VALUES ('container-1', 'BasicContainer', 'Duplicate', 'Duplicate')
	`)
	if err == nil {
		t.Error("Expected primary key constraint violation")
	}

	// Test memberships table constraints
	// Insert valid membership
	_, err = db.Exec(`
		INSERT INTO memberships (container_id, member_id, member_type) 
		VALUES ('container-1', 'resource-1', 'Resource')
	`)
	if err != nil {
		t.Fatalf("Failed to insert valid membership: %v", err)
	}

	// Test primary key constraint on memberships
	_, err = db.Exec(`
		INSERT INTO memberships (container_id, member_id, member_type) 
		VALUES ('container-1', 'resource-1', 'Resource')
	`)
	if err == nil {
		t.Error("Expected primary key constraint violation on memberships")
	}
}

func TestDatabaseMigration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test initial migration
	err := migrateDatabase(db)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Verify migration was applied
	var version int
	err = db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected migration version 1, got %d", version)
	}

	// Test idempotent migration (running again should not fail)
	err = migrateDatabase(db)
	if err != nil {
		t.Errorf("Migration should be idempotent: %v", err)
	}
}

func TestSchemaVersioning(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test schema migrations table creation
	err := createSchemaMigrationsTable(db)
	if err != nil {
		t.Fatalf("Failed to create schema migrations table: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Schema migrations table not found: %v", err)
	}

	// Test recording migration
	err = recordMigration(db, 1, "Initial container schema")
	if err != nil {
		t.Fatalf("Failed to record migration: %v", err)
	}

	// Verify migration was recorded
	var version int
	var description string
	err = db.QueryRow("SELECT version, description FROM schema_migrations WHERE version = 1").Scan(&version, &description)
	if err != nil {
		t.Fatalf("Failed to retrieve migration record: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}
	if description != "Initial container schema" {
		t.Errorf("Expected description 'Initial container schema', got '%s'", description)
	}
}

func TestSchemaValidation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create full schema with migrations
	err := migrateDatabase(db)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Test schema validation
	err = validateSchema(db)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}
}

func TestSchemaUpgrade(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create initial schema
	err := migrateDatabase(db)
	if err != nil {
		t.Fatalf("Failed to create initial schema: %v", err)
	}

	// Test that schema can be upgraded (future migrations)
	currentVersion, err := getCurrentSchemaVersion(db)
	if err != nil {
		t.Fatalf("Failed to get current schema version: %v", err)
	}

	if currentVersion != 1 {
		t.Errorf("Expected current version 1, got %d", currentVersion)
	}
}

// Helper functions for testing
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_schema_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Open database connection
	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}
