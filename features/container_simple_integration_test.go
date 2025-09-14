package features

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
)

// SimpleContainerIntegrationTest provides basic integration testing for container functionality
type SimpleContainerIntegrationTest struct {
	containerRepo domain.ContainerRepository
	tempDir       string
	t             *testing.T
}

func NewSimpleContainerIntegrationTest(t *testing.T) *SimpleContainerIntegrationTest {
	return &SimpleContainerIntegrationTest{t: t}
}

func (test *SimpleContainerIntegrationTest) Setup() error {
	// Create temporary directory for test storage
	tempDir, err := os.MkdirTemp("", "container-simple-test-*")
	if err != nil {
		return err
	}
	test.tempDir = tempDir

	// Initialize basic repository
	indexer, err := infrastructure.NewSQLiteMembershipIndexer(filepath.Join(tempDir, "index.db"))
	if err != nil {
		return err
	}

	containerRepo, err := infrastructure.NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		return err
	}

	test.containerRepo = containerRepo
	return nil
}

func (test *SimpleContainerIntegrationTest) Cleanup() {
	if test.tempDir != "" {
		os.RemoveAll(test.tempDir)
	}
}

// TestBasicContainerOperations tests basic container CRUD operations
func (test *SimpleContainerIntegrationTest) TestBasicContainerOperations(t *testing.T) {
	ctx := context.Background()

	// Test container creation using domain constructor
	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	container.SetMetadata("title", "Test Container")

	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Test container retrieval
	retrieved, err := test.containerRepo.GetContainer(ctx, "test-container")
	require.NoError(t, err, "Failed to retrieve container")
	assert.Equal(t, "test-container", retrieved.ID())
	assert.Equal(t, domain.BasicContainer, retrieved.ContainerType)

	// Test container existence check
	exists, err := test.containerRepo.ContainerExists(ctx, "test-container")
	require.NoError(t, err, "Failed to check container existence")
	assert.True(t, exists, "Container should exist")

	// Test non-existent container
	exists, err = test.containerRepo.ContainerExists(ctx, "non-existent")
	require.NoError(t, err, "Failed to check non-existent container")
	assert.False(t, exists, "Non-existent container should not exist")

	// Test container deletion
	err = test.containerRepo.DeleteContainer(ctx, "test-container")
	require.NoError(t, err, "Failed to delete container")

	// Verify deletion
	exists, err = test.containerRepo.ContainerExists(ctx, "test-container")
	require.NoError(t, err, "Failed to check deleted container")
	assert.False(t, exists, "Deleted container should not exist")
}

// TestContainerHierarchy tests container hierarchy operations
func (test *SimpleContainerIntegrationTest) TestContainerHierarchy(t *testing.T) {
	ctx := context.Background()

	// Create root container
	root := domain.NewContainer("root", "", domain.BasicContainer)
	err := test.containerRepo.CreateContainer(ctx, root)
	require.NoError(t, err, "Failed to create root container")

	// Create child container
	child := domain.NewContainer("child", "root", domain.BasicContainer)
	err = test.containerRepo.CreateContainer(ctx, child)
	require.NoError(t, err, "Failed to create child container")

	// Test parent-child relationship
	retrievedChild, err := test.containerRepo.GetContainer(ctx, "child")
	require.NoError(t, err, "Failed to retrieve child container")
	assert.Equal(t, "root", retrievedChild.ParentID, "Child should have correct parent")

	// Test getting children
	children, err := test.containerRepo.GetChildren(ctx, "root")
	require.NoError(t, err, "Failed to get children")
	assert.Len(t, children, 1, "Root should have one child")
	assert.Equal(t, "child", children[0].ID(), "Child should be correct")

	// Test getting parent
	parent, err := test.containerRepo.GetParent(ctx, "child")
	require.NoError(t, err, "Failed to get parent")
	assert.Equal(t, "root", parent.ID(), "Parent should be correct")
}

// TestContainerMembership tests container membership operations
func (test *SimpleContainerIntegrationTest) TestContainerMembership(t *testing.T) {
	ctx := context.Background()

	// Create container
	container := domain.NewContainer("membership-test", "", domain.BasicContainer)
	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Add members
	err = test.containerRepo.AddMember(ctx, "membership-test", "resource1")
	require.NoError(t, err, "Failed to add member 1")

	err = test.containerRepo.AddMember(ctx, "membership-test", "resource2")
	require.NoError(t, err, "Failed to add member 2")

	// List members
	members, err := test.containerRepo.ListMembers(ctx, "membership-test", domain.PaginationOptions{})
	require.NoError(t, err, "Failed to list members")
	assert.Len(t, members, 2, "Container should have 2 members")

	// Check specific members
	memberIDs := make([]string, len(members))
	for i, member := range members {
		memberIDs[i] = member
	}
	assert.Contains(t, memberIDs, "resource1", "Should contain resource1")
	assert.Contains(t, memberIDs, "resource2", "Should contain resource2")

	// Remove member
	err = test.containerRepo.RemoveMember(ctx, "membership-test", "resource1")
	require.NoError(t, err, "Failed to remove member")

	// Verify removal
	members, err = test.containerRepo.ListMembers(ctx, "membership-test", domain.PaginationOptions{})
	require.NoError(t, err, "Failed to list members after removal")
	assert.Len(t, members, 1, "Container should have 1 member after removal")
	assert.Equal(t, "resource2", members[0], "Remaining member should be resource2")
}

