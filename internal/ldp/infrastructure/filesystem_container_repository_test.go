package infrastructure

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

func TestFileSystemContainerRepository_CreateContainer(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	tests := []struct {
		name      string
		container *domain.Container
		wantErr   bool
	}{
		{
			name:      "create basic container",
			container: domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer),
			wantErr:   false,
		},
		{
			name:      "create container with non-existent parent",
			container: domain.NewContainer(context.Background(), "child-container", "parent-id", domain.BasicContainer),
			wantErr:   true, // Should fail because parent doesn't exist
		},
		{
			name:      "create container with nil",
			container: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateContainer(context.Background(), tt.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify container was created
				exists, err := repo.ContainerExists(context.Background(), tt.container.ID())
				if err != nil {
					t.Errorf("ContainerExists() error = %v", err)
				}
				if !exists {
					t.Errorf("Container should exist after creation")
				}

				// Verify container metadata file exists
				containerDir := repo.getContainerPath(tt.container.ID())
				metadataPath := filepath.Join(containerDir, "container.json")
				if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
					t.Errorf("Container metadata file should exist")
				}
			}
		})
	}
}

func TestFileSystemContainerRepository_CreateContainerWithParent(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create parent container first
	parentContainer := domain.NewContainer(context.Background(), "parent-container", "", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), parentContainer)
	if err != nil {
		t.Fatalf("Failed to create parent container: %v", err)
	}

	// Create child container
	childContainer := domain.NewContainer(context.Background(), "child-container", "parent-container", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), childContainer)
	if err != nil {
		t.Errorf("Failed to create child container: %v", err)
	}

	// Verify child container was created
	exists, err := repo.ContainerExists(context.Background(), "child-container")
	if err != nil {
		t.Errorf("ContainerExists() error = %v", err)
	}
	if !exists {
		t.Errorf("Child container should exist after creation")
	}

	// Verify parent-child relationship
	retrievedChild, err := repo.GetContainer(context.Background(), "child-container")
	if err != nil {
		t.Errorf("GetContainer() error = %v", err)
	}
	if retrievedChild.ParentID != "parent-container" {
		t.Errorf("Child container ParentID = %v, want %v", retrievedChild.ParentID, "parent-container")
	}
}

func TestFileSystemContainerRepository_GetContainer(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create test container
	testContainer := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	testContainer.SetTitle("Test Container")
	testContainer.SetDescription("A test container")

	err = repo.CreateContainer(context.Background(), testContainer)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	tests := []struct {
		name        string
		containerID string
		wantErr     bool
	}{
		{
			name:        "get existing container",
			containerID: "test-container",
			wantErr:     false,
		},
		{
			name:        "get non-existent container",
			containerID: "non-existent",
			wantErr:     true,
		},
		{
			name:        "get container with empty ID",
			containerID: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container, err := repo.GetContainer(context.Background(), tt.containerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if container == nil {
					t.Errorf("GetContainer() should return container")
					return
				}
				if container.ID() != tt.containerID {
					t.Errorf("GetContainer() ID = %v, want %v", container.ID(), tt.containerID)
				}
				if container.GetTitle() != "Test Container" {
					t.Errorf("GetContainer() Title = %v, want %v", container.GetTitle(), "Test Container")
				}
			}
		})
	}
}

