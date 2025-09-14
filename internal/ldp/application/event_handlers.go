package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// ResourceEventHandler handles resource events and updates the repository accordingly
type ResourceEventHandler struct {
	repo domain.ResourceRepository
}

// NewResourceEventHandler creates a new resource event handler
func NewResourceEventHandler(repo domain.ResourceRepository) *ResourceEventHandler {
	return &ResourceEventHandler{
		repo: repo,
	}
}

// EventTypes returns the list of event types this handler can process
func (h *ResourceEventHandler) EventTypes() []string {
	return []string{
		"resource.created",
		"resource.updated",
		"resource.deleted",
		"resource.created_with_relations",
		"resource.linked",
		"resource.relationship_updated",
	}
}

// Handle processes resource events and updates the repository
func (h *ResourceEventHandler) Handle(ctx context.Context, envelope pericarpdomain.Envelope) error {
	event := envelope.Event()

	// Check if this is a resource event
	if entityEvent, ok := event.(*pericarpdomain.EntityEvent); ok {
		if entityEvent.EntityType != "resource" {
			// Not a resource event, ignore
			return nil
		}

		switch entityEvent.Type {
		case domain.EventTypeResourceCreated:
			return h.handleResourceCreated(ctx, entityEvent)
		case domain.EventTypeResourceUpdated:
			return h.handleResourceUpdated(ctx, entityEvent)
		case domain.EventTypeResourceDeleted:
			return h.handleResourceDeleted(ctx, entityEvent)
		case domain.EventTypeResourceCreatedWithRelations:
			return h.handleResourceCreatedWithRelations(ctx, entityEvent)
		case domain.EventTypeResourceLinked:
			return h.handleResourceLinked(ctx, entityEvent)
		case domain.EventTypeResourceRelationshipUpdated:
			return h.handleResourceRelationshipUpdated(ctx, entityEvent)
		default:
			// Unknown event type, log and ignore
			fmt.Printf("Unknown resource event type: %s\n", entityEvent.Type)
			return nil
		}
	}

	return nil
}

// handleResourceCreated handles resource created events
func (h *ResourceEventHandler) handleResourceCreated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract resource data from event payload
	resource, err := h.reconstructResourceFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to reconstruct resource from created event: %w", err)
	}

	// Store the resource in the repository
	if err := h.repo.Store(ctx, resource); err != nil {
		return fmt.Errorf("failed to store resource in repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s created\n", event.AggregateID())
	return nil
}

// handleResourceUpdated handles resource updated events
func (h *ResourceEventHandler) handleResourceUpdated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract resource data from event payload
	resource, err := h.reconstructResourceFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to reconstruct resource from updated event: %w", err)
	}

	// Update the resource in the repository
	if err := h.repo.Store(ctx, resource); err != nil {
		return fmt.Errorf("failed to update resource in repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s updated\n", event.AggregateID())
	return nil
}

// handleResourceDeleted handles resource deleted events
func (h *ResourceEventHandler) handleResourceDeleted(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Delete the resource from the repository
	if err := h.repo.Delete(ctx, event.AggregateID()); err != nil {
		return fmt.Errorf("failed to delete resource from repository: %w", err)
	}

	fmt.Printf("Repository updated: resource %s deleted\n", event.AggregateID())
	return nil
}

// handleResourceCreatedWithRelations handles resource created with relations events
func (h *ResourceEventHandler) handleResourceCreatedWithRelations(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// This event indicates that a resource needs relationship processing
	// In a production system, this might trigger the orchestration service
	fmt.Printf("Resource created with relationships: %s (requires orchestration)\n", event.AggregateID())

	// For now, we'll just log the event - the orchestration should be handled
	// by a separate service that listens for these events
	return nil
}

// handleResourceLinked handles resource linked events
func (h *ResourceEventHandler) handleResourceLinked(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract linked resource information from event payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		fmt.Printf("Failed to parse resource linked event payload: %v\n", err)
		return nil // Don't fail the event processing
	}

	linkedResourceID, _ := payload["linkedResourceID"].(string)
	alreadyExists, _ := payload["alreadyExists"].(bool)

	if alreadyExists {
		fmt.Printf("Resource linked to existing resource: %s -> %s\n", event.AggregateID(), linkedResourceID)
	} else {
		fmt.Printf("Resource linked to new resource: %s -> %s\n", event.AggregateID(), linkedResourceID)
	}

	return nil
}

