package domain

import (
	"context"
	"fmt"
	"strings"
)

// ContainerValidator provides validation functionality for containers
type ContainerValidator struct{}

// NewContainerValidator creates a new container validator
func NewContainerValidator() *ContainerValidator {
	return &ContainerValidator{}
}

// ValidateContainer performs comprehensive validation on a container
func (v *ContainerValidator) ValidateContainer(ctx context.Context, container ContainerResource) error {
	if container == nil {
		return NewContainerError("INVALID_CONTAINER", "container cannot be nil")
	}

	// Additional check for nil underlying value in interface
	if containerPtr, ok := container.(*Container); ok && containerPtr == nil {
		return NewContainerError("INVALID_CONTAINER", "container cannot be nil")
	}

	// Validate container ID
	if err := v.ValidateContainerID(container.ID()); err != nil {
		return err
	}

	// Validate container type
	if err := v.ValidateContainerType(container.GetContainerType()); err != nil {
		return err
	}

	// Validate parent ID if present
	if container.GetParentID() != "" {
		if err := v.ValidateContainerID(container.GetParentID()); err != nil {
			return WrapContainerError(err, "INVALID_PARENT_ID", "invalid parent container ID")
		}
	}

	// Validate members
	if err := v.ValidateMembers(container.GetMembers()); err != nil {
		return err
	}

	return nil
}

