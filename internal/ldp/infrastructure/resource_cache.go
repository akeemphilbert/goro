package infrastructure

import (
	"context"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// CacheEntry represents a cached resource with metadata
type CacheEntry struct {
	Resource *domain.Resource
	AccessAt time.Time
	HitCount int64
	Size     int
}

// ResourceCache provides in-memory caching for frequently accessed resources
type ResourceCache struct {
	entries     map[string]*CacheEntry
	mu          sync.RWMutex
	maxSize     int64         // Maximum cache size in bytes
	currentSize int64         // Current cache size in bytes
	maxEntries  int           // Maximum number of entries
	ttl         time.Duration // Time to live for cache entries
}

// CacheConfig holds configuration for the resource cache
type CacheConfig struct {
	MaxSize    int64         // Maximum cache size in bytes (default: 100MB)
	MaxEntries int           // Maximum number of entries (default: 1000)
	TTL        time.Duration // Time to live (default: 1 hour)
}

// NewResourceCache creates a new resource cache with the given configuration
func NewResourceCache(config CacheConfig) *ResourceCache {
	// Set defaults if not provided
	if config.MaxSize <= 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = 1000
	}
	if config.TTL <= 0 {
		config.TTL = time.Hour
	}

	cache := &ResourceCache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    config.MaxSize,
		maxEntries: config.MaxEntries,
		ttl:        config.TTL,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a resource from the cache
func (rc *ResourceCache) Get(ctx context.Context, id string) (*domain.Resource, bool) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	entry, exists := rc.entries[id]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.AccessAt) > rc.ttl {
		delete(rc.entries, id)
		rc.currentSize -= int64(entry.Size)
		return nil, false
	}

	// Update access time and hit count
	entry.AccessAt = time.Now()
	entry.HitCount++

	return entry.Resource, true
}

// Put stores a resource in the cache
func (rc *ResourceCache) Put(ctx context.Context, resource *domain.Resource) {
	if resource == nil {
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	resourceSize := resource.GetSize()

	// Check if resource is too large for cache
	if int64(resourceSize) > rc.maxSize/2 {
		return // Don't cache resources larger than half the cache size
	}

	// Remove existing entry if it exists
	if existing, exists := rc.entries[resource.ID()]; exists {
		rc.currentSize -= int64(existing.Size)
	}

	// Ensure we have space for the new entry
	rc.ensureSpace(int64(resourceSize))

	// Create new cache entry
	entry := &CacheEntry{
		Resource: resource,
		AccessAt: time.Now(),
		HitCount: 1,
		Size:     resourceSize,
	}

	rc.entries[resource.ID()] = entry
	rc.currentSize += int64(resourceSize)
}

// Remove removes a resource from the cache
func (rc *ResourceCache) Remove(ctx context.Context, id string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if entry, exists := rc.entries[id]; exists {
		delete(rc.entries, id)
		rc.currentSize -= int64(entry.Size)
	}
}

// Clear clears all entries from the cache
func (rc *ResourceCache) Clear(ctx context.Context) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.entries = make(map[string]*CacheEntry)
	rc.currentSize = 0
}

// GetStats returns cache statistics
func (rc *ResourceCache) GetStats() map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	totalHits := int64(0)
	oldestAccess := time.Now()
	newestAccess := time.Time{}

	for _, entry := range rc.entries {
		totalHits += entry.HitCount
		if entry.AccessAt.Before(oldestAccess) {
			oldestAccess = entry.AccessAt
		}
		if entry.AccessAt.After(newestAccess) {
			newestAccess = entry.AccessAt
		}
	}

	stats := map[string]interface{}{
		"entryCount":      len(rc.entries),
		"maxEntries":      rc.maxEntries,
		"currentSize":     rc.currentSize,
		"maxSize":         rc.maxSize,
		"sizeUtilization": float64(rc.currentSize) / float64(rc.maxSize),
		"totalHits":       totalHits,
		"averageHits":     float64(0),
		"ttl":             rc.ttl.String(),
	}

	if len(rc.entries) > 0 {
		stats["averageHits"] = float64(totalHits) / float64(len(rc.entries))
		stats["oldestAccess"] = oldestAccess
		stats["newestAccess"] = newestAccess
	}

	return stats
}

// GetMostAccessed returns the most frequently accessed resources
func (rc *ResourceCache) GetMostAccessed(limit int) []*CacheEntry {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	// Convert map to slice
	entries := make([]*CacheEntry, 0, len(rc.entries))
	for _, entry := range rc.entries {
		entries = append(entries, entry)
	}

	// Sort by hit count (descending)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].HitCount < entries[j].HitCount {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Return top entries
	if limit > len(entries) {
		limit = len(entries)
	}
	return entries[:limit]
}

// ensureSpace ensures there's enough space for a new entry
func (rc *ResourceCache) ensureSpace(requiredSize int64) {
	// Check if we need to make space
	for (rc.currentSize+requiredSize > rc.maxSize || len(rc.entries) >= rc.maxEntries) && len(rc.entries) > 0 {
		rc.evictLeastRecentlyUsed()
	}
}

// evictLeastRecentlyUsed removes the least recently used entry
func (rc *ResourceCache) evictLeastRecentlyUsed() {
	var oldestID string
	var oldestTime time.Time = time.Now()

	// Find the least recently used entry
	for id, entry := range rc.entries {
		if entry.AccessAt.Before(oldestTime) {
			oldestTime = entry.AccessAt
			oldestID = id
		}
	}

	// Remove the oldest entry
	if oldestID != "" {
		if entry, exists := rc.entries[oldestID]; exists {
			delete(rc.entries, oldestID)
			rc.currentSize -= int64(entry.Size)
		}
	}
}

// cleanupExpired removes expired entries periodically
func (rc *ResourceCache) cleanupExpired() {
	ticker := time.NewTicker(rc.ttl / 4) // Cleanup every quarter of TTL
	defer ticker.Stop()

	for range ticker.C {
		rc.mu.Lock()
		now := time.Now()

		for id, entry := range rc.entries {
			if now.Sub(entry.AccessAt) > rc.ttl {
				delete(rc.entries, id)
				rc.currentSize -= int64(entry.Size)
			}
		}

		rc.mu.Unlock()
	}
}

// Warmup preloads frequently accessed resources into the cache
func (rc *ResourceCache) Warmup(ctx context.Context, resources []*domain.Resource) {
	for _, resource := range resources {
		if resource != nil {
			rc.Put(ctx, resource)
		}
	}
}
