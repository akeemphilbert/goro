package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

// RDFTriple represents a subject-predicate-object triple in RDF
type RDFTriple struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	IsLiteral bool   `json:"is_literal"` // true if object is a literal value, false if it's a resource
}

// ResourceRelationship represents a relationship between resources
type ResourceRelationship struct {
	Subject    string `json:"subject"`
	Predicate  string `json:"predicate"`
	Object     string `json:"object"`
	IsExternal bool   `json:"is_external"` // true if the linked resource is external to this pod
	IsLiteral  bool   `json:"is_literal"`  // true if object is a literal value
}

// ResourceEventData represents enhanced event data for resource operations
type ResourceEventData struct {
	// Basic resource information
	Format      string `json:"format"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	DeletedAt   string `json:"deleted_at,omitempty"`

	// RDF-specific information
	Relationships      []ResourceRelationship `json:"relationships,omitempty"`
	LinkedResourceURIs []string               `json:"linked_resource_uris,omitempty"`
	Triples            []RDFTriple            `json:"triples,omitempty"`

	// Processing metadata
	RequiresOrchestration bool     `json:"requires_orchestration"`
	ProcessingErrors      []string `json:"processing_errors,omitempty"`
}

// IsRDFFormat checks if the given content type is an RDF format
func IsRDFFormat(contentType string) bool {
	rdfFormats := []string{
		"application/rdf+xml",
		"text/turtle",
		"application/ld+json",
		"application/n-triples",
		"application/n-quads",
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))
	for _, format := range rdfFormats {
		if contentType == format {
			return true
		}
	}
	return false
}

// ParseRelationshipsFromRDF extracts relationships from RDF data
func ParseRelationshipsFromRDF(ctx context.Context, data []byte, contentType string) (*ResourceEventData, error) {
	log.Context(ctx).Debugf("[ParseRelationshipsFromRDF] Parsing relationships from %s data, size=%d", contentType, len(data))

	eventData := &ResourceEventData{
		Format:                contentType,
		Size:                  len(data),
		ContentType:           contentType,
		Relationships:         make([]ResourceRelationship, 0),
		Triples:               make([]RDFTriple, 0),
		LinkedResourceURIs:    make([]string, 0),
		RequiresOrchestration: false,
	}

	if !IsRDFFormat(contentType) {
		log.Context(ctx).Debug("[ParseRelationshipsFromRDF] Not an RDF format, skipping relationship parsing")
		return eventData, nil
	}

	// Parse based on content type
	switch contentType {
	case "application/rdf+xml":
		return parseRDFXML(ctx, data, eventData)
	case "text/turtle":
		return parseTurtle(ctx, data, eventData)
	case "application/ld+json":
		return parseJSONLD(ctx, data, eventData)
	default:
		log.Context(ctx).Warnf("[ParseRelationshipsFromRDF] Unsupported RDF format: %s", contentType)
		return eventData, fmt.Errorf("unsupported RDF format: %s", contentType)
	}
}

// parseRDFXML parses RDF/XML data to extract relationships
func parseRDFXML(ctx context.Context, data []byte, eventData *ResourceEventData) (*ResourceEventData, error) {
	log.Context(ctx).Debug("[parseRDFXML] Parsing RDF/XML relationships")

	// For now, implement basic parsing - in production, you'd use a proper RDF library
	dataStr := string(data)

	// Look for basic RDF patterns
	if strings.Contains(dataStr, "rdf:about") || strings.Contains(dataStr, "rdf:resource") {
		eventData.RequiresOrchestration = true
		log.Context(ctx).Debug("[parseRDFXML] Found RDF references, marking for orchestration")
	}

	// TODO: Implement proper RDF/XML parsing with a library like rdf2go or similar
	// For now, return basic structure
	return eventData, nil
}

// parseTurtle parses Turtle data to extract relationships
func parseTurtle(ctx context.Context, data []byte, eventData *ResourceEventData) (*ResourceEventData, error) {
	log.Context(ctx).Debug("[parseTurtle] Parsing Turtle relationships")

	dataStr := string(data)
	lines := strings.Split(dataStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse @prefix declarations
		if strings.HasPrefix(line, "@prefix") {
			log.Context(ctx).Debugf("[parseTurtle] Found prefix declaration: %s", line)
			continue
		}

		// Parse triples (basic implementation)
		if strings.Contains(line, " ") && (strings.Contains(line, "<") || strings.Contains(line, ":")) {
			triple := parseTurtleLine(ctx, line)
			if triple != nil {
				eventData.Triples = append(eventData.Triples, *triple)

				// Check if this creates a relationship to another resource
				if !triple.IsLiteral && isResourceURI(triple.Object) {
					relationship := ResourceRelationship{
						Subject:    triple.Subject,
						Predicate:  triple.Predicate,
						Object:     triple.Object,
						IsExternal: isExternalURI(triple.Object),
						IsLiteral:  false,
					}
					eventData.Relationships = append(eventData.Relationships, relationship)

					if !relationship.IsExternal {
						eventData.LinkedResourceURIs = append(eventData.LinkedResourceURIs, triple.Object)
						eventData.RequiresOrchestration = true
					}
				}
			}
		}
	}

	log.Context(ctx).Debugf("[parseTurtle] Parsed %d triples, %d relationships", len(eventData.Triples), len(eventData.Relationships))
	return eventData, nil
}

// parseJSONLD parses JSON-LD data to extract relationships
func parseJSONLD(ctx context.Context, data []byte, eventData *ResourceEventData) (*ResourceEventData, error) {
	log.Context(ctx).Debug("[parseJSONLD] Parsing JSON-LD relationships")

	// TODO: Implement JSON-LD parsing
	// For now, just mark as requiring orchestration if it looks like JSON-LD
	dataStr := string(data)
	if strings.Contains(dataStr, "@context") || strings.Contains(dataStr, "@id") {
		eventData.RequiresOrchestration = true
		log.Context(ctx).Debug("[parseJSONLD] Found JSON-LD patterns, marking for orchestration")
	}

	return eventData, nil
}

// parseTurtleLine parses a single line of Turtle data into a triple
func parseTurtleLine(ctx context.Context, line string) *RDFTriple {
	// Basic Turtle parsing - in production use a proper parser
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	subject := strings.TrimSpace(parts[0])
	predicate := strings.TrimSpace(parts[1])
	object := strings.TrimSpace(strings.Join(parts[2:], " "))

	// Remove trailing dot if present
	object = strings.TrimSuffix(object, " .")
	object = strings.TrimSuffix(object, ".")

	// Determine if object is a literal
	isLiteral := strings.HasPrefix(object, "\"") ||
		strings.HasPrefix(object, "'") ||
		(!strings.HasPrefix(object, "<") && !strings.Contains(object, ":"))

	return &RDFTriple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
		IsLiteral: isLiteral,
	}
}

// isResourceURI checks if a string looks like a resource URI
func isResourceURI(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") ||
		strings.Contains(s, ":") && !strings.HasPrefix(s, "\"")
}

// isExternalURI checks if a URI is external to this pod
func isExternalURI(uri string) bool {
	uri = strings.TrimSpace(uri)
	if strings.HasPrefix(uri, "<") && strings.HasSuffix(uri, ">") {
		uri = uri[1 : len(uri)-1]
	}

	// Check for common external URI patterns
	return strings.HasPrefix(uri, "http://") ||
		strings.HasPrefix(uri, "https://") ||
		strings.HasPrefix(uri, "ftp://") ||
		strings.Contains(uri, "://")
}

// ValidateRelationships validates the parsed relationships
func ValidateRelationships(ctx context.Context, relationships []ResourceRelationship) []string {
	log.Context(ctx).Debugf("[ValidateRelationships] Validating %d relationships", len(relationships))

	var errors []string

	for i, rel := range relationships {
		if strings.TrimSpace(rel.Subject) == "" {
			errors = append(errors, fmt.Sprintf("relationship %d: subject cannot be empty", i))
		}
		if strings.TrimSpace(rel.Predicate) == "" {
			errors = append(errors, fmt.Sprintf("relationship %d: predicate cannot be empty", i))
		}
		if strings.TrimSpace(rel.Object) == "" {
			errors = append(errors, fmt.Sprintf("relationship %d: object cannot be empty", i))
		}
	}

	log.Context(ctx).Debugf("[ValidateRelationships] Found %d validation errors", len(errors))
	return errors
}
