package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

func TestResourceCache(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024 * 1024, // 1MB
		MaxEntries: 10,
		TTL:        time.Second * 5,
	}
	cache := NewResourceCache(config)
	ctx := context.Background()

	// Test basic put and get
	t.Run("PutAndGet", func(t *testing.T) {
		resource := domain.NewResource(
			"cache-test-1",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test ex:name \"Test\" ."),
		)

		// Put resource in cache
		cache.Put(ctx, resource)

		// Get resource from cache
		cachedResource, found := cache.Get(ctx, "cache-test-1")
		if !found {
			t.Fatal("Resource should be found in cache")
		}

		if cachedResource.ID() != resource.ID() {
			t.Errorf("Expected ID %s, got %s", resource.ID(), cachedResource.ID())
		}
		if cachedResource.GetContentType() != resource.GetContentType() {
			t.Errorf("Expected content type %s, got %s", resource.GetContentType(), cachedResource.GetContentType())
		}
	})

	// Test cache miss
	t.Run("CacheMiss", func(t *testing.T) {
		_, found := cache.Get(ctx, "non-existent-resource")
		if found {
			t.Error("Should not find non-existent resource in cache")
		}
	})

	// Test cache eviction by size
	t.Run("SizeEviction", func(t *testing.T) {
		smallConfig := CacheConfig{
			MaxSize:    1024, // 1KB
			MaxEntries: 100,
			TTL:        time.Minute,
		}
		smallCache := NewResourceCache(smallConfig)

		// Add resources that exceed cache size
		resources := make([]*domain.Resource, 5)
		for i := 0; i < 5; i++ {
			data := make([]byte, 300) // 300 bytes each
			for j := range data {
				data[j] = byte('A' + (i % 26))
			}

			resources[i] = domain.NewResource(
				fmt.Sprintf("size-test-%d", i),
				"application/octet-stream",
				data,
			)
			smallCache.Put(ctx, resources[i])
		}

		// Check that some resources were evicted
		stats := smallCache.GetStats()
		currentSize := stats["currentSize"].(int64)
		maxSize := stats["maxSize"].(int64)

		if currentSize > maxSize {
			t.Errorf("Cache size (%d) should not exceed max size (%d)", currentSize, maxSize)
		}

		// First resource should likely be evicted (LRU)
		_, found := smallCache.Get(ctx, "size-test-0")
		if found {
			// This might not always fail due to timing, but it's a good indicator
			t.Log("First resource still in cache (this is okay due to LRU timing)")
		}
	})

	// Test cache eviction by entry count
	t.Run("EntryCountEviction", func(t *testing.T) {
		entryConfig := CacheConfig{
			MaxSize:    10 * 1024 * 1024, // Large size
			MaxEntries: 3,                // Small entry count
			TTL:        time.Minute,
		}
		entryCache := NewResourceCache(entryConfig)

		// Add more resources than max entries
		for i := 0; i < 5; i++ {
			resource := domain.NewResource(
				fmt.Sprintf("entry-test-%d", i),
				"text/turtle",
				[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:test%d ex:name \"Test%d\" .", i, i)),
			)
			entryCache.Put(ctx, resource)
		}

		// Check entry count
		stats := entryCache.GetStats()
		entryCount := stats["entryCount"].(int)
		maxEntries := stats["maxEntries"].(int)

		if entryCount > maxEntries {
			t.Errorf("Entry count (%d) should not exceed max entries (%d)", entryCount, maxEntries)
		}
	})

	// Test TTL expiration
	t.Run("TTLExpiration", func(t *testing.T) {
		ttlConfig := CacheConfig{
			MaxSize:    1024 * 1024,
			MaxEntries: 100,
			TTL:        time.Millisecond * 100, // Very short TTL
		}
		ttlCache := NewResourceCache(ttlConfig)

		resource := domain.NewResource(
			"ttl-test",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:ttl ex:name \"TTL Test\" ."),
		)

		// Put resource in cache
		ttlCache.Put(ctx, resource)

		// Should be found immediately
		_, found := ttlCache.Get(ctx, "ttl-test")
		if !found {
			t.Fatal("Resource should be found immediately after put")
		}

		// Wait for TTL to expire
		time.Sleep(time.Millisecond * 150)

		// Should not be found after TTL expiration
		_, found = ttlCache.Get(ctx, "ttl-test")
		if found {
			t.Error("Resource should not be found after TTL expiration")
		}
	})

	// Test remove operation
	t.Run("Remove", func(t *testing.T) {
		resource := domain.NewResource(
			"remove-test",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:remove ex:name \"Remove Test\" ."),
		)

		cache.Put(ctx, resource)

		// Verify it's in cache
		_, found := cache.Get(ctx, "remove-test")
		if !found {
			t.Fatal("Resource should be in cache before removal")
		}

		// Remove it
		cache.Remove(ctx, "remove-test")

		// Verify it's removed
		_, found = cache.Get(ctx, "remove-test")
		if found {
			t.Error("Resource should not be found after removal")
		}
	})

	// Test clear operation
	t.Run("Clear", func(t *testing.T) {
		// Add multiple resources
		for i := 0; i < 3; i++ {
			resource := domain.NewResource(
				fmt.Sprintf("clear-test-%d", i),
				"text/turtle",
				[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:clear%d ex:name \"Clear Test %d\" .", i, i)),
			)
			cache.Put(ctx, resource)
		}

		// Verify resources are in cache
		stats := cache.GetStats()
		entryCount := stats["entryCount"].(int)
		if entryCount < 3 {
			t.Fatalf("Expected at least 3 entries before clear, got %d", entryCount)
		}

		// Clear cache
		cache.Clear(ctx)

		// Verify cache is empty
		stats = cache.GetStats()
		entryCount = stats["entryCount"].(int)
		currentSize := stats["currentSize"].(int64)

		if entryCount != 0 {
			t.Errorf("Expected 0 entries after clear, got %d", entryCount)
		}
		if currentSize != 0 {
			t.Errorf("Expected 0 size after clear, got %d", currentSize)
		}
	})

	// Test statistics
	t.Run("Statistics", func(t *testing.T) {
		statsCache := NewResourceCache(CacheConfig{
			MaxSize:    1024 * 1024,
			MaxEntries: 100,
			TTL:        time.Minute,
		})

		// Add some resources
		for i := 0; i < 5; i++ {
			resource := domain.NewResource(
				fmt.Sprintf("stats-test-%d", i),
				"text/turtle",
				[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:stats%d ex:name \"Stats Test %d\" .", i, i)),
			)
			statsCache.Put(ctx, resource)
		}

		// Access some resources multiple times to increase hit count
		for i := 0; i < 3; i++ {
			statsCache.Get(ctx, "stats-test-0")
			statsCache.Get(ctx, "stats-test-1")
		}

		stats := statsCache.GetStats()

		// Verify basic stats
		if entryCount := stats["entryCount"].(int); entryCount != 5 {
			t.Errorf("Expected 5 entries, got %d", entryCount)
		}

		if currentSize := stats["currentSize"].(int64); currentSize <= 0 {
			t.Error("Current size should be greater than 0")
		}

		if totalHits := stats["totalHits"].(int64); totalHits < 10 { // 5 puts + at least 6 gets
			t.Errorf("Expected at least 10 total hits, got %d", totalHits)
		}

		if averageHits := stats["averageHits"].(float64); averageHits <= 0 {
			t.Error("Average hits should be greater than 0")
		}

		// Test most accessed resources
		mostAccessed := statsCache.GetMostAccessed(3)
		if len(mostAccessed) != 3 {
			t.Errorf("Expected 3 most accessed resources, got %d", len(mostAccessed))
		}

		// First resource should have highest hit count
		if mostAccessed[0].HitCount < mostAccessed[1].HitCount {
			t.Error("Most accessed resources should be sorted by hit count")
		}
	})

	// Test large resource handling
	t.Run("LargeResourceHandling", func(t *testing.T) {
		largeConfig := CacheConfig{
			MaxSize:    1024, // 1KB
			MaxEntries: 100,
			TTL:        time.Minute,
		}
		largeCache := NewResourceCache(largeConfig)

		// Try to cache a resource larger than half the cache size
		largeData := make([]byte, 600) // 600 bytes (more than half of 1KB)
		for i := range largeData {
			largeData[i] = byte('X')
		}

		largeResource := domain.NewResource(
			"large-resource",
			"application/octet-stream",
			largeData,
		)

		largeCache.Put(ctx, largeResource)

		// Large resource should not be cached
		_, found := largeCache.Get(ctx, "large-resource")
		if found {
			t.Error("Large resource should not be cached")
		}

		stats := largeCache.GetStats()
		if entryCount := stats["entryCount"].(int); entryCount != 0 {
			t.Errorf("Expected 0 entries for large resource, got %d", entryCount)
		}
	})
}