// handleResourceRelationshipUpdated handles resource relationship updated events
func (h *ResourceEventHandler) handleResourceRelationshipUpdated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract relationship change information from event payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		fmt.Printf("Failed to parse relationship updated event payload: %v\n", err)
		return nil // Don't fail the event processing
	}

	// Handle different types of relationship updates
	if removedLinkTo, exists := payload["removedLinkTo"]; exists {
		fmt.Printf("Resource relationship removed: %s no longer links to %s\n", event.AggregateID(), removedLinkTo)
	}

	if removedLinkFrom, exists := payload["removedLinkFrom"]; exists {
		reason, _ := payload["reason"].(string)
		fmt.Printf("Resource relationship removed: %s no longer linked from %s (reason: %s)\n",
			event.AggregateID(), removedLinkFrom, reason)
	}

	return nil
}

// reconstructResourceFromEvent reconstructs a resource from an event payload
func (h *ResourceEventHandler) reconstructResourceFromEvent(event *pericarpdomain.EntityEvent) (domain.Resource, error) {
	// The payload contains the resource data encoded as JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// Extract resource information from payload
	contentType, ok := payload["contentType"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid contentType in event payload")
	}

	size, ok := payload["size"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid size in event payload")
	}

	// For now, we'll create a placeholder resource since we don't have the actual data
	// In a real implementation, you might store the data in the event or have a separate mechanism
	// to retrieve the current state of the resource
	data := make([]byte, int(size))

	// Create the resource
	resource := domain.NewResource(context.Background(), event.AggregateID(), contentType, data)

	// Clear events since this is a reconstruction from events
	resource.MarkEventsAsCommitted()

	return resource, nil
}

// EventHandlerRegistrar registers event handlers with the event dispatcher
type EventHandlerRegistrar struct {
	dispatcher pericarpdomain.EventDispatcher
	handlers   []pericarpdomain.EventHandler
}

// NewEventHandlerRegistrar creates a new event handler registrar
func NewEventHandlerRegistrar(dispatcher pericarpdomain.EventDispatcher) *EventHandlerRegistrar {
	return &EventHandlerRegistrar{
		dispatcher: dispatcher,
		handlers:   make([]pericarpdomain.EventHandler, 0),
	}
}

// RegisterResourceEventHandler registers the resource event handler for all resource events
func (r *EventHandlerRegistrar) RegisterResourceEventHandler(handler *ResourceEventHandler) error {
	// Register for all resource event types
	eventTypes := []string{
		"resource.created",
		"resource.updated",
		"resource.deleted",
	}

	for _, eventType := range eventTypes {
		if err := r.dispatcher.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to subscribe to event type %s: %w", eventType, err)
		}
	}

	r.handlers = append(r.handlers, handler)
	fmt.Printf("Registered resource event handler for events: %v\n", eventTypes)
	return nil
}

// ContainerEventHandler handles container events and updates the repository accordingly
type ContainerEventHandler struct {
	containerRepo domain.ContainerRepository
}

// NewContainerEventHandler creates a new container event handler
func NewContainerEventHandler(containerRepo domain.ContainerRepository) *ContainerEventHandler {
	return &ContainerEventHandler{
		containerRepo: containerRepo,
	}
}

// EventTypes returns the list of event types this handler can process
func (h *ContainerEventHandler) EventTypes() []string {
	return []string{
		"container.created",
		"container.updated",
		"container.deleted",
		"container.member_added",
		"container.member_removed",
	}
}

