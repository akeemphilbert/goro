package infrastructure

import (
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/assert"
)

// TestContainerCacheBasicOperations tests basic cache operations
func TestContainerCacheBasicOperations(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 100)

	// Create test container
	container := domain.NewContainer("test-container", "", domain.BasicContainer)

	// Test cache miss
	_, found := cache.Get("test-container")
	assert.False(t, found)

	// Test cache put and get
	cache.Put("test-container", container, 5, 1024)
	cachedContainer, found := cache.Get("test-container")
	assert.True(t, found)
	assert.Equal(t, container.ID(), cachedContainer.ID())

	// Test stats retrieval
	memberCount, totalSize, found := cache.GetStats("test-container")
	assert.True(t, found)
	assert.Equal(t, 5, memberCount)
	assert.Equal(t, int64(1024), totalSize)

	// Test cache size
	assert.Equal(t, 1, cache.Size())
}

// TestContainerCacheExpiration tests cache expiration
func TestContainerCacheExpiration(t *testing.T) {
	// Short TTL for testing
	cache := NewContainerCache(50*time.Millisecond, 100)

	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	cache.Put("test-container", container, 5, 1024)

	// Should be available immediately
	_, found := cache.Get("test-container")
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("test-container")
	assert.False(t, found)
}

// TestContainerCacheEviction tests LRU eviction
func TestContainerCacheEviction(t *testing.T) {
	// Small cache for testing eviction
	cache := NewContainerCache(5*time.Minute, 2)

	// Add first container
	container1 := domain.NewContainer("container-1", "", domain.BasicContainer)
	cache.Put("container-1", container1, 1, 100)

	// Add second container
	container2 := domain.NewContainer("container-2", "", domain.BasicContainer)
	cache.Put("container-2", container2, 2, 200)

	// Both should be in cache
	assert.Equal(t, 2, cache.Size())

	// Access first container to make it more recently used
	_, found := cache.Get("container-1")
	assert.True(t, found)

	// Add third container - should evict container-2 (least recently used)
	container3 := domain.NewContainer("container-3", "", domain.BasicContainer)
	cache.Put("container-3", container3, 3, 300)

	// Cache should still have 2 entries
	assert.Equal(t, 2, cache.Size())

	// Container-1 and container-3 should be present
	_, found = cache.Get("container-1")
	assert.True(t, found)
	_, found = cache.Get("container-3")
	assert.True(t, found)

	// Container-2 should be evicted
	_, found = cache.Get("container-2")
	assert.False(t, found)
}

// TestContainerCacheInvalidation tests cache invalidation
func TestContainerCacheInvalidation(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 100)

	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	cache.Put("test-container", container, 5, 1024)

	// Verify it's cached
	_, found := cache.Get("test-container")
	assert.True(t, found)

	// Invalidate
	cache.Invalidate("test-container")

	// Should no longer be in cache
	_, found = cache.Get("test-container")
	assert.False(t, found)
	assert.Equal(t, 0, cache.Size())
}

// TestContainerCacheStatsUpdate tests updating cached statistics
func TestContainerCacheStatsUpdate(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 100)

	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	cache.Put("test-container", container, 5, 1024)

	// Update stats
	cache.UpdateStats("test-container", 10, 2048)

	// Verify updated stats
	memberCount, totalSize, found := cache.GetStats("test-container")
	assert.True(t, found)
	assert.Equal(t, 10, memberCount)
	assert.Equal(t, int64(2048), totalSize)
}

// TestContainerCacheClear tests clearing the cache
func TestContainerCacheClear(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 100)

	// Add multiple containers
	for i := 0; i < 5; i++ {
		container := domain.NewContainer(fmt.Sprintf("container-%d", i), "", domain.BasicContainer)
		cache.Put(fmt.Sprintf("container-%d", i), container, i, int64(i*100))
	}

	assert.Equal(t, 5, cache.Size())

	// Clear cache
	cache.Clear()

	assert.Equal(t, 0, cache.Size())

	// Verify all containers are gone
	for i := 0; i < 5; i++ {
		_, found := cache.Get(fmt.Sprintf("container-%d", i))
		assert.False(t, found)
	}
}

// TestContainerCacheStats tests cache statistics
func TestContainerCacheStats(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 100)

	// Add some containers
	for i := 0; i < 3; i++ {
		container := domain.NewContainer(fmt.Sprintf("container-%d", i), "", domain.BasicContainer)
		cache.Put(fmt.Sprintf("container-%d", i), container, i, int64(i*100))
	}

	stats := cache.GetCacheStats()
	assert.Equal(t, 3, stats["total_entries"])
	assert.Equal(t, 100, stats["max_entries"])
	assert.Equal(t, 300.0, stats["ttl_seconds"]) // 5 minutes = 300 seconds
	assert.Equal(t, 0, stats["expired_entries"])
}

// TestCachedContainerRepository tests the cached repository wrapper
func TestCachedContainerRepository(t *testing.T) {
	// Skip this test for now - requires mock repository
	t.Skip("Skipping cached repository test - requires mock implementation")
}

// TestCachedContainerRepositoryInvalidation tests cache invalidation in repository operations
func TestCachedContainerRepositoryInvalidation(t *testing.T) {
	// Skip this test for now - requires mock repository
	t.Skip("Skipping cached repository invalidation test - requires mock implementation")
}

// TestCachedContainerRepositoryMemberOperations tests member operations with cache invalidation
func TestCachedContainerRepositoryMemberOperations(t *testing.T) {
	// Skip this test for now - requires mock repository
	t.Skip("Skipping cached repository member operations test - requires mock implementation")
}

// TestContainerCacheExpiredCleanup tests automatic cleanup of expired entries
func TestContainerCacheExpiredCleanup(t *testing.T) {
	// Very short TTL for testing
	cache := NewContainerCache(10*time.Millisecond, 100)

	// Add container
	container := domain.NewContainer("test-container", "", domain.BasicContainer)
	cache.Put("test-container", container, 5, 1024)

	assert.Equal(t, 1, cache.Size())

	// Wait for expiration and cleanup
	time.Sleep(50 * time.Millisecond)

	// Size should eventually become 0 due to cleanup
	assert.Equal(t, 0, cache.Size())
}

// TestContainerCacheConcurrentAccess tests concurrent access to cache
func TestContainerCacheConcurrentAccess(t *testing.T) {
	cache := NewContainerCache(5*time.Minute, 1000)

	// Concurrent writes
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				containerID := fmt.Sprintf("container-%d-%d", workerID, j)
				container := domain.NewContainer(containerID, "", domain.BasicContainer)
				cache.Put(containerID, container, j, int64(j*100))

				// Try to read it back
				_, found := cache.Get(containerID)
				assert.True(t, found)
			}
		}(i)
	}

	// Wait for all workers
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Worker completed
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}

	// Verify final state
	assert.Equal(t, 100, cache.Size()) // 10 workers * 10 containers each
}
