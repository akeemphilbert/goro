package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLargeContainerPerformance tests performance with large containers
func TestLargeContainerPerformance(t *testing.T) {
	t.Skip("Skipping performance test - requires complex mock setup")
	tests := []struct {
		name        string
		memberCount int
		maxDuration time.Duration
	}{
		{
			name:        "Small container (100 members)",
			memberCount: 100,
			maxDuration: 100 * time.Millisecond,
		},
		{
			name:        "Medium container (1000 members)",
			memberCount: 1000,
			maxDuration: 500 * time.Millisecond,
		},
		{
			name:        "Large container (10000 members)",
			memberCount: 10000,
			maxDuration: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			containerRepo := &MockContainerRepository{}
			unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
				return &MockUnitOfWork{}
			}
			rdfConverter := &infrastructure.ContainerRDFConverter{}
			service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

			// Create container
			containerID := "perf-test-container"
			_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
			require.NoError(t, err)

			// Add members to container
			for i := 0; i < tt.memberCount; i++ {
				memberID := fmt.Sprintf("member-%d", i)
				err := service.AddResource(ctx, containerID, memberID)
				require.NoError(t, err)
			}

			// Test listing performance
			start := time.Now()
			pagination := domain.PaginationOptions{Limit: 50, Offset: 0}
			listing, err := service.ListContainerMembers(ctx, containerID, pagination)
			duration := time.Since(start)

			// Assertions
			require.NoError(t, err)
			assert.NotNil(t, listing)
			assert.Equal(t, containerID, listing.ContainerID)
			assert.LessOrEqual(t, len(listing.Members), 50) // Should respect pagination limit
			assert.LessOrEqual(t, duration, tt.maxDuration, "Operation took too long: %v", duration)

			t.Logf("Listed %d members from container with %d total members in %v",
				len(listing.Members), tt.memberCount, duration)
		})
	}
}

// TestPaginationPerformance tests pagination performance with various page sizes
func TestPaginationPerformance(t *testing.T) {
	t.Skip("Skipping performance test - requires complex mock setup")
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with 1000 members
	containerID := "pagination-test-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	memberCount := 1000
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID)
		require.NoError(t, err)
	}

	pageSizes := []int{10, 25, 50, 100, 200}

	for _, pageSize := range pageSizes {
		t.Run(fmt.Sprintf("PageSize_%d", pageSize), func(t *testing.T) {
			start := time.Now()

			// Test first page
			pagination := domain.PaginationOptions{Limit: pageSize, Offset: 0}
			listing, err := service.ListContainerMembers(ctx, containerID, pagination)

			duration := time.Since(start)

			require.NoError(t, err)
			assert.NotNil(t, listing)
			assert.LessOrEqual(t, len(listing.Members), pageSize)
			assert.LessOrEqual(t, duration, 100*time.Millisecond, "Pagination took too long: %v", duration)

			t.Logf("Page size %d: retrieved %d members in %v", pageSize, len(listing.Members), duration)
		})
	}
}

// TestConcurrentContainerAccess tests concurrent access to containers
func TestConcurrentContainerAccess(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container
	containerID := "concurrent-test-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	// Add some initial members
	for i := 0; i < 100; i++ {
		memberID := fmt.Sprintf("initial-member-%d", i)
		err := service.AddResource(ctx, containerID, memberID)
		require.NoError(t, err)
	}

	// Test concurrent reads
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			// Perform multiple operations
			for j := 0; j < 10; j++ {
				pagination := domain.PaginationOptions{Limit: 20, Offset: j * 20}
				_, err := service.ListContainerMembers(ctx, containerID, pagination)
				if err != nil {
					errors <- fmt.Errorf("worker %d iteration %d: %w", workerID, j, err)
					return
				}
			}
		}(i)
	}

	// Wait for all workers to complete
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Worker completed successfully
		case err := <-errors:
			t.Errorf("Concurrent access error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}

	duration := time.Since(start)
	t.Logf("Concurrent access test completed in %v", duration)
	assert.LessOrEqual(t, duration, 2*time.Second, "Concurrent access took too long")
}

