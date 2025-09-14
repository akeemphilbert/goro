package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// ContainerCacheEntry represents a cached container entry
type ContainerCacheEntry struct {
	Container    *domain.Container
	MemberCount  int
	TotalSize    int64
	LastAccessed time.Time
	ExpiresAt    time.Time
}

// IsExpired checks if the cache entry has expired
func (e *ContainerCacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ContainerCache provides caching for container data and statistics
type ContainerCache struct {
	entries    map[string]*ContainerCacheEntry
	mu         sync.RWMutex
	ttl        time.Duration
	maxEntries int
}

// NewContainerCache creates a new container cache
func NewContainerCache(ttl time.Duration, maxEntries int) *ContainerCache {
	cache := &ContainerCache{
		entries:    make(map[string]*ContainerCacheEntry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a container from cache
func (c *ContainerCache) Get(containerID string) (*domain.Container, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[containerID]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	// Update last accessed time
	entry.LastAccessed = time.Now()
	return entry.Container, true
}

// Put stores a container in cache
func (c *ContainerCache) Put(containerID string, container *domain.Container, memberCount int, totalSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.maxEntries {
		c.evictLRU()
	}

	entry := &ContainerCacheEntry{
		Container:    container,
		MemberCount:  memberCount,
		TotalSize:    totalSize,
		LastAccessed: time.Now(),
		ExpiresAt:    time.Now().Add(c.ttl),
	}

	c.entries[containerID] = entry
}

// GetStats retrieves cached statistics for a container
func (c *ContainerCache) GetStats(containerID string) (memberCount int, totalSize int64, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[containerID]
	if !exists || entry.IsExpired() {
		return 0, 0, false
	}

	entry.LastAccessed = time.Now()
	return entry.MemberCount, entry.TotalSize, true
}

// UpdateStats updates cached statistics for a container
func (c *ContainerCache) UpdateStats(containerID string, memberCount int, totalSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[containerID]
	if exists && !entry.IsExpired() {
		entry.MemberCount = memberCount
		entry.TotalSize = totalSize
		entry.LastAccessed = time.Now()
		entry.ExpiresAt = time.Now().Add(c.ttl) // Refresh expiration
	}
}

// Invalidate removes a container from cache
func (c *ContainerCache) Invalidate(containerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, containerID)
}

// InvalidatePattern removes containers matching a pattern from cache
func (c *ContainerCache) InvalidatePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for containerID := range c.entries {
		// Simple pattern matching - could be enhanced with regex
		if containerID == pattern {
			delete(c.entries, containerID)
		}
	}
}

// Clear removes all entries from cache
func (c *ContainerCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*ContainerCacheEntry)
}

// Size returns the number of entries in cache
func (c *ContainerCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// GetCacheStats returns cache statistics
func (c *ContainerCache) GetCacheStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_entries"] = len(c.entries)
	stats["max_entries"] = c.maxEntries
	stats["ttl_seconds"] = c.ttl.Seconds()

	// Count expired entries
	expiredCount := 0
	for _, entry := range c.entries {
		if entry.IsExpired() {
			expiredCount++
		}
	}
	stats["expired_entries"] = expiredCount

	return stats
}

// evictLRU evicts the least recently used entry
func (c *ContainerCache) evictLRU() {
	var oldestID string
	var oldestTime time.Time

	for containerID, entry := range c.entries {
		if oldestID == "" || entry.LastAccessed.Before(oldestTime) {
			oldestID = containerID
			oldestTime = entry.LastAccessed
		}
	}

	if oldestID != "" {
		delete(c.entries, oldestID)
	}
}

// cleanupExpired removes expired entries periodically
func (c *ContainerCache) cleanupExpired() {
	ticker := time.NewTicker(c.ttl / 2) // Cleanup every half TTL
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for containerID, entry := range c.entries {
			if entry.IsExpired() {
				delete(c.entries, containerID)
			}
		}
		c.mu.Unlock()
	}
}

