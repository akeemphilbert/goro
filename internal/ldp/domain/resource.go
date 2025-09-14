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

// BasicResource represents a stored resource in the pod using pericarp domain entity
type BasicResource struct {
	*pericarpdomain.BasicEntity
	ContentType string                 `json:"contentType"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewResource creates a new BasicResource entity with pericarp integration
func NewResource(ctx context.Context, id, contentType string, data []byte) *BasicResource {
	log.Context(ctx).Debugf("[NewResource] Creating new resource: id=%s, contentType=%s, dataSize=%d", id, contentType, len(data))

	if id == "" {
		id = uuid.New().String()
		log.Context(ctx).Debugf("[NewResource] Generated new ID: %s", id)
	} else {
		log.Context(ctx).Debugf("[NewResource] Using provided ID: %s", id)
	}

	log.Context(ctx).Debug("[NewResource] Creating resource entity")
	resource := &BasicResource{
		BasicEntity: pericarpdomain.NewEntity(id),
		ContentType: contentType,
		Data:        data,
		Metadata:    make(map[string]interface{}),
	}

	log.Context(ctx).Debug("[NewResource] Creating resource created event")

	// Check if this is an RDF format that needs relationship processing
	if IsRDFFormat(contentType) {
		log.Context(ctx).Debug("[NewResource] RDF format detected, creating enhanced event")
		// For new resources with RDF content, we'll emit the enhanced event
		// The actual relationship parsing will be done by the relationship service
		eventData := &ResourceEventData{
			Format:                contentType,
			Size:                  len(data),
			ContentType:           contentType,
			CreatedAt:             time.Now().Format(time.RFC3339),
			RequiresOrchestration: true, // Mark for processing by relationship service
		}
		event := NewResourceCreatedWithRelationsEvent(id, eventData)
		resource.AddEvent(event)
	} else {
		// Emit standard created event for non-RDF resources
		event := NewResourceCreatedEvent(id, map[string]interface{}{
			"contentType": contentType,
			"size":        len(data),
			"createdAt":   time.Now(),
		})
		resource.AddEvent(event)
	}

	log.Context(ctx).Infof("Resource created successfully: resourceID=%s, contentType=%s, size=%d", id, contentType, len(data))
	return resource
}

// FromJSONLD creates a Resource from JSON-LD data
func (r *BasicResource) FromJSONLD(ctx context.Context, data []byte) *BasicResource {
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
	// Emit enhanced update event for JSON-LD (RDF format)
	eventData := &ResourceEventData{
		Format:                "application/ld+json",
		Size:                  len(data),
		ContentType:           "application/ld+json",
		UpdatedAt:             time.Now().Format(time.RFC3339),
		RequiresOrchestration: true, // Mark for processing by relationship service
	}
	event := NewResourceUpdatedWithRelationsEvent(r.ID(), eventData)
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with JSON-LD data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// FromRDF creates a Resource from RDF/XML data
func (r *BasicResource) FromRDF(ctx context.Context, data []byte) *BasicResource {
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
	// Emit enhanced update event for RDF/XML (RDF format)
	eventData := &ResourceEventData{
		Format:                "application/rdf+xml",
		Size:                  len(data),
		ContentType:           "application/rdf+xml",
		UpdatedAt:             time.Now().Format(time.RFC3339),
		RequiresOrchestration: true, // Mark for processing by relationship service
	}
	event := NewResourceUpdatedWithRelationsEvent(r.ID(), eventData)
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with RDF/XML data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// FromXML is an alias for FromRDF for backward compatibility
func (r *BasicResource) FromXML(ctx context.Context, data []byte) *BasicResource {
	log.Context(ctx).Debugf("[FromXML] Delegating to FromRDF for resource: resourceID=%s, dataSize=%d", r.ID(), len(data))
	return r.FromRDF(ctx, data)
}

// FromTurtle creates a Resource from Turtle data
func (r *BasicResource) FromTurtle(ctx context.Context, data []byte) *BasicResource {
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
	// Emit enhanced update event for Turtle (RDF format)
	eventData := &ResourceEventData{
		Format:                "text/turtle",
		Size:                  len(data),
		ContentType:           "text/turtle",
		UpdatedAt:             time.Now().Format(time.RFC3339),
		RequiresOrchestration: true, // Mark for processing by relationship service
	}
	event := NewResourceUpdatedWithRelationsEvent(r.ID(), eventData)
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource updated with Turtle data: resourceID=%s, size=%d", r.ID(), len(data))
	return r
}

// ToFormat converts the resource data to the specified format
func (r *BasicResource) ToFormat(format string) ([]byte, error) {
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
func (r *BasicResource) Update(ctx context.Context, data []byte, contentType string) {
	log.Context(ctx).Debugf("[Update] Updating resource: resourceID=%s, contentType=%s, dataSize=%d", r.ID(), contentType, len(data))

	log.Context(ctx).Debug("[Update] Setting resource data and content type")
	r.Data = data
	r.ContentType = contentType
	r.Metadata["updatedAt"] = time.Now()

	log.Context(ctx).Debug("[Update] Creating resource updated event")

	// Check if this is an RDF format that needs relationship processing
	if IsRDFFormat(contentType) {
		log.Context(ctx).Debug("[Update] RDF format detected, creating enhanced event")
		// Emit enhanced update event for RDF resources
		eventData := &ResourceEventData{
			Format:                contentType,
			Size:                  len(data),
			ContentType:           contentType,
			UpdatedAt:             time.Now().Format(time.RFC3339),
			RequiresOrchestration: true, // Mark for processing by relationship service
		}
		event := NewResourceUpdatedWithRelationsEvent(r.ID(), eventData)
		r.AddEvent(event)
	} else {
		// Emit standard update event for non-RDF resources
		event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
			"contentType": contentType,
			"size":        len(data),
			"updatedAt":   time.Now(),
		})
		r.AddEvent(event)
	}

	log.Context(ctx).Infof("Resource updated successfully: resourceID=%s, contentType=%s, size=%d", r.ID(), contentType, len(data))
}

// Delete marks the resource as deleted and emits a delete event
func (r *BasicResource) Delete(ctx context.Context) {
	log.Context(ctx).Debugf("[Delete] Deleting resource: resourceID=%s", r.ID())

	log.Context(ctx).Debug("[Delete] Creating resource deleted event")
	// Emit delete event
	event := NewResourceDeletedEvent(r.ID(), map[string]interface{}{
		"deletedAt": time.Now(),
	})
	r.AddEvent(event)

	log.Context(ctx).Infof("Resource deleted successfully: resourceID=%s", r.ID())
}

// GetSize returns the size of the resource data
func (r *BasicResource) GetSize() int {
	return len(r.Data)
}

// GetContentType returns the content type of the resource
func (r *BasicResource) GetContentType() string {
	return r.ContentType
}

// GetData returns the resource data
func (r *BasicResource) GetData() []byte {
	return r.Data
}

// GetMetadata returns the resource metadata
func (r *BasicResource) GetMetadata() map[string]interface{} {
	return r.Metadata
}

// SetMetadata sets a metadata value
func (r *BasicResource) SetMetadata(key string, value interface{}) {
	r.Metadata[key] = value
}

// ClearEvents clears all uncommitted events
func (r *BasicResource) ClearEvents() {
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
	Store(ctx context.Context, resource Resource) error
	Retrieve(ctx context.Context, id string) (Resource, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// ContainerRepository extends ResourceRepository with container-specific operations
type ContainerRepository interface {
	ResourceRepository

	// Container-specific operations
	CreateContainer(ctx context.Context, container ContainerResource) error
	GetContainer(ctx context.Context, id string) (ContainerResource, error)
	UpdateContainer(ctx context.Context, container ContainerResource) error
	DeleteContainer(ctx context.Context, id string) error
	ContainerExists(ctx context.Context, id string) (bool, error)

	// Membership operations
	AddMember(ctx context.Context, containerID, memberID string) error
	RemoveMember(ctx context.Context, containerID, memberID string) error
	ListMembers(ctx context.Context, containerID string, pagination PaginationOptions) ([]string, error)

	// Hierarchy navigation
	GetChildren(ctx context.Context, containerID string) ([]ContainerResource, error)
	GetParent(ctx context.Context, containerID string) (ContainerResource, error)
	GetPath(ctx context.Context, containerID string) ([]string, error)
	FindByPath(ctx context.Context, path string) (ContainerResource, error)
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
