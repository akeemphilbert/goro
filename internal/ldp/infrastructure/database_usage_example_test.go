package infrastructure

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Example: Using SQLite for development
func ExampleSQLiteUsage() {
	// Create temporary database for example
	tmpFile, err := os.CreateTemp("", "example_sqlite_*.db")
	if err != nil {
		panic(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Configure SQLite
	config := DefaultSQLiteConfig(tmpFile.Name())

	// Create indexer
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		panic(err)
	}
	defer indexer.Close()

	// Use the indexer
	ctx := context.Background()

	// Create a container (in real usage, this would be done by the container service)
	err = createTestContainerGeneric(indexer, "documents")
	if err != nil {
		panic(err)
	}

	// Add a member to the container
	err = indexer.IndexMembership(ctx, "documents", "document1.pdf")
	if err != nil {
		panic(err)
	}

	// Get container members
	members, err := indexer.GetMembers(ctx, "documents", PaginationOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Container 'documents' has %d members\n", len(members))
	// Output: Container 'documents' has 1 members
}

// Example: Using PostgreSQL for production (requires PostgreSQL server)
func ExamplePostgreSQLUsage() {
	// Skip if PostgreSQL is not available
	if !isPostgreSQLAvailable() {
		fmt.Println("PostgreSQL not available, skipping example")
		return
	}

	// Configure PostgreSQL
	config := DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")

	// Create indexer
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		fmt.Printf("PostgreSQL not available: %v\n", err)
		return
	}
	defer indexer.Close()

	// Use the indexer (same API as SQLite)
	ctx := context.Background()

	// Create a container
	err = createTestContainerGeneric(indexer, "images")
	if err != nil {
		panic(err)
	}

	// Add multiple members
	members := []string{"photo1.jpg", "photo2.png", "photo3.gif"}
	for _, member := range members {
		err = indexer.IndexMembership(ctx, "images", member)
		if err != nil {
			panic(err)
		}
	}

	// Get container statistics
	stats, err := indexer.GetContainerStats(ctx, "images")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Container 'images' statistics: %+v\n", stats)
}

// Example: Environment-based configuration
func ExampleEnvironmentConfiguration() {
	// Set environment variables (in real usage, these would be set externally)
	os.Setenv("DB_DRIVER", "sqlite3")
	os.Setenv("DB_PATH", ":memory:")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_PATH")
	}()

	// Get configuration from environment
	config := DatabaseConfigFromEnv()

	// Create indexer
	indexer, err := NewGenericMembershipIndexer(config)
	if err != nil {
		panic(err)
	}
	defer indexer.Close()

	fmt.Printf("Using database driver: %s\n", config.Driver)
	// Output: Using database driver: sqlite3
}

// Test demonstrating database switching
func TestDatabaseSwitching(t *testing.T) {
	testCases := []struct {
		name   string
		config DatabaseConfig
	}{
		{
			name:   "SQLite",
			config: DefaultSQLiteConfig(":memory:"),
		},
	}

	// Add PostgreSQL test case if available
	if isPostgreSQLAvailable() {
		testCases = append(testCases, struct {
			name   string
			config DatabaseConfig
		}{
			name:   "PostgreSQL",
			config: DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password"),
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indexer, err := NewGenericMembershipIndexer(tc.config)
			if err != nil {
				t.Skipf("Failed to create %s indexer: %v", tc.name, err)
			}
			defer indexer.Close()

			// Test basic functionality
			ctx := context.Background()
			containerID := "test-container"

			// Create container
			err = createTestContainerGeneric(indexer, containerID)
			if err != nil {
				t.Fatalf("Failed to create container: %v", err)
			}

			// Add member
			err = indexer.IndexMembership(ctx, containerID, "test-resource")
			if err != nil {
				t.Fatalf("Failed to add member: %v", err)
			}

			// Verify member
			members, err := indexer.GetMembers(ctx, containerID, PaginationOptions{})
			if err != nil {
				t.Fatalf("Failed to get members: %v", err)
			}

			if len(members) != 1 {
				t.Errorf("Expected 1 member, got %d", len(members))
			}

			t.Logf("%s database test completed successfully", tc.name)
		})
	}
}

// Test demonstrating migration compatibility
func TestMigrationCompatibility(t *testing.T) {
	// Test that both databases create the same logical schema
	databases := []struct {
		name   string
		config DatabaseConfig
	}{
		{"SQLite", DefaultSQLiteConfig(":memory:")},
	}

	if isPostgreSQLAvailable() {
		databases = append(databases, struct {
			name   string
			config DatabaseConfig
		}{"PostgreSQL", DefaultPostgresConfig("localhost", "test_container_db", "test_user", "test_password")})
	}

	for _, db := range databases {
		t.Run(db.name, func(t *testing.T) {
			// Create connection
			conn, err := NewDatabaseConnection(db.config)
			if err != nil {
				t.Skipf("Failed to connect to %s: %v", db.name, err)
			}
			defer conn.Close()

			// Create schema provider
			provider, err := NewSchemaProvider(db.config.Driver)
			if err != nil {
				t.Fatalf("Failed to create schema provider: %v", err)
			}

			// Apply migrations
			err = MigrateDatabaseWithProvider(conn, provider)
			if err != nil {
				t.Fatalf("Failed to apply migrations: %v", err)
			}

			// Validate schema
			err = provider.ValidateSchema(conn)
			if err != nil {
				t.Fatalf("Schema validation failed: %v", err)
			}

			// Check migration version
			version, err := provider.GetCurrentSchemaVersion(conn)
			if err != nil {
				t.Fatalf("Failed to get schema version: %v", err)
			}

			if version != 1 {
				t.Errorf("Expected schema version 1, got %d", version)
			}

			t.Logf("%s migration test completed successfully", db.name)
		})
	}
}
