package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSystemRepository(t *testing.T) {
	tests := []struct {
		name        string
		basePath    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid base path",
			basePath:    t.TempDir(),
			expectError: false,
		},
		{
			name:        "empty base path",
			basePath:    "",
			expectError: true,
			errorMsg:    "base path cannot be empty",
		},
		{
			name:        "creates directory if not exists",
			basePath:    filepath.Join(t.TempDir(), "new", "directory"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewFileSystemRepository(tt.basePath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tt.basePath, repo.basePath)

				// Verify directory was created
				_, err := os.Stat(tt.basePath)
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileSystemRepository_Store(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		resource    *domain.Resource
		expectError bool
		errorCode   string
	}{
		{
			name:        "store valid RDF resource",
			resource:    createTestResource("test-1", "application/ld+json", `{"@context": "http://example.org", "name": "test"}`),
			expectError: false,
		},
		{
			name:        "store binary resource",
			resource:    createTestResource("test-2", "image/png", "binary-data-here"),
			expectError: false,
		},
		{
			name:        "store nil resource",
			resource:    nil,
			expectError: true,
			errorCode:   domain.ErrInvalidResource.Code,
		},
		{
			name:        "store resource with empty ID",
			resource:    createTestResourceWithID("", "text/plain", "test data"),
			expectError: true,
			errorCode:   domain.ErrInvalidID.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Store(ctx, tt.resource)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCode != "" {
					storageErr, ok := domain.GetStorageError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, storageErr.Code)
				}
			} else {
				assert.NoError(t, err)

				// Verify files were created
				resourceDir := repo.getResourcePath(tt.resource.ID())
				contentPath := filepath.Join(resourceDir, "content")
				metadataPath := filepath.Join(resourceDir, "metadata.json")

				assert.FileExists(t, contentPath)
				assert.FileExists(t, metadataPath)

				// Verify content
				content, err := os.ReadFile(contentPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.resource.GetData(), content)

				// Verify metadata
				metadataBytes, err := os.ReadFile(metadataPath)
				assert.NoError(t, err)

				var metadata ResourceMetadata
				err = json.Unmarshal(metadataBytes, &metadata)
				assert.NoError(t, err)
				assert.Equal(t, tt.resource.ID(), metadata.ID)
				assert.Equal(t, tt.resource.GetContentType(), metadata.ContentType)
				assert.Equal(t, len(tt.resource.GetData()), metadata.Size)
				assert.NotEmpty(t, metadata.Checksum)
			}
		})
	}
}

func TestFileSystemRepository_Retrieve(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Store a test resource first
	testResource := createTestResource("test-retrieve", "application/ld+json", `{"@context": "http://example.org", "name": "test"}`)
	err = repo.Store(ctx, testResource)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorCode   string
	}{
		{
			name:        "retrieve existing resource",
			id:          "test-retrieve",
			expectError: false,
		},
		{
			name:        "retrieve non-existent resource",
			id:          "non-existent",
			expectError: true,
			errorCode:   domain.ErrResourceNotFound.Code,
		},
		{
			name:        "retrieve with empty ID",
			id:          "",
			expectError: true,
			errorCode:   domain.ErrInvalidID.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := repo.Retrieve(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resource)
				if tt.errorCode != "" {
					storageErr, ok := domain.GetStorageError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, storageErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resource)
				assert.Equal(t, tt.id, resource.ID())
				assert.Equal(t, testResource.GetContentType(), resource.GetContentType())
				assert.Equal(t, testResource.GetData(), resource.GetData())
			}
		})
	}
}

func TestFileSystemRepository_Delete(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Store a test resource first
	testResource := createTestResource("test-delete", "text/plain", "test data")
	err = repo.Store(ctx, testResource)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorCode   string
	}{
		{
			name:        "delete existing resource",
			id:          "test-delete",
			expectError: false,
		},
		{
			name:        "delete non-existent resource",
			id:          "non-existent",
			expectError: true,
			errorCode:   domain.ErrResourceNotFound.Code,
		},
		{
			name:        "delete with empty ID",
			id:          "",
			expectError: true,
			errorCode:   domain.ErrInvalidID.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCode != "" {
					storageErr, ok := domain.GetStorageError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, storageErr.Code)
				}
			} else {
				assert.NoError(t, err)

				// Verify resource directory was deleted
				resourceDir := repo.getResourcePath(tt.id)
				_, err := os.Stat(resourceDir)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestFileSystemRepository_Exists(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Store a test resource first
	testResource := createTestResource("test-exists", "text/plain", "test data")
	err = repo.Store(ctx, testResource)
	require.NoError(t, err)

	tests := []struct {
		name         string
		id           string
		expectError  bool
		expectExists bool
		errorCode    string
	}{
		{
			name:         "existing resource",
			id:           "test-exists",
			expectError:  false,
			expectExists: true,
		},
		{
			name:         "non-existent resource",
			id:           "non-existent",
			expectError:  false,
			expectExists: false,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
			errorCode:   domain.ErrInvalidID.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.Exists(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCode != "" {
					storageErr, ok := domain.GetStorageError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errorCode, storageErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectExists, exists)
			}
		})
	}
}

func TestFileSystemRepository_ChecksumValidation(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Store a test resource
	testResource := createTestResource("test-checksum", "text/plain", "test data")
	err = repo.Store(ctx, testResource)
	require.NoError(t, err)

	// Corrupt the content file
	resourceDir := repo.getResourcePath("test-checksum")
	contentPath := filepath.Join(resourceDir, "content")
	err = os.WriteFile(contentPath, []byte("corrupted data"), 0644)
	require.NoError(t, err)

	// Try to retrieve - should fail with checksum mismatch
	_, err = repo.Retrieve(ctx, "test-checksum")
	assert.Error(t, err)

	storageErr, ok := domain.GetStorageError(err)
	require.True(t, ok)
	assert.Equal(t, domain.ErrChecksumMismatch.Code, storageErr.Code)
}

