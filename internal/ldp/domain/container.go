package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ContainerType represents the type of LDP container
type ContainerType string

const (
	BasicContainer  ContainerType = "BasicContainer"
	DirectContainer ContainerType = "DirectContainer"
)

// String returns the string representation of the container type
func (ct ContainerType) String() string {
	return string(ct)
}

// IsValid checks if the container type is valid
func (ct ContainerType) IsValid() bool {
	switch ct {
	case BasicContainer, DirectContainer:
		return true
	default:
		return false
	}
}

// Container represents a container resource that can hold other resources
type Container struct {
	*Resource                   // Inherits from Resource
	Members       []string      // Resource IDs contained in this container
	ParentID      string        // Parent container ID (empty for root)
	ContainerType ContainerType // BasicContainer, DirectContainer, etc.
}

// NewContainer creates a new Container entity
func NewContainer(id, parentID string, containerType ContainerType) *Container {
	if id == "" {
		id = uuid.New().String()
	}

	// Create the underlying resource with container-specific content type
	resource := NewResource(id, "application/ld+json", []byte("{}"))

	container := &Container{
		Resource:      resource,
		Members:       make([]string, 0),
		ParentID:      parentID,
		ContainerType: containerType,
	}

	// Set container-specific metadata
	container.SetMetadata("type", "Container")
	container.SetMetadata("containerType", containerType.String())
	container.SetMetadata("parentID", parentID)
	container.SetMetadata("createdAt", time.Now())

	// Emit container created event (resource created event will also be present)
	event := NewContainerCreatedEvent(id, map[string]interface{}{
		"parentID":      parentID,
		"containerType": containerType.String(),
		"createdAt":     time.Now(),
	})
	container.AddEvent(event)

	return container
}

// AddMember adds a resource to the container
func (c *Container) AddMember(memberID string) error {
	// Check if member already exists
	for _, member := range c.Members {
		if member == memberID {
			return fmt.Errorf("member already exists: %s", memberID)
		}
	}

	c.Members = append(c.Members, memberID)
	c.SetMetadata("updatedAt", time.Now())

	// Emit member added event
	event := NewMemberAddedEvent(c.ID(), map[string]interface{}{
		"memberID":   memberID,
		"memberType": "Resource", // Default to Resource, could be enhanced to detect type
		"addedAt":    time.Now(),
	})
	c.AddEvent(event)

	return nil
}

// RemoveMember removes a resource from the container
func (c *Container) RemoveMember(memberID string) error {
	for i, member := range c.Members {
		if member == memberID {
			// Remove member from slice
			c.Members = append(c.Members[:i], c.Members[i+1:]...)
			c.SetMetadata("updatedAt", time.Now())

			// Emit member removed event
			event := NewMemberRemovedEvent(c.ID(), map[string]interface{}{
				"memberID":   memberID,
				"memberType": "Resource",
				"removedAt":  time.Now(),
			})
			c.AddEvent(event)

			return nil
		}
	}

	return fmt.Errorf("member not found: %s", memberID)
}

// HasMember checks if a resource is a member of the container
func (c *Container) HasMember(memberID string) bool {
	for _, member := range c.Members {
		if member == memberID {
			return true
		}
	}
	return false
}

// GetMemberCount returns the number of members in the container
func (c *Container) GetMemberCount() int {
	return len(c.Members)
}

// IsEmpty checks if the container has no members
func (c *Container) IsEmpty() bool {
	return len(c.Members) == 0
}

// ValidateHierarchy validates the container hierarchy to prevent circular references
func (c *Container) ValidateHierarchy(ancestorPath []string) error {
	// Check for self-reference
	if c.ParentID == c.ID() {
		return fmt.Errorf("container cannot be its own parent")
	}

	// Check for circular reference in the ancestor path
	for _, ancestorID := range ancestorPath {
		if ancestorID == c.ID() {
			return fmt.Errorf("circular reference detected: container %s is already in the hierarchy path", c.ID())
		}
	}

	return nil
}

// SetTitle sets the container title
func (c *Container) SetTitle(title string) {
	c.SetMetadata("title", title)
	c.SetMetadata("updatedAt", time.Now())

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"title":     title,
		"updatedAt": time.Now(),
	})
	c.AddEvent(event)
}

// GetTitle returns the container title
func (c *Container) GetTitle() string {
	if title, exists := c.GetMetadata()["title"]; exists {
		if titleStr, ok := title.(string); ok {
			return titleStr
		}
	}
	return ""
}

