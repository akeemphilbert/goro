package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// ResourceIndex represents an index entry for a resource
type ResourceIndex struct {
	ID          string            `json:"id"`
	ContentType string            `json:"contentType"`
	Size        int               `json:"size"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Tags        map[string]string `json:"tags"`
	Path        string            `json:"path"`
}

// ResourceIndexer provides fast lookup capabilities for resources
type ResourceIndexer struct {
	basePath string
	index    map[string]*ResourceIndex
	mu       sync.RWMutex
}

// NewResourceIndexer creates a new resource indexer
func NewResourceIndexer(basePath string) (*ResourceIndexer, error) {
	indexer := &ResourceIndexer{
		basePath: basePath,
		index:    make(map[string]*ResourceIndex),
	}

	// Load existing index
	if err := indexer.loadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return indexer, nil
}

// AddResource adds a resource to the index
func (ri *ResourceIndexer) AddResource(resource domain.Resource) error {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	// Extract tags from metadata
	tags := make(map[string]string)
	for key, value := range resource.GetMetadata() {
		if strValue, ok := value.(string); ok {
			tags[key] = strValue
		}
	}

	// Create index entry
	entry := &ResourceIndex{
		ID:          resource.ID(),
		ContentType: resource.GetContentType(),
		Size:        resource.GetSize(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        tags,
		Path:        ri.getResourcePath(resource.ID()),
	}

	// Check if resource already exists in index
	if existing, exists := ri.index[resource.ID()]; exists {
		entry.CreatedAt = existing.CreatedAt
	}

	ri.index[resource.ID()] = entry

	// Persist index
	return ri.saveIndex()
}

// RemoveResource removes a resource from the index
func (ri *ResourceIndexer) RemoveResource(id string) error {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	delete(ri.index, id)
	return ri.saveIndex()
}

// FindByID finds a resource by ID
func (ri *ResourceIndexer) FindByID(id string) (*ResourceIndex, bool) {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	entry, exists := ri.index[id]
	return entry, exists
}

// FindByContentType finds resources by content type
func (ri *ResourceIndexer) FindByContentType(contentType string) []*ResourceIndex {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	var results []*ResourceIndex
	for _, entry := range ri.index {
		if entry.ContentType == contentType {
			results = append(results, entry)
		}
	}
	return results
}

// FindByTag finds resources by tag
func (ri *ResourceIndexer) FindByTag(key, value string) []*ResourceIndex {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	var results []*ResourceIndex
	for _, entry := range ri.index {
		if tagValue, exists := entry.Tags[key]; exists && tagValue == value {
			results = append(results, entry)
		}
	}
	return results
}

// FindBySizeRange finds resources within a size range
func (ri *ResourceIndexer) FindBySizeRange(minSize, maxSize int) []*ResourceIndex {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	var results []*ResourceIndex
	for _, entry := range ri.index {
		if entry.Size >= minSize && entry.Size <= maxSize {
			results = append(results, entry)
		}
	}
	return results
}

// ListAll returns all indexed resources
func (ri *ResourceIndexer) ListAll() []*ResourceIndex {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	var results []*ResourceIndex
	for _, entry := range ri.index {
		results = append(results, entry)
	}
	return results
}

// GetStats returns indexing statistics
func (ri *ResourceIndexer) GetStats() map[string]interface{} {
	ri.mu.RLock()
	defer ri.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["totalResources"] = len(ri.index)

	// Count by content type
	contentTypes := make(map[string]int)
	totalSize := 0
	for _, entry := range ri.index {
		contentTypes[entry.ContentType]++
		totalSize += entry.Size
	}

	stats["contentTypes"] = contentTypes
	stats["totalSize"] = totalSize
	stats["averageSize"] = 0
	if len(ri.index) > 0 {
		stats["averageSize"] = totalSize / len(ri.index)
	}

	return stats
}

// Rebuild rebuilds the index by scanning the file system
func (ri *ResourceIndexer) Rebuild(ctx context.Context) error {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	// Clear existing index
	ri.index = make(map[string]*ResourceIndex)

	// Scan resources directory
	resourcesDir := filepath.Join(ri.basePath, "resources")
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		return ri.saveIndex() // No resources directory, save empty index
	}

	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("failed to read resources directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		resourceID := entry.Name()
		metadataPath := filepath.Join(resourcesDir, resourceID, "metadata.json")

		// Read metadata file
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			continue // Skip resources without metadata
		}

		var metadata ResourceMetadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			continue // Skip resources with invalid metadata
		}

		// Create index entry
		tags := make(map[string]string)
		for key, value := range metadata.Tags {
			if strValue, ok := value.(string); ok {
				tags[key] = strValue
			}
		}

		indexEntry := &ResourceIndex{
			ID:          metadata.ID,
			ContentType: metadata.ContentType,
			Size:        metadata.Size,
			CreatedAt:   metadata.CreatedAt,
			UpdatedAt:   metadata.UpdatedAt,
			Tags:        tags,
			Path:        ri.getResourcePath(metadata.ID),
		}

		ri.index[metadata.ID] = indexEntry
	}

	return ri.saveIndex()
}

// loadIndex loads the index from disk
func (ri *ResourceIndexer) loadIndex() error {
	indexPath := ri.getIndexPath()

	// Check if index file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// No index file exists, start with empty index
		return nil
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index file: %w", err)
	}

	var indexData map[string]*ResourceIndex
	if err := json.Unmarshal(data, &indexData); err != nil {
		return fmt.Errorf("failed to unmarshal index data: %w", err)
	}

	ri.index = indexData
	return nil
}

// saveIndex saves the index to disk
func (ri *ResourceIndexer) saveIndex() error {
	indexPath := ri.getIndexPath()

	// Ensure index directory exists
	indexDir := filepath.Dir(indexPath)
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	data, err := json.MarshalIndent(ri.index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index data: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

// getIndexPath returns the path to the index file
func (ri *ResourceIndexer) getIndexPath() string {
	return filepath.Join(ri.basePath, "index", "resources.json")
}

// getResourcePath returns the file system path for a resource
func (ri *ResourceIndexer) getResourcePath(id string) string {
	sanitizedID := ri.sanitizeID(id)
	return filepath.Join(ri.basePath, "resources", sanitizedID)
}

// sanitizeID sanitizes a resource ID for safe file system usage
func (ri *ResourceIndexer) sanitizeID(id string) string {
	sanitized := strings.ReplaceAll(id, "..", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	return sanitized
}
