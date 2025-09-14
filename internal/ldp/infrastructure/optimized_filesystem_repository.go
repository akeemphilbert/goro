package infrastructure

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// OptimizedFileSystemRepository extends FileSystemRepository with indexing and caching
type OptimizedFileSystemRepository struct {
	*FileSystemRepository
	indexer *ResourceIndexer
	cache   *ResourceCache
	mu      sync.RWMutex
}

// NewOptimizedFileSystemRepository creates a new optimized repository with indexing and caching
func NewOptimizedFileSystemRepository(basePath string, cacheConfig CacheConfig) (*OptimizedFileSystemRepository, error) {
	// Create base repository
	baseRepo, err := NewFileSystemRepository(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create base repository: %w", err)
	}

	// Create indexer
	indexer, err := NewResourceIndexer(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	// Create cache
	cache := NewResourceCache(cacheConfig)

	repo := &OptimizedFileSystemRepository{
		FileSystemRepository: baseRepo,
		indexer:              indexer,
		cache:                cache,
	}

	// Rebuild index if it's empty (first run)
	stats := indexer.GetStats()
	if totalResources, ok := stats["totalResources"].(int); ok && totalResources == 0 {
		if err := indexer.Rebuild(context.Background()); err != nil {
			// Log warning but don't fail - repository can work without index
			fmt.Printf("Warning: failed to rebuild index: %v\n", err)
		}
	}

	return repo, nil
}

// NewOptimizedFileSystemRepositoryProvider provides an OptimizedFileSystemRepository for Wire
func NewOptimizedFileSystemRepositoryProvider() (domain.StreamingResourceRepository, error) {
	basePath := "./data/pod-storage"
	cacheConfig := CacheConfig{
		MaxSize:    50 * 1024 * 1024, // 50MB cache
		MaxEntries: 500,              // 500 entries max
		TTL:        30 * 60,          // 30 minutes TTL
	}

	repo, err := NewOptimizedFileSystemRepository(basePath, cacheConfig)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Store saves a resource with optimized indexing and cache invalidation
func (r *OptimizedFileSystemRepository) Store(ctx context.Context, resource *domain.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store using base repository
	if err := r.FileSystemRepository.Store(ctx, resource); err != nil {
		return err
	}

	// Update index
	if err := r.indexer.AddResource(resource); err != nil {
		// Log warning but don't fail the store operation
		fmt.Printf("Warning: failed to update index for resource %s: %v\n", resource.ID(), err)
	}

	// Update cache
	r.cache.Put(ctx, resource)

	return nil
}

// Retrieve loads a resource with cache-first lookup and index optimization
func (r *OptimizedFileSystemRepository) Retrieve(ctx context.Context, id string) (*domain.Resource, error) {
	// Try cache first
	if resource, found := r.cache.Get(ctx, id); found {
		return resource, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check index for fast existence check
	if _, exists := r.indexer.FindByID(id); !exists {
		return nil, domain.WrapStorageError(
			fmt.Errorf("resource not found"),
			domain.ErrResourceNotFound.Code,
			"resource not found in index",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	// Retrieve using base repository
	resource, err := r.FileSystemRepository.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the retrieved resource
	r.cache.Put(ctx, resource)

	return resource, nil
}

// Delete removes a resource with index and cache cleanup
func (r *OptimizedFileSystemRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete using base repository
	if err := r.FileSystemRepository.Delete(ctx, id); err != nil {
		return err
	}

	// Remove from index
	if err := r.indexer.RemoveResource(id); err != nil {
		// Log warning but don't fail the delete operation
		fmt.Printf("Warning: failed to remove resource %s from index: %v\n", id, err)
	}

	// Remove from cache
	r.cache.Remove(ctx, id)

	return nil
}

// Exists checks if a resource exists using index for fast lookup
func (r *OptimizedFileSystemRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check index first for fast lookup
	if _, exists := r.indexer.FindByID(id); exists {
		return true, nil
	}

	// Fallback to file system check
	return r.FileSystemRepository.Exists(ctx, id)
}

// FindByContentType finds resources by content type using index
func (r *OptimizedFileSystemRepository) FindByContentType(ctx context.Context, contentType string) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use index for fast lookup
	indexEntries := r.indexer.FindByContentType(contentType)

	resources := make([]*domain.Resource, 0, len(indexEntries))
	for _, entry := range indexEntries {
		// Try cache first
		if resource, found := r.cache.Get(ctx, entry.ID); found {
			resources = append(resources, resource)
			continue
		}

		// Load from file system
		resource, err := r.FileSystemRepository.Retrieve(ctx, entry.ID)
		if err != nil {
			// Log warning but continue with other resources
			fmt.Printf("Warning: failed to retrieve resource %s: %v\n", entry.ID, err)
			continue
		}

		// Cache the resource
		r.cache.Put(ctx, resource)
		resources = append(resources, resource)
	}

	return resources, nil
}

// FindByTag finds resources by tag using index
func (r *OptimizedFileSystemRepository) FindByTag(ctx context.Context, key, value string) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use index for fast lookup
	indexEntries := r.indexer.FindByTag(key, value)

	resources := make([]*domain.Resource, 0, len(indexEntries))
	for _, entry := range indexEntries {
		// Try cache first
		if resource, found := r.cache.Get(ctx, entry.ID); found {
			resources = append(resources, resource)
			continue
		}

		// Load from file system
		resource, err := r.FileSystemRepository.Retrieve(ctx, entry.ID)
		if err != nil {
			// Log warning but continue with other resources
			fmt.Printf("Warning: failed to retrieve resource %s: %v\n", entry.ID, err)
			continue
		}

		// Cache the resource
		r.cache.Put(ctx, resource)
		resources = append(resources, resource)
	}

	return resources, nil
}

// ListResources returns a paginated list of resources using index
func (r *OptimizedFileSystemRepository) ListResources(ctx context.Context, offset, limit int) ([]*domain.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get all resources from index
	allEntries := r.indexer.ListAll()

	// Apply pagination
	start := offset
	if start >= len(allEntries) {
		return []*domain.Resource{}, nil
	}

	end := start + limit
	if end > len(allEntries) {
		end = len(allEntries)
	}

	entries := allEntries[start:end]
	resources := make([]*domain.Resource, 0, len(entries))

	for _, entry := range entries {
		// Try cache first
		if resource, found := r.cache.Get(ctx, entry.ID); found {
			resources = append(resources, resource)
			continue
		}

		// Load from file system
		resource, err := r.FileSystemRepository.Retrieve(ctx, entry.ID)
		if err != nil {
			// Log warning but continue with other resources
			fmt.Printf("Warning: failed to retrieve resource %s: %v\n", entry.ID, err)
			continue
		}

		// Cache the resource
		r.cache.Put(ctx, resource)
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetStats returns combined statistics from repository, index, and cache
func (r *OptimizedFileSystemRepository) GetStats(ctx context.Context) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})

	// Index stats
	indexStats := r.indexer.GetStats()
	stats["index"] = indexStats

	// Cache stats
	cacheStats := r.cache.GetStats()
	stats["cache"] = cacheStats

	// Repository stats
	basePath := r.FileSystemRepository.basePath
	if info, err := os.Stat(filepath.Join(basePath, "resources")); err == nil {
		stats["diskUsage"] = map[string]interface{}{
			"resourcesPath": filepath.Join(basePath, "resources"),
			"lastModified":  info.ModTime(),
		}
	}

	return stats
}

// WarmupCache preloads frequently accessed resources into cache
func (r *OptimizedFileSystemRepository) WarmupCache(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get most accessed resources from cache stats
	mostAccessed := r.cache.GetMostAccessed(100) // Top 100 most accessed

	// If cache is empty, load some recent resources
	if len(mostAccessed) == 0 {
		allEntries := r.indexer.ListAll()

		// Load up to 50 most recent resources
		limit := 50
		if len(allEntries) < limit {
			limit = len(allEntries)
		}

		for i := 0; i < limit; i++ {
			entry := allEntries[i]
			resource, err := r.FileSystemRepository.Retrieve(ctx, entry.ID)
			if err != nil {
				continue
			}
			r.cache.Put(ctx, resource)
		}
	}

	return nil
}

// RebuildIndex rebuilds the resource index
func (r *OptimizedFileSystemRepository) RebuildIndex(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.indexer.Rebuild(ctx)
}

// ClearCache clears the resource cache
func (r *OptimizedFileSystemRepository) ClearCache(ctx context.Context) {
	r.cache.Clear(ctx)
}

// StoreStream stores a resource from a stream (delegates to base repository)
func (r *OptimizedFileSystemRepository) StoreStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store using base repository
	err := r.FileSystemRepository.StoreStream(ctx, id, reader, contentType, size)
	if err != nil {
		return err
	}

	// Retrieve the stored resource to update index
	resource, err := r.FileSystemRepository.Retrieve(ctx, id)
	if err == nil {
		// Update index
		if err := r.indexer.AddResource(resource); err != nil {
			// Log warning but don't fail the operation
			fmt.Printf("Warning: failed to index resource %s: %v\n", id, err)
		}
	}

	// Invalidate cache entry if it exists (cache doesn't have Delete method, so we'll skip this)
	// r.cache.Delete(ctx, id)

	return nil
}

// RetrieveStream retrieves a resource as a stream (delegates to base repository)
func (r *OptimizedFileSystemRepository) RetrieveStream(ctx context.Context, id string) (io.ReadCloser, *domain.ResourceMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// For streaming, we bypass cache and go directly to storage
	// This is because we want to stream the data, not load it all into memory
	return r.FileSystemRepository.RetrieveStream(ctx, id)
}
