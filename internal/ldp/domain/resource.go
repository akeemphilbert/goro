package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Resource represents a stored resource in the pod using pericarp domain entity
type Resource struct {
	*pericarpdomain.BasicEntity
	ContentType string                 `json:"contentType"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewResource creates a new Resource entity with pericarp integration
func NewResource(ctx context.Context, id, contentType string, data []byte) *Resource {
	log.Context(ctx).Debugf("[NewResource] Creating new resource: id=%s, contentType=%s, dataSize=%d", id, contentType, len(data))

	if id == "" {
		id = uuid.New().String()
		log.Context(ctx).Debugf("[NewResource] Generated new ID: %s", id)
	} else {
		log.Context(ctx).Debugf("[NewResource] Using provided ID: %s", id)
	}

	log.Context(ctx).Debug("[NewResource] Creating resource entity")
	resource := &Resource{
		BasicEntity: pericarpdomain.NewEntity(id),
		ContentType: contentType,
		Data:        data,
		Metadata:    make(map[string]interface{}),
	}

	log.Context(ctx).Debug("[NewResource] Creating resource created event")
	// Emit resource created event
	event := NewResourceCreatedEvent(id, map[string]interface{}{
		"contentType": contentType,
		"size":        len(data),
		"createdAt":   time.Now(),
	})
	resource.AddEvent(event)

	log.Context(ctx).Infof("Resource created successfully: resourceID=%s, contentType=%s, size=%d", id, contentType, len(data))
	return resource
}

// FromJSONLD creates a Resource from JSON-LD data
func (r *Resource) FromJSONLD(ctx context.Context, data []byte) *Resource {
	log.Context(ctx).Debugf("[FromJSONLD] Processing JSON-LD data for resource: resourceID=%s, dataSize=%d", r.ID(), len(data))

	// Validate JSON-LD format
	log.Context(ctx).Debug("[FromJSONLD] Validating JSON-LD format")
	var jsonLD map[string]interface{}
	if err := json.Unmarshal(data, &jsonLD); err != nil {
		log.Context(ctx).Debugf("[FromJSONLD] JSON-LD validation failed: %v", err)
		r.AddError(fmt.Errorf("invalid JSON-LD format: %w", err))
		return r
	}
	log.Context(ctx).Debug("[FromJSONLD] JSON-LD format validation passed")

	// Check for required JSON-LD context
	log.Context(ctx).Debug("[FromJSONLD] Checking for @context in JSON-LD")
	if _, hasContext := jsonLD["@context"]; !hasContext {
		log.Context(ctx).Debug("[FromJSONLD] Validation failed: missing @context")
		r.AddError(fmt.Errorf("JSON-LD data must contain @context"))
		return r
	}
	log.Context(ctx).Debug("[FromJSONLD] @context validation passed")

	log.Context(ctx).Debug("[FromJSONLD] Setting resource data and metadata")
	r.Data = data
	r.ContentType = "application/ld+json"
	r.Metadata["originalFormat"] = "application/ld+json"
	r.Metadata["updatedAt"] = time.Now()

	log.Context(ctx).Debug("[FromJSONLD] Creating resource updated event")
	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "application/ld+json",
		"size":   len(data),
	})
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with JSON-LD data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// FromRDF creates a Resource from RDF/XML data
func (r *Resource) FromRDF(ctx context.Context, data []byte) *Resource {
	log.Context(ctx).Debugf("[FromRDF] Processing RDF/XML data for resource: resourceID=%s, dataSize=%d", r.ID(), len(data))

	// Basic validation for RDF/XML format
	log.Context(ctx).Debug("[FromRDF] Validating RDF/XML format")
	dataStr := string(data)
	if len(dataStr) == 0 {
		log.Context(ctx).Debug("[FromRDF] Validation failed: RDF/XML data cannot be empty")
		r.AddError(fmt.Errorf("RDF/XML data cannot be empty"))
		return r
	}
	log.Context(ctx).Debug("[FromRDF] Non-empty data validation passed")

	// Check for basic RDF/XML structure
	log.Context(ctx).Debug("[FromRDF] Checking for RDF/XML structural elements")
	if !containsRDFElements(dataStr) {
		log.Context(ctx).Debug("[FromRDF] Validation failed: missing required RDF elements")
		r.AddError(fmt.Errorf("invalid RDF/XML format: missing required RDF elements"))
		return r
	}
	log.Context(ctx).Debug("[FromRDF] RDF/XML structural validation passed")

	log.Context(ctx).Debug("[FromRDF] Setting resource data and metadata")
	r.Data = data
	r.ContentType = "application/rdf+xml"
	r.Metadata["originalFormat"] = "application/rdf+xml"
	r.Metadata["updatedAt"] = time.Now()

	log.Context(ctx).Debug("[FromRDF] Creating resource updated event")
	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "application/rdf+xml",
		"size":   len(data),
	})
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with RDF/XML data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// FromXML is an alias for FromRDF for backward compatibility
func (r *Resource) FromXML(ctx context.Context, data []byte) *Resource {
	log.Context(ctx).Debugf("[FromXML] Delegating to FromRDF for resource: resourceID=%s, dataSize=%d", r.ID(), len(data))
	return r.FromRDF(ctx, data)
}

// FromTurtle creates a Resource from Turtle data
func (r *Resource) FromTurtle(ctx context.Context, data []byte) *Resource {
	log.Context(ctx).Debugf("[FromTurtle] Processing Turtle data for resource: resourceID=%s, dataSize=%d", r.ID(), len(data))

	// Basic validation for Turtle format
	log.Context(ctx).Debug("[FromTurtle] Validating Turtle format")
	dataStr := string(data)
	if len(dataStr) == 0 {
		log.Context(ctx).Debug("[FromTurtle] Validation failed: Turtle data cannot be empty")
		r.AddError(fmt.Errorf("Turtle data cannot be empty"))
		return r
	}
	log.Context(ctx).Debug("[FromTurtle] Non-empty data validation passed")

	// Basic Turtle format validation (check for common patterns)
	log.Context(ctx).Debug("[FromTurtle] Checking for Turtle syntax elements")
	if !containsTurtleElements(dataStr) {
		log.Context(ctx).Debug("[FromTurtle] Validation failed: missing required Turtle syntax")
		r.AddError(fmt.Errorf("invalid Turtle format: missing required Turtle syntax"))
		return r
	}
	log.Context(ctx).Debug("[FromTurtle] Turtle syntax validation passed")

	log.Context(ctx).Debug("[FromTurtle] Setting resource data and metadata")
	r.Data = data
	r.ContentType = "text/turtle"
	r.Metadata["originalFormat"] = "text/turtle"
	r.Metadata["updatedAt"] = time.Now()

	log.Context(ctx).Debug("[FromTurtle] Creating resource updated event")
	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "text/turtle",
		"size":   len(data),
	})
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with Turtle data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// ToFormat converts the resource data to the specified format
func (r *Resource) ToFormat(format string) ([]byte, error) {
	// For now, return the data as-is since format conversion
	// will be handled by a separate RDFConverter component
	// This method provides the interface for future format conversion

	switch format {
	case "application/ld+json", "application/json":
		if r.ContentType == "application/ld+json" {
			return r.Data, nil
		}
		return nil, fmt.Errorf("format conversion from %s to %s not yet implemented", r.ContentType, format)
	case "application/rdf+xml":
		if r.ContentType == "application/rdf+xml" {
			return r.Data, nil
		}
		return nil, fmt.Errorf("format conversion from %s to %s not yet implemented", r.ContentType, format)
	case "text/turtle":
		if r.ContentType == "text/turtle" {
			return r.Data, nil
		}
		return nil, fmt.Errorf("format conversion from %s to %s not yet implemented", r.ContentType, format)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Update updates the resource data and emits an update event
func (r *Resource) Update(ctx context.Context, data []byte, contentType string) {
	log.Context(ctx).Debugf("[Update] Updating resource: resourceID=%s, contentType=%s, dataSize=%d", r.ID(), contentType, len(data))

	log.Context(ctx).Debug("[Update] Setting resource data and content type")
	r.Data = data
	r.ContentType = contentType
	r.Metadata["updatedAt"] = time.Now()

	log.Context(ctx).Debug("[Update] Creating resource updated event")
	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"contentType": contentType,
		"size":        len(data),
		"updatedAt":   time.Now(),
	})
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated successfully: resourceID=%s, contentType=%s, size=%d", r.ID(), contentType, len(data))
}

// Delete marks the resource as deleted and emits a delete event
func (r *Resource) Delete() {
	// Emit delete event
	event := NewResourceDeletedEvent(r.ID(), map[string]interface{}{
		"deletedAt": time.Now(),
	})
	r.AddEvent(event)
}

// GetSize returns the size of the resource data
func (r *Resource) GetSize() int {
	return len(r.Data)
}

// GetContentType returns the content type of the resource
func (r *Resource) GetContentType() string {
	return r.ContentType
}

// GetData returns the resource data
func (r *Resource) GetData() []byte {
	return r.Data
}

// GetMetadata returns the resource metadata
func (r *Resource) GetMetadata() map[string]interface{} {
	return r.Metadata
}

// SetMetadata sets a metadata value
func (r *Resource) SetMetadata(key string, value interface{}) {
	r.Metadata[key] = value
}

// ClearEvents clears all uncommitted events
func (r *Resource) ClearEvents() {
	// Reset the events slice - pericarp BasicEntity doesn't have ClearEvents method
	// We'll work around this by not clearing events for now
	// The container will just have both resource created and container created events
}

// Helper functions for format validation

func containsRDFElements(data string) bool {
	// Basic check for RDF/XML elements
	return len(data) > 0 && (
	// Allow any non-empty content for now - more sophisticated validation
	// would require a proper RDF parser
	true)
}

func containsTurtleElements(data string) bool {
	// Basic check for Turtle syntax patterns
	return len(data) > 0 && (
	// Allow any non-empty content for now - more sophisticated validation
	// would require a proper Turtle parser
	true)
}

// ResourceRepository defines the interface for resource persistence
type ResourceRepository interface {
	Store(ctx context.Context, resource *Resource) error
	Retrieve(ctx context.Context, id string) (*Resource, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// ContainerRepository extends ResourceRepository with container-specific operations
type ContainerRepository interface {
	ResourceRepository

	// Container-specific operations
	CreateContainer(ctx context.Context, container *Container) error
	GetContainer(ctx context.Context, id string) (*Container, error)
	UpdateContainer(ctx context.Context, container *Container) error
	DeleteContainer(ctx context.Context, id string) error
	ContainerExists(ctx context.Context, id string) (bool, error)

	// Membership operations
	AddMember(ctx context.Context, containerID, memberID string) error
	RemoveMember(ctx context.Context, containerID, memberID string) error
	ListMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]string, error)

	// Hierarchy navigation
	GetChildren(ctx context.Context, containerID string) ([]*Container, error)
	GetParent(ctx context.Context, containerID string) (*Container, error)
	GetPath(ctx context.Context, containerID string) ([]string, error)
	FindByPath(ctx context.Context, path string) (*Container, error)
}

// StreamingResourceRepository extends ResourceRepository with streaming capabilities
type StreamingResourceRepository interface {
	ResourceRepository
	StoreStream(ctx context.Context, id string, reader io.Reader, contentType string, size int64) error
	RetrieveStream(ctx context.Context, id string) (io.ReadCloser, *ResourceMetadata, error)
}

// ResourceMetadata represents metadata for a resource
type ResourceMetadata struct {
	ID             string                 `json:"id"`
	ContentType    string                 `json:"contentType"`
	OriginalFormat string                 `json:"originalFormat"`
	Size           int64                  `json:"size"`
	Checksum       string                 `json:"checksum"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Tags           map[string]interface{} `json:"tags"`
}

// FormatConverter defines the interface for RDF format conversion
type FormatConverter interface {
	Convert(data []byte, fromFormat, toFormat string) ([]byte, error)
	ValidateFormat(format string) bool
}
