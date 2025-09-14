package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// FileSystemRepository implements StreamingResourceRepository using file system storage
type FileSystemRepository struct {
	basePath string
}

// NewFileSystemRepository creates a new FileSystemRepository
func NewFileSystemRepository(basePath string) (*FileSystemRepository, error) {
	if basePath == "" {
		return nil, fmt.Errorf("base path cannot be empty")
	}

	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &FileSystemRepository{
		basePath: basePath,
	}, nil
}

// NewFileSystemRepositoryProvider provides a FileSystemRepository for Wire dependency injection
// This function uses a default base path for the repository
func NewFileSystemRepositoryProvider() (domain.StreamingResourceRepository, error) {
	// Use a default base path - in production this should come from configuration
	basePath := "./data/pod-storage"
	repo, err := NewFileSystemRepository(basePath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// NewFileSystemRepositoryWithPath provides a FileSystemRepository with a specific base path
func NewFileSystemRepositoryWithPath(basePath string) (domain.StreamingResourceRepository, error) {
	repo, err := NewFileSystemRepository(basePath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// Store saves a resource to the file system with metadata and checksum validation
func (r *FileSystemRepository) Store(ctx context.Context, resource *domain.Resource) error {
	if resource == nil {
		return domain.WrapStorageError(
			fmt.Errorf("resource cannot be nil"),
			domain.ErrInvalidResource.Code,
			"resource cannot be nil",
		).WithOperation("Store")
	}

	if resource.ID() == "" {
		return domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("Store")
	}

	// Create resource directory
	resourceDir := r.getResourcePath(resource.ID())
	if err := os.MkdirAll(resourceDir, 0755); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to create resource directory",
		).WithOperation("Store").WithContext("resourceID", resource.ID())
	}

	// Generate checksum for data integrity
	checksum := r.generateChecksum(resource.GetData())

	// Create metadata
	metadata := r.createMetadata(resource, checksum)

	// Store content file
	contentPath := filepath.Join(resourceDir, "content")
	if err := r.writeFile(contentPath, resource.GetData()); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to write content file",
		).WithOperation("Store").WithContext("resourceID", resource.ID())
	}

	// Store metadata file
	metadataPath := filepath.Join(resourceDir, "metadata.json")
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to marshal metadata",
		).WithOperation("Store").WithContext("resourceID", resource.ID())
	}

	if err := r.writeFile(metadataPath, metadataBytes); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to write metadata file",
		).WithOperation("Store").WithContext("resourceID", resource.ID())
	}

	return nil
}

// Retrieve loads a resource from the file system with checksum validation
func (r *FileSystemRepository) Retrieve(ctx context.Context, id string) (*domain.Resource, error) {
	if id == "" {
		return nil, domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("Retrieve")
	}

	resourceDir := r.getResourcePath(id)

	// Check if resource exists
	if !r.resourceExists(resourceDir) {
		return nil, domain.WrapStorageError(
			fmt.Errorf("resource not found"),
			domain.ErrResourceNotFound.Code,
			"resource not found",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	// Read metadata
	metadataPath := filepath.Join(resourceDir, "metadata.json")
	metadataBytes, err := r.readFile(metadataPath)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to read metadata file",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	var metadata ResourceMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to unmarshal metadata",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	// Read content
	contentPath := filepath.Join(resourceDir, "content")
	data, err := r.readFile(contentPath)
	if err != nil {
		return nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to read content file",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	// Validate checksum for data integrity
	expectedChecksum := metadata.Checksum
	actualChecksum := r.generateChecksum(data)
	if expectedChecksum != actualChecksum {
		return nil, domain.WrapStorageError(
			fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum),
			domain.ErrChecksumMismatch.Code,
			"data integrity check failed",
		).WithOperation("Retrieve").WithContext("resourceID", id)
	}

	// Create resource from stored data
	resource := domain.NewResource(id, metadata.ContentType, data)

	// Restore metadata
	for key, value := range metadata.Tags {
		resource.SetMetadata(key, value)
	}
	resource.SetMetadata("originalFormat", metadata.OriginalFormat)
	resource.SetMetadata("size", metadata.Size)
	resource.SetMetadata("checksum", metadata.Checksum)
	resource.SetMetadata("createdAt", metadata.CreatedAt)
	resource.SetMetadata("updatedAt", metadata.UpdatedAt)

	return resource, nil
}

// Delete removes a resource from the file system
func (r *FileSystemRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("Delete")
	}

	resourceDir := r.getResourcePath(id)

	// Check if resource exists
	if !r.resourceExists(resourceDir) {
		return domain.WrapStorageError(
			fmt.Errorf("resource not found"),
			domain.ErrResourceNotFound.Code,
			"resource not found",
		).WithOperation("Delete").WithContext("resourceID", id)
	}

	// Remove the entire resource directory
	if err := os.RemoveAll(resourceDir); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to delete resource directory",
		).WithOperation("Delete").WithContext("resourceID", id)
	}

	return nil
}

