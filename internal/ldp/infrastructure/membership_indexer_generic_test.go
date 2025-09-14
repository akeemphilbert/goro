package infrastructure

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestGenericMembershipIndexer_SQLite(t *testing.T) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_generic_sqlite_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := DefaultSQLiteConfig(tmpFile.Name())
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite indexer: %v", err)
	}
	defer indexer.Close()

	testGenericMembershipIndexer(t, indexer)
}

func TestGenericMembershipIndexer_PostgreSQL(t *testing.T) {
	// Skip if PostgreSQL is not available
	if !isPostgreSQLAvailable() {
		t.Skip("PostgreSQL not available, skipping test")
	}

	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		t.Skipf("Failed to create PostgreSQL indexer (database may not be available): %v", err)
	}
	defer indexer.Close()

	testGenericMembershipIndexer(t, indexer)
}

func testGenericMembershipIndexer(t *testing.T, indexer *GenericMembershipIndexer) {
	ctx := context.Background()
	containerID := "test-container"
	memberID := "test-resource"

	// Create test container first
	err := createTestContainerGeneric(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Test IndexMembership
	err = indexer.IndexMembership(ctx, containerID, memberID)
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}

	// Test GetMembers
	members, err := indexer.GetMembers(ctx, containerID, PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("Expected 1 member, got %d", len(members))
	}

	if members[0].ID != memberID {
		t.Errorf("Expected member ID %s, got %s", memberID, members[0].ID)
	}

	// Test GetContainers
	containers, err := indexer.GetContainers(ctx, memberID)
	if err != nil {
		t.Fatalf("Failed to get containers: %v", err)
	}

	if len(containers) != 1 {
		t.Fatalf("Expected 1 container, got %d", len(containers))
	}

	if containers[0] != containerID {
		t.Errorf("Expected container ID %s, got %s", containerID, containers[0])
	}

	// Test RemoveMembership
	err = indexer.RemoveMembership(ctx, containerID, memberID)
	if err != nil {
		t.Fatalf("Failed to remove membership: %v", err)
	}

	// Verify removal
	members, err = indexer.GetMembers(ctx, containerID, PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members after removal: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("Expected 0 members after removal, got %d", len(members))
	}
}

func TestGenericMembershipIndexer_Pagination_SQLite(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_pagination_sqlite_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := DefaultSQLiteConfig(tmpFile.Name())
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite indexer: %v", err)
	}
	defer indexer.Close()

	testGenericMembershipIndexerPagination(t, indexer)
}

func TestGenericMembershipIndexer_Pagination_PostgreSQL(t *testing.T) {
	if !isPostgreSQLAvailable() {
		t.Skip("PostgreSQL not available, skipping test")
	}

	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		t.Skipf("Failed to create PostgreSQL indexer: %v", err)
	}
	defer indexer.Close()

	testGenericMembershipIndexerPagination(t, indexer)
}

func testGenericMembershipIndexerPagination(t *testing.T, indexer *GenericMembershipIndexer) {
	ctx := context.Background()
	containerID := "pagination-container"

	// Create test container
	err := createTestContainerGeneric(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Add multiple members
	memberCount := 10
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("resource-%d", i)
		err := indexer.IndexMembership(ctx, containerID, memberID)
		if err != nil {
			t.Fatalf("Failed to index membership %d: %v", i, err)
		}
	}

	// Test pagination
	pageSize := 3
	members, err := indexer.GetMembers(ctx, containerID, PaginationOptions{
		Limit:  pageSize,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get paginated members: %v", err)
	}

	if len(members) != pageSize {
		t.Errorf("Expected %d members in first page, got %d", pageSize, len(members))
	}

	// Test second page
	members, err = indexer.GetMembers(ctx, containerID, PaginationOptions{
		Limit:  pageSize,
		Offset: pageSize,
	})
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}

	if len(members) != pageSize {
		t.Errorf("Expected %d members in second page, got %d", pageSize, len(members))
	}

	// Test member count
	count, err := indexer.GetMemberCount(ctx, containerID)
	if err != nil {
		t.Fatalf("Failed to get member count: %v", err)
	}

	if count != memberCount {
		t.Errorf("Expected member count %d, got %d", memberCount, count)
	}
}

func TestDatabaseProvider_SQLite(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_provider_sqlite_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := DefaultSQLiteConfig(tmpFile.Name())
	db, err := NewDatabaseConnection(config)
	if err != nil {
		t.Fatalf("Failed to create SQLite connection: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping SQLite database: %v", err)
	}
}

func TestDatabaseProvider_PostgreSQL(t *testing.T) {
	if !isPostgreSQLAvailable() {
		t.Skip("PostgreSQL not available, skipping test")
	}

	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")
	db, err := NewDatabaseConnection(config)
	if err != nil {
		t.Skipf("Failed to create PostgreSQL connection: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}
}

func TestSchemaProvider_SQLite(t *testing.T) {
	provider, err := NewSchemaProvider("sqlite3")
	if err != nil {
		t.Fatalf("Failed to create SQLite schema provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Schema provider should not be nil")
	}

	// Test with actual database
	tmpFile, err := os.CreateTemp("", "test_schema_sqlite_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := DefaultSQLiteConfig(tmpFile.Name())
	db, err := NewDatabaseConnection(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test migration
	err = MigrateDatabaseWithProvider(db, provider)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Test validation
	err = provider.ValidateSchema(db)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}
}

func TestSchemaProvider_PostgreSQL(t *testing.T) {
	if !isPostgreSQLAvailable() {
		t.Skip("PostgreSQL not available, skipping test")
	}

	provider, err := NewSchemaProvider("postgres")
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL schema provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Schema provider should not be nil")
	}

	// Test with actual database
	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")
	db, err := NewDatabaseConnection(config)
	if err != nil {
		t.Skipf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test migration
	err = MigrateDatabaseWithProvider(db, provider)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Test validation
	err = provider.ValidateSchema(db)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}
}

// Helper functions
func createTestContainerGeneric(indexer *GenericMembershipIndexer, containerID string) error {
	query := `INSERT OR IGNORE INTO containers (id, type, title, description) VALUES (?, 'BasicContainer', 'Test Container', 'Test Description')`

	// Adjust query for PostgreSQL
	if indexer.driver == "postgres" || indexer.driver == "postgresql" {
		query = `INSERT INTO containers (id, type, title, description) VALUES ($1, 'BasicContainer', 'Test Container', 'Test Description') ON CONFLICT (id) DO NOTHING`
	}

	_, err := indexer.db.Exec(query, containerID)
	return err
}

func isPostgreSQLAvailable() bool {
	// Check if PostgreSQL environment variables are set or if we can connect
	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")
	db, err := NewDatabaseConnection(config)
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	return err == nil
}
