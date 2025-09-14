package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMembershipIndexerFiltering tests the filtering functionality
func TestMembershipIndexerFiltering(t *testing.T) {
	// Create in-memory SQLite database for testing
	indexer, err := NewSQLiteMembershipIndexer(":memory:")
	require.NoError(t, err)
	defer indexer.Close()

	ctx := context.Background()
	containerID := "test-container"

	// First add the container itself to the containers table
	_, err = indexer.db.ExecContext(ctx,
		"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		containerID, "BasicContainer")
	require.NoError(t, err)

	// Add test members
	testMembers := []struct {
		memberID   string
		memberType ResourceType
	}{
		{"document-1", ResourceTypeResource},
		{"document-2", ResourceTypeResource},
		{"image-1", ResourceTypeResource},
		{"subfolder-1", ResourceTypeContainer},
		{"subfolder-2", ResourceTypeContainer},
	}

	for _, member := range testMembers {
		// For containers, we need to add them to the containers table first
		if member.memberType == ResourceTypeContainer {
			_, err := indexer.db.ExecContext(ctx,
				"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
				member.memberID, "BasicContainer")
			require.NoError(t, err)
		}

		err := indexer.IndexMembership(ctx, containerID, member.memberID)
		require.NoError(t, err)
	}

	// Test filtering by member type - Resources only
	filter := FilterOptions{MemberType: string(ResourceTypeResource)}
	sort := SortOptions{Field: "name", Direction: "asc"}
	pagination := PaginationOptions{Limit: 10, Offset: 0}

	members, err := indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	// Should only return resources
	assert.Len(t, members, 3) // document-1, document-2, image-1
	for _, member := range members {
		assert.Equal(t, ResourceTypeResource, member.Type)
	}

	// Test filtering by member type - Containers only
	filter = FilterOptions{MemberType: string(ResourceTypeContainer)}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	// Should only return containers
	assert.Len(t, members, 2) // subfolder-1, subfolder-2
	for _, member := range members {
		assert.Equal(t, ResourceTypeContainer, member.Type)
	}

	// Test filtering by name pattern
	filter = FilterOptions{NamePattern: "document"}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	// Should only return documents
	assert.Len(t, members, 2) // document-1, document-2
	for _, member := range members {
		assert.Contains(t, member.ID, "document")
	}

	// Test no filtering - should return all members
	filter = FilterOptions{}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 5) // All members
}

// TestMembershipIndexerSorting tests the sorting functionality
func TestMembershipIndexerSorting(t *testing.T) {
	// Create in-memory SQLite database for testing
	indexer, err := NewSQLiteMembershipIndexer(":memory:")
	require.NoError(t, err)
	defer indexer.Close()

	ctx := context.Background()
	containerID := "test-container"

	// First add the container itself to the containers table
	_, err = indexer.db.ExecContext(ctx,
		"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		containerID, "BasicContainer")
	require.NoError(t, err)

	// Add test members with controlled timing
	testMembers := []string{"zebra", "alpha", "beta", "gamma"}

	for _, memberID := range testMembers {
		err := indexer.IndexMembership(ctx, containerID, memberID)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Test sorting by name ascending
	filter := FilterOptions{}
	sort := SortOptions{Field: "name", Direction: "asc"}
	pagination := PaginationOptions{Limit: 10, Offset: 0}

	members, err := indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 4)
	// Should be sorted alphabetically
	assert.Equal(t, "alpha", members[0].ID)
	assert.Equal(t, "beta", members[1].ID)
	assert.Equal(t, "gamma", members[2].ID)
	assert.Equal(t, "zebra", members[3].ID)

	// Test sorting by name descending
	sort = SortOptions{Field: "name", Direction: "desc"}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 4)
	// Should be sorted reverse alphabetically
	assert.Equal(t, "zebra", members[0].ID)
	assert.Equal(t, "gamma", members[1].ID)
	assert.Equal(t, "beta", members[2].ID)
	assert.Equal(t, "alpha", members[3].ID)

	// Test sorting by creation time ascending (default)
	sort = SortOptions{Field: "createdAt", Direction: "asc"}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 4)
	// Should be in insertion order
	assert.Equal(t, "zebra", members[0].ID)
	assert.Equal(t, "alpha", members[1].ID)
	assert.Equal(t, "beta", members[2].ID)
	assert.Equal(t, "gamma", members[3].ID)
}

