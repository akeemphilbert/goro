package application

import (
	"context"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// ResourceOrchestrationService handles complex resource creation and relationship management
type ResourceOrchestrationService struct {
	resourceRepo        domain.ResourceRepository
	relationshipService *domain.ResourceRelationshipService
	unitOfWorkFactory   UnitOfWorkFactory
	eventDispatcher     domain.EventDispatcher
}

// NewResourceOrchestrationService creates a new resource orchestration service
func NewResourceOrchestrationService(
	resourceRepo domain.ResourceRepository,
	relationshipService *domain.ResourceRelationshipService,
	unitOfWorkFactory UnitOfWorkFactory,
	eventDispatcher domain.EventDispatcher,
) *ResourceOrchestrationService {
	return &ResourceOrchestrationService{
		resourceRepo:        resourceRepo,
		relationshipService: relationshipService,
		unitOfWorkFactory:   unitOfWorkFactory,
		eventDispatcher:     eventDispatcher,
	}
}

// OrchestratResourceCreation handles the creation of a resource and all its related resources
func (s *ResourceOrchestrationService) OrchestratResourceCreation(ctx context.Context, resource domain.Resource, eventData *domain.ResourceEventData) error {
	log.Context(ctx).Debugf("[OrchestratResourceCreation] Starting orchestration for resource: resourceID=%s, requiresOrchestration=%t",
		resource.ID(), eventData.RequiresOrchestration)

	if !eventData.RequiresOrchestration {
		log.Context(ctx).Debug("[OrchestratResourceCreation] No orchestration required")
		return nil
	}

	// Create unit of work for transactional consistency
	unitOfWork := s.unitOfWorkFactory()

	// Process relationships and create linked resources
	err := s.processResourceRelationships(ctx, resource, eventData, unitOfWork)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceCreation] Relationship processing failed: %v", err)
		return fmt.Errorf("failed to process resource relationships: %w", err)
	}

	// Commit the unit of work
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceCreation] Unit of work commit failed: %v", err)
		return fmt.Errorf("failed to commit orchestration changes: %w", err)
	}

	log.Context(ctx).Infof("Resource orchestration completed successfully: resourceID=%s, linkedResources=%d",
		resource.ID(), len(eventData.LinkedResourceURIs))

	return nil
}

// processResourceRelationships processes the relationships for a resource
func (s *ResourceOrchestrationService) processResourceRelationships(ctx context.Context, resource domain.Resource, eventData *domain.ResourceEventData, unitOfWork pericarpdomain.UnitOfWork) error {
	log.Context(ctx).Debugf("[processResourceRelationships] Processing %d relationships for resource: %s",
		len(eventData.Relationships), resource.ID())

	// Process each relationship
	for _, relationship := range eventData.Relationships {
		if relationship.IsExternal || relationship.IsLiteral {
			log.Context(ctx).Debugf("[processResourceRelationships] Skipping external/literal relationship: %s -> %s",
				relationship.Subject, relationship.Object)
			continue
		}

		log.Context(ctx).Debugf("[processResourceRelationships] Processing internal relationship: %s -> %s -> %s",
			relationship.Subject, relationship.Predicate, relationship.Object)

		// Check if linked resource already exists
		linkedResourceID := extractResourceIDFromURI(relationship.Object)
		if linkedResourceID == "" {
			log.Context(ctx).Warnf("[processResourceRelationships] Could not extract resource ID from URI: %s", relationship.Object)
			continue
		}

		exists, err := s.resourceRepo.Exists(ctx, linkedResourceID)
		if err != nil {
			log.Context(ctx).Debugf("[processResourceRelationships] Error checking resource existence: %v", err)
			return fmt.Errorf("failed to check existence of linked resource %s: %w", linkedResourceID, err)
		}

		if !exists {
			log.Context(ctx).Debugf("[processResourceRelationships] Creating linked resource: %s", linkedResourceID)
			linkedResource, err := s.relationshipService.CreateResourceFromRelationship(ctx, relationship, resource.ID())
			if err != nil {
				log.Context(ctx).Debugf("[processResourceRelationships] Failed to create linked resource: %v", err)
				return fmt.Errorf("failed to create linked resource %s: %w", linkedResourceID, err)
			}

			// Store the linked resource
			err = s.resourceRepo.Store(ctx, linkedResource)
			if err != nil {
				log.Context(ctx).Debugf("[processResourceRelationships] Failed to store linked resource: %v", err)
				return fmt.Errorf("failed to store linked resource %s: %w", linkedResourceID, err)
			}

			// Register events from the linked resource
			events := linkedResource.UncommittedEvents()
			if len(events) > 0 {
				unitOfWork.RegisterEvents(events)
			}

			// Emit resource linked event
			linkedEvent := domain.NewResourceLinkedEvent(resource.ID(), map[string]interface{}{
				"linkedResourceID": linkedResourceID,
				"relationship":     relationship,
				"createdAt":        time.Now(),
			})
			unitOfWork.RegisterEvents([]pericarpdomain.Event{linkedEvent})

			log.Context(ctx).Infof("Created and linked resource: %s -> %s", resource.ID(), linkedResourceID)
		} else {
			log.Context(ctx).Debugf("[processResourceRelationships] Linked resource already exists: %s", linkedResourceID)

			// Still emit a linked event for existing resources
			linkedEvent := domain.NewResourceLinkedEvent(resource.ID(), map[string]interface{}{
				"linkedResourceID": linkedResourceID,
				"relationship":     relationship,
				"alreadyExists":    true,
				"linkedAt":         time.Now(),
			})
			unitOfWork.RegisterEvents([]pericarpdomain.Event{linkedEvent})
		}
	}

	return nil
}