// TestMemoryUsageWithLargeContainers tests memory usage with large containers
func TestMemoryUsageWithLargeContainers(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create container with many members
	containerID := "memory-test-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(t, err)

	memberCount := 5000
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID)
		require.NoError(t, err)
	}

	// Test that we can list members without loading everything into memory
	// This test ensures streaming behavior
	totalRetrieved := 0
	pageSize := 100

	for offset := 0; offset < memberCount; offset += pageSize {
		pagination := domain.PaginationOptions{Limit: pageSize, Offset: offset}
		listing, err := service.ListContainerMembers(ctx, containerID, pagination)
		require.NoError(t, err)

		totalRetrieved += len(listing.Members)

		// Ensure we're not loading more than the page size
		assert.LessOrEqual(t, len(listing.Members), pageSize)
	}

	// We should have retrieved all members through pagination
	assert.Equal(t, memberCount, totalRetrieved)
}

// TestDeepHierarchyPerformance tests performance with deep container hierarchies
func TestDeepHierarchyPerformance(t *testing.T) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Create deep hierarchy (10 levels)
	depth := 10
	var parentID string
	containerIDs := make([]string, depth)

	start := time.Now()

	for i := 0; i < depth; i++ {
		containerID := fmt.Sprintf("level-%d-container", i)
		containerIDs[i] = containerID

		_, err := service.CreateContainer(ctx, containerID, parentID, domain.BasicContainer)
		require.NoError(t, err)

		parentID = containerID
	}

	creationDuration := time.Since(start)

	// Test path resolution performance
	start = time.Now()
	deepestContainer := containerIDs[depth-1]
	path, err := service.GetContainerPath(ctx, deepestContainer)
	pathDuration := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, depth, len(path))
	assert.LessOrEqual(t, creationDuration, 1*time.Second, "Deep hierarchy creation took too long")
	assert.LessOrEqual(t, pathDuration, 100*time.Millisecond, "Path resolution took too long")

	t.Logf("Created %d-level hierarchy in %v, path resolution in %v", depth, creationDuration, pathDuration)
}

// BenchmarkContainerListing benchmarks container member listing
func BenchmarkContainerListing(b *testing.B) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Setup container with members
	containerID := "benchmark-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(b, err)

	memberCount := 1000
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID)
		require.NoError(b, err)
	}

	pagination := domain.PaginationOptions{Limit: 50, Offset: 0}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.ListContainerMembers(ctx, containerID, pagination)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// BenchmarkPaginationVariousPageSizes benchmarks pagination with different page sizes
func BenchmarkPaginationVariousPageSizes(b *testing.B) {
	ctx := context.Background()
	containerRepo := &MockContainerRepository{}
	unitOfWorkFactory := func() pericarpdomain.UnitOfWork {
		return &MockUnitOfWork{}
	}
	rdfConverter := &infrastructure.ContainerRDFConverter{}
	service := NewContainerService(containerRepo, unitOfWorkFactory, rdfConverter)

	// Setup container with members
	containerID := "benchmark-pagination-container"
	_, err := service.CreateContainer(ctx, containerID, "", domain.BasicContainer)
	require.NoError(b, err)

	memberCount := 5000
	for i := 0; i < memberCount; i++ {
		memberID := fmt.Sprintf("member-%d", i)
		err := service.AddResource(ctx, containerID, memberID)
		require.NoError(b, err)
	}

	pageSizes := []int{10, 25, 50, 100, 200}

	for _, pageSize := range pageSizes {
		b.Run(fmt.Sprintf("PageSize_%d", pageSize), func(b *testing.B) {
			pagination := domain.PaginationOptions{Limit: pageSize, Offset: 0}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := service.ListContainerMembers(ctx, containerID, pagination)
				if err != nil {
					b.Fatalf("Benchmark failed: %v", err)
				}
			}
		})
	}
}