// ValidateContainerID validates a container ID
func (v *ContainerValidator) ValidateContainerID(id string) error {
	if id == "" {
		return NewContainerError("EMPTY_CONTAINER_ID", "container ID cannot be empty")
	}

	if len(id) > 255 {
		return NewContainerError("CONTAINER_ID_TOO_LONG", "container ID cannot exceed 255 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(id, "/\\:*?\"<>|") {
		return NewContainerError("INVALID_CONTAINER_ID_CHARS", "container ID contains invalid characters")
	}

	return nil
}

// ValidateContainerType validates a container type
func (v *ContainerValidator) ValidateContainerType(containerType ContainerType) error {
	if !containerType.IsValid() {
		return NewContainerError("INVALID_CONTAINER_TYPE", fmt.Sprintf("unsupported container type: %s", containerType))
	}
	return nil
}

// ValidateMembers validates container members
func (v *ContainerValidator) ValidateMembers(members []string) error {
	seen := make(map[string]bool)

	for _, memberID := range members {
		if memberID == "" {
			return NewContainerError("EMPTY_MEMBER_ID", "member ID cannot be empty")
		}

		if seen[memberID] {
			return NewContainerError("DUPLICATE_MEMBER", fmt.Sprintf("duplicate member ID: %s", memberID))
		}

		seen[memberID] = true
	}

	return nil
}

// ValidateHierarchy validates container hierarchy to prevent circular references
func (v *ContainerValidator) ValidateHierarchy(ctx context.Context, containerID, parentID string, repo ContainerRepository) error {
	if containerID == parentID {
		return NewContainerError("SELF_REFERENCE", "container cannot be its own parent")
	}

	if parentID == "" {
		return nil // Root container, no hierarchy to validate
	}

	// Build ancestor path to check for circular references
	ancestorPath, err := v.buildAncestorPath(ctx, parentID, repo)
	if err != nil {
		return WrapContainerError(err, "HIERARCHY_VALIDATION_FAILED", "failed to validate container hierarchy")
	}

	// Check for circular reference
	for _, ancestorID := range ancestorPath {
		if ancestorID == containerID {
			return NewContainerError("CIRCULAR_REFERENCE", fmt.Sprintf("circular reference detected: container %s is already in the hierarchy path", containerID))
		}
	}

	return nil
}

// buildAncestorPath builds the complete ancestor path for a container
func (v *ContainerValidator) buildAncestorPath(ctx context.Context, containerID string, repo ContainerRepository) ([]string, error) {
	var path []string
	currentID := containerID
	visited := make(map[string]bool)

	for currentID != "" {
		// Check for cycles in the existing hierarchy
		if visited[currentID] {
			return nil, NewContainerError("EXISTING_CIRCULAR_REFERENCE", fmt.Sprintf("circular reference detected in existing hierarchy at container %s", currentID))
		}
		visited[currentID] = true
		path = append(path, currentID)

		// Get parent container
		container, err := repo.GetContainer(ctx, currentID)
		if err != nil {
			if IsContainerNotFound(err) {
				break // Parent doesn't exist, end of path
			}
			return nil, err
		}

		currentID = container.GetParentID()
	}

	return path, nil
}

// ValidateContainerForDeletion validates that a container can be safely deleted
func (v *ContainerValidator) ValidateContainerForDeletion(ctx context.Context, container *Container, repo ContainerRepository) error {
	if container == nil {
		return NewContainerError("INVALID_CONTAINER", "container cannot be nil")
	}

	// Check if container is empty
	if !container.IsEmpty() {
		return NewContainerError("CONTAINER_NOT_EMPTY", fmt.Sprintf("container contains %d members and cannot be deleted", container.GetMemberCount()))
	}

	// Check if container has child containers
	children, err := repo.GetChildren(ctx, container.ID())
	if err != nil {
		return WrapContainerError(err, "DELETION_VALIDATION_FAILED", "failed to check for child containers")
	}

	if len(children) > 0 {
		return NewContainerError("CONTAINER_HAS_CHILDREN", fmt.Sprintf("container has %d child containers and cannot be deleted", len(children)))
	}

	return nil
}

// ValidateMembershipOperation validates adding or removing a member from a container
func (v *ContainerValidator) ValidateMembershipOperation(ctx context.Context, containerID, memberID string, operation string, repo ContainerRepository) error {
	if containerID == "" {
		return NewContainerError("EMPTY_CONTAINER_ID", "container ID cannot be empty")
	}

	if memberID == "" {
		return NewContainerError("EMPTY_MEMBER_ID", "member ID cannot be empty")
	}

	// Prevent container from being a member of itself
	if containerID == memberID {
		return NewContainerError("SELF_MEMBERSHIP", "container cannot be a member of itself")
	}

	// Get the container
	container, err := repo.GetContainer(ctx, containerID)
	if err != nil {
		return WrapContainerError(err, "MEMBERSHIP_VALIDATION_FAILED", "failed to retrieve container for membership validation")
	}

	switch operation {
	case "add":
		// Check if member already exists
		if container.HasMember(memberID) {
			return NewContainerError("MEMBER_ALREADY_EXISTS", fmt.Sprintf("member %s already exists in container %s", memberID, containerID))
		}

		// If the member is a container, validate hierarchy
		if memberContainer, err := repo.GetContainer(ctx, memberID); err == nil {
			// Member is a container, validate hierarchy
			if err := v.ValidateHierarchy(ctx, memberID, containerID, repo); err != nil {
				return WrapContainerError(err, "HIERARCHY_VALIDATION_FAILED", "adding member would create circular reference")
			}

			// Ensure the member container's parent would be valid
			memberParentID := memberContainer.GetParentID()
			if memberParentID != "" && memberParentID != containerID {
				return NewContainerError("INVALID_PARENT_CHANGE", fmt.Sprintf("container %s already has parent %s", memberID, memberParentID))
			}
		}

	case "remove":
		// Check if member exists
		if !container.HasMember(memberID) {
			return NewContainerError("MEMBER_NOT_FOUND", fmt.Sprintf("member %s not found in container %s", memberID, containerID))
		}

	default:
		return NewContainerError("INVALID_OPERATION", fmt.Sprintf("invalid membership operation: %s", operation))
	}

	return nil
}

// ValidateListingOptions validates options for listing container contents
func (v *ContainerValidator) ValidateListingOptions(options ListingOptions) error {
	if !options.IsValid() {
		return NewContainerError("INVALID_LISTING_OPTIONS", "invalid listing options provided")
	}

	// Additional validation for filter options
	if options.Filter.SizeMin != nil && options.Filter.SizeMax != nil {
		if *options.Filter.SizeMin > *options.Filter.SizeMax {
			return NewContainerError("INVALID_SIZE_RANGE", "minimum size cannot be greater than maximum size")
		}
	}

	if options.Filter.CreatedAfter != nil && options.Filter.CreatedBefore != nil {
		if options.Filter.CreatedAfter.After(*options.Filter.CreatedBefore) {
			return NewContainerError("INVALID_DATE_RANGE", "created after date cannot be later than created before date")
		}
	}

	return nil
}

// ValidateContainerConstraints validates container-specific constraints
func (v *ContainerValidator) ValidateContainerConstraints(ctx context.Context, container *Container, repo ContainerRepository) error {
	// Validate maximum depth
	depth, err := v.calculateContainerDepth(ctx, container.ID(), repo)
	if err != nil {
		return WrapContainerError(err, "DEPTH_CALCULATION_FAILED", "failed to calculate container depth")
	}

	const maxDepth = 100 // Configurable maximum depth
	if depth > maxDepth {
		return NewContainerError("MAX_DEPTH_EXCEEDED", fmt.Sprintf("container depth %d exceeds maximum allowed depth %d", depth, maxDepth))
	}

	// Validate maximum members
	const maxMembers = 10000 // Configurable maximum members
	memberCount := container.GetMemberCount()
	if memberCount > maxMembers {
		return NewContainerError("MAX_MEMBERS_EXCEEDED", fmt.Sprintf("container has %d members, exceeds maximum allowed %d", memberCount, maxMembers))
	}

	return nil
}

// calculateContainerDepth calculates the depth of a container in the hierarchy
func (v *ContainerValidator) calculateContainerDepth(ctx context.Context, containerID string, repo ContainerRepository) (int, error) {
	depth := 0
	currentID := containerID
	visited := make(map[string]bool)

	for currentID != "" {
		if visited[currentID] {
			return 0, NewContainerError("CIRCULAR_REFERENCE", "circular reference detected while calculating depth")
		}
		visited[currentID] = true

		container, err := repo.GetContainer(ctx, currentID)
		if err != nil {
			if IsContainerNotFound(err) {
				break
			}
			return 0, err
		}

		if container.GetParentID() == "" {
			break // Reached root
		}

		depth++
		currentID = container.GetParentID()
	}

	return depth, nil
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidateContainerBatch validates multiple containers in a batch
func (v *ContainerValidator) ValidateContainerBatch(ctx context.Context, containers []*Container, repo ContainerRepository) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []string{}}

	for i, container := range containers {
		if err := v.ValidateContainer(ctx, container); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("container %d (%s): %v", i, container.ID(), err))
		}

		if err := v.ValidateContainerConstraints(ctx, container, repo); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("container %d (%s) constraints: %v", i, container.ID(), err))
		}
	}

	return result
}