// Handle processes container events and updates the repository
func (h *ContainerEventHandler) Handle(ctx context.Context, envelope pericarpdomain.Envelope) error {
	event := envelope.Event()

	// Check if this is a container event
	if entityEvent, ok := event.(*pericarpdomain.EntityEvent); ok {
		if entityEvent.EntityType != "container" {
			// Not a container event, ignore
			return nil
		}

		switch entityEvent.Type {
		case domain.EventTypeContainerCreated:
			return h.handleContainerCreated(ctx, entityEvent)
		case domain.EventTypeContainerUpdated:
			return h.handleContainerUpdated(ctx, entityEvent)
		case domain.EventTypeContainerDeleted:
			return h.handleContainerDeleted(ctx, entityEvent)
		case domain.EventTypeMemberAdded:
			return h.handleMemberAdded(ctx, entityEvent)
		case domain.EventTypeMemberRemoved:
			return h.handleMemberRemoved(ctx, entityEvent)
		default:
			// Unknown event type, log and ignore
			fmt.Printf("Unknown container event type: %s\n", entityEvent.Type)
			return nil
		}
	}

	return nil
}

// handleContainerCreated handles container created events
func (h *ContainerEventHandler) handleContainerCreated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Reconstruct container from event
	container, err := h.reconstructContainerFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to reconstruct container from created event: %w", err)
	}

	// Store the container in the repository
	if err := h.containerRepo.CreateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to store container in repository: %w", err)
	}

	fmt.Printf("Repository updated: container %s created\n", event.AggregateID())
	return nil
}

// handleContainerUpdated handles container updated events
func (h *ContainerEventHandler) handleContainerUpdated(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Get existing container and apply updates
	container, err := h.containerRepo.GetContainer(ctx, event.AggregateID())
	if err != nil {
		return fmt.Errorf("failed to get container for update: %w", err)
	}

	// Apply updates from event payload
	if err := h.applyContainerUpdatesFromEvent(container, event); err != nil {
		return fmt.Errorf("failed to apply container updates from event: %w", err)
	}

	// Update the container in the repository
	if err := h.containerRepo.UpdateContainer(ctx, container); err != nil {
		return fmt.Errorf("failed to update container in repository: %w", err)
	}

	fmt.Printf("Repository updated: container %s updated\n", event.AggregateID())
	return nil
}

// handleContainerDeleted handles container deleted events
func (h *ContainerEventHandler) handleContainerDeleted(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Delete the container from the repository
	if err := h.containerRepo.DeleteContainer(ctx, event.AggregateID()); err != nil {
		return fmt.Errorf("failed to delete container from repository: %w", err)
	}

	fmt.Printf("Repository updated: container %s deleted\n", event.AggregateID())
	return nil
}

// handleMemberAdded handles member added events
func (h *ContainerEventHandler) handleMemberAdded(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract member information from event payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal member added event payload: %w", err)
	}

	memberID, ok := payload["memberID"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid memberID in member added event payload")
	}

	// Extract additional resource information if available
	resourceType, _ := payload["resourceType"].(string)
	contentType, _ := payload["contentType"].(string)
	size, _ := payload["size"].(float64)

	// Add member to container via repository
	if err := h.containerRepo.AddMember(ctx, event.AggregateID(), memberID); err != nil {
		return fmt.Errorf("failed to add member to container in repository: %w", err)
	}

	fmt.Printf("Repository updated: member %s (type: %s, contentType: %s, size: %.0f) added to container %s\n",
		memberID, resourceType, contentType, size, event.AggregateID())
	return nil
}

// handleMemberRemoved handles member removed events
func (h *ContainerEventHandler) handleMemberRemoved(ctx context.Context, event *pericarpdomain.EntityEvent) error {
	// Extract member ID from event payload
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal member removed event payload: %w", err)
	}

	memberID, ok := payload["memberID"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid memberID in member removed event payload")
	}

	// Remove member from container via repository
	if err := h.containerRepo.RemoveMember(ctx, event.AggregateID(), memberID); err != nil {
		return fmt.Errorf("failed to remove member from container in repository: %w", err)
	}

	fmt.Printf("Repository updated: member %s removed from container %s\n", memberID, event.AggregateID())
	return nil
}