// Exists checks if a resource exists in the file system
func (r *FileSystemRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("Exists")
	}

	resourceDir := r.getResourcePath(id)
	return r.resourceExists(resourceDir), nil
}

// Helper methods

// getResourcePath returns the file system path for a resource
func (r *FileSystemRepository) getResourcePath(id string) string {
	// Sanitize the ID to prevent directory traversal attacks
	sanitizedID := r.sanitizeID(id)
	return filepath.Join(r.basePath, "resources", sanitizedID)
}

// sanitizeID sanitizes a resource ID for safe file system usage
func (r *FileSystemRepository) sanitizeID(id string) string {
	// Replace any potentially dangerous characters
	sanitized := strings.ReplaceAll(id, "..", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	return sanitized
}

// resourceExists checks if a resource directory exists and contains required files
func (r *FileSystemRepository) resourceExists(resourceDir string) bool {
	// Check if directory exists
	if _, err := os.Stat(resourceDir); os.IsNotExist(err) {
		return false
	}

	// Check if required files exist
	contentPath := filepath.Join(resourceDir, "content")
	metadataPath := filepath.Join(resourceDir, "metadata.json")

	if _, err := os.Stat(contentPath); os.IsNotExist(err) {
		return false
	}

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// generateChecksum generates a SHA-256 checksum for data integrity
func (r *FileSystemRepository) generateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// createMetadata creates metadata for a resource
func (r *FileSystemRepository) createMetadata(resource *domain.Resource, checksum string) ResourceMetadata {
	now := time.Now()

	// Extract original format from metadata if available
	originalFormat := resource.GetContentType()
	if format, exists := resource.GetMetadata()["originalFormat"]; exists {
		if formatStr, ok := format.(string); ok {
			originalFormat = formatStr
		}
	}

	return ResourceMetadata{
		ID:             resource.ID(),
		ContentType:    resource.GetContentType(),
		OriginalFormat: originalFormat,
		Size:           len(resource.GetData()),
		Checksum:       checksum,
		CreatedAt:      now,
		UpdatedAt:      now,
		Tags:           resource.GetMetadata(),
	}
}

// writeFile writes data to a file with proper error handling
func (r *FileSystemRepository) writeFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", path, err)
	}

	return nil
}

// readFile reads data from a file with proper error handling
func (r *FileSystemRepository) readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return data, nil
}

