package domain

import (
	"testing"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestampManager_SetCreatedTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Execute
	tm.SetCreatedTimestamp(container)

	// Assert
	metadata := container.GetMetadata()
	assert.Equal(t, fixedTime, metadata["createdAt"])
	assert.Equal(t, fixedTime, metadata["updatedAt"])
}

func TestTimestampManager_UpdateTimestamp(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)

	tm := NewTimestampManagerWithProvider(func() time.Time { return createdTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Set initial timestamps
	tm.SetCreatedTimestamp(container)

	// Change time provider to simulate time passing
	tm.timeProvider = func() time.Time { return updatedTime }

	// Execute
	tm.UpdateTimestamp(container)

	// Assert
	metadata := container.GetMetadata()
	assert.Equal(t, createdTime, metadata["createdAt"]) // Should remain unchanged
	assert.Equal(t, updatedTime, metadata["updatedAt"]) // Should be updated
}

func TestTimestampManager_GetCreatedTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	tm.SetCreatedTimestamp(container)

	// Execute
	createdAt, exists := tm.GetCreatedTimestamp(container)

	// Assert
	assert.True(t, exists)
	assert.Equal(t, fixedTime, createdAt)
}

func TestTimestampManager_GetCreatedTimestamp_Missing(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Remove the createdAt metadata that was set during construction
	metadata := container.GetMetadata()
	delete(metadata, "createdAt")

	// Execute
	createdAt, exists := tm.GetCreatedTimestamp(container)

	// Assert
	assert.False(t, exists)
	assert.True(t, createdAt.IsZero())
}

func TestTimestampManager_GetUpdatedTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	tm.UpdateTimestamp(container)

	// Execute
	updatedAt, exists := tm.GetUpdatedTimestamp(container)

	// Assert
	assert.True(t, exists)
	assert.Equal(t, fixedTime, updatedAt)
}

func TestTimestampManager_ValidateTimestamps_Valid(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)

	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	container.SetMetadata("createdAt", createdTime)
	container.SetMetadata("updatedAt", updatedTime)

	// Execute
	err := tm.ValidateTimestamps(container)

	// Assert
	assert.NoError(t, err)
}

func TestTimestampManager_ValidateTimestamps_MissingCreated(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Remove createdAt but keep updatedAt
	metadata := container.GetMetadata()
	delete(metadata, "createdAt")
	container.SetMetadata("updatedAt", time.Now())

	// Execute
	err := tm.ValidateTimestamps(container)

	// Assert
	assert.Error(t, err)
	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, "MISSING_CREATED_TIMESTAMP", domainErr.Code)
}

func TestTimestampManager_ValidateTimestamps_MissingUpdated(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Remove updatedAt but keep createdAt
	metadata := container.GetMetadata()
	delete(metadata, "updatedAt")
	container.SetMetadata("createdAt", time.Now())

	// Execute
	err := tm.ValidateTimestamps(container)

	// Assert
	assert.Error(t, err)
	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, "MISSING_UPDATED_TIMESTAMP", domainErr.Code)
}

func TestTimestampManager_ValidateTimestamps_InvalidOrder(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC) // Before created

	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	container.SetMetadata("createdAt", createdTime)
	container.SetMetadata("updatedAt", updatedTime)

	// Execute
	err := tm.ValidateTimestamps(container)

	// Assert
	assert.Error(t, err)
	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, "INVALID_TIMESTAMP_ORDER", domainErr.Code)
}

func TestTimestampManager_RepairTimestamps_MissingCreated(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Remove createdAt
	metadata := container.GetMetadata()
	delete(metadata, "createdAt")

	// Execute
	repaired := tm.RepairTimestamps(container)

	// Assert
	assert.True(t, repaired)
	createdAt, exists := tm.GetCreatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, fixedTime, createdAt)
}

func TestTimestampManager_RepairTimestamps_MissingUpdated(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	container.SetMetadata("createdAt", createdTime)
	// Remove updatedAt
	metadata := container.GetMetadata()
	delete(metadata, "updatedAt")

	// Execute
	repaired := tm.RepairTimestamps(container)

	// Assert
	assert.True(t, repaired)
	updatedAt, exists := tm.GetUpdatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, createdTime, updatedAt) // Should be set to creation time
}

