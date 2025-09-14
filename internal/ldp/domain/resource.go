package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/google/uuid"
)

// Resource represents a stored resource in the pod using pericarp domain entity
type Resource struct {
	pericarpdomain.BasicEntity
	ContentType string                 `json:"contentType"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewResource creates a new Resource entity with pericarp integration
func NewResource(id, contentType string, data []byte) *Resource {
	if id == "" {
		id = uuid.New().String()
	}

	resource := &Resource{
		BasicEntity: pericarpdomain.NewEntity(id),
		ContentType: contentType,
		Data:        data,
		Metadata:    make(map[string]interface{}),
	}

	// Emit resource created event
	event := NewResourceCreatedEvent(id, map[string]interface{}{
		"contentType": contentType,
		"size":        len(data),
		"createdAt":   time.Now(),
	})
	resource.AddEvent(event)

	return resource
}

// FromJSONLD creates a Resource from JSON-LD data
func (r *Resource) FromJSONLD(data []byte) *Resource {
	// Validate JSON-LD format
	var jsonLD map[string]interface{}
	if err := json.Unmarshal(data, &jsonLD); err != nil {
		r.AddError(fmt.Errorf("invalid JSON-LD format: %w", err))
		return r
	}

	// Check for required JSON-LD context
	if _, hasContext := jsonLD["@context"]; !hasContext {
		r.AddError(fmt.Errorf("JSON-LD data must contain @context"))
		return r
	}

	r.Data = data
	r.ContentType = "application/ld+json"
	r.Metadata["originalFormat"] = "application/ld+json"
	r.Metadata["updatedAt"] = time.Now()

	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "application/ld+json",
		"size":   len(data),
	})
	r.AddEvent(event)

	return r
}

// FromRDF creates a Resource from RDF/XML data
func (r *Resource) FromRDF(data []byte) *Resource {
	// Basic validation for RDF/XML format
	dataStr := string(data)
	if len(dataStr) == 0 {
		r.AddError(fmt.Errorf("RDF/XML data cannot be empty"))
		return r
	}

	// Check for basic RDF/XML structure
	if !containsRDFElements(dataStr) {
		r.AddError(fmt.Errorf("invalid RDF/XML format: missing required RDF elements"))
		return r
	}

	r.Data = data
	r.ContentType = "application/rdf+xml"
	r.Metadata["originalFormat"] = "application/rdf+xml"
	r.Metadata["updatedAt"] = time.Now()

	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "application/rdf+xml",
		"size":   len(data),
	})
	r.AddEvent(event)

	return r
}

// FromXML is an alias for FromRDF for backward compatibility
func (r *Resource) FromXML(data []byte) *Resource {
	return r.FromRDF(data)
}

// FromTurtle creates a Resource from Turtle data
func (r *Resource) FromTurtle(data []byte) *Resource {
	// Basic validation for Turtle format
	dataStr := string(data)
	if len(dataStr) == 0 {
		r.AddError(fmt.Errorf("Turtle data cannot be empty"))
		return r
	}

	// Basic Turtle format validation (check for common patterns)
	if !containsTurtleElements(dataStr) {
		r.AddError(fmt.Errorf("invalid Turtle format: missing required Turtle syntax"))
		return r
	}

	r.Data = data
	r.ContentType = "text/turtle"
	r.Metadata["originalFormat"] = "text/turtle"
	r.Metadata["updatedAt"] = time.Now()

	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"format": "text/turtle",
		"size":   len(data),
	})
	r.AddEvent(event)

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
func (r *Resource) Update(data []byte, contentType string) {
	r.Data = data
	r.ContentType = contentType
	r.Metadata["updatedAt"] = time.Now()

	// Emit update event
	event := NewResourceUpdatedEvent(r.ID(), map[string]interface{}{
		"contentType": contentType,
		"size":        len(data),
		"updatedAt":   time.Now(),
	})
	r.AddEvent(event)
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
