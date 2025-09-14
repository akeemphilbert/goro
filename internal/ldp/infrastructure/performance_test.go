package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// BenchmarkRepository benchmarks repository operations
func BenchmarkRepository(b *testing.B) {
	// Setup test directory
	tempDir := b.TempDir()

	// Test both regular and optimized repositories
	b.Run("FileSystemRepository", func(b *testing.B) {
		benchmarkRepositoryOperations(b, tempDir, false)
	})

	b.Run("OptimizedFileSystemRepository", func(b *testing.B) {
		benchmarkRepositoryOperations(b, tempDir, true)
	})
}

func benchmarkRepositoryOperations(b *testing.B, tempDir string, useOptimized bool) {
	var repo domain.ResourceRepository
	var err error

	if useOptimized {
		cacheConfig := CacheConfig{
			MaxSize:    10 * 1024 * 1024, // 10MB
			MaxEntries: 100,
			TTL:        time.Minute * 5,
		}
		repo, err = NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	} else {
		repo, err = NewFileSystemRepository(tempDir)
	}

	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Benchmark Store operations
	b.Run("Store", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resource := domain.NewResource(
				fmt.Sprintf("resource-%d", i),
				"text/turtle",
				[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d ex:name \"Resource %d\" .", i, i)),
			)

			if err := repo.Store(ctx, resource); err != nil {
				b.Fatalf("Failed to store resource: %v", err)
			}
		}
	})

	// Pre-populate repository for retrieve benchmarks
	for i := 0; i < 100; i++ {
		resource := domain.NewResource(
			fmt.Sprintf("bench-resource-%d", i),
			"text/turtle",
			[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d ex:name \"Resource %d\" .", i, i)),
		)
		if err := repo.Store(ctx, resource); err != nil {
			b.Fatalf("Failed to pre-populate resource: %v", err)
		}
	}

	// Benchmark Retrieve operations
	b.Run("Retrieve", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resourceID := fmt.Sprintf("bench-resource-%d", i%100)
			_, err := repo.Retrieve(ctx, resourceID)
			if err != nil {
				b.Fatalf("Failed to retrieve resource: %v", err)
			}
		}
	})

	// Benchmark Exists operations
	b.Run("Exists", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resourceID := fmt.Sprintf("bench-resource-%d", i%100)
			_, err := repo.Exists(ctx, resourceID)
			if err != nil {
				b.Fatalf("Failed to check resource existence: %v", err)
			}
		}
	})
}

// TestPerformanceRequirements tests that operations meet sub-second response time requirements
func TestPerformanceRequirements(t *testing.T) {
	tempDir := t.TempDir()

	// Test both repository types
	t.Run("FileSystemRepository", func(t *testing.T) {
		testPerformanceRequirements(t, tempDir, false)
	})

	t.Run("OptimizedFileSystemRepository", func(t *testing.T) {
		testPerformanceRequirements(t, tempDir, true)
	})
}