// StoreStream stores a resource from a stream with efficient memory usage
func (r *FileSystemRepository) StoreStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) error {
	if id == "" {
		return domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("StoreStream")
	}

	if reader == nil {
		return domain.WrapStorageError(
			fmt.Errorf("reader cannot be nil"),
			domain.ErrInvalidResource.Code,
			"reader cannot be nil",
		).WithOperation("StoreStream")
	}

	// Create resource directory
	resourceDir := r.getResourcePath(id)
	if err := os.MkdirAll(resourceDir, 0755); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to create resource directory",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}

	// Create content file for streaming write
	contentPath := filepath.Join(resourceDir, "content")
	contentFile, err := os.Create(contentPath)
	if err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to create content file",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}
	defer contentFile.Close()

	// Create a hash writer to calculate checksum while streaming
	hasher := sha256.New()
	multiWriter := io.MultiWriter(contentFile, hasher)

	// Stream data from reader to file and hasher
	bytesWritten, err := io.Copy(multiWriter, reader)
	if err != nil {
		// Clean up on error
		os.Remove(contentPath)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to stream data to file",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}

	// Sync file to ensure data is written to disk
	if err := contentFile.Sync(); err != nil {
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to sync content file",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}

	// Generate checksum from hasher
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Create metadata
	now := time.Now()
	metadata := domain.ResourceMetadata{
		ID:             id,
		ContentType:    contentType,
		OriginalFormat: contentType,
		Size:           bytesWritten,
		Checksum:       checksum,
		CreatedAt:      now,
		UpdatedAt:      now,
		Tags:           make(map[string]interface{}),
	}

	// Store metadata file
	metadataPath := filepath.Join(resourceDir, "metadata.json")
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		// Clean up on error
		os.Remove(contentPath)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to marshal metadata",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}

	if err := r.writeFile(metadataPath, metadataBytes); err != nil {
		// Clean up on error
		os.Remove(contentPath)
		return domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to write metadata file",
		).WithOperation("StoreStream").WithContext("resourceID", id)
	}

	return nil
}

// RetrieveStream retrieves a resource as a stream for efficient memory usage
func (r *FileSystemRepository) RetrieveStream(ctx context.Context, id string) (io.ReadCloser, *domain.ResourceMetadata, error) {
	if id == "" {
		return nil, nil, domain.WrapStorageError(
			fmt.Errorf("resource ID cannot be empty"),
			domain.ErrInvalidID.Code,
			"resource ID cannot be empty",
		).WithOperation("RetrieveStream")
	}

	resourceDir := r.getResourcePath(id)

	// Check if resource exists
	if !r.resourceExists(resourceDir) {
		return nil, nil, domain.WrapStorageError(
			fmt.Errorf("resource not found"),
			domain.ErrResourceNotFound.Code,
			"resource not found",
		).WithOperation("RetrieveStream").WithContext("resourceID", id)
	}

	// Read metadata first
	metadataPath := filepath.Join(resourceDir, "metadata.json")
	metadataBytes, err := r.readFile(metadataPath)
	if err != nil {
		return nil, nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to read metadata file",
		).WithOperation("RetrieveStream").WithContext("resourceID", id)
	}

	var metadata domain.ResourceMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to unmarshal metadata",
		).WithOperation("RetrieveStream").WithContext("resourceID", id)
	}

	// Open content file for streaming read
	contentPath := filepath.Join(resourceDir, "content")
	contentFile, err := os.Open(contentPath)
	if err != nil {
		return nil, nil, domain.WrapStorageError(
			err,
			domain.ErrStorageOperation.Code,
			"failed to open content file",
		).WithOperation("RetrieveStream").WithContext("resourceID", id)
	}

	// Create a validating reader that checks checksum while streaming
	validatingReader := &checksumValidatingReader{
		reader:           contentFile,
		hasher:           sha256.New(),
		expectedChecksum: metadata.Checksum,
		resourceID:       id,
	}

	return validatingReader, &metadata, nil
}

// checksumValidatingReader wraps a reader to validate checksum while streaming
type checksumValidatingReader struct {
	reader           io.ReadCloser
	hasher           hash.Hash
	expectedChecksum string
	resourceID       string
	validated        bool
}

func (r *checksumValidatingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		// Write to hasher for checksum calculation
		r.hasher.Write(p[:n])
	}

	// If we've reached EOF, validate the checksum
	if err == io.EOF && !r.validated {
		r.validated = true
		actualChecksum := hex.EncodeToString(r.hasher.Sum(nil))
		if actualChecksum != r.expectedChecksum {
			return n, domain.WrapStorageError(
				fmt.Errorf("checksum mismatch: expected %s, got %s", r.expectedChecksum, actualChecksum),
				domain.ErrChecksumMismatch.Code,
				"data integrity check failed during streaming",
			).WithOperation("RetrieveStream").WithContext("resourceID", r.resourceID)
		}
	}

	return n, err
}

func (r *checksumValidatingReader) Close() error {
	return r.reader.Close()
}

// ResourceMetadata represents the metadata stored alongside resource content
type ResourceMetadata struct {
	ID             string                 `json:"id"`
	ContentType    string                 `json:"contentType"`
	OriginalFormat string                 `json:"originalFormat"`
	Size           int                    `json:"size"`
	Checksum       string                 `json:"checksum"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Tags           map[string]interface{} `json:"tags"`
}