func TestFileSystemRepository_IDSanitization(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "normal ID",
			id:       "normal-id-123",
			expected: "normal-id-123",
		},
		{
			name:     "ID with directory traversal",
			id:       "../../../etc/passwd",
			expected: "______etc_passwd",
		},
		{
			name:     "ID with backslashes",
			id:       "test\\path\\id",
			expected: "test_path_id",
		},
		{
			name:     "ID with forward slashes",
			id:       "test/path/id",
			expected: "test_path_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := repo.sanitizeID(tt.id)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestFileSystemRepository_BinaryFileSupport(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with binary data (simulated image data)
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
	testResource := createTestResource("binary-test", "image/png", string(binaryData))

	// Store binary resource
	err = repo.Store(ctx, testResource)
	assert.NoError(t, err)

	// Retrieve binary resource
	retrieved, err := repo.Retrieve(ctx, "binary-test")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify binary data integrity
	assert.Equal(t, binaryData, retrieved.GetData())
	assert.Equal(t, "image/png", retrieved.GetContentType())
}

func TestFileSystemRepository_RDFFormatSupport(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	rdfFormats := []struct {
		name        string
		contentType string
		data        string
	}{
		{
			name:        "JSON-LD",
			contentType: "application/ld+json",
			data:        `{"@context": "http://example.org", "name": "test"}`,
		},
		{
			name:        "RDF/XML",
			contentType: "application/rdf+xml",
			data:        `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="http://example.org/test"><name>test</name></rdf:Description></rdf:RDF>`,
		},
		{
			name:        "Turtle",
			contentType: "text/turtle",
			data:        `@prefix ex: <http://example.org/> . ex:test ex:name "test" .`,
		},
	}

	for i, format := range rdfFormats {
		t.Run(format.name, func(t *testing.T) {
			resourceID := fmt.Sprintf("rdf-test-%d", i)
			testResource := createTestResource(resourceID, format.contentType, format.data)

			// Store RDF resource
			err := repo.Store(ctx, testResource)
			assert.NoError(t, err)

			// Retrieve RDF resource
			retrieved, err := repo.Retrieve(ctx, resourceID)
			assert.NoError(t, err)
			assert.NotNil(t, retrieved)

			// Verify RDF data integrity
			assert.Equal(t, []byte(format.data), retrieved.GetData())
			assert.Equal(t, format.contentType, retrieved.GetContentType())
		})
	}
}

func TestFileSystemRepository_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 10

	// Test concurrent stores
	t.Run("concurrent stores", func(t *testing.T) {
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				resourceID := fmt.Sprintf("concurrent-store-%d", id)
				testResource := createTestResource(resourceID, "text/plain", fmt.Sprintf("data-%d", id))

				if err := repo.Store(ctx, testResource); err != nil {
					errors <- err
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Check for errors
		select {
		case err := <-errors:
			t.Fatalf("Concurrent store failed: %v", err)
		default:
			// No errors
		}

		// Verify all resources were stored
		for i := 0; i < numGoroutines; i++ {
			resourceID := fmt.Sprintf("concurrent-store-%d", i)
			exists, err := repo.Exists(ctx, resourceID)
			assert.NoError(t, err)
			assert.True(t, exists)
		}
	})
}

func TestFileSystemRepository_MetadataPreservation(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileSystemRepository(tempDir)
	require.NoError(t, err)

	ctx := context.Background()

	// Create resource with custom metadata
	testResource := createTestResource("metadata-test", "application/ld+json", `{"@context": "http://example.org", "name": "test"}`)
	testResource.SetMetadata("customField", "customValue")
	testResource.SetMetadata("tags", []string{"tag1", "tag2"})

	// Store resource
	err = repo.Store(ctx, testResource)
	assert.NoError(t, err)

	// Retrieve resource
	retrieved, err := repo.Retrieve(ctx, "metadata-test")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify metadata preservation
	metadata := retrieved.GetMetadata()
	assert.Equal(t, "customValue", metadata["customField"])

	// Note: tags might be stored as interface{} so we need to handle type assertion
	if tags, ok := metadata["tags"]; ok {
		assert.NotNil(t, tags)
	}

	// Verify system metadata
	assert.NotNil(t, metadata["createdAt"])
	assert.NotNil(t, metadata["updatedAt"])
	assert.NotNil(t, metadata["checksum"])
	assert.Equal(t, len(testResource.GetData()), metadata["size"])
}

// Helper functions for tests

func createTestResource(id, contentType, data string) *domain.Resource {
	return domain.NewResource(id, contentType, []byte(data))
}

func createTestResourceWithID(id, contentType, data string) *domain.Resource {
	resource := &domain.Resource{}
	// This is a bit of a hack to create a resource with empty ID for testing
	// In real usage, NewResource should always be used
	return resource
}

func TestNewFileSystemRepositoryProvider(t *testing.T) {
	// Clean up any existing test data
	defer func() {
		os.RemoveAll("./data")
	}()

	repo, err := NewFileSystemRepositoryProvider()
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// Verify it implements the ResourceRepository interface
	var _ domain.ResourceRepository = repo

	// Test basic functionality
	ctx := context.Background()
	testResource := createTestResource("provider-test", "text/plain", "test data")

	err = repo.Store(ctx, testResource)
	assert.NoError(t, err)

	retrieved, err := repo.Retrieve(ctx, "provider-test")
	assert.NoError(t, err)
	assert.Equal(t, testResource.GetData(), retrieved.GetData())
}
