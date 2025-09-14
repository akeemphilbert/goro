package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// ResourceRelationshipService handles resource relationship parsing and validation
type ResourceRelationshipService struct {
	resourceRepo ResourceRepository
}

// NewResourceRelationshipService creates a new resource relationship service
func NewResourceRelationshipService(resourceRepo ResourceRepository) *ResourceRelationshipService {
	return &ResourceRelationshipService{
		resourceRepo: resourceRepo,
	}
}

// ProcessResourceWithRelationships processes a resource and extracts its relationships
func (s *ResourceRelationshipService) ProcessResourceWithRelationships(ctx context.Context, resource Resource) (*ResourceEventData, error) {
	log.Context(ctx).Debugf("[ProcessResourceWithRelationships] Processing resource: resourceID=%s, contentType=%s",
		resource.ID(), resource.GetContentType())

	// Create base event data
	eventData := &ResourceEventData{
		Format:                resource.GetContentType(),
		Size:                  resource.GetSize(),
		ContentType:           resource.GetContentType(),
		Relationships:         make([]ResourceRelationship, 0),
		Triples:               make([]RDFTriple, 0),
		LinkedResourceURIs:    make([]string, 0),
		RequiresOrchestration: false,
		ProcessingErrors:      make([]string, 0),
		CreatedAt:             time.Now().Format(time.RFC3339),
	}

	// Only process RDF formats
	if !IsRDFFormat(resource.GetContentType()) {
		log.Context(ctx).Debug("[ProcessResourceWithRelationships] Not an RDF format, returning basic event data")
		return eventData, nil
	}

	log.Context(ctx).Debug("[ProcessResourceWithRelationships] Processing RDF relationships")

	// Parse relationships from the resource data
	parsedEventData, err := ParseRelationshipsFromRDF(ctx, resource.GetData(), resource.GetContentType())
	if err != nil {
		log.Context(ctx).Debugf("[ProcessResourceWithRelationships] Error parsing relationships: %v", err)
		eventData.ProcessingErrors = append(eventData.ProcessingErrors, fmt.Sprintf("parsing error: %v", err))
		return eventData, nil // Return partial data instead of failing completely
	}

	// Merge parsed data with base event data
	eventData.Relationships = parsedEventData.Relationships
	eventData.Triples = parsedEventData.Triples
	eventData.LinkedResourceURIs = parsedEventData.LinkedResourceURIs
	eventData.RequiresOrchestration = parsedEventData.RequiresOrchestration

	// Validate relationships
	validationErrors := ValidateRelationships(ctx, eventData.Relationships)
	if len(validationErrors) > 0 {
		log.Context(ctx).Debugf("[ProcessResourceWithRelationships] Validation errors found: %v", validationErrors)
		eventData.ProcessingErrors = append(eventData.ProcessingErrors, validationErrors...)
	}

	// Check for circular references
	circularRefs := s.checkCircularReferences(ctx, resource.ID(), eventData.LinkedResourceURIs)
	if len(circularRefs) > 0 {
		log.Context(ctx).Warnf("[ProcessResourceWithRelationships] Circular references detected: %v", circularRefs)
		eventData.ProcessingErrors = append(eventData.ProcessingErrors, circularRefs...)
	}

	// Verify linked resources exist or can be created
	err = s.validateLinkedResources(ctx, eventData.LinkedResourceURIs)
	if err != nil {
		log.Context(ctx).Debugf("[ProcessResourceWithRelationships] Linked resource validation failed: %v", err)
		eventData.ProcessingErrors = append(eventData.ProcessingErrors, fmt.Sprintf("linked resource validation: %v", err))
	}

	log.Context(ctx).Infof("Resource relationship processing completed: resourceID=%s, relationships=%d, linkedResources=%d, requiresOrchestration=%t, errors=%d",
		resource.ID(), len(eventData.Relationships), len(eventData.LinkedResourceURIs), eventData.RequiresOrchestration, len(eventData.ProcessingErrors))

	return eventData, nil
}

// CreateResourceFromRelationship creates a new resource based on a relationship
func (s *ResourceRelationshipService) CreateResourceFromRelationship(ctx context.Context, relationship ResourceRelationship, sourceResourceID string) (Resource, error) {
	log.Context(ctx).Debugf("[CreateResourceFromRelationship] Creating resource from relationship: subject=%s, predicate=%s, object=%s",
		relationship.Subject, relationship.Predicate, relationship.Object)

	if relationship.IsExternal || relationship.IsLiteral {
		log.Context(ctx).Debug("[CreateResourceFromRelationship] Relationship is external or literal, skipping creation")
		return nil, fmt.Errorf("cannot create resource from external or literal relationship")
	}

	// Extract resource ID from URI
	resourceID := extractResourceIDFromURI(relationship.Object)
	if resourceID == "" {
		log.Context(ctx).Debug("[CreateResourceFromRelationship] Could not extract resource ID from URI")
		return nil, fmt.Errorf("invalid resource URI: %s", relationship.Object)
	}

	// Check if resource already exists
	exists, err := s.resourceRepo.Exists(ctx, resourceID)
	if err != nil {
		log.Context(ctx).Debugf("[CreateResourceFromRelationship] Error checking resource existence: %v", err)
		return nil, fmt.Errorf("failed to check resource existence: %w", err)
	}

	if exists {
		log.Context(ctx).Debugf("[CreateResourceFromRelationship] Resource already exists: %s", resourceID)
		// Return existing resource
		return s.resourceRepo.Retrieve(ctx, resourceID)
	}

	// Create a basic placeholder resource
	log.Context(ctx).Debug("[CreateResourceFromRelationship] Creating placeholder resource")
	placeholderData := s.generatePlaceholderRDF(ctx, resourceID, relationship, sourceResourceID)
	resource := NewResource(ctx, resourceID, "text/turtle", placeholderData)

	log.Context(ctx).Infof("Created placeholder resource from relationship: resourceID=%s, sourceResource=%s", resourceID, sourceResourceID)
	return resource, nil
}

