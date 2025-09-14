package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMembershipIndexerProvider(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test_membership_provider_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test provider
	indexer, err := MembershipIndexerProvider(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create membership indexer: %v", err)
	}
	defer indexer.Close()

	if indexer == nil {
		t.Fatal("Indexer should not be nil")
	}

	// Verify database file was created
	dbPath := filepath.Join(tmpDir, "membership.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}

	// Test that indexer is functional
	ctx := context.Background()

	// Create a test container first
	err = createTestContainer(indexer, "test-container")
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	// Test basic functionality
	err = indexer.IndexMembership(ctx, "test-container", "test-resource")
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}

	members, err := indexer.GetMembers(ctx, "test-container", PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(members))
	}
}

func TestMembershipIndexerProvider_InvalidPath(t *testing.T) {
	// Test with invalid path (should still work as SQLite creates the file)
	invalidPath := "/nonexistent/path"

	// This should fail because the directory doesn't exist
	_, err := MembershipIndexerProvider(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, but got none")
	}
}

func TestMembershipIndexerProvider_ExistingDatabase(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test_membership_existing_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create first indexer
	indexer1, err := MembershipIndexerProvider(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create first indexer: %v", err)
	}

	// Add some data
	ctx := context.Background()
	err = createTestContainer(indexer1, "test-container")
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	err = indexer1.IndexMembership(ctx, "test-container", "test-resource")
	if err != nil {
		t.Fatalf("Failed to index membership: %v", err)
	}
	indexer1.Close()

	// Create second indexer with same path (should reuse existing database)
	indexer2, err := MembershipIndexerProvider(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create second indexer: %v", err)
	}
	defer indexer2.Close()

	// Verify data persisted
	members, err := indexer2.GetMembers(ctx, "test-container", PaginationOptions{})
	if err != nil {
		t.Fatalf("Failed to get members from second indexer: %v", err)
	}

	if len(members) != 1 {
		t.Errorf("Expected 1 member from persisted data, got %d", len(members))
	}
}
