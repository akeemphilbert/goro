package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

func TestOptimizedFileSystemRepository(t *testing.T) {
	tempDir := t.TempDir()

	cacheConfig := CacheConfig{
		MaxSize:    1024 * 1024, // 1MB
		MaxEntries: 100,
		TTL:        time.Minute * 5,
	}

	repo, err := NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create optimized repository: %v", err)
	}

	ctx := context.Background()

	// Test basic operations
	t.Run("BasicOperations", func(t *testing.T) {
		resource := domain.NewResource(
			"optimized-test-1",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test ex:name \"Optimized Test\" ."),
		)
		resource.SetMetadata("category", "test")

		// Store resource
		err := repo.Store(ctx, resource)
		if err != nil {
			t.Fatalf("Failed to store resource: %v", err)
		}

		// Check existence (should use index)
		exists, err := repo.Exists(ctx, "optimized-test-1")
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Error("Resource should exist")
		}

		// Retrieve resource (should use cache after first retrieval)
		retrievedResource, err := repo.Retrieve(ctx, "optimized-test-1")
		if err != nil {
			t.Fatalf("Failed to retrieve resource: %v", err)
		}

		if retrievedResource.ID() != resource.ID() {
			t.Errorf("Expected ID %s, got %s", resource.ID(), retrievedResource.ID())
		}

		// Delete resource
		err = repo.Delete(ctx, "optimized-test-1")
		if err != nil {
			t.Fatalf("Failed to delete resource: %v", err)
		}

		// Verify deletion
		exists, err = repo.Exists(ctx, "optimized-test-1")
		if err != nil {
			t.Fatalf("Failed to check existence after deletion: %v", err)
		}
		if exists {
			t.Error("Resource should not exist after deletion")
		}
	})

	// Test cache effectiveness
	t.Run("CacheEffectiveness", func(t *testing.T) {
		resource := domain.NewResource(
			"cache-test",
			"application/ld+json",
			[]byte(`{"@context": "http://example.org/", "@id": "cache-test", "name": "Cache Test"}`),
		)

		// Store resource
		err := repo.Store(ctx, resource)
		if err != nil {
			t.Fatalf("Failed to store resource: %v", err)
		}

		// First retrieval (should populate cache)
		start := time.Now()
		_, err = repo.Retrieve(ctx, "cache-test")
		firstDuration := time.Since(start)
		if err != nil {
			t.Fatalf("Failed to retrieve resource first time: %v", err)
		}

		// Second retrieval (should use cache)
		start = time.Now()
		_, err = repo.Retrieve(ctx, "cache-test")
		secondDuration := time.Since(start)
		if err != nil {
			t.Fatalf("Failed to retrieve resource second time: %v", err)
		}

		// Cache hit should be faster (though this might not always be true in tests)
		t.Logf("First retrieval: %v, Second retrieval: %v", firstDuration, secondDuration)

		// Verify resource is in cache by checking stats
		stats := repo.GetStats(ctx)
		cacheStats, ok := stats["cache"].(map[string]interface{})
		if !ok {
			t.Fatal("Cache stats should be available")
		}

		entryCount := cacheStats["entryCount"].(int)
		if entryCount == 0 {
			t.Error("Cache should contain entries after retrieval")
		}
	})

	// Test index-based queries
	t.Run("IndexQueries", func(t *testing.T) {
		// Create a fresh repository for this test
		queryTempDir := t.TempDir()
		queryRepo, err := NewOptimizedFileSystemRepository(queryTempDir, cacheConfig)
		if err != nil {
			t.Fatalf("Failed to create query test repository: %v", err)
		}

		// Add multiple resources with different content types
		resources := []*domain.Resource{
			domain.NewResource("index-test-1", "text/turtle", []byte("@prefix ex: <http://example.org/> .\nex:test1 ex:name \"Test1\" .")),
			domain.NewResource("index-test-2", "application/ld+json", []byte(`{"@context": "http://example.org/", "@id": "test2", "name": "Test2"}`)),
			domain.NewResource("index-test-3", "text/turtle", []byte("@prefix ex: <http://example.org/> .\nex:test3 ex:name \"Test3\" .")),
		}

		// Set metadata for tag-based queries
		resources[0].SetMetadata("category", "turtle")
		resources[1].SetMetadata("category", "jsonld")
		resources[2].SetMetadata("category", "turtle")

		// Store all resources
		for _, resource := range resources {
			if err := queryRepo.Store(ctx, resource); err != nil {
				t.Fatalf("Failed to store resource %s: %v", resource.ID(), err)
			}
		}

		// Test FindByContentType
		turtleResources, err := queryRepo.FindByContentType(ctx, "text/turtle")
		if err != nil {
			t.Fatalf("Failed to find resources by content type: %v", err)
		}
		if len(turtleResources) != 2 {
			t.Errorf("Expected 2 turtle resources, got %d", len(turtleResources))
		}

		jsonLDResources, err := queryRepo.FindByContentType(ctx, "application/ld+json")
		if err != nil {
			t.Fatalf("Failed to find JSON-LD resources: %v", err)
		}
		if len(jsonLDResources) != 1 {
			t.Errorf("Expected 1 JSON-LD resource, got %d", len(jsonLDResources))
		}

		// Test FindByTag
		turtleCategoryResources, err := queryRepo.FindByTag(ctx, "category", "turtle")
		if err != nil {
			t.Fatalf("Failed to find resources by tag: %v", err)
		}
		if len(turtleCategoryResources) != 2 {
			t.Errorf("Expected 2 resources with category 'turtle', got %d", len(turtleCategoryResources))
		}
	})

	// Test pagination
	t.Run("Pagination", func(t *testing.T) {
		// Add more resources for pagination test
		for i := 0; i < 15; i++ {
			resource := domain.NewResource(
				fmt.Sprintf("pagination-test-%d", i),
				"text/turtle",
				[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:page%d ex:name \"Page Test %d\" .", i, i)),
			)
			if err := repo.Store(ctx, resource); err != nil {
				t.Fatalf("Failed to store pagination test resource: %v", err)
			}
		}

		// Test first page
		firstPage, err := repo.ListResources(ctx, 0, 5)
		if err != nil {
			t.Fatalf("Failed to get first page: %v", err)
		}
		if len(firstPage) != 5 {
			t.Errorf("Expected 5 resources in first page, got %d", len(firstPage))
		}

		// Test second page
		secondPage, err := repo.ListResources(ctx, 5, 5)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}
		if len(secondPage) != 5 {
			t.Errorf("Expected 5 resources in second page, got %d", len(secondPage))
		}

		// Test beyond available resources
		beyondPage, err := repo.ListResources(ctx, 1000, 5)
		if err != nil {
			t.Fatalf("Failed to get page beyond available resources: %v", err)
		}
		if len(beyondPage) != 0 {
			t.Errorf("Expected 0 resources beyond available range, got %d", len(beyondPage))
		}
	})

	// Test statistics
	t.Run("Statistics", func(t *testing.T) {
		stats := repo.GetStats(ctx)

		// Check index stats
		indexStats, ok := stats["index"].(map[string]interface{})
		if !ok {
			t.Fatal("Index stats should be available")
		}

		totalResources := indexStats["totalResources"].(int)
		if totalResources <= 0 {
			t.Error("Index should contain resources")
		}

		// Check cache stats
		cacheStats, ok := stats["cache"].(map[string]interface{})
		if !ok {
			t.Fatal("Cache stats should be available")
		}

		maxSize := cacheStats["maxSize"].(int64)
		if maxSize != int64(cacheConfig.MaxSize) {
			t.Errorf("Expected max cache size %d, got %d", cacheConfig.MaxSize, maxSize)
		}

		t.Logf("Repository stats: %+v", stats)
	})

	// Test cache warmup
	t.Run("CacheWarmup", func(t *testing.T) {
		// Clear cache first
		repo.ClearCache(ctx)

		// Verify cache is empty
		stats := repo.GetStats(ctx)
		cacheStats := stats["cache"].(map[string]interface{})
		entryCount := cacheStats["entryCount"].(int)
		if entryCount != 0 {
			t.Errorf("Expected empty cache, got %d entries", entryCount)
		}

		// Warmup cache
		err := repo.WarmupCache(ctx)
		if err != nil {
			t.Fatalf("Failed to warmup cache: %v", err)
		}

		// Verify cache has entries after warmup
		stats = repo.GetStats(ctx)
		cacheStats = stats["cache"].(map[string]interface{})
		entryCount = cacheStats["entryCount"].(int)
		if entryCount == 0 {
			t.Error("Cache should contain entries after warmup")
		}

		t.Logf("Cache entries after warmup: %d", entryCount)
	})

	// Test index rebuild
	t.Run("IndexRebuild", func(t *testing.T) {
		// Get current index stats
		stats := repo.GetStats(ctx)
		indexStats := stats["index"].(map[string]interface{})
		originalCount := indexStats["totalResources"].(int)

		// Rebuild index
		err := repo.RebuildIndex(ctx)
		if err != nil {
			t.Fatalf("Failed to rebuild index: %v", err)
		}

		// Verify index still has the same number of resources
		stats = repo.GetStats(ctx)
		indexStats = stats["index"].(map[string]interface{})
		rebuiltCount := indexStats["totalResources"].(int)

		if rebuiltCount != originalCount {
			t.Errorf("Expected %d resources after rebuild, got %d", originalCount, rebuiltCount)
		}
	})
}