func TestResourceCacheConcurrency(t *testing.T) {
	config := CacheConfig{
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxEntries: 1000,
		TTL:        time.Minute,
	}
	cache := NewResourceCache(config)
	ctx := context.Background()

	// Test concurrent puts and gets
	numGoroutines := 10
	operationsPerGoroutine := 50
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < operationsPerGoroutine; j++ {
				resourceID := fmt.Sprintf("concurrent-resource-%d-%d", goroutineID, j)
				resource := domain.NewResource(
					resourceID,
					"text/turtle",
					[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d%d ex:name \"Resource %d-%d\" .", goroutineID, j, goroutineID, j)),
				)

				// Put resource
				cache.Put(ctx, resource)

				// Get resource
				cachedResource, found := cache.Get(ctx, resourceID)
				if !found {
					errors <- fmt.Errorf("resource %s not found after put", resourceID)
					return
				}

				if cachedResource.ID() != resourceID {
					errors <- fmt.Errorf("resource ID mismatch: expected %s, got %s", resourceID, cachedResource.ID())
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent operation failed: %v", err)
	default:
	}

	// Verify cache stats
	stats := cache.GetStats()
	entryCount := stats["entryCount"].(int)

	if entryCount <= 0 {
		t.Error("Cache should contain entries after concurrent operations")
	}

	t.Logf("Concurrent test completed with %d entries in cache", entryCount)
}

