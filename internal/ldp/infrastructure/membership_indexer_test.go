package infrastructure

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteMembershipIndexer_IndexMembership(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()
	containerID := "container-1"
	memberID := "resource-1"

	// Create container first
	err := createTestContainer(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Test indexing membership
	err = indexer.IndexMembership(ctx, containerID, memberID)
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}

	// Verify membership was indexed
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
}

func TestSQLiteMembershipIndexer_RemoveMembership(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()
	containerID := "container-1"
	memberID := "resource-1"

	// Create container first
	err := createTestContainer(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Index membership first
	err = indexer.IndexMembership(ctx, containerID, memberID)
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}

	// Remove membership
	err = indexer.RemoveMembership(ctx, containerID, memberID)
	if err != nil {
		t.Fatalf("Failed to remove membership: %v", err)
	}

	// Verify membership was removed
	members, err := indexer.GetMembers(ctx, containerID, PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("Expected 0 members after removal, got %d", len(members))
	}
}

func TestSQLiteMembershipIndexer_GetMembers(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()
	containerID := "container-1"

	// Create containers first
	err := createTestContainer(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}
	err = createTestContainer(indexer, "container-2")
	if err != nil {
		t.Fatalf("Failed to create test container-2: %v", err)
	}

	// Index multiple members
	members := []string{"resource-1", "resource-2", "container-2"}
	for _, memberID := range members {
		err := indexer.IndexMembership(ctx, containerID, memberID)
		if err != nil {
			t.Fatalf("Failed to index membership for %s: %v", memberID, err)
		}
	}

	// Test getting all members
	result, err := indexer.GetMembers(ctx, containerID, PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 members, got %d", len(result))
	}

	// Test pagination
	paginatedResult, err := indexer.GetMembers(ctx, containerID, PaginationOptions{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to get paginated members: %v", err)
	}

	if len(paginatedResult) != 2 {
		t.Errorf("Expected 2 members with pagination, got %d", len(paginatedResult))
	}
}

func TestSQLiteMembershipIndexer_GetContainers(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()
	memberID := "resource-1"

	// Create containers first
	containers := []string{"container-1", "container-2"}
	for _, containerID := range containers {
		err := createTestContainer(indexer, containerID)
		if err != nil {
			t.Fatalf("Failed to create test container %s: %v", containerID, err)
		}
	}

	// Index member in multiple containers
	for _, containerID := range containers {
		err := indexer.IndexMembership(ctx, containerID, memberID)
		if err != nil {
			t.Fatalf("Failed to index membership in %s: %v", containerID, err)
		}
	}

	// Get containers for member
	result, err := indexer.GetContainers(ctx, memberID)
	if err != nil {
		t.Fatalf("Failed to get containers: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 containers, got %d", len(result))
	}

	// Verify container IDs
	containerMap := make(map[string]bool)
	for _, containerID := range result {
		containerMap[containerID] = true
	}

	for _, expectedID := range containers {
		if !containerMap[expectedID] {
			t.Errorf("Expected container %s not found in results", expectedID)
		}
	}
}

func TestSQLiteMembershipIndexer_RebuildIndex(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()

	// Create container first
	err := createTestContainer(indexer, "container-1")
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Index some memberships
	err = indexer.IndexMembership(ctx, "container-1", "resource-1")
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}

	// Rebuild index
	err = indexer.RebuildIndex(ctx)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Verify index is still functional (should be empty after rebuild)
	members, err := indexer.GetMembers(ctx, "container-1", PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members after rebuild: %v", err)
	}

	if len(members) != 0 {
		t.Errorf("Expected 0 members after rebuild (index cleared), got %d", len(members))
	}

	// Test that we can still add memberships after rebuild
	err = indexer.IndexMembership(ctx, "container-1", "resource-2")
	if err != nil {
		t.Fatalf("Failed to index membership after rebuild: %v", err)
	}

	members, err = indexer.GetMembers(ctx, "container-1", PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members after re-indexing: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 member after re-indexing, got %d", len(members))
	}
}

func TestSQLiteMembershipIndexer_ConcurrentOperations(t *testing.T) {
	indexer, cleanup := setupTestIndexer(t)
	defer cleanup()

	ctx := context.Background()
	containerID := "container-1"

	// Create container first
	err := createTestContainer(indexer, containerID)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Test concurrent indexing
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			memberID := fmt.Sprintf("resource-%d", id)
			err := indexer.IndexMembership(ctx, containerID, memberID)
			if err != nil {
				t.Errorf("Failed to index membership %s: %v", memberID, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all memberships were indexed
	members, err := indexer.GetMembers(ctx, containerID, PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) != 10 {
		t.Errorf("Expected 10 members after concurrent operations, got %d", len(members))
	}
}

// Helper functions for testing
func setupTestIndexer(t *testing.T) (*SQLiteMembershipIndexer, func()) {
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test_membership_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// Create indexer
	indexer, err := NewSQLiteMembershipIndexer(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create indexer: %v", err)
	}

	cleanup := func() {
		indexer.Close()
		os.Remove(tmpFile.Name())
	}

	return indexer, cleanup
}

// createTestContainer creates a test container in the database
func createTestContainer(indexer *SQLiteMembershipIndexer, containerID string) error {
	query := `
		INSERT OR IGNORE INTO containers (id, type, title, description) 
		VALUES (?, 'BasicContainer', 'Test Container', 'Test Description')`

	_, err := indexer.db.Exec(query, containerID)
	return err
}