func testPerformanceRequirements(t *testing.T, tempDir string, useOptimized bool) {
	var repo domain.ResourceRepository
	var err error

	if useOptimized {
		cacheConfig := CacheConfig{
			MaxSize:    10 * 1024 * 1024, // 10MB
			MaxEntries: 100,
			TTL:        time.Minute * 5,
		}
		repo, err = NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	} else {
		repo, err = NewFileSystemRepository(tempDir)
	}

	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Pre-populate with test data
	testResources := make([]*domain.Resource, 50)
	for i := 0; i < 50; i++ {
		resource := domain.NewResource(
			fmt.Sprintf("perf-test-resource-%d", i),
			"application/ld+json",
			[]byte(fmt.Sprintf(`{
				"@context": "http://example.org/context",
				"@id": "http://example.org/resource%d",
				"name": "Performance Test Resource %d",
				"description": "This is a test resource for performance validation",
				"data": "%s"
			}`, i, i, generateTestData(1024))), // 1KB of test data
		)
		testResources[i] = resource

		if err := repo.Store(ctx, resource); err != nil {
			t.Fatalf("Failed to store test resource: %v", err)
		}
	}

	// Test Store operation performance (Requirement 3.1: sub-second response times)
	t.Run("StorePerformance", func(t *testing.T) {
		resource := domain.NewResource(
			"perf-store-test",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test ex:name \"Performance Test\" ."),
		)

		start := time.Now()
		err := repo.Store(ctx, resource)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Store operation failed: %v", err)
		}

		if duration > time.Second {
			t.Errorf("Store operation took %v, expected < 1 second", duration)
		}

		t.Logf("Store operation completed in %v", duration)
	})

	// Test Retrieve operation performance (Requirement 3.1: sub-second response times)
	t.Run("RetrievePerformance", func(t *testing.T) {
		start := time.Now()
		_, err := repo.Retrieve(ctx, "perf-test-resource-0")
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Retrieve operation failed: %v", err)
		}

		if duration > time.Second {
			t.Errorf("Retrieve operation took %v, expected < 1 second", duration)
		}

		t.Logf("Retrieve operation completed in %v", duration)
	})

	// Test multiple concurrent retrievals (Requirement 3.4: concurrent access)
	t.Run("ConcurrentRetrievePerformance", func(t *testing.T) {
		const numGoroutines = 10
		const retrievalsPerGoroutine = 5

		start := time.Now()

		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines*retrievalsPerGoroutine)

		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < retrievalsPerGoroutine; j++ {
					resourceID := fmt.Sprintf("perf-test-resource-%d", (goroutineID*retrievalsPerGoroutine+j)%50)
					_, err := repo.Retrieve(ctx, resourceID)
					if err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		duration := time.Since(start)

		// Check for errors
		select {
		case err := <-errors:
			t.Fatalf("Concurrent retrieve operation failed: %v", err)
		default:
		}

		// All operations should complete within reasonable time
		expectedMaxDuration := time.Second * 2 // Allow 2 seconds for concurrent operations
		if duration > expectedMaxDuration {
			t.Errorf("Concurrent retrieve operations took %v, expected < %v", duration, expectedMaxDuration)
		}

		t.Logf("Concurrent retrieve operations (%d goroutines, %d retrievals each) completed in %v",
			numGoroutines, retrievalsPerGoroutine, duration)
	})

	// Test cache effectiveness (for optimized repository)
	if useOptimized {
		t.Run("CacheEffectiveness", func(t *testing.T) {
			optimizedRepo := repo.(*OptimizedFileSystemRepository)

			// First retrieval (cache miss)
			start := time.Now()
			_, err := repo.Retrieve(ctx, "perf-test-resource-0")
			firstDuration := time.Since(start)

			if err != nil {
				t.Fatalf("First retrieve failed: %v", err)
			}

			// Second retrieval (cache hit)
			start = time.Now()
			_, err = repo.Retrieve(ctx, "perf-test-resource-0")
			secondDuration := time.Since(start)

			if err != nil {
				t.Fatalf("Second retrieve failed: %v", err)
			}

			// Cache hit should be significantly faster
			if secondDuration >= firstDuration {
				t.Errorf("Cache hit (%v) should be faster than cache miss (%v)", secondDuration, firstDuration)
			}

			// Get cache stats
			stats := optimizedRepo.GetStats(ctx)
			t.Logf("Cache stats: %+v", stats["cache"])
			t.Logf("First retrieve (cache miss): %v", firstDuration)
			t.Logf("Second retrieve (cache hit): %v", secondDuration)
		})
	}
}