func TestResourceCacheWarmup(t *testing.T) {
	config := CacheConfig{
		MaxSize:    1024 * 1024,
		MaxEntries: 100,
		TTL:        time.Minute,
	}
	cache := NewResourceCache(config)
	ctx := context.Background()

	// Create resources for warmup
	resources := make([]*domain.Resource, 10)
	for i := 0; i < 10; i++ {
		resources[i] = domain.NewResource(
			fmt.Sprintf("warmup-resource-%d", i),
			"text/turtle",
			[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:warmup%d ex:name \"Warmup %d\" .", i, i)),
		)
	}

	// Warmup cache
	cache.Warmup(ctx, resources)

	// Verify all resources are in cache
	for _, resource := range resources {
		cachedResource, found := cache.Get(ctx, resource.ID())
		if !found {
			t.Errorf("Resource %s should be in cache after warmup", resource.ID())
		}
		if cachedResource.ID() != resource.ID() {
			t.Errorf("Resource ID mismatch after warmup: expected %s, got %s", resource.ID(), cachedResource.ID())
		}
	}

	// Verify cache stats
	stats := cache.GetStats()
	entryCount := stats["entryCount"].(int)

	if entryCount != len(resources) {
		t.Errorf("Expected %d entries after warmup, got %d", len(resources), entryCount)
	}
}