func TestOptimizedRepositoryPerformance(t *testing.T) {
	tempDir := t.TempDir()

	cacheConfig := CacheConfig{
		MaxSize:    5 * 1024 * 1024, // 5MB
		MaxEntries: 200,
		TTL:        time.Minute * 10,
	}

	repo, err := NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create optimized repository: %v", err)
	}

	ctx := context.Background()

	// Pre-populate repository
	numResources := 50
	for i := 0; i < numResources; i++ {
		resource := domain.NewResource(
			fmt.Sprintf("perf-resource-%d", i),
			"text/turtle",
			[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:perf%d ex:name \"Performance Test %d\" .", i, i)),
		)
		resource.SetMetadata("index", fmt.Sprintf("%d", i))

		if err := repo.Store(ctx, resource); err != nil {
			t.Fatalf("Failed to store performance test resource: %v", err)
		}
	}

	// Test existence check performance (should use index)
	t.Run("ExistenceCheckPerformance", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < numResources; i++ {
			resourceID := fmt.Sprintf("perf-resource-%d", i)
			exists, err := repo.Exists(ctx, resourceID)
			if err != nil {
				t.Fatalf("Existence check failed: %v", err)
			}
			if !exists {
				t.Errorf("Resource %s should exist", resourceID)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(numResources)

		t.Logf("Existence checks: %d resources in %v (avg: %v per check)", numResources, duration, avgDuration)

		// Each existence check should be very fast with indexing
		if avgDuration > time.Millisecond*10 {
			t.Errorf("Average existence check took %v, expected < 10ms", avgDuration)
		}
	})

	// Test retrieval performance with caching
	t.Run("RetrievalPerformanceWithCaching", func(t *testing.T) {
		// First round - populate cache
		start := time.Now()
		for i := 0; i < 10; i++ {
			resourceID := fmt.Sprintf("perf-resource-%d", i)
			_, err := repo.Retrieve(ctx, resourceID)
			if err != nil {
				t.Fatalf("First retrieval failed: %v", err)
			}
		}
		firstRoundDuration := time.Since(start)

		// Second round - should use cache
		start = time.Now()
		for i := 0; i < 10; i++ {
			resourceID := fmt.Sprintf("perf-resource-%d", i)
			_, err := repo.Retrieve(ctx, resourceID)
			if err != nil {
				t.Fatalf("Second retrieval failed: %v", err)
			}
		}
		secondRoundDuration := time.Since(start)

		t.Logf("First round (cache miss): %v, Second round (cache hit): %v", firstRoundDuration, secondRoundDuration)

		// Second round should be faster or at least not significantly slower
		if secondRoundDuration > firstRoundDuration*2 {
			t.Errorf("Cache hits should not be significantly slower than cache misses")
		}
	})

	// Test concurrent access performance
	t.Run("ConcurrentAccessPerformance", func(t *testing.T) {
		numGoroutines := 5
		operationsPerGoroutine := 10

		start := time.Now()
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines*operationsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < operationsPerGoroutine; j++ {
					resourceID := fmt.Sprintf("perf-resource-%d", (goroutineID*operationsPerGoroutine+j)%numResources)

					// Mix of operations
					switch j % 3 {
					case 0:
						_, err := repo.Exists(ctx, resourceID)
						if err != nil {
							errors <- err
							return
						}
					case 1:
						_, err := repo.Retrieve(ctx, resourceID)
						if err != nil {
							errors <- err
							return
						}
					case 2:
						_, err := repo.FindByContentType(ctx, "text/turtle")
						if err != nil {
							errors <- err
							return
						}
					}
				}
			}(i)
		}

		// Wait for completion
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		duration := time.Since(start)

		// Check for errors
		select {
		case err := <-errors:
			t.Fatalf("Concurrent operation failed: %v", err)
		default:
		}

		totalOperations := numGoroutines * operationsPerGoroutine
		avgDuration := duration / time.Duration(totalOperations)

		t.Logf("Concurrent operations: %d operations in %v (avg: %v per operation)",
			totalOperations, duration, avgDuration)

		// Operations should complete in reasonable time
		if duration > time.Second*5 {
			t.Errorf("Concurrent operations took %v, expected < 5 seconds", duration)
		}
	})
}