func TestFileSystemContainerRepository_AddMember(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create test container
	testContainer := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), testContainer)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	tests := []struct {
		name        string
		containerID string
		memberID    string
		wantErr     bool
	}{
		{
			name:        "add member to existing container",
			containerID: "test-container",
			memberID:    "member-1",
			wantErr:     false,
		},
		{
			name:        "add member to non-existent container",
			containerID: "non-existent",
			memberID:    "member-2",
			wantErr:     true,
		},
		{
			name:        "add member with empty container ID",
			containerID: "",
			memberID:    "member-3",
			wantErr:     true,
		},
		{
			name:        "add member with empty member ID",
			containerID: "test-container",
			memberID:    "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.AddMember(context.Background(), tt.containerID, tt.memberID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify member was added to container
				container, err := repo.GetContainer(context.Background(), tt.containerID)
				if err != nil {
					t.Errorf("GetContainer() error = %v", err)
					return
				}
				if !container.HasMember(tt.memberID) {
					t.Errorf("Container should have member %s", tt.memberID)
				}

				// Verify membership was indexed
				members, err := indexer.GetMembers(context.Background(), tt.containerID, PaginationOptions{Limit: 100})
				if err != nil {
					t.Errorf("GetMembers() error = %v", err)
					return
				}
				found := false
				for _, member := range members {
					if member.ID == tt.memberID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Member %s should be indexed", tt.memberID)
				}
			}
		})
	}
}

func TestFileSystemContainerRepository_RemoveMember(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create test container and add a member
	testContainer := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), testContainer)
	if err != nil {
		t.Fatalf("Failed to create test container: %v", err)
	}

	err = repo.AddMember(context.Background(), "test-container", "member-1")
	if err != nil {
		t.Fatalf("Failed to add test member: %v", err)
	}

	tests := []struct {
		name        string
		containerID string
		memberID    string
		wantErr     bool
	}{
		{
			name:        "remove existing member",
			containerID: "test-container",
			memberID:    "member-1",
			wantErr:     false,
		},
		{
			name:        "remove non-existent member",
			containerID: "test-container",
			memberID:    "non-existent",
			wantErr:     true,
		},
		{
			name:        "remove member from non-existent container",
			containerID: "non-existent",
			memberID:    "member-1",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RemoveMember(context.Background(), tt.containerID, tt.memberID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify member was removed from container
				container, err := repo.GetContainer(context.Background(), tt.containerID)
				if err != nil {
					t.Errorf("GetContainer() error = %v", err)
					return
				}
				if container.HasMember(tt.memberID) {
					t.Errorf("Container should not have member %s", tt.memberID)
				}
			}
		})
	}
}

func TestFileSystemContainerRepository_HierarchicalDirectoryStructure(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create parent container
	parentContainer := domain.NewContainer(context.Background(), "parent", "", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), parentContainer)
	if err != nil {
		t.Fatalf("Failed to create parent container: %v", err)
	}

	// Create child container
	childContainer := domain.NewContainer(context.Background(), "child", "parent", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), childContainer)
	if err != nil {
		t.Fatalf("Failed to create child container: %v", err)
	}

	// Verify hierarchical directory structure
	parentDir := repo.getContainerPath("parent")
	childDir := repo.getContainerPath("child")

	// Check that directories exist
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent container directory should exist")
	}

	if _, err := os.Stat(childDir); os.IsNotExist(err) {
		t.Errorf("Child container directory should exist")
	}

	// Verify container metadata files exist
	parentMetadata := filepath.Join(parentDir, "container.json")
	childMetadata := filepath.Join(childDir, "container.json")

	if _, err := os.Stat(parentMetadata); os.IsNotExist(err) {
		t.Errorf("Parent container metadata should exist")
	}

	if _, err := os.Stat(childMetadata); os.IsNotExist(err) {
		t.Errorf("Child container metadata should exist")
	}

	// Verify parent-child relationship
	retrievedChild, err := repo.GetContainer(context.Background(), "child")
	if err != nil {
		t.Fatalf("Failed to retrieve child container: %v", err)
	}

	if retrievedChild.ParentID != "parent" {
		t.Errorf("Child container ParentID = %v, want %v", retrievedChild.ParentID, "parent")
	}
}