// TestMembershipIndexerPagination tests pagination with filtering
func TestMembershipIndexerPagination(t *testing.T) {
	// Create in-memory SQLite database for testing
	indexer, err := NewSQLiteMembershipIndexer(":memory:")
	require.NoError(t, err)
	defer indexer.Close()

	ctx := context.Background()
	containerID := "test-container"

	// First add the container itself to the containers table
	_, err = indexer.db.ExecContext(ctx,
		"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		containerID, "BasicContainer")
	require.NoError(t, err)

	// Add 10 test members
	for i := 0; i < 10; i++ {
		memberID := fmt.Sprintf("member-%02d", i)
		err := indexer.IndexMembership(ctx, containerID, memberID)
		require.NoError(t, err)
	}

	// Test first page
	filter := FilterOptions{}
	sort := SortOptions{Field: "name", Direction: "asc"}
	pagination := PaginationOptions{Limit: 3, Offset: 0}

	members, err := indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 3)
	assert.Equal(t, "member-00", members[0].ID)
	assert.Equal(t, "member-01", members[1].ID)
	assert.Equal(t, "member-02", members[2].ID)

	// Test second page
	pagination = PaginationOptions{Limit: 3, Offset: 3}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 3)
	assert.Equal(t, "member-03", members[0].ID)
	assert.Equal(t, "member-04", members[1].ID)
	assert.Equal(t, "member-05", members[2].ID)

	// Test last page (partial)
	pagination = PaginationOptions{Limit: 3, Offset: 9}
	members, err = indexer.GetMembersWithFiltering(ctx, containerID, pagination, filter, sort)
	require.NoError(t, err)

	assert.Len(t, members, 1)
	assert.Equal(t, "member-09", members[0].ID)
}

// TestGetFilteredMemberCount tests filtered member count functionality
func TestGetFilteredMemberCount(t *testing.T) {
	// Create in-memory SQLite database for testing
	indexer, err := NewSQLiteMembershipIndexer(":memory:")
	require.NoError(t, err)
	defer indexer.Close()

	ctx := context.Background()
	containerID := "test-container"

	// First add the container itself to the containers table
	_, err = indexer.db.ExecContext(ctx,
		"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		containerID, "BasicContainer")
	require.NoError(t, err)

	// Add test members
	testMembers := []struct {
		memberID   string
		memberType ResourceType
	}{
		{"document-1", ResourceTypeResource},
		{"document-2", ResourceTypeResource},
		{"image-1", ResourceTypeResource},
		{"subfolder-1", ResourceTypeContainer},
		{"subfolder-2", ResourceTypeContainer},
	}

	for _, member := range testMembers {
		// For containers, add to containers table first
		if member.memberType == ResourceTypeContainer {
			_, err := indexer.db.ExecContext(ctx,
				"INSERT INTO containers (id, type, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
				member.memberID, "BasicContainer")
			require.NoError(t, err)
		}

		err := indexer.IndexMembership(ctx, containerID, member.memberID)
		require.NoError(t, err)
	}

	// Test total count (no filter)
	filter := FilterOptions{}
	count, err := indexer.GetFilteredMemberCount(ctx, containerID, filter)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Test count with resource filter
	filter = FilterOptions{MemberType: string(ResourceTypeResource)}
	count, err = indexer.GetFilteredMemberCount(ctx, containerID, filter)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// Test count with container filter
	filter = FilterOptions{MemberType: string(ResourceTypeContainer)}
	count, err = indexer.GetFilteredMemberCount(ctx, containerID, filter)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Test count with name pattern filter
	filter = FilterOptions{NamePattern: "document"}
	count, err = indexer.GetFilteredMemberCount(ctx, containerID, filter)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}
