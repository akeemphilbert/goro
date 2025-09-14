package domain

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataCorruptionDetector_DetectCorruption_HealthyContainer(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Ensure container has proper timestamps (NewContainer sets createdAt but not updatedAt)
	tm.SetCreatedTimestamp(container)

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.False(t, report.IsCorrupted)
	assert.Empty(t, report.Issues)
	assert.Equal(t, container.ID(), report.ContainerID)
}

func TestMetadataCorruptionDetector_DetectCorruption_InvalidTimestamp(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Corrupt timestamp
	container.SetMetadata("createdAt", "not-a-timestamp")

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.True(t, report.IsCorrupted)
	assert.NotEmpty(t, report.Issues)

	// Find the timestamp corruption issue
	found := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionTypeInvalidTimestamp && issue.Field == "createdAt" {
			found = true
			assert.Equal(t, "high", issue.Severity)
			assert.Contains(t, issue.Description, "invalid type")
			break
		}
	}
	assert.True(t, found, "Should find invalid timestamp corruption")
}

func TestMetadataCorruptionDetector_DetectCorruption_MissingTimestamp(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Remove required timestamp
	metadata := container.GetMetadata()
	delete(metadata, "createdAt")

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.True(t, report.IsCorrupted)
	assert.NotEmpty(t, report.Issues)

	// Find the missing field issue
	found := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionTypeMissingRequiredField && issue.Field == "createdAt" {
			found = true
			assert.Equal(t, "high", issue.Severity)
			break
		}
	}
	assert.True(t, found, "Should find missing createdAt field")
}

func TestMetadataCorruptionDetector_DetectCorruption_InvalidContainerType(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Corrupt container type
	container.ContainerType = ContainerType("InvalidType")

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.True(t, report.IsCorrupted)
	assert.NotEmpty(t, report.Issues)

	// Find the container type corruption issue
	found := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionTypeInvalidContainerType && issue.Field == "containerType" {
			found = true
			assert.Equal(t, "critical", issue.Severity)
			assert.Contains(t, issue.Description, "invalid container type")
			break
		}
	}
	assert.True(t, found, "Should find invalid container type corruption")
}

func TestMetadataCorruptionDetector_DetectCorruption_DuplicateMembers(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Add duplicate members
	container.Members = []string{"member1", "member2", "member1", "member3"}

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.True(t, report.IsCorrupted)
	assert.NotEmpty(t, report.Issues)

	// Find the duplicate member issue
	found := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionTypeInvalidMemberList && issue.Field == "members" {
			found = true
			assert.Equal(t, "medium", issue.Severity)
			assert.Contains(t, issue.Description, "duplicate member")
			break
		}
	}
	assert.True(t, found, "Should find duplicate member corruption")
}

func TestMetadataCorruptionDetector_DetectCorruption_SelfParent(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Set container as its own parent
	container.ParentID = container.ID()

	// Execute
	report := detector.DetectCorruption(container)

	// Assert
	assert.True(t, report.IsCorrupted)
	assert.NotEmpty(t, report.Issues)

	// Find the self-parent issue
	found := false
	for _, issue := range report.Issues {
		if issue.Type == CorruptionTypeInvalidParentID && issue.Field == "parentID" {
			found = true
			assert.Equal(t, "critical", issue.Severity)
			assert.Contains(t, issue.Description, "cannot be its own parent")
			break
		}
	}
	assert.True(t, found, "Should find self-parent corruption")
}