func TestFileSystemContainerRepository_ContainerMetadataPersistence(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create container with metadata (no parent for this test)
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container with metadata")
	container.AddMember("member-1")
	container.AddMember("member-2")

	// Store container
	err = repo.CreateContainer(context.Background(), container)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Retrieve container
	retrievedContainer, err := repo.GetContainer(context.Background(), "test-container")
	if err != nil {
		t.Fatalf("Failed to retrieve container: %v", err)
	}

	// Verify metadata persistence
	if retrievedContainer.ID() != "test-container" {
		t.Errorf("Container ID = %v, want %v", retrievedContainer.ID(), "test-container")
	}

	if retrievedContainer.ParentID != "" {
		t.Errorf("Container ParentID = %v, want %v", retrievedContainer.ParentID, "")
	}

	if retrievedContainer.ContainerType != domain.BasicContainer {
		t.Errorf("Container Type = %v, want %v", retrievedContainer.ContainerType, domain.BasicContainer)
	}

	if retrievedContainer.GetTitle() != "Test Container" {
		t.Errorf("Container Title = %v, want %v", retrievedContainer.GetTitle(), "Test Container")
	}

	if retrievedContainer.GetDescription() != "A test container with metadata" {
		t.Errorf("Container Description = %v, want %v", retrievedContainer.GetDescription(), "A test container with metadata")
	}

	if len(retrievedContainer.Members) != 2 {
		t.Errorf("Container Members count = %v, want %v", len(retrievedContainer.Members), 2)
	}

	// Verify JSON serialization format
	containerDir := repo.getContainerPath("test-container")
	metadataPath := filepath.Join(containerDir, "container.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	var metadata map[string]interface{}
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	// Verify JSON structure
	if metadata["id"] != "test-container" {
		t.Errorf("JSON metadata ID = %v, want %v", metadata["id"], "test-container")
	}

	if metadata["containerType"] != "BasicContainer" {
		t.Errorf("JSON metadata containerType = %v, want %v", metadata["containerType"], "BasicContainer")
	}
}

func TestFileSystemContainerRepository_MembershipTrackingIntegration(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "container_repo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create membership indexer
	indexer, err := NewSQLiteMembershipIndexer(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}
	defer indexer.Close()

	// Create repository
	repo, err := NewFileSystemContainerRepository(tempDir, indexer)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Create container
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	err = repo.CreateContainer(context.Background(), container)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}

	// Add members
	members := []string{"member-1", "member-2", "member-3"}
	for _, memberID := range members {
		err = repo.AddMember(context.Background(), "test-container", memberID)
		if err != nil {
			t.Fatalf("Failed to add member %s: %v", memberID, err)
		}
	}

	// Verify membership tracking in indexer
	indexedMembers, err := indexer.GetMembers(context.Background(), "test-container", PaginationOptions{Limit: 100})
	if err != nil {
		t.Fatalf("Failed to get indexed members: %v", err)
	}

	if len(indexedMembers) != len(members) {
		t.Errorf("Indexed members count = %v, want %v", len(indexedMembers), len(members))
	}

	// Verify each member is indexed
	for _, expectedMember := range members {
		found := false
		for _, indexedMember := range indexedMembers {
			if indexedMember.ID == expectedMember {
				found = true
				if indexedMember.Type != ResourceTypeResource {
					t.Errorf("Member %s type = %v, want %v", expectedMember, indexedMember.Type, ResourceTypeResource)
				}
				break
			}
		}
		if !found {
			t.Errorf("Member %s should be indexed", expectedMember)
		}
	}

	// Test member removal
	err = repo.RemoveMember(context.Background(), "test-container", "member-2")
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}

	// Verify member was removed from index
	updatedMembers, err := indexer.GetMembers(context.Background(), "test-container", PaginationOptions{Limit: 100})
	if err != nil {
		t.Fatalf("Failed to get updated members: %v", err)
	}

	if len(updatedMembers) != 2 {
		t.Errorf("Updated members count = %v, want %v", len(updatedMembers), 2)
	}

	// Verify member-2 is not in the index
	for _, member := range updatedMembers {
		if member.ID == "member-2" {
			t.Errorf("Member member-2 should not be in index after removal")
		}
	}
}
