package domain

import (
	"time"
)

// TimestampManager handles automatic timestamp management for containers
type TimestampManager struct {
	timeProvider func() time.Time
}

// NewTimestampManager creates a new timestamp manager
func NewTimestampManager() *TimestampManager {
	return &TimestampManager{
		timeProvider: time.Now,
	}
}

// NewTimestampManagerWithProvider creates a timestamp manager with a custom time provider
func NewTimestampManagerWithProvider(timeProvider func() time.Time) *TimestampManager {
	return &TimestampManager{
		timeProvider: timeProvider,
	}
}

// SetCreatedTimestamp sets the creation timestamp on a container
func (tm *TimestampManager) SetCreatedTimestamp(container *Container) {
	now := tm.timeProvider()
	container.SetMetadata("createdAt", now)
	container.SetMetadata("updatedAt", now)
}

// UpdateTimestamp updates the modification timestamp on a container
func (tm *TimestampManager) UpdateTimestamp(container *Container) {
	now := tm.timeProvider()
	container.SetMetadata("updatedAt", now)
}

// GetCreatedTimestamp retrieves the creation timestamp from a container
func (tm *TimestampManager) GetCreatedTimestamp(container *Container) (time.Time, bool) {
	metadata := container.GetMetadata()
	if createdAt, exists := metadata["createdAt"]; exists {
		if timestamp, ok := createdAt.(time.Time); ok {
			return timestamp, true
		}
	}
	return time.Time{}, false
}

// GetUpdatedTimestamp retrieves the modification timestamp from a container
func (tm *TimestampManager) GetUpdatedTimestamp(container *Container) (time.Time, bool) {
	metadata := container.GetMetadata()
	if updatedAt, exists := metadata["updatedAt"]; exists {
		if timestamp, ok := updatedAt.(time.Time); ok {
			return timestamp, true
		}
	}
	return time.Time{}, false
}

// ValidateTimestamps validates that timestamps are consistent
func (tm *TimestampManager) ValidateTimestamps(container *Container) error {
	createdAt, hasCreated := tm.GetCreatedTimestamp(container)
	updatedAt, hasUpdated := tm.GetUpdatedTimestamp(container)

	if !hasCreated {
		return NewDomainError("MISSING_CREATED_TIMESTAMP", "container missing creation timestamp")
	}

	if !hasUpdated {
		return NewDomainError("MISSING_UPDATED_TIMESTAMP", "container missing update timestamp")
	}

	if updatedAt.Before(createdAt) {
		return NewDomainError("INVALID_TIMESTAMP_ORDER", "update timestamp cannot be before creation timestamp")
	}

	return nil
}

// RepairTimestamps repairs missing or invalid timestamps
func (tm *TimestampManager) RepairTimestamps(container *Container) bool {
	now := tm.timeProvider()
	repaired := false

	createdAt, hasCreated := tm.GetCreatedTimestamp(container)
	updatedAt, hasUpdated := tm.GetUpdatedTimestamp(container)

	// If no creation timestamp, set it to now
	if !hasCreated {
		container.SetMetadata("createdAt", now)
		repaired = true
		createdAt = now
	}

	// If no update timestamp, set it to creation timestamp or now
	if !hasUpdated {
		if hasCreated {
			container.SetMetadata("updatedAt", createdAt)
		} else {
			container.SetMetadata("updatedAt", now)
		}
		repaired = true
	} else if updatedAt.Before(createdAt) {
		// If update timestamp is before creation, fix it
		container.SetMetadata("updatedAt", createdAt)
		repaired = true
	}

	return repaired
}

// Enhanced container methods with automatic timestamp management

// SetTitleWithTimestamp sets the container title and updates timestamp
func (c *Container) SetTitleWithTimestamp(title string, tm *TimestampManager) {
	c.SetMetadata("title", title)
	tm.UpdateTimestamp(c)

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"title":     title,
		"updatedAt": tm.timeProvider(),
	})
	c.AddEvent(event)
}

// SetDescriptionWithTimestamp sets the container description and updates timestamp
func (c *Container) SetDescriptionWithTimestamp(description string, tm *TimestampManager) {
	c.SetMetadata("description", description)
	tm.UpdateTimestamp(c)

	// Emit update event
	event := NewContainerUpdatedEvent(c.ID(), map[string]interface{}{
		"description": description,
		"updatedAt":   tm.timeProvider(),
	})
	c.AddEvent(event)
}

// AddMemberWithTimestamp adds a member and updates timestamp
func (c *Container) AddMemberWithTimestamp(memberID string, tm *TimestampManager) error {
	// Check if member already exists
	for _, member := range c.Members {
		if member == memberID {
			return NewDomainError("MEMBER_ALREADY_EXISTS", "member already exists: "+memberID)
		}
	}

	c.Members = append(c.Members, memberID)
	tm.UpdateTimestamp(c)

	// Emit member added event
	event := NewMemberAddedEvent(c.ID(), map[string]interface{}{
		"memberID":   memberID,
		"memberType": "Resource",
		"addedAt":    tm.timeProvider(),
	})
	c.AddEvent(event)

	return nil
}

// RemoveMemberWithTimestamp removes a member and updates timestamp
func (c *Container) RemoveMemberWithTimestamp(memberID string, tm *TimestampManager) error {
	for i, member := range c.Members {
		if member == memberID {
			// Remove member from slice
			c.Members = append(c.Members[:i], c.Members[i+1:]...)
			tm.UpdateTimestamp(c)

			// Emit member removed event
			event := NewMemberRemovedEvent(c.ID(), map[string]interface{}{
				"memberID":   memberID,
				"memberType": "Resource",
				"removedAt":  tm.timeProvider(),
			})
			c.AddEvent(event)

			return nil
		}
	}

	return NewDomainError("MEMBER_NOT_FOUND", "member not found: "+memberID)
}
