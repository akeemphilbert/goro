package infrastructure

import (
	"context"
	"fmt"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

func TestResourceIndexer(t *testing.T) {
	tempDir := t.TempDir()

	indexer, err := NewResourceIndexer(tempDir)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	// Test adding resources
	t.Run("AddResource", func(t *testing.T) {
		resource := domain.NewResource(
			"test-resource-1",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test ex:name \"Test\" ."),
		)
		resource.SetMetadata("category", "test")
		resource.SetMetadata("priority", "high")

		err := indexer.AddResource(resource)
		if err != nil {
			t.Fatalf("Failed to add resource to index: %v", err)
		}

		// Verify resource is in index
		entry, exists := indexer.FindByID("test-resource-1")
		if !exists {
			t.Fatal("Resource not found in index")
		}

		if entry.ID != "test-resource-1" {
			t.Errorf("Expected ID 'test-resource-1', got '%s'", entry.ID)
		}
		if entry.ContentType != "text/turtle" {
			t.Errorf("Expected content type 'text/turtle', got '%s'", entry.ContentType)
		}
		if entry.Size != len(resource.GetData()) {
			t.Errorf("Expected size %d, got %d", len(resource.GetData()), entry.Size)
		}
	})

	// Test finding by content type
	t.Run("FindByContentType", func(t *testing.T) {
		// Add more resources
		resource2 := domain.NewResource(
			"test-resource-2",
			"application/ld+json",
			[]byte(`{"@context": "http://example.org/", "@id": "test", "name": "Test"}`),
		)
		resource3 := domain.NewResource(
			"test-resource-3",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test2 ex:name \"Test2\" ."),
		)

		indexer.AddResource(resource2)
		indexer.AddResource(resource3)

		// Find Turtle resources
		turtleResources := indexer.FindByContentType("text/turtle")
		if len(turtleResources) != 2 {
			t.Errorf("Expected 2 turtle resources, got %d", len(turtleResources))
		}

		// Find JSON-LD resources
		jsonLDResources := indexer.FindByContentType("application/ld+json")
		if len(jsonLDResources) != 1 {
			t.Errorf("Expected 1 JSON-LD resource, got %d", len(jsonLDResources))
		}
	})

	// Test finding by tag
	t.Run("FindByTag", func(t *testing.T) {
		resource4 := domain.NewResource(
			"test-resource-4",
			"text/turtle",
			[]byte("@prefix ex: <http://example.org/> .\nex:test3 ex:name \"Test3\" ."),
		)
		resource4.SetMetadata("category", "test")
		resource4.SetMetadata("priority", "low")

		indexer.AddResource(resource4)

		// Find by category tag
		categoryResources := indexer.FindByTag("category", "test")
		if len(categoryResources) != 2 { // test-resource-1 and test-resource-4
			t.Errorf("Expected 2 resources with category 'test', got %d", len(categoryResources))
		}

		// Find by priority tag
		highPriorityResources := indexer.FindByTag("priority", "high")
		if len(highPriorityResources) != 1 {
			t.Errorf("Expected 1 resource with priority 'high', got %d", len(highPriorityResources))
		}
	})

	// Test size range queries
	t.Run("FindBySizeRange", func(t *testing.T) {
		smallResources := indexer.FindBySizeRange(0, 100)
		largeResources := indexer.FindBySizeRange(100, 1000)

		totalResources := len(smallResources) + len(largeResources)
		allResources := indexer.ListAll()

		if totalResources != len(allResources) {
			t.Errorf("Size range queries should cover all resources")
		}
	})

	// Test removing resources
	t.Run("RemoveResource", func(t *testing.T) {
		err := indexer.RemoveResource("test-resource-1")
		if err != nil {
			t.Fatalf("Failed to remove resource from index: %v", err)
		}

		// Verify resource is removed
		_, exists := indexer.FindByID("test-resource-1")
		if exists {
			t.Error("Resource should have been removed from index")
		}
	})

	// Test statistics
	t.Run("GetStats", func(t *testing.T) {
		stats := indexer.GetStats()

		totalResources, ok := stats["totalResources"].(int)
		if !ok || totalResources < 0 {
			t.Error("Stats should include valid totalResources")
		}

		contentTypes, ok := stats["contentTypes"].(map[string]int)
		if !ok {
			t.Error("Stats should include contentTypes map")
		}

		if len(contentTypes) == 0 {
			t.Error("Stats should include content type counts")
		}

		totalSize, ok := stats["totalSize"].(int)
		if !ok || totalSize < 0 {
			t.Error("Stats should include valid totalSize")
		}
	})
}