// CachedContainerRepository wraps a container repository with caching
type CachedContainerRepository struct {
	repo  domain.ContainerRepository
	cache *ContainerCache
}

// NewCachedContainerRepository creates a new cached container repository
func NewCachedContainerRepository(repo domain.ContainerRepository, cache *ContainerCache) *CachedContainerRepository {
	return &CachedContainerRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetContainer retrieves a container with caching
func (r *CachedContainerRepository) GetContainer(ctx context.Context, id string) (domain.ContainerResource, error) {
	// Try cache first
	if container, found := r.cache.Get(id); found {
		return container, nil
	}

	// Cache miss - get from repository
	container, err := r.repo.GetContainer(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get member count for caching
	memberCount := container.GetMemberCount()
	totalSize := int64(0) // Would be calculated from actual member sizes

	// Cache the result - type assert to concrete Container for caching
	if concreteContainer, ok := container.(*domain.Container); ok {
		r.cache.Put(id, concreteContainer, memberCount, totalSize)
	}

	return container, nil
}

// CreateContainer creates a container and invalidates cache
func (r *CachedContainerRepository) CreateContainer(ctx context.Context, container domain.ContainerResource) error {
	err := r.repo.CreateContainer(ctx, container)
	if err != nil {
		return err
	}

	// Invalidate parent container cache if it exists
	if container.GetParentID() != "" {
		r.cache.Invalidate(container.GetParentID())
	}

	return nil
}

// UpdateContainer updates a container and invalidates cache
func (r *CachedContainerRepository) UpdateContainer(ctx context.Context, container *domain.Container) error {
	err := r.repo.UpdateContainer(ctx, container)
	if err != nil {
		return err
	}

	// Invalidate cache
	r.cache.Invalidate(container.ID())

	return nil
}

// DeleteContainer deletes a container and invalidates cache
func (r *CachedContainerRepository) DeleteContainer(ctx context.Context, id string) error {
	err := r.repo.DeleteContainer(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache
	r.cache.Invalidate(id)

	return nil
}

// AddMember adds a member and invalidates container cache
func (r *CachedContainerRepository) AddMember(ctx context.Context, containerID, memberID string) error {
	err := r.repo.AddMember(ctx, containerID, memberID)
	if err != nil {
		return err
	}

	// Invalidate container cache
	r.cache.Invalidate(containerID)

	return nil
}

// RemoveMember removes a member and invalidates container cache
func (r *CachedContainerRepository) RemoveMember(ctx context.Context, containerID, memberID string) error {
	err := r.repo.RemoveMember(ctx, containerID, memberID)
	if err != nil {
		return err
	}

	// Invalidate container cache
	r.cache.Invalidate(containerID)

	return nil
}

// ListMembers lists members (no caching for dynamic results)
func (r *CachedContainerRepository) ListMembers(ctx context.Context, containerID string, pagination domain.PaginationOptions) ([]string, error) {
	return r.repo.ListMembers(ctx, containerID, pagination)
}

// GetChildren gets children (no caching for dynamic results)
func (r *CachedContainerRepository) GetChildren(ctx context.Context, containerID string) ([]domain.ContainerResource, error) {
	return r.repo.GetChildren(ctx, containerID)
}

// GetParent gets parent with caching
func (r *CachedContainerRepository) GetParent(ctx context.Context, containerID string) (domain.ContainerResource, error) {
	return r.repo.GetParent(ctx, containerID)
}

// GetPath gets path (no caching for dynamic results)
func (r *CachedContainerRepository) GetPath(ctx context.Context, containerID string) ([]string, error) {
	return r.repo.GetPath(ctx, containerID)
}

// FindByPath finds by path (no caching for dynamic results)
func (r *CachedContainerRepository) FindByPath(ctx context.Context, path string) (domain.ContainerResource, error) {
	return r.repo.FindByPath(ctx, path)
}

// ContainerExists checks existence (no caching for dynamic results)
func (r *CachedContainerRepository) ContainerExists(ctx context.Context, id string) (bool, error) {
	return r.repo.ContainerExists(ctx, id)
}