// OrchestratResourceUpdate handles the update of a resource and its relationships
func (s *ResourceOrchestrationService) OrchestratResourceUpdate(ctx context.Context, resource domain.Resource, eventData *domain.ResourceEventData) error {
	log.Context(ctx).Debugf("[OrchestratResourceUpdate] Starting update orchestration for resource: resourceID=%s", resource.ID())

	if !eventData.RequiresOrchestration {
		log.Context(ctx).Debug("[OrchestratResourceUpdate] No orchestration required")
		return nil
	}

	// Create unit of work for transactional consistency
	unitOfWork := s.unitOfWorkFactory()

	// Get existing relationships for comparison
	existingRelatedResources, err := s.relationshipService.GetRelatedResources(ctx, resource.ID())
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceUpdate] Failed to get existing relationships: %v", err)
		// Don't fail the update, just log the warning
		log.Context(ctx).Warnf("Could not retrieve existing relationships for comparison: %v", err)
		existingRelatedResources = []string{}
	}

	// Process new relationships
	err = s.processResourceRelationships(ctx, resource, eventData, unitOfWork)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceUpdate] Relationship processing failed: %v", err)
		return fmt.Errorf("failed to process updated relationships: %w", err)
	}

	// Find relationships that were removed
	newRelatedResources := make(map[string]bool)
	for _, uri := range eventData.LinkedResourceURIs {
		resourceID := extractResourceIDFromURI(uri)
		if resourceID != "" {
			newRelatedResources[resourceID] = true
		}
	}

	// Emit relationship updated events for removed relationships
	for _, existingResourceID := range existingRelatedResources {
		if !newRelatedResources[existingResourceID] {
			log.Context(ctx).Debugf("[OrchestratResourceUpdate] Relationship removed: %s -> %s", resource.ID(), existingResourceID)

			relationshipUpdatedEvent := domain.NewResourceRelationshipUpdatedEvent(resource.ID(), map[string]interface{}{
				"removedLinkTo": existingResourceID,
				"updatedAt":     time.Now(),
			})
			unitOfWork.RegisterEvents([]pericarpdomain.Event{relationshipUpdatedEvent})
		}
	}

	// Commit the unit of work
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceUpdate] Unit of work commit failed: %v", err)
		return fmt.Errorf("failed to commit update orchestration changes: %w", err)
	}

	log.Context(ctx).Infof("Resource update orchestration completed: resourceID=%s, relationships=%d",
		resource.ID(), len(eventData.Relationships))

	return nil
}

// ValidateResourceDeletion checks if a resource can be safely deleted
func (s *ResourceOrchestrationService) ValidateResourceDeletion(ctx context.Context, resourceID string) error {
	log.Context(ctx).Debugf("[ValidateResourceDeletion] Validating deletion for resource: %s", resourceID)

	// Get resources that depend on this resource
	// This would require a reverse relationship index in a production system
	// For now, we'll just log and allow deletion
	log.Context(ctx).Warnf("Resource deletion validation not fully implemented - allowing deletion of resource: %s", resourceID)

	return nil
}

// OrchestratResourceDeletion handles the deletion of a resource and cleanup of relationships
func (s *ResourceOrchestrationService) OrchestratResourceDeletion(ctx context.Context, resourceID string) error {
	log.Context(ctx).Debugf("[OrchestratResourceDeletion] Starting deletion orchestration for resource: %s", resourceID)

	// Validate deletion
	err := s.ValidateResourceDeletion(ctx, resourceID)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceDeletion] Deletion validation failed: %v", err)
		return fmt.Errorf("resource deletion validation failed: %w", err)
	}

	// Create unit of work for transactional consistency
	unitOfWork := s.unitOfWorkFactory()

	// Get related resources that might need cleanup
	relatedResources, err := s.relationshipService.GetRelatedResources(ctx, resourceID)
	if err != nil {
		log.Context(ctx).Warnf("Could not retrieve related resources for cleanup: %v", err)
		relatedResources = []string{}
	}

	// Emit relationship cleanup events
	for _, relatedResourceID := range relatedResources {
		log.Context(ctx).Debugf("[OrchestratResourceDeletion] Notifying relationship cleanup: %s -> %s", resourceID, relatedResourceID)

		relationshipUpdatedEvent := domain.NewResourceRelationshipUpdatedEvent(relatedResourceID, map[string]interface{}{
			"removedLinkFrom": resourceID,
			"updatedAt":       time.Now(),
			"reason":          "source_resource_deleted",
		})
		unitOfWork.RegisterEvents([]pericarpdomain.Event{relationshipUpdatedEvent})
	}

	// Commit the unit of work
	_, err = unitOfWork.Commit(ctx)
	if err != nil {
		log.Context(ctx).Debugf("[OrchestratResourceDeletion] Unit of work commit failed: %v", err)
		return fmt.Errorf("failed to commit deletion orchestration changes: %w", err)
	}

	log.Context(ctx).Infof("Resource deletion orchestration completed: resourceID=%s, notifiedResources=%d",
		resourceID, len(relatedResources))

	return nil
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