func TestResourceIndexerPersistence(t *testing.T) {
	tempDir := t.TempDir()

	// Create indexer and add resources
	indexer1, err := NewResourceIndexer(tempDir)
	if err != nil {
		t.Fatalf("Failed to create first indexer: %v", err)
	}

	resource := domain.NewResource(
		"persistent-resource",
		"text/turtle",
		[]byte("@prefix ex: <http://example.org/> .\nex:persistent ex:name \"Persistent\" ."),
	)
	resource.SetMetadata("persistent", "true")

	err = indexer1.AddResource(resource)
	if err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	// Create new indexer instance (should load existing index)
	indexer2, err := NewResourceIndexer(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second indexer: %v", err)
	}

	// Verify resource is still in index
	entry, exists := indexer2.FindByID("persistent-resource")
	if !exists {
		t.Fatal("Resource should persist across indexer instances")
	}

	if entry.ID != "persistent-resource" {
		t.Errorf("Expected ID 'persistent-resource', got '%s'", entry.ID)
	}
}

func TestResourceIndexerRebuild(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file system repository and add some resources
	repo, err := NewFileSystemRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()
	resources := []*domain.Resource{
		domain.NewResource("rebuild-1", "text/turtle", []byte("@prefix ex: <http://example.org/> .\nex:rebuild1 ex:name \"Rebuild1\" .")),
		domain.NewResource("rebuild-2", "application/ld+json", []byte(`{"@context": "http://example.org/", "@id": "rebuild2", "name": "Rebuild2"}`)),
		domain.NewResource("rebuild-3", "text/turtle", []byte("@prefix ex: <http://example.org/> .\nex:rebuild3 ex:name \"Rebuild3\" .")),
	}

	for _, resource := range resources {
		if err := repo.Store(ctx, resource); err != nil {
			t.Fatalf("Failed to store resource: %v", err)
		}
	}

	// Create indexer and rebuild
	indexer, err := NewResourceIndexer(tempDir)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	err = indexer.Rebuild(ctx)
	if err != nil {
		t.Fatalf("Failed to rebuild index: %v", err)
	}

	// Verify all resources are indexed
	for _, resource := range resources {
		entry, exists := indexer.FindByID(resource.ID())
		if !exists {
			t.Errorf("Resource %s not found after rebuild", resource.ID())
		}
		if entry.ContentType != resource.GetContentType() {
			t.Errorf("Content type mismatch for %s: expected %s, got %s",
				resource.ID(), resource.GetContentType(), entry.ContentType)
		}
	}

	// Verify stats
	stats := indexer.GetStats()
	totalResources := stats["totalResources"].(int)
	if totalResources != len(resources) {
		t.Errorf("Expected %d resources in index, got %d", len(resources), totalResources)
	}
}

func TestResourceIndexerConcurrency(t *testing.T) {
	tempDir := t.TempDir()

	indexer, err := NewResourceIndexer(tempDir)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	// Test concurrent adds
	numGoroutines := 10
	resourcesPerGoroutine := 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*resourcesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < resourcesPerGoroutine; j++ {
				resourceID := fmt.Sprintf("concurrent-resource-%d-%d", goroutineID, j)
				resource := domain.NewResource(
					resourceID,
					"text/turtle",
					[]byte(fmt.Sprintf("@prefix ex: <http://example.org/> .\nex:resource%d%d ex:name \"Resource %d-%d\" .", goroutineID, j, goroutineID, j)),
				)

				if err := indexer.AddResource(resource); err != nil {
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

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent add operation failed: %v", err)
	default:
	}

	// Verify all resources were added
	stats := indexer.GetStats()
	totalResources := stats["totalResources"].(int)
	expectedTotal := numGoroutines * resourcesPerGoroutine

	if totalResources != expectedTotal {
		t.Errorf("Expected %d resources after concurrent adds, got %d", expectedTotal, totalResources)
	}
}
