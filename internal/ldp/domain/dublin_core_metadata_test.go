package domain

import (
	"encoding/json"
	"testing"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_SetDublinCoreMetadata(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	testDate := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	dc := DublinCoreMetadata{
		Title:       "Test Container",
		Description: "A test container for Dublin Core metadata",
		Creator:     "Test Creator",
		Subject:     "Testing",
		Publisher:   "Test Publisher",
		Contributor: "Test Contributor",
		Date:        testDate,
		Type:        "Container",
		Format:      "application/ld+json",
		Identifier:  "test-container-id",
		Source:      "Test Source",
		Language:    "en",
		Relation:    "Test Relation",
		Coverage:    "Test Coverage",
		Rights:      "Test Rights",
	}

	// Execute
	container.SetDublinCoreMetadata(dc)

	// Assert
	metadata := container.GetMetadata()
	assert.Equal(t, dc.Title, metadata["dc:title"])
	assert.Equal(t, dc.Description, metadata["dc:description"])
	assert.Equal(t, dc.Creator, metadata["dc:creator"])
	assert.Equal(t, dc.Subject, metadata["dc:subject"])
	assert.Equal(t, dc.Publisher, metadata["dc:publisher"])
	assert.Equal(t, dc.Contributor, metadata["dc:contributor"])
	assert.Equal(t, dc.Date, metadata["dc:date"])
	assert.Equal(t, dc.Type, metadata["dc:type"])
	assert.Equal(t, dc.Format, metadata["dc:format"])
	assert.Equal(t, dc.Identifier, metadata["dc:identifier"])
	assert.Equal(t, dc.Source, metadata["dc:source"])
	assert.Equal(t, dc.Language, metadata["dc:language"])
	assert.Equal(t, dc.Relation, metadata["dc:relation"])
	assert.Equal(t, dc.Coverage, metadata["dc:coverage"])
	assert.Equal(t, dc.Rights, metadata["dc:rights"])

	// Check that updatedAt was set
	assert.Contains(t, metadata, "updatedAt")

	// Check that event was emitted
	events := container.UncommittedEvents()
	assert.Len(t, events, 1)

	// Cast to EntityEvent to access Type field
	if entityEvent, ok := events[0].(*pericarpdomain.EntityEvent); ok {
		assert.Equal(t, EventTypeContainerUpdated, entityEvent.Type)
	} else {
		t.Errorf("Expected EntityEvent, got %T", events[0])
	}
}

func TestContainer_SetDublinCoreMetadata_PartialData(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	dc := DublinCoreMetadata{
		Title:       "Test Container",
		Description: "A test container",
		// Only set title and description, leave others empty
	}

	// Execute
	container.SetDublinCoreMetadata(dc)

	// Assert
	metadata := container.GetMetadata()
	assert.Equal(t, dc.Title, metadata["dc:title"])
	assert.Equal(t, dc.Description, metadata["dc:description"])

	// Check that empty fields are not set
	assert.NotContains(t, metadata, "dc:creator")
	assert.NotContains(t, metadata, "dc:subject")
	assert.NotContains(t, metadata, "dc:publisher")
	assert.NotContains(t, metadata, "dc:date")
}

func TestContainer_GetDublinCoreMetadata(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	testDate := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	originalDC := DublinCoreMetadata{
		Title:       "Test Container",
		Description: "A test container for Dublin Core metadata",
		Creator:     "Test Creator",
		Subject:     "Testing",
		Publisher:   "Test Publisher",
		Contributor: "Test Contributor",
		Date:        testDate,
		Type:        "Container",
		Format:      "application/ld+json",
		Identifier:  "test-container-id",
		Source:      "Test Source",
		Language:    "en",
		Relation:    "Test Relation",
		Coverage:    "Test Coverage",
		Rights:      "Test Rights",
	}

	// Set the metadata first
	container.SetDublinCoreMetadata(originalDC)
	container.MarkEventsAsCommitted() // Clear update events

	// Execute
	retrievedDC := container.GetDublinCoreMetadata()

	// Assert
	assert.Equal(t, originalDC.Title, retrievedDC.Title)
	assert.Equal(t, originalDC.Description, retrievedDC.Description)
	assert.Equal(t, originalDC.Creator, retrievedDC.Creator)
	assert.Equal(t, originalDC.Subject, retrievedDC.Subject)
	assert.Equal(t, originalDC.Publisher, retrievedDC.Publisher)
	assert.Equal(t, originalDC.Contributor, retrievedDC.Contributor)
	assert.Equal(t, originalDC.Date, retrievedDC.Date)
	assert.Equal(t, originalDC.Type, retrievedDC.Type)
	assert.Equal(t, originalDC.Format, retrievedDC.Format)
	assert.Equal(t, originalDC.Identifier, retrievedDC.Identifier)
	assert.Equal(t, originalDC.Source, retrievedDC.Source)
	assert.Equal(t, originalDC.Language, retrievedDC.Language)
	assert.Equal(t, originalDC.Relation, retrievedDC.Relation)
	assert.Equal(t, originalDC.Coverage, retrievedDC.Coverage)
	assert.Equal(t, originalDC.Rights, retrievedDC.Rights)
}

func TestContainer_GetDublinCoreMetadata_EmptyContainer(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Execute
	dc := container.GetDublinCoreMetadata()

	// Assert - all fields should be empty/zero values
	assert.Empty(t, dc.Title)
	assert.Empty(t, dc.Description)
	assert.Empty(t, dc.Creator)
	assert.Empty(t, dc.Subject)
	assert.Empty(t, dc.Publisher)
	assert.Empty(t, dc.Contributor)
	assert.True(t, dc.Date.IsZero())
	assert.Empty(t, dc.Type)
	assert.Empty(t, dc.Format)
	assert.Empty(t, dc.Identifier)
	assert.Empty(t, dc.Source)
	assert.Empty(t, dc.Language)
	assert.Empty(t, dc.Relation)
	assert.Empty(t, dc.Coverage)
	assert.Empty(t, dc.Rights)
}

func TestContainer_GetDublinCoreMetadata_PartialData(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Manually set only some Dublin Core fields
	container.SetMetadata("dc:title", "Test Title")
	container.SetMetadata("dc:creator", "Test Creator")

	// Execute
	dc := container.GetDublinCoreMetadata()

	// Assert
	assert.Equal(t, "Test Title", dc.Title)
	assert.Equal(t, "Test Creator", dc.Creator)

	// Other fields should be empty
	assert.Empty(t, dc.Description)
	assert.Empty(t, dc.Subject)
	assert.Empty(t, dc.Publisher)
}

func TestContainer_DublinCoreMetadata_TypeSafety(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Set invalid type for a Dublin Core field
	container.SetMetadata("dc:title", 123)         // Should be string
	container.SetMetadata("dc:date", "not-a-date") // Should be time.Time

	// Execute
	dc := container.GetDublinCoreMetadata()

	// Assert - invalid types should result in zero values
	assert.Empty(t, dc.Title)        // 123 is not a string, so should be empty
	assert.True(t, dc.Date.IsZero()) // "not-a-date" is not time.Time, so should be zero
}

func TestContainer_DublinCoreMetadata_Integration(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	// Test setting and getting Dublin Core metadata multiple times
	dc1 := DublinCoreMetadata{
		Title:       "First Title",
		Description: "First Description",
		Creator:     "First Creator",
	}

	dc2 := DublinCoreMetadata{
		Title:       "Second Title",
		Description: "Second Description",
		Subject:     "New Subject",
	}

	// Execute - Set first metadata
	container.SetDublinCoreMetadata(dc1)
	retrieved1 := container.GetDublinCoreMetadata()

	// Assert first set
	assert.Equal(t, dc1.Title, retrieved1.Title)
	assert.Equal(t, dc1.Description, retrieved1.Description)
	assert.Equal(t, dc1.Creator, retrieved1.Creator)
	assert.Empty(t, retrieved1.Subject)

	// Execute - Set second metadata (should overwrite)
	container.MarkEventsAsCommitted() // Clear events from first set
	container.SetDublinCoreMetadata(dc2)
	retrieved2 := container.GetDublinCoreMetadata()

	// Assert second set
	assert.Equal(t, dc2.Title, retrieved2.Title)
	assert.Equal(t, dc2.Description, retrieved2.Description)
	assert.Equal(t, dc2.Subject, retrieved2.Subject)
	// Creator should still be there from first set (not overwritten because dc2 doesn't set it)
	assert.Equal(t, dc1.Creator, retrieved2.Creator)
}

func TestDublinCoreMetadata_EventEmission(t *testing.T) {
	// Setup
	container := NewContainer("test-container", "", BasicContainer)
	container.MarkEventsAsCommitted() // Clear creation events

	dc := DublinCoreMetadata{
		Title:       "Test Container",
		Description: "Test Description",
	}

	// Execute
	container.SetDublinCoreMetadata(dc)

	// Assert
	events := container.UncommittedEvents()
	require.Len(t, events, 1)

	// Cast to EntityEvent to access fields
	entityEvent, ok := events[0].(*pericarpdomain.EntityEvent)
	require.True(t, ok, "Expected EntityEvent")

	assert.Equal(t, EventTypeContainerUpdated, entityEvent.Type)
	assert.Equal(t, "container", entityEvent.EntityType)
	assert.Equal(t, container.ID(), entityEvent.AggregateID())

	// Check event payload contains Dublin Core data
	var payload map[string]interface{}
	err := json.Unmarshal(entityEvent.Payload(), &payload)
	require.NoError(t, err)
	assert.Contains(t, payload, "dublinCore")
	assert.Contains(t, payload, "updatedAt")
}
