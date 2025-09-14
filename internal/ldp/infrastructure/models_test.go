package infrastructure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestContainerModel_TableName(t *testing.T) {
	model := ContainerModel{}
	assert.Equal(t, "containers", model.TableName())
}

func TestResourceModel_TableName(t *testing.T) {
	model := ResourceModel{}
	assert.Equal(t, "resources", model.TableName())
}

func TestMembershipModel_TableName(t *testing.T) {
	model := MembershipModel{}
	assert.Equal(t, "memberships", model.TableName())
}

func TestContainerModel_CRUD(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Test Create
	container := &ContainerModel{
		ID:          "test-container",
		ParentID:    nil,
		Type:        "BasicContainer",
		Title:       "Test Container",
		Description: "A test container",
	}

	result := db.Create(container)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.False(t, container.CreatedAt.IsZero())
	assert.False(t, container.UpdatedAt.IsZero())

	// Test Read
	var retrieved ContainerModel
	err = db.First(&retrieved, "id = ?", "test-container").Error
	assert.NoError(t, err)
	assert.Equal(t, "test-container", retrieved.ID)
	assert.Equal(t, "BasicContainer", retrieved.Type)
	assert.Equal(t, "Test Container", retrieved.Title)
	assert.Equal(t, "A test container", retrieved.Description)
	assert.Nil(t, retrieved.ParentID)

	// Test Update
	retrieved.Title = "Updated Container"
	err = db.Save(&retrieved).Error
	assert.NoError(t, err)
	assert.True(t, retrieved.UpdatedAt.After(retrieved.CreatedAt))

	// Verify update
	var updated ContainerModel
	err = db.First(&updated, "id = ?", "test-container").Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Container", updated.Title)

	// Test Delete
	err = db.Delete(&ContainerModel{}, "id = ?", "test-container").Error
	assert.NoError(t, err)

	// Verify deletion
	var count int64
	err = db.Model(&ContainerModel{}).Where("id = ?", "test-container").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestContainerModel_Relationships(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Create parent container
	parent := &ContainerModel{
		ID:          "parent",
		Type:        "BasicContainer",
		Title:       "Parent Container",
		Description: "Parent container",
	}
	err = db.Create(parent).Error
	assert.NoError(t, err)

	// Create child container
	parentID := "parent"
	child := &ContainerModel{
		ID:          "child",
		ParentID:    &parentID,
		Type:        "BasicContainer",
		Title:       "Child Container",
		Description: "Child container",
	}
	err = db.Create(child).Error
	assert.NoError(t, err)

	// Test preloading parent
	var childWithParent ContainerModel
	err = db.Preload("Parent").First(&childWithParent, "id = ?", "child").Error
	assert.NoError(t, err)
	assert.NotNil(t, childWithParent.Parent)
	assert.Equal(t, "parent", childWithParent.Parent.ID)
	assert.Equal(t, "Parent Container", childWithParent.Parent.Title)

	// Test preloading children
	var parentWithChildren ContainerModel
	err = db.Preload("Children").First(&parentWithChildren, "id = ?", "parent").Error
	assert.NoError(t, err)
	assert.Len(t, parentWithChildren.Children, 1)
	assert.Equal(t, "child", parentWithChildren.Children[0].ID)
	assert.Equal(t, "Child Container", parentWithChildren.Children[0].Title)
}

func TestResourceModel_CRUD(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Create container first
	container := &ContainerModel{
		ID:   "test-container",
		Type: "BasicContainer",
	}
	err = db.Create(container).Error
	assert.NoError(t, err)

	// Test Create Resource
	resource := &ResourceModel{
		ID:          "test-resource",
		ContainerID: "test-container",
		ContentType: "text/plain",
		Size:        100,
		FilePath:    "/path/to/file",
		Metadata:    `{"key": "value"}`,
	}

	result := db.Create(resource)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.False(t, resource.CreatedAt.IsZero())
	assert.False(t, resource.UpdatedAt.IsZero())

	// Test Read with relationship
	var retrieved ResourceModel
	err = db.Preload("Container").First(&retrieved, "id = ?", "test-resource").Error
	assert.NoError(t, err)
	assert.Equal(t, "test-resource", retrieved.ID)
	assert.Equal(t, "test-container", retrieved.ContainerID)
	assert.Equal(t, "text/plain", retrieved.ContentType)
	assert.Equal(t, int64(100), retrieved.Size)
	assert.Equal(t, "/path/to/file", retrieved.FilePath)
	assert.Equal(t, `{"key": "value"}`, retrieved.Metadata)
	assert.NotNil(t, retrieved.Container)
	assert.Equal(t, "test-container", retrieved.Container.ID)

	// Test Update
	retrieved.Size = 200
	retrieved.ContentType = "application/json"
	err = db.Save(&retrieved).Error
	assert.NoError(t, err)
	assert.True(t, retrieved.UpdatedAt.After(retrieved.CreatedAt))

	// Verify update
	var updated ResourceModel
	err = db.First(&updated, "id = ?", "test-resource").Error
	assert.NoError(t, err)
	assert.Equal(t, int64(200), updated.Size)
	assert.Equal(t, "application/json", updated.ContentType)

	// Test Delete
	err = db.Delete(&ResourceModel{}, "id = ?", "test-resource").Error
	assert.NoError(t, err)

	// Verify deletion
	var count int64
	err = db.Model(&ResourceModel{}).Where("id = ?", "test-resource").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMembershipModel_CRUD(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Create container first
	container := &ContainerModel{
		ID:   "test-container",
		Type: "BasicContainer",
	}
	err = db.Create(container).Error
	assert.NoError(t, err)

	// Test Create Membership
	membership := &MembershipModel{
		ContainerID: "test-container",
		MemberID:    "test-member",
		MemberType:  "Resource",
	}

	result := db.Create(membership)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1), result.RowsAffected)
	assert.False(t, membership.CreatedAt.IsZero())

	// Test Read with relationship
	var retrieved MembershipModel
	err = db.Preload("Container").First(&retrieved,
		"container_id = ? AND member_id = ?", "test-container", "test-member").Error
	assert.NoError(t, err)
	assert.Equal(t, "test-container", retrieved.ContainerID)
	assert.Equal(t, "test-member", retrieved.MemberID)
	assert.Equal(t, "Resource", retrieved.MemberType)
	assert.NotNil(t, retrieved.Container)
	assert.Equal(t, "test-container", retrieved.Container.ID)

	// Test Delete
	err = db.Delete(&MembershipModel{},
		"container_id = ? AND member_id = ?", "test-container", "test-member").Error
	assert.NoError(t, err)

	// Verify deletion
	var count int64
	err = db.Model(&MembershipModel{}).Where(
		"container_id = ? AND member_id = ?", "test-container", "test-member").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestModel_BeforeCreateHooks(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Test ContainerModel BeforeCreate hook
	container := &ContainerModel{
		ID:   "test-container",
		Type: "BasicContainer",
	}
	assert.True(t, container.CreatedAt.IsZero())
	assert.True(t, container.UpdatedAt.IsZero())

	err = db.Create(container).Error
	assert.NoError(t, err)
	assert.False(t, container.CreatedAt.IsZero())
	assert.False(t, container.UpdatedAt.IsZero())

	// Test ResourceModel BeforeCreate hook
	resource := &ResourceModel{
		ID:          "test-resource",
		ContainerID: "test-container",
		ContentType: "text/plain",
		Size:        100,
	}
	assert.True(t, resource.CreatedAt.IsZero())
	assert.True(t, resource.UpdatedAt.IsZero())

	err = db.Create(resource).Error
	assert.NoError(t, err)
	assert.False(t, resource.CreatedAt.IsZero())
	assert.False(t, resource.UpdatedAt.IsZero())

	// Test MembershipModel BeforeCreate hook
	membership := &MembershipModel{
		ContainerID: "test-container",
		MemberID:    "test-member",
		MemberType:  "Resource",
	}
	assert.True(t, membership.CreatedAt.IsZero())

	err = db.Create(membership).Error
	assert.NoError(t, err)
	assert.False(t, membership.CreatedAt.IsZero())
}

func TestModel_BeforeUpdateHooks(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the models
	err = db.AutoMigrate(&ContainerModel{}, &ResourceModel{}, &MembershipModel{})
	assert.NoError(t, err)

	// Create and update container
	container := &ContainerModel{
		ID:    "test-container",
		Type:  "BasicContainer",
		Title: "Original Title",
	}
	err = db.Create(container).Error
	assert.NoError(t, err)
	originalUpdatedAt := container.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update container
	container.Title = "Updated Title"
	err = db.Save(container).Error
	assert.NoError(t, err)
	assert.True(t, container.UpdatedAt.After(originalUpdatedAt))

	// Create and update resource
	resource := &ResourceModel{
		ID:          "test-resource",
		ContainerID: "test-container",
		ContentType: "text/plain",
		Size:        100,
	}
	err = db.Create(resource).Error
	assert.NoError(t, err)
	originalResourceUpdatedAt := resource.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update resource
	resource.Size = 200
	err = db.Save(resource).Error
	assert.NoError(t, err)
	assert.True(t, resource.UpdatedAt.After(originalResourceUpdatedAt))
}