func TestMetadataCorruptionDetector_RepairCorruption_TimestampRepair(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Corrupt timestamp
	container.SetMetadata("createdAt", "not-a-timestamp")

	// Detect corruption
	report := detector.DetectCorruption(container)
	require.True(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.True(t, repaired)

	// Verify repair
	newReport := detector.DetectCorruption(container)
	assert.False(t, newReport.IsCorrupted)

	// Check that timestamps are now valid
	createdAt, exists := tm.GetCreatedTimestamp(container)
	assert.True(t, exists)
	assert.False(t, createdAt.IsZero())
}

func TestMetadataCorruptionDetector_RepairCorruption_ContainerTypeRepair(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Corrupt container type
	container.ContainerType = ContainerType("InvalidType")

	// Detect corruption
	report := detector.DetectCorruption(container)
	require.True(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.True(t, repaired)

	// Verify repair
	assert.Equal(t, BasicContainer, container.ContainerType)
	assert.Equal(t, BasicContainer.String(), container.GetMetadata()["containerType"])

	// Check that corruption is fixed
	newReport := detector.DetectCorruption(container)
	assert.False(t, newReport.IsCorrupted)
}

func TestMetadataCorruptionDetector_RepairCorruption_MemberListRepair(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Add duplicate and empty members
	container.Members = []string{"member1", "", "member2", "member1", "member3"}

	// Detect corruption
	report := detector.DetectCorruption(container)
	require.True(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.True(t, repaired)

	// Verify repair - should have unique, non-empty members
	expectedMembers := []string{"member1", "member2", "member3"}
	assert.Equal(t, expectedMembers, container.Members)

	// Check that corruption is fixed
	newReport := detector.DetectCorruption(container)
	assert.False(t, newReport.IsCorrupted)
}

func TestMetadataCorruptionDetector_RepairCorruption_SelfParentRepair(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Set container as its own parent
	container.ParentID = container.ID()

	// Detect corruption
	report := detector.DetectCorruption(container)
	require.True(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.True(t, repaired)

	// Verify repair - parent ID should be cleared
	assert.Empty(t, container.ParentID)
	assert.Equal(t, "", container.GetMetadata()["parentID"])

	// Check that corruption is fixed
	newReport := detector.DetectCorruption(container)
	assert.False(t, newReport.IsCorrupted)
}

func TestMetadataCorruptionDetector_RepairCorruption_NoCorruption(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Ensure container has proper timestamps
	tm.SetCreatedTimestamp(container)

	// Detect corruption on healthy container
	report := detector.DetectCorruption(container)
	require.False(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.False(t, repaired) // Nothing to repair
}

func TestMetadataCorruptionDetector_RepairCorruption_EventEmission(t *testing.T) {
	// Setup
	tm := NewTimestampManager()
	detector := NewMetadataCorruptionDetector(tm)
	ctx := context.Background()
	container := NewContainer(ctx, "test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Corrupt container
	container.ContainerType = ContainerType("InvalidType")

	// Detect corruption
	report := detector.DetectCorruption(container)
	require.True(t, report.IsCorrupted)

	// Execute repair
	repaired, err := detector.RepairCorruption(container, report)

	// Assert
	assert.NoError(t, err)
	assert.True(t, repaired)

	// Check that repair event was emitted
	events := container.UncommittedEvents()
	require.Len(t, events, 1)

	// Cast to EntityEvent to access fields
	entityEvent, ok := events[0].(*pericarpdomain.EntityEvent)
	require.True(t, ok, "Expected EntityEvent")
	assert.Equal(t, EventTypeContainerUpdated, entityEvent.Type)

	// Check event payload
	var payload map[string]interface{}
	err = json.Unmarshal(entityEvent.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, true, payload["repaired"])
	assert.Contains(t, payload, "issues")
	assert.Contains(t, payload, "updatedAt")
}

func TestCorruptionReport_JSON(t *testing.T) {
	// Setup
	report := &CorruptionReport{
		ContainerID: "test-container",
		Issues: []CorruptionIssue{
			{
				Type:        CorruptionTypeInvalidTimestamp,
				Field:       "createdAt",
				Description: "Invalid timestamp type",
				Severity:    "high",
			},
		},
		IsCorrupted: true,
		Timestamp:   time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC),
	}

	// Execute
	jsonData, err := json.Marshal(report)
	require.NoError(t, err)

	var unmarshaled CorruptionReport
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, report.ContainerID, unmarshaled.ContainerID)
	assert.Equal(t, report.IsCorrupted, unmarshaled.IsCorrupted)
	assert.Len(t, unmarshaled.Issues, 1)
	assert.Equal(t, report.Issues[0].Type, unmarshaled.Issues[0].Type)
	assert.Equal(t, report.Issues[0].Field, unmarshaled.Issues[0].Field)
	assert.Equal(t, report.Issues[0].Description, unmarshaled.Issues[0].Description)
	assert.Equal(t, report.Issues[0].Severity, unmarshaled.Issues[0].Severity)
}
