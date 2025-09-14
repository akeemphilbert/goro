package infrastructure

import (
	"time"

	"gorm.io/gorm"
)

// ContainerModel represents the GORM model for containers
type ContainerModel struct {
	ID          string          `gorm:"primaryKey;type:varchar(255)"`
	ParentID    *string         `gorm:"type:varchar(255);index"`
	Parent      *ContainerModel `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
	Type        string          `gorm:"not null;type:varchar(50);default:'BasicContainer'"`
	Title       string          `gorm:"type:varchar(255)"`
	Description string          `gorm:"type:text"`
	CreatedAt   time.Time       `gorm:"not null"`
	UpdatedAt   time.Time       `gorm:"not null"`

	// Relationships
	Children  []ContainerModel `gorm:"foreignKey:ParentID;references:ID"`
	Resources []ResourceModel  `gorm:"foreignKey:ContainerID;references:ID"`
}

// TableName specifies the table name for ContainerModel
func (ContainerModel) TableName() string {
	return "containers"
}

// BeforeCreate GORM hook for validation before creation
func (c *ContainerModel) BeforeCreate(tx *gorm.DB) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate GORM hook for updating timestamps
func (c *ContainerModel) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// ResourceModel represents the GORM model for resources
type ResourceModel struct {
	ID          string         `gorm:"primaryKey;type:varchar(255)"`
	ContainerID string         `gorm:"not null;type:varchar(255);index"`
	Container   ContainerModel `gorm:"foreignKey:ContainerID;references:ID;constraint:OnDelete:CASCADE"`
	ContentType string         `gorm:"not null;type:varchar(255)"`
	Size        int64          `gorm:"not null;default:0"`
	FilePath    string         `gorm:"type:varchar(500)"` // Path to file on filesystem
	Metadata    string         `gorm:"type:text"`         // JSON serialized metadata
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
}

// TableName specifies the table name for ResourceModel
func (ResourceModel) TableName() string {
	return "resources"
}

// BeforeCreate GORM hook for validation before creation
func (r *ResourceModel) BeforeCreate(tx *gorm.DB) error {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate GORM hook for updating timestamps
func (r *ResourceModel) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// MembershipModel represents the GORM model for container membership relationships
type MembershipModel struct {
	ContainerID string    `gorm:"primaryKey;type:varchar(255)"`
	MemberID    string    `gorm:"primaryKey;type:varchar(255)"`
	MemberType  string    `gorm:"not null;type:varchar(50)"` // "Container" or "Resource"
	CreatedAt   time.Time `gorm:"not null"`

	// Relationships
	Container ContainerModel `gorm:"foreignKey:ContainerID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for MembershipModel
func (MembershipModel) TableName() string {
	return "memberships"
}

// BeforeCreate GORM hook for validation before creation
func (m *MembershipModel) BeforeCreate(tx *gorm.DB) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}