// checkCircularReferences checks for circular references in linked resources
func (s *ResourceRelationshipService) checkCircularReferences(ctx context.Context, resourceID string, linkedURIs []string) []string {
	log.Context(ctx).Debugf("[checkCircularReferences] Checking circular references for resource: %s", resourceID)

	var circularRefs []string
	for _, uri := range linkedURIs {
		linkedResourceID := extractResourceIDFromURI(uri)
		if linkedResourceID == resourceID {
			circularRefs = append(circularRefs, fmt.Sprintf("circular reference detected: %s -> %s", resourceID, linkedResourceID))
		}
	}

	return circularRefs
}

// validateLinkedResources validates that linked resources can be resolved
func (s *ResourceRelationshipService) validateLinkedResources(ctx context.Context, linkedURIs []string) error {
	log.Context(ctx).Debugf("[validateLinkedResources] Validating %d linked resources", len(linkedURIs))

	for _, uri := range linkedURIs {
		resourceID := extractResourceIDFromURI(uri)
		if resourceID == "" {
			return fmt.Errorf("invalid resource URI: %s", uri)
		}

		// For now, we assume all resources can be created if they don't exist
		// In a more sophisticated system, you might check URI schemes, permissions, etc.
		log.Context(ctx).Debugf("[validateLinkedResources] Validated linked resource URI: %s -> %s", uri, resourceID)
	}

	return nil
}

// generatePlaceholderRDF generates basic RDF content for a placeholder resource
func (s *ResourceRelationshipService) generatePlaceholderRDF(ctx context.Context, resourceID string, relationship ResourceRelationship, sourceResourceID string) []byte {
	log.Context(ctx).Debugf("[generatePlaceholderRDF] Generating placeholder RDF for resource: %s", resourceID)

	// Create basic Turtle content
	rdfContent := fmt.Sprintf(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix ldp: <http://www.w3.org/ns/ldp#> .

<%s> rdf:type ldp:Resource ;
    rdfs:label "Placeholder resource created from relationship" ;
    ldp:createdBy <%s> ;
    ldp:createdFrom "%s" .
`, resourceID, sourceResourceID, relationship.Predicate)

	return []byte(rdfContent)
}

// extractResourceIDFromURI extracts a resource ID from a URI
func extractResourceIDFromURI(uri string) string {
	// Remove angle brackets if present
	if len(uri) > 2 && uri[0] == '<' && uri[len(uri)-1] == '>' {
		uri = uri[1 : len(uri)-1]
	}

	// For now, use the URI as-is as the resource ID
	// In a more sophisticated system, you might:
	// - Extract the fragment (#resource-id)
	// - Extract the last path segment
	// - Use a mapping service
	return uri
}

// UpdateResourceRelationships updates the relationships for an existing resource
func (s *ResourceRelationshipService) UpdateResourceRelationships(ctx context.Context, resource Resource, newRelationships []ResourceRelationship) error {
	log.Context(ctx).Debugf("[UpdateResourceRelationships] Updating relationships for resource: %s", resource.ID())

	// Validate new relationships
	validationErrors := ValidateRelationships(ctx, newRelationships)
	if len(validationErrors) > 0 {
		log.Context(ctx).Debugf("[UpdateResourceRelationships] Validation errors: %v", validationErrors)
		return fmt.Errorf("relationship validation failed: %v", validationErrors)
	}

	// Check for circular references
	linkedURIs := make([]string, 0)
	for _, rel := range newRelationships {
		if !rel.IsExternal && !rel.IsLiteral {
			linkedURIs = append(linkedURIs, rel.Object)
		}
	}

	circularRefs := s.checkCircularReferences(ctx, resource.ID(), linkedURIs)
	if len(circularRefs) > 0 {
		log.Context(ctx).Warnf("[UpdateResourceRelationships] Circular references detected: %v", circularRefs)
		return fmt.Errorf("circular references detected: %v", circularRefs)
	}

	log.Context(ctx).Infof("Resource relationships updated successfully: resourceID=%s, relationships=%d", resource.ID(), len(newRelationships))
	return nil
}

// GetRelatedResources retrieves all resources related to a given resource
func (s *ResourceRelationshipService) GetRelatedResources(ctx context.Context, resourceID string) ([]string, error) {
	log.Context(ctx).Debugf("[GetRelatedResources] Getting related resources for: %s", resourceID)

	// Retrieve the resource
	resource, err := s.resourceRepo.Retrieve(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	// Process relationships
	eventData, err := s.ProcessResourceWithRelationships(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to process relationships: %w", err)
	}

	relatedIDs := make([]string, 0)
	for _, uri := range eventData.LinkedResourceURIs {
		relatedID := extractResourceIDFromURI(uri)
		if relatedID != "" && relatedID != resourceID {
			relatedIDs = append(relatedIDs, relatedID)
		}
	}

	log.Context(ctx).Debugf("[GetRelatedResources] Found %d related resources", len(relatedIDs))
	return relatedIDs, nil
}