func TestTimestampManager_RepairTimestamps_InvalidOrder(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC) // Before created

	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	container.SetMetadata("createdAt", createdTime)
	container.SetMetadata("updatedAt", updatedTime)

	// Execute
	repaired := tm.RepairTimestamps(container)

	// Assert
	assert.True(t, repaired)
	updatedAt, exists := tm.GetUpdatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, createdTime, updatedAt) // Should be fixed to creation time
}

func TestTimestampManager_RepairTimestamps_NoRepairNeeded(t *testing.T) {
	// Setup
	createdTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)

	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	container.SetMetadata("createdAt", createdTime)
	container.SetMetadata("updatedAt", updatedTime)

	// Execute
	repaired := tm.RepairTimestamps(container)

	// Assert
	assert.False(t, repaired) // No repair needed
}

func TestContainer_SetTitleWithTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	title := "Test Title"

	// Execute
	container.SetTitleWithTimestamp(title, tm)

	// Assert
	assert.Equal(t, title, container.GetTitle())
	updatedAt, exists := tm.GetUpdatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, fixedTime, updatedAt)

	// Check event emission
	events := container.UncommittedEvents()
	require.Len(t, events, 1)

	// Cast to EntityEvent to access Type field
	if entityEvent, ok := events[0].(*pericarpdomain.EntityEvent); ok {
		assert.Equal(t, EventTypeContainerUpdated, entityEvent.Type)
	} else {
		t.Errorf("Expected EntityEvent, got %T", events[0])
	}
}

func TestContainer_AddMemberWithTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	memberID := "test-member"

	// Execute
	err := container.AddMemberWithTimestamp(memberID, tm)

	// Assert
	assert.NoError(t, err)
	assert.True(t, container.HasMember(memberID))
	updatedAt, exists := tm.GetUpdatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, fixedTime, updatedAt)

	// Check event emission
	events := container.UncommittedEvents()
	require.Len(t, events, 1)

	// Cast to EntityEvent to access Type field
	if entityEvent, ok := events[0].(*pericarpdomain.EntityEvent); ok {
		assert.Equal(t, EventTypeMemberAdded, entityEvent.Type)
	} else {
		t.Errorf("Expected EntityEvent, got %T", events[0])
	}
}

func TestContainer_AddMemberWithTimestamp_DuplicateMember(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	memberID := "test-member"

	// Add member first time
	err := container.AddMemberWithTimestamp(memberID, tm)
	require.NoError(t, err)
	container.MarkEventsAsCommitted() // Clear events

	// Execute - try to add same member again
	err = container.AddMemberWithTimestamp(memberID, tm)

	// Assert
	assert.Error(t, err)
	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, "MEMBER_ALREADY_EXISTS", domainErr.Code)

	// No new events should be emitted
	events := container.UncommittedEvents()
	assert.Len(t, events, 0)
}

func TestContainer_RemoveMemberWithTimestamp(t *testing.T) {
	// Setup
	fixedTime := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	tm := NewTimestampManagerWithProvider(func() time.Time { return fixedTime })
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	memberID := "test-member"

	// Add member first
	err := container.AddMemberWithTimestamp(memberID, tm)
	require.NoError(t, err)
	container.MarkEventsAsCommitted() // Clear add events

	// Execute
	err = container.RemoveMemberWithTimestamp(memberID, tm)

	// Assert
	assert.NoError(t, err)
	assert.False(t, container.HasMember(memberID))
	updatedAt, exists := tm.GetUpdatedTimestamp(container)
	assert.True(t, exists)
	assert.Equal(t, fixedTime, updatedAt)

	// Check event emission
	events := container.UncommittedEvents()
	require.Len(t, events, 1)

	// Cast to EntityEvent to access Type field
	if entityEvent, ok := events[0].(*pericarpdomain.EntityEvent); ok {
		assert.Equal(t, EventTypeMemberRemoved, entityEvent.Type)
	} else {
		t.Errorf("Expected EntityEvent, got %T", events[0])
	}
}

func TestContainer_RemoveMemberWithTimestamp_MemberNotFound(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	memberID := "non-existent-member"

	// Execute
	err := container.RemoveMemberWithTimestamp(memberID, tm)

	// Assert
	assert.Error(t, err)
	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, "MEMBER_NOT_FOUND", domainErr.Code)

	// No events should be emitted
	events := container.UncommittedEvents()
	assert.Len(t, events, 0)
}