// reconstructContainerFromEvent reconstructs a container from a created event payload
func (h *ContainerEventHandler) reconstructContainerFromEvent(event *pericarpdomain.EntityEvent) (*domain.Container, error) {
	// The payload contains the container data encoded as JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// Extract container information from payload
	parentID, _ := payload["parentID"].(string) // Optional, can be empty
	containerTypeStr, ok := payload["containerType"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid containerType in event payload")
	}

	containerType := domain.ContainerType(containerTypeStr)
	if !containerType.IsValid() {
		return nil, fmt.Errorf("invalid container type: %s", containerTypeStr)
	}

	// Create the container
	container := domain.NewContainer(context.Background(), event.AggregateID(), parentID, containerType)

	// Apply any additional metadata from payload
	if title, ok := payload["title"].(string); ok && title != "" {
		container.SetTitle(title)
	}
	if description, ok := payload["description"].(string); ok && description != "" {
		container.SetDescription(description)
	}

	// Clear events since this is a reconstruction from events
	container.MarkEventsAsCommitted()

	return container, nil
}

// applyContainerUpdatesFromEvent applies updates to a container from an event payload
func (h *ContainerEventHandler) applyContainerUpdatesFromEvent(container *domain.Container, event *pericarpdomain.EntityEvent) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	// Apply updates based on what's in the payload
	if title, ok := payload["title"].(string); ok {
		container.SetTitle(title)
	}
	if description, ok := payload["description"].(string); ok {
		container.SetDescription(description)
	}

	// Clear events since we're applying from events
	container.MarkEventsAsCommitted()

	return nil
}

// RegisterContainerEventHandler registers the container event handler for all container events
func (r *EventHandlerRegistrar) RegisterContainerEventHandler(handler *ContainerEventHandler) error {
	// Register for all container event types
	eventTypes := []string{
		"container.created",
		"container.updated",
		"container.deleted",
		"container.member_added",
		"container.member_removed",
	}

	for _, eventType := range eventTypes {
		if err := r.dispatcher.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to subscribe to event type %s: %w", eventType, err)
		}
	}

	r.handlers = append(r.handlers, handler)
	fmt.Printf("Registered container event handler for events: %v\n", eventTypes)
	return nil
}

// RegisterAllHandlers registers all event handlers
func (r *EventHandlerRegistrar) RegisterAllHandlers(repo domain.ResourceRepository) error {
	// Create and register resource event handler
	resourceHandler := NewResourceEventHandler(repo)
	if err := r.RegisterResourceEventHandler(resourceHandler); err != nil {
		return fmt.Errorf("failed to register resource event handler: %w", err)
	}

	return nil
}

// RegisterEventPersistenceHandler registers the event persistence handler for all events
func (r *EventHandlerRegistrar) RegisterEventPersistenceHandler(handler *EventPersistenceHandler) error {
	// Subscribe to all events by registering for all known event types
	allEventTypes := []string{
		// Resource events
		"resource.created",
		"resource.updated",
		"resource.deleted",
		"resource.created_with_relations",
		"resource.linked",
		"resource.relationship_updated",
		// Container events
		"container.created",
		"container.updated",
		"container.deleted",
		"container.member_added",
		"container.member_removed",
	}

	for _, eventType := range allEventTypes {
		if err := r.dispatcher.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to subscribe persistence handler to event type %s: %w", eventType, err)
		}
	}

	r.handlers = append(r.handlers, handler)
	fmt.Printf("Registered event persistence handler for all events: %v\n", allEventTypes)
	return nil
}

// RegisterAllHandlersWithContainer registers all event handlers including container handlers
func (r *EventHandlerRegistrar) RegisterAllHandlersWithContainer(repo domain.ResourceRepository, containerRepo domain.ContainerRepository) error {
	// First, register the event persistence handler to capture all events
	persistenceHandler := NewEventPersistenceHandler()
	if err := r.RegisterEventPersistenceHandler(persistenceHandler); err != nil {
		return fmt.Errorf("failed to register event persistence handler: %w", err)
	}

	// Create and register resource event handler
	resourceHandler := NewResourceEventHandler(repo)
	if err := r.RegisterResourceEventHandler(resourceHandler); err != nil {
		return fmt.Errorf("failed to register resource event handler: %w", err)
	}

	// Create and register container event handler
	containerHandler := NewContainerEventHandler(containerRepo)
	if err := r.RegisterContainerEventHandler(containerHandler); err != nil {
		return fmt.Errorf("failed to register container event handler: %w", err)
	}

	return nil
}