// SetDescription sets the container description
func (c *Container) SetDescription(description string) {
	c.SetMetadata("description", description)
	c.SetMetadata("updatedAt", time.Now())

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"description": description,
		"updatedAt":   time.Now(),
	})
	c.AddEvent(event)
}

// GetDescription returns the container description
func (c *Container) GetDescription() string {
	if description, exists := c.GetMetadata()["description"]; exists {
		if descStr, ok := description.(string); ok {
			return descStr
		}
	}
	return ""
}

// GetPath returns the path representation of the container
func (c *Container) GetPath() string {
	if c.ParentID == "" {
		return c.ID()
	}
	return c.ParentID + "/" + c.ID()
}

// DublinCoreMetadata represents Dublin Core metadata fields for containers
type DublinCoreMetadata struct {
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Creator     string    `json:"creator,omitempty"`
	Subject     string    `json:"subject,omitempty"`
	Publisher   string    `json:"publisher,omitempty"`
	Contributor string    `json:"contributor,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Type        string    `json:"type,omitempty"`
	Format      string    `json:"format,omitempty"`
	Identifier  string    `json:"identifier,omitempty"`
	Source      string    `json:"source,omitempty"`
	Language    string    `json:"language,omitempty"`
	Relation    string    `json:"relation,omitempty"`
	Coverage    string    `json:"coverage,omitempty"`
	Rights      string    `json:"rights,omitempty"`
}

// SetDublinCoreMetadata sets Dublin Core metadata on a container
func (c *Container) SetDublinCoreMetadata(dc DublinCoreMetadata) {
	if dc.Title != "" {
		c.SetMetadata("dc:title", dc.Title)
	}
	if dc.Description != "" {
		c.SetMetadata("dc:description", dc.Description)
	}
	if dc.Creator != "" {
		c.SetMetadata("dc:creator", dc.Creator)
	}
	if dc.Subject != "" {
		c.SetMetadata("dc:subject", dc.Subject)
	}
	if dc.Publisher != "" {
		c.SetMetadata("dc:publisher", dc.Publisher)
	}
	if dc.Contributor != "" {
		c.SetMetadata("dc:contributor", dc.Contributor)
	}
	if !dc.Date.IsZero() {
		c.SetMetadata("dc:date", dc.Date)
	}
	if dc.Type != "" {
		c.SetMetadata("dc:type", dc.Type)
	}
	if dc.Format != "" {
		c.SetMetadata("dc:format", dc.Format)
	}
	if dc.Identifier != "" {
		c.SetMetadata("dc:identifier", dc.Identifier)
	}
	if dc.Source != "" {
		c.SetMetadata("dc:source", dc.Source)
	}
	if dc.Language != "" {
		c.SetMetadata("dc:language", dc.Language)
	}
	if dc.Relation != "" {
		c.SetMetadata("dc:relation", dc.Relation)
	}
	if dc.Coverage != "" {
		c.SetMetadata("dc:coverage", dc.Coverage)
	}
	if dc.Rights != "" {
		c.SetMetadata("dc:rights", dc.Rights)
	}

	// Update modification timestamp
	c.SetMetadata("updatedAt", time.Now())

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"dublinCore": dc,
		"updatedAt":  time.Now(),
	})
	c.AddEvent(event)
}

// GetDublinCoreMetadata retrieves Dublin Core metadata from a container
func (c *Container) GetDublinCoreMetadata() DublinCoreMetadata {
	metadata := c.GetMetadata()
	dc := DublinCoreMetadata{}

	if title, exists := metadata["dc:title"]; exists {
		if titleStr, ok := title.(string); ok {
			dc.Title = titleStr
		}
	}
	if description, exists := metadata["dc:description"]; exists {
		if descStr, ok := description.(string); ok {
			dc.Description = descStr
		}
	}
	if creator, exists := metadata["dc:creator"]; exists {
		if creatorStr, ok := creator.(string); ok {
			dc.Creator = creatorStr
		}
	}
	if subject, exists := metadata["dc:subject"]; exists {
		if subjectStr, ok := subject.(string); ok {
			dc.Subject = subjectStr
		}
	}
	if publisher, exists := metadata["dc:publisher"]; exists {
		if publisherStr, ok := publisher.(string); ok {
			dc.Publisher = publisherStr
		}
	}
	if contributor, exists := metadata["dc:contributor"]; exists {
		if contributorStr, ok := contributor.(string); ok {
			dc.Contributor = contributorStr
		}
	}
	if date, exists := metadata["dc:date"]; exists {
		if dateTime, ok := date.(time.Time); ok {
			dc.Date = dateTime
		}
	}
	if dcType, exists := metadata["dc:type"]; exists {
		if typeStr, ok := dcType.(string); ok {
			dc.Type = typeStr
		}
	}
	if format, exists := metadata["dc:format"]; exists {
		if formatStr, ok := format.(string); ok {
			dc.Format = formatStr
		}
	}
	if identifier, exists := metadata["dc:identifier"]; exists {
		if identifierStr, ok := identifier.(string); ok {
			dc.Identifier = identifierStr
		}
	}
	if source, exists := metadata["dc:source"]; exists {
		if sourceStr, ok := source.(string); ok {
			dc.Source = sourceStr
		}
	}
	if language, exists := metadata["dc:language"]; exists {
		if languageStr, ok := language.(string); ok {
			dc.Language = languageStr
		}
	}
	if relation, exists := metadata["dc:relation"]; exists {
		if relationStr, ok := relation.(string); ok {
			dc.Relation = relationStr
		}
	}
	if coverage, exists := metadata["dc:coverage"]; exists {
		if coverageStr, ok := coverage.(string); ok {
			dc.Coverage = coverageStr
		}
	}
	if rights, exists := metadata["dc:rights"]; exists {
		if rightsStr, ok := rights.(string); ok {
			dc.Rights = rightsStr
		}
	}

	return dc
}

// Delete marks the container as deleted and emits a delete event
func (c *Container) Delete() error {
	// Check if container is empty
	if !c.IsEmpty() {
		return fmt.Errorf("container is not empty: cannot delete container with %d members", len(c.Members))
	}

	// Emit delete event
	event := NewContainerDeletedEvent(c.ID(), map[string]interface{}{
		"deletedAt": time.Now(),
	})
	c.AddEvent(event)

	return nil
}

// PaginationOptions represents pagination parameters for listing operations
type PaginationOptions struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// IsValid validates the pagination options
func (p PaginationOptions) IsValid() bool {
	return p.Limit > 0 && p.Limit <= 1000 && p.Offset >= 0
}

// GetDefaultPagination returns default pagination options
func GetDefaultPagination() PaginationOptions {
	return PaginationOptions{
		Limit:  50,
		Offset: 0,
	}
}

// FilterOptions represents filtering options for container members
type FilterOptions struct {
	MemberType    string     `json:"memberType,omitempty"`    // "Container" or "Resource"
	ContentType   string     `json:"contentType,omitempty"`   // MIME type filter
	NamePattern   string     `json:"namePattern,omitempty"`   // Name pattern matching
	CreatedAfter  *time.Time `json:"createdAfter,omitempty"`  // Created after timestamp
	CreatedBefore *time.Time `json:"createdBefore,omitempty"` // Created before timestamp
	SizeMin       *int64     `json:"sizeMin,omitempty"`       // Minimum size in bytes
	SizeMax       *int64     `json:"sizeMax,omitempty"`       // Maximum size in bytes
}

// SortOptions represents sorting options for container members
type SortOptions struct {
	Field     string `json:"field"`     // "name", "createdAt", "updatedAt", "size", "type"
	Direction string `json:"direction"` // "asc" or "desc"
}

// IsValid validates sort options
func (s SortOptions) IsValid() bool {
	validFields := map[string]bool{
		"name":      true,
		"createdAt": true,
		"updatedAt": true,
		"size":      true,
		"type":      true,
	}

	validDirections := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	return validFields[s.Field] && validDirections[s.Direction]
}

// GetDefaultSort returns default sort options
func GetDefaultSort() SortOptions {
	return SortOptions{
		Field:     "createdAt",
		Direction: "asc",
	}
}

// ListingOptions combines pagination, filtering, and sorting options
type ListingOptions struct {
	Pagination PaginationOptions `json:"pagination"`
	Filter     FilterOptions     `json:"filter,omitempty"`
	Sort       SortOptions       `json:"sort"`
}

// GetDefaultListingOptions returns default listing options
func GetDefaultListingOptions() ListingOptions {
	return ListingOptions{
		Pagination: GetDefaultPagination(),
		Filter:     FilterOptions{},
		Sort:       GetDefaultSort(),
	}
}

// IsValid validates listing options
func (l ListingOptions) IsValid() bool {
	return l.Pagination.IsValid() && l.Sort.IsValid()
}