// TestContainerPagination tests pagination functionality
func (test *SimpleContainerIntegrationTest) TestContainerPagination(t *testing.T) {
	ctx := context.Background()

	// Create container
	container := domain.NewContainer("pagination-test", "", domain.BasicContainer)
	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Add many members
	const totalMembers = 25
	for i := 0; i < totalMembers; i++ {
		memberID := fmt.Sprintf("resource-%d", i)
		err = test.containerRepo.AddMember(ctx, "pagination-test", memberID)
		require.NoError(t, err, "Failed to add member %d", i)
	}

	// Test pagination
	pageSize := 10
	members, err := test.containerRepo.ListMembers(ctx, "pagination-test", domain.PaginationOptions{
		Limit:  pageSize,
		Offset: 0,
	})
	require.NoError(t, err, "Failed to get first page")
	assert.Len(t, members, pageSize, "First page should have correct size")

	// Test second page
	members, err = test.containerRepo.ListMembers(ctx, "pagination-test", domain.PaginationOptions{
		Limit:  pageSize,
		Offset: pageSize,
	})
	require.NoError(t, err, "Failed to get second page")
	assert.Len(t, members, pageSize, "Second page should have correct size")

	// Test last page
	members, err = test.containerRepo.ListMembers(ctx, "pagination-test", domain.PaginationOptions{
		Limit:  pageSize,
		Offset: 2 * pageSize,
	})
	require.NoError(t, err, "Failed to get last page")
	assert.Len(t, members, totalMembers-2*pageSize, "Last page should have remaining items")
}

// TestContainerPathResolution tests path-based container resolution
func (test *SimpleContainerIntegrationTest) TestContainerPathResolution(t *testing.T) {
	ctx := context.Background()

	// Create hierarchy: root -> level1 -> level2
	containers := []struct {
		id       string
		parentID string
	}{
		{"root", ""},
		{"level1", "root"},
		{"level2", "level1"},
	}

	for _, c := range containers {
		container := domain.NewContainer(c.id, c.parentID, domain.BasicContainer)
		err := test.containerRepo.CreateContainer(ctx, container)
		require.NoError(t, err, "Failed to create container %s", c.id)
	}

	// Test path resolution
	path, err := test.containerRepo.GetPath(ctx, "level2")
	require.NoError(t, err, "Failed to get path")
	assert.Equal(t, []string{"root", "level1", "level2"}, path, "Path should be correct")

	// Test path-based lookup
	container, err := test.containerRepo.FindByPath(ctx, "root/level1/level2")
	require.NoError(t, err, "Failed to find by path")
	assert.Equal(t, "level2", container.ID(), "Found container should be correct")
}

// TestContainerConcurrency tests basic concurrent operations
func (test *SimpleContainerIntegrationTest) TestContainerConcurrency(t *testing.T) {
	ctx := context.Background()

	// Create container
	container := domain.NewContainer("concurrent-test", "", domain.BasicContainer)
	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Test concurrent member additions
	const numGoroutines = 5
	const membersPerGoroutine = 10

	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			var err error
			for j := 0; j < membersPerGoroutine; j++ {
				memberID := fmt.Sprintf("resource-%d-%d", goroutineID, j)
				if addErr := test.containerRepo.AddMember(ctx, "concurrent-test", memberID); addErr != nil {
					err = addErr
					break
				}
			}
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err, "Concurrent operation should not fail")
	}

	// Verify all members were added
	members, err := test.containerRepo.ListMembers(ctx, "concurrent-test", domain.PaginationOptions{})
	require.NoError(t, err, "Failed to list members")

	expectedCount := numGoroutines * membersPerGoroutine
	assert.Len(t, members, expectedCount, "All members should be added")
}

// TestContainerPerformance tests basic performance characteristics
func (test *SimpleContainerIntegrationTest) TestContainerPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()

	// Create container
	container := domain.NewContainer("performance-test", "", domain.BasicContainer)
	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Add many members and measure performance
	const numMembers = 1000

	start := time.Now()
	for i := 0; i < numMembers; i++ {
		memberID := fmt.Sprintf("perf-resource-%d", i)
		err = test.containerRepo.AddMember(ctx, "performance-test", memberID)
		require.NoError(t, err, "Failed to add member %d", i)
	}
	addDuration := time.Since(start)

	// Test listing performance
	start = time.Now()
	members, err := test.containerRepo.ListMembers(ctx, "performance-test", domain.PaginationOptions{})
	require.NoError(t, err, "Failed to list members")
	listDuration := time.Since(start)

	assert.Len(t, members, numMembers, "All members should be listed")

	t.Logf("Performance results:")
	t.Logf("  Added %d members in %v (%.2f members/sec)", numMembers, addDuration, float64(numMembers)/addDuration.Seconds())
	t.Logf("  Listed %d members in %v", numMembers, listDuration)

	// Basic performance assertions
	assert.Less(t, addDuration, 30*time.Second, "Adding members should be reasonably fast")
	assert.Less(t, listDuration, 5*time.Second, "Listing should be fast")
}

// Integration test runner
func TestSimpleContainerIntegration(t *testing.T) {
	test := NewSimpleContainerIntegrationTest(t)

	err := test.Setup()
	require.NoError(t, err, "Failed to setup test environment")
	defer test.Cleanup()

	t.Run("BasicOperations", test.TestBasicContainerOperations)
	t.Run("Hierarchy", test.TestContainerHierarchy)
	t.Run("Membership", test.TestContainerMembership)
	t.Run("Pagination", test.TestContainerPagination)
	t.Run("PathResolution", test.TestContainerPathResolution)
	t.Run("Concurrency", test.TestContainerConcurrency)
	t.Run("Performance", test.TestContainerPerformance)
}