// TestLargeFilePerformance tests performance with larger files (Requirement 3.2: streaming support)
func TestLargeFilePerformance(t *testing.T) {
	tempDir := t.TempDir()

	cacheConfig := CacheConfig{
		MaxSize:    50 * 1024 * 1024, // 50MB cache
		MaxEntries: 10,               // Small number for large files
		TTL:        time.Minute * 5,
	}

	repo, err := NewOptimizedFileSystemRepository(tempDir, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Test with different file sizes
	fileSizes := []int{
		1024,    // 1KB
		10240,   // 10KB
		102400,  // 100KB
		1048576, // 1MB
	}

	for _, size := range fileSizes {
		t.Run(fmt.Sprintf("FileSize_%dB", size), func(t *testing.T) {
			// Create large resource
			largeData := generateTestData(size)
			resource := domain.NewResource(
				fmt.Sprintf("large-file-%d", size),
				"application/octet-stream",
				largeData,
			)

			// Test store performance
			start := time.Now()
			err := repo.Store(ctx, resource)
			storeDuration := time.Since(start)

			if err != nil {
				t.Fatalf("Failed to store large file: %v", err)
			}

			// Test retrieve performance
			start = time.Now()
			retrievedResource, err := repo.Retrieve(ctx, resource.ID())
			retrieveDuration := time.Since(start)

			if err != nil {
				t.Fatalf("Failed to retrieve large file: %v", err)
			}

			// Verify data integrity
			if len(retrievedResource.GetData()) != len(largeData) {
				t.Errorf("Retrieved data size mismatch: expected %d, got %d",
					len(largeData), len(retrievedResource.GetData()))
			}

			// Performance should still be reasonable for larger files
			maxDuration := time.Second * 2 // Allow 2 seconds for large files
			if storeDuration > maxDuration {
				t.Errorf("Store operation for %d bytes took %v, expected < %v",
					size, storeDuration, maxDuration)
			}
			if retrieveDuration > maxDuration {
				t.Errorf("Retrieve operation for %d bytes took %v, expected < %v",
					size, retrieveDuration, maxDuration)
			}

			t.Logf("File size: %d bytes, Store: %v, Retrieve: %v",
				size, storeDuration, retrieveDuration)
		})
	}
}

// TestIndexingPerformance tests the performance benefits of indexing
func TestIndexingPerformance(t *testing.T) {
	tempDir := t.TempDir()

	// Create repositories
	regularRepo, err := NewFileSystemRepository(tempDir + "/regular")
	if err != nil {
		t.Fatalf("Failed to create regular repository: %v", err)
	}

	cacheConfig := CacheConfig{
		MaxSize:    10 * 1024 * 1024,
		MaxEntries: 100,
		TTL:        time.Minute * 5,
	}
	optimizedRepo, err := NewOptimizedFileSystemRepository(tempDir+"/optimized", cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create optimized repository: %v", err)
	}

	ctx := context.Background()

	// Pre-populate both repositories with the same data
	numResources := 100
	for i := 0; i < numResources; i++ {
		resource := domain.NewResource(
			fmt.Sprintf("index-test-resource-%d", i),
			"text/turtle",
			[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d ex:name \"Resource %d\" .", i, i)),
		)

		if err := regularRepo.Store(ctx, resource); err != nil {
			t.Fatalf("Failed to store resource in regular repo: %v", err)
		}
		if err := optimizedRepo.Store(ctx, resource); err != nil {
			t.Fatalf("Failed to store resource in optimized repo: %v", err)
		}
	}

	// Test existence checks (where indexing should provide significant benefit)
	t.Run("ExistenceChecks", func(t *testing.T) {
		// Test regular repository
		start := time.Now()
		for i := 0; i < numResources; i++ {
			resourceID := fmt.Sprintf("index-test-resource-%d", i)
			_, err := regularRepo.Exists(ctx, resourceID)
			if err != nil {
				t.Fatalf("Regular repo existence check failed: %v", err)
			}
		}
		regularDuration := time.Since(start)

		// Test optimized repository
		start = time.Now()
		for i := 0; i < numResources; i++ {
			resourceID := fmt.Sprintf("index-test-resource-%d", i)
			_, err := optimizedRepo.Exists(ctx, resourceID)
			if err != nil {
				t.Fatalf("Optimized repo existence check failed: %v", err)
			}
		}
		optimizedDuration := time.Since(start)

		t.Logf("Regular repository existence checks: %v", regularDuration)
		t.Logf("Optimized repository existence checks: %v", optimizedDuration)

		// Optimized should be faster (though the difference might be small for this test size)
		if optimizedDuration > regularDuration*2 {
			t.Errorf("Optimized repository should not be significantly slower than regular repository")
		}
	})
}

// generateTestData generates test data of specified size
func generateTestData(size int) []byte {
	data := make([]byte, size)
	pattern := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	for i := 0; i < size; i++ {
		data[i] = pattern[i%len(pattern)]
	}

	return data
}

// BenchmarkCacheOperations benchmarks cache operations specifically
func BenchmarkCacheOperations(b *testing.B) {
	cacheConfig := CacheConfig{
		MaxSize:    10 * 1024 * 1024,
		MaxEntries: 1000,
		TTL:        time.Minute * 5,
	}
	cache := NewResourceCache(cacheConfig)
	ctx := context.Background()

	// Create test resources
	resources := make([]*domain.Resource, 100)
	for i := 0; i < 100; i++ {
		resources[i] = domain.NewResource(
			fmt.Sprintf("cache-bench-resource-%d", i),
			"text/turtle",
			[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d ex:name \"Resource %d\" .", i, i)),
		)
	}

	b.Run("CachePut", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resource := resources[i%len(resources)]
			cache.Put(ctx, resource)
		}
	})

	// Pre-populate cache for get benchmarks
	for _, resource := range resources {
		cache.Put(ctx, resource)
	}

	b.Run("CacheGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resourceID := fmt.Sprintf("cache-bench-resource-%d", i%len(resources))
			cache.Get(ctx, resourceID)
		}
	})
}
