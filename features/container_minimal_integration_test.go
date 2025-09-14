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

// MinimalContainerIntegrationTest provides minimal integration testing for container functionality
type MinimalContainerIntegrationTest struct {
	containerRepo domain.ContainerRepository
	tempDir       string
	t             *testing.T
}

func NewMinimalContainerIntegrationTest(t *testing.T) *MinimalContainerIntegrationTest {
	return &MinimalContainerIntegrationTest{t: t}
}

func (test *MinimalContainerIntegrationTest) Setup() error {
	// Create temporary directory for test storage
	tempDir, err := os.MkdirTemp("", "container-minimal-test-*")
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

func (test *MinimalContainerIntegrationTest) Cleanup() {
	if test.tempDir != "" {
		os.RemoveAll(test.tempDir)
	}
}

// TestBasicContainerOperations tests basic container CRUD operations
func (test *MinimalContainerIntegrationTest) TestBasicContainerOperations(t *testing.T) {
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

// TestContainerMembership tests container membership operations
func (test *MinimalContainerIntegrationTest) TestContainerMembership(t *testing.T) {
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
func (test *MinimalContainerIntegrationTest) TestContainerPagination(t *testing.T) {
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

// TestContainerConcurrency tests basic concurrent operations
func (test *MinimalContainerIntegrationTest) TestContainerConcurrency(t *testing.T) {
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
func (test *MinimalContainerIntegrationTest) TestContainerPerformance(t *testing.T) {
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

// TestContainerEventEmission tests that container operations emit events
func (test *MinimalContainerIntegrationTest) TestContainerEventEmission(t *testing.T) {
	ctx := context.Background()

	// Create container and check events
	container := domain.NewContainer("event-test", "", domain.BasicContainer)

	// Check that creation events were emitted
	events := container.UncommittedEvents()
	assert.GreaterOrEqual(t, len(events), 1, "Container creation should emit events")

	// Store the container
	err := test.containerRepo.CreateContainer(ctx, container)
	require.NoError(t, err, "Failed to create container")

	// Add member and check for events
	err = container.AddMember("test-resource")
	require.NoError(t, err, "Failed to add member")

	events = container.UncommittedEvents()
	assert.GreaterOrEqual(t, len(events), 2, "Adding member should emit additional events")
}

// Integration test runner
func TestMinimalContainerIntegration(t *testing.T) {
	test := NewMinimalContainerIntegrationTest(t)

	err := test.Setup()
	require.NoError(t, err, "Failed to setup test environment")
	defer test.Cleanup()

	t.Run("BasicOperations", test.TestBasicContainerOperations)
	t.Run("Membership", test.TestContainerMembership)
	t.Run("Pagination", test.TestContainerPagination)
	t.Run("Concurrency", test.TestContainerConcurrency)
	t.Run("Performance", test.TestContainerPerformance)
	t.Run("EventEmission", test.TestContainerEventEmission)
}
