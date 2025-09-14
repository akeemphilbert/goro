package infrastructure

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	require.NoError(t, err)

	return db
}

func TestNewGORMContainerRepository(t *testing.T) {
	db := setupTestDB(t)

	repo, err := NewGORMContainerRepository(db)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestNewGORMContainerRepository_NilDB(t *testing.T) {
	repo, err := NewGORMContainerRepository(nil)
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "database cannot be nil")
}

func TestGORMContainerRepository_CreateContainer(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create a test container
	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container")

	// Store the container
	err = repo.CreateContainer(ctx, container)
	assert.NoError(t, err)

	// Verify it was stored
	exists, err := repo.ContainerExists(ctx, "test-container")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestGORMContainerRepository_CreateContainer_AlreadyExists(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create a test container
	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)

	// Store the container
	err = repo.CreateContainer(ctx, container)
	assert.NoError(t, err)

	// Try to store it again
	err = repo.CreateContainer(ctx, container)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container already exists")
}

func TestGORMContainerRepository_GetContainer(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create and store a test container
	original := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	original.SetTitle("Test Container")
	original.SetDescription("A test container")

	err = repo.CreateContainer(ctx, original)
	require.NoError(t, err)

	// Retrieve the container
	retrieved, err := repo.GetContainer(ctx, "test-container")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-container", retrieved.ID())
	assert.Equal(t, "Test Container", retrieved.GetTitle())
	assert.Equal(t, "A test container", retrieved.GetDescription())
	assert.Equal(t, domain.BasicContainer, retrieved.GetContainerType())
	assert.Equal(t, "", retrieved.GetParentID())
}

func TestGORMContainerRepository_GetContainer_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to retrieve non-existent container
	_, err = repo.GetContainer(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container not found")
}

func TestGORMContainerRepository_UpdateContainer(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create and store a test container
	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	container.SetTitle("Original Title")

	err = repo.CreateContainer(ctx, container)
	require.NoError(t, err)

	// Update the container
	container.SetTitle("Updated Title")
	container.SetDescription("Updated description")

	err = repo.UpdateContainer(ctx, container)
	assert.NoError(t, err)

	// Retrieve and verify the update
	updated, err := repo.GetContainer(ctx, "test-container")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.GetTitle())
	assert.Equal(t, "Updated description", updated.GetDescription())
}

func TestGORMContainerRepository_DeleteContainer(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create and store a test container
	container := domain.NewContainer(ctx, "test-container", "", domain.BasicContainer)
	err = repo.CreateContainer(ctx, container)
	require.NoError(t, err)

	// Verify it exists
	exists, err := repo.ContainerExists(ctx, "test-container")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete the container
	err = repo.DeleteContainer(ctx, "test-container")
	assert.NoError(t, err)

	// Verify it no longer exists
	exists, err = repo.ContainerExists(ctx, "test-container")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGORMContainerRepository_ContainerHierarchy(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create parent container
	parent := domain.NewContainer(ctx, "parent", "", domain.BasicContainer)
	parent.SetTitle("Parent Container")
	err = repo.CreateContainer(ctx, parent)
	require.NoError(t, err)

	// Create child container
	child := domain.NewContainer(ctx, "child", "parent", domain.BasicContainer)
	child.SetTitle("Child Container")
	err = repo.CreateContainer(ctx, child)
	require.NoError(t, err)

	// Test GetChildren
	children, err := repo.GetChildren(ctx, "parent")
	assert.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, "child", children[0].ID())
	assert.Equal(t, "Child Container", children[0].GetTitle())

	// Test GetParent
	parentRetrieved, err := repo.GetParent(ctx, "child")
	assert.NoError(t, err)
	assert.Equal(t, "parent", parentRetrieved.ID())
	assert.Equal(t, "Parent Container", parentRetrieved.GetTitle())

	// Test GetPath
	path, err := repo.GetPath(ctx, "child")
	assert.NoError(t, err)
	assert.Equal(t, []string{"parent", "child"}, path)
}

func TestGORMContainerRepository_Membership(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create containers
	container := domain.NewContainer(ctx, "container", "", domain.BasicContainer)
	err = repo.CreateContainer(ctx, container)
	require.NoError(t, err)

	member := domain.NewContainer(ctx, "member", "", domain.BasicContainer)
	err = repo.CreateContainer(ctx, member)
	require.NoError(t, err)

	// Test AddMember
	err = repo.AddMember(ctx, "container", "member")
	assert.NoError(t, err)

	// Test ListMembers
	members, err := repo.ListMembers(ctx, "container", domain.PaginationOptions{})
	assert.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "member", members[0])

	// Test RemoveMember
	err = repo.RemoveMember(ctx, "container", "member")
	assert.NoError(t, err)

	// Verify member was removed
	members, err = repo.ListMembers(ctx, "container", domain.PaginationOptions{})
	assert.NoError(t, err)
	assert.Len(t, members, 0)
}

func TestGORMContainerRepository_Store_BasicResource(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create a basic resource
	resource := domain.NewResource(ctx, "test-resource", "text/plain", []byte("test data"))
	resource.SetMetadata("key", "value")

	// Store the resource
	err = repo.Store(ctx, resource)
	assert.NoError(t, err)

	// Verify it exists
	exists, err := repo.Exists(ctx, "test-resource")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Retrieve the resource
	retrieved, err := repo.Retrieve(ctx, "test-resource")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test-resource", retrieved.ID())
	assert.Equal(t, "text/plain", retrieved.GetContentType())
}

func TestGORMContainerRepository_Retrieve_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to retrieve non-existent resource
	_, err = repo.Retrieve(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGORMContainerRepository_Delete_BasicResource(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create and store a basic resource
	resource := domain.NewResource(ctx, "test-resource", "text/plain", []byte("test data"))
	err = repo.Store(ctx, resource)
	require.NoError(t, err)

	// Verify it exists
	exists, err := repo.Exists(ctx, "test-resource")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Delete the resource
	err = repo.Delete(ctx, "test-resource")
	assert.NoError(t, err)

	// Verify it no longer exists
	exists, err = repo.Exists(ctx, "test-resource")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGORMContainerRepository_FindByPath(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Create a test container
	container := domain.NewContainer(ctx, "test-path", "", domain.BasicContainer)
	err = repo.CreateContainer(ctx, container)
	require.NoError(t, err)

	// Test FindByPath (currently just uses container ID)
	found, err := repo.FindByPath(ctx, "test-path")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "test-path", found.ID())
}

func TestGORMContainerRepository_EmptyInputs(t *testing.T) {
	db := setupTestDB(t)
	repo, err := NewGORMContainerRepository(db)
	require.NoError(t, err)

	ctx := context.Background()

	// Test empty container ID
	_, err = repo.GetContainer(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container ID cannot be empty")

	exists, err := repo.ContainerExists(ctx, "")
	assert.Error(t, err)
	assert.False(t, exists)

	err = repo.DeleteContainer(ctx, "")
	assert.Error(t, err)

	// Test nil container
	err = repo.CreateContainer(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container cannot be nil")

	err = repo.UpdateContainer(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container cannot be nil")
}
