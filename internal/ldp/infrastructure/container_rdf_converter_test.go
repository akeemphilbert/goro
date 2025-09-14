package infrastructure

import (
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

func TestContainerRDFConverter_ConvertToTurtle(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a test container with members
	container := domain.NewContainer("container-1", "parent-1", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container for unit testing")
	container.AddMember("resource-1")
	container.AddMember("resource-2")

	// Convert to Turtle
	result, err := converter.ConvertToTurtle(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert container to Turtle: %v", err)
	}

	// Verify Turtle output contains expected elements
	turtleStr := string(result)

	// Check for LDP namespace
	if !contains(turtleStr, "@prefix ldp:") {
		t.Error("Turtle output should contain LDP namespace declaration")
	}

	// Check for container type
	if !contains(turtleStr, "BasicContainer") {
		t.Errorf("Turtle output should contain BasicContainer type. Got: %s", turtleStr)
	}

	// Check for title
	if !contains(turtleStr, "Test Container") {
		t.Error("Turtle output should contain container title")
	}

	// Check for membership triples
	if !contains(turtleStr, "ldp:contains") {
		t.Error("Turtle output should contain membership triples")
	}

	// Check for member resources
	if !contains(turtleStr, "resource-1") {
		t.Error("Turtle output should contain member resource-1")
	}
	if !contains(turtleStr, "resource-2") {
		t.Error("Turtle output should contain member resource-2")
	}
}

func TestContainerRDFConverter_ConvertToJSONLD(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a test container with members
	container := domain.NewContainer("container-1", "parent-1", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container for unit testing")
	container.AddMember("resource-1")
	container.AddMember("resource-2")

	// Convert to JSON-LD
	result, err := converter.ConvertToJSONLD(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert container to JSON-LD: %v", err)
	}

	// Verify JSON-LD output contains expected elements
	jsonldStr := string(result)

	// Check for JSON-LD context
	if !contains(jsonldStr, "@context") {
		t.Error("JSON-LD output should contain @context")
	}

	// Check for container ID
	if !contains(jsonldStr, "@id") {
		t.Error("JSON-LD output should contain @id")
	}

	// Check for container type
	if !contains(jsonldStr, "BasicContainer") {
		t.Error("JSON-LD output should contain BasicContainer type")
	}

	// Check for title
	if !contains(jsonldStr, "Test Container") {
		t.Error("JSON-LD output should contain container title")
	}

	// Check for contains property
	if !contains(jsonldStr, "contains") {
		t.Error("JSON-LD output should contain contains property")
	}
}

func TestContainerRDFConverter_ConvertToRDFXML(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a test container with members
	container := domain.NewContainer("container-1", "parent-1", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container for unit testing")
	container.AddMember("resource-1")
	container.AddMember("resource-2")

	// Convert to RDF/XML
	result, err := converter.ConvertToRDFXML(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert container to RDF/XML: %v", err)
	}

	// Verify RDF/XML output contains expected elements
	rdfxmlStr := string(result)

	// Check for XML declaration
	if !contains(rdfxmlStr, "<?xml") {
		t.Error("RDF/XML output should contain XML declaration")
	}

	// Check for RDF namespace
	if !contains(rdfxmlStr, "xmlns:rdf") {
		t.Error("RDF/XML output should contain RDF namespace")
	}

	// Check for LDP namespace
	if !contains(rdfxmlStr, "xmlns:ldp") {
		t.Error("RDF/XML output should contain LDP namespace")
	}

	// Check for container type
	if !contains(rdfxmlStr, "BasicContainer") {
		t.Error("RDF/XML output should contain BasicContainer type")
	}

	// Check for title
	if !contains(rdfxmlStr, "Test Container") {
		t.Error("RDF/XML output should contain container title")
	}
}

func TestContainerRDFConverter_GenerateMembershipTriples(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a test container with members
	container := domain.NewContainer("container-1", "", domain.BasicContainer)
	container.AddMember("resource-1")
	container.AddMember("resource-2")
	container.AddMember("sub-container-1")

	// Generate membership triples
	triples := converter.GenerateMembershipTriples(container, "http://example.org/")

	// Verify we have the expected number of triples
	expectedTriples := 3 // One for each member
	if len(triples) != expectedTriples {
		t.Errorf("Expected %d membership triples, got %d", expectedTriples, len(triples))
	}

	// Verify each triple has the correct structure
	for i, triple := range triples {
		if triple.Subject != "http://example.org/container-1" {
			t.Errorf("Triple %d: expected subject 'http://example.org/container-1', got '%s'", i, triple.Subject)
		}

		if triple.Predicate != "http://www.w3.org/ns/ldp#contains" {
			t.Errorf("Triple %d: expected predicate 'http://www.w3.org/ns/ldp#contains', got '%s'", i, triple.Predicate)
		}

		if triple.ObjectType != "uri" {
			t.Errorf("Triple %d: expected object type 'uri', got '%s'", i, triple.ObjectType)
		}
	}

	// Verify specific member URIs
	memberURIs := make(map[string]bool)
	for _, triple := range triples {
		memberURIs[triple.Object] = true
	}

	expectedMembers := []string{
		"http://example.org/resource-1",
		"http://example.org/resource-2",
		"http://example.org/sub-container-1",
	}

	for _, expectedMember := range expectedMembers {
		if !memberURIs[expectedMember] {
			t.Errorf("Expected member URI '%s' not found in membership triples", expectedMember)
		}
	}
}

func TestContainerRDFConverter_EmptyContainer(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create an empty container
	container := domain.NewContainer("empty-container", "", domain.BasicContainer)
	container.SetTitle("Empty Container")

	// Test Turtle conversion
	turtle, err := converter.ConvertToTurtle(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert empty container to Turtle: %v", err)
	}

	turtleStr := string(turtle)
	if !contains(turtleStr, "Empty Container") {
		t.Error("Turtle output should contain container title even for empty container")
	}

	// Should not contain any membership triples
	if contains(turtleStr, "ldp:contains") {
		t.Error("Empty container should not contain membership triples")
	}

	// Test JSON-LD conversion
	jsonld, err := converter.ConvertToJSONLD(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert empty container to JSON-LD: %v", err)
	}

	jsonldStr := string(jsonld)
	if !contains(jsonldStr, "Empty Container") {
		t.Error("JSON-LD output should contain container title even for empty container")
	}
}

func TestContainerRDFConverter_DirectContainer(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a DirectContainer
	container := domain.NewContainer("direct-container", "", domain.DirectContainer)
	container.SetTitle("Direct Container")
	container.AddMember("resource-1")

	// Convert to Turtle
	result, err := converter.ConvertToTurtle(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert DirectContainer to Turtle: %v", err)
	}

	// Verify DirectContainer type is present
	turtleStr := string(result)
	if !contains(turtleStr, "DirectContainer") {
		t.Error("Turtle output should contain DirectContainer type")
	}
}

func TestContainerRDFConverter_WithTimestamps(t *testing.T) {
	converter := NewContainerRDFConverter()

	// Create a container and set specific timestamps
	container := domain.NewContainer("timestamped-container", "", domain.BasicContainer)
	container.SetTitle("Timestamped Container")

	// Set specific timestamps in metadata
	createdAt := time.Date(2025, 9, 13, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 9, 13, 11, 0, 0, 0, time.UTC)
	container.SetMetadata("createdAt", createdAt)
	container.SetMetadata("updatedAt", updatedAt)

	// Convert to Turtle
	result, err := converter.ConvertToTurtle(container, "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to convert container with timestamps to Turtle: %v", err)
	}

	// Verify timestamps are included
	turtleStr := string(result)
	if !contains(turtleStr, "dcterms:created") {
		t.Error("Turtle output should contain creation timestamp")
	}
	if !contains(turtleStr, "dcterms:modified") {
		t.Error("Turtle output should contain modification timestamp")
	}
	if !contains(turtleStr, "2025-09-13T10:00:00Z") {
		t.Error("Turtle output should contain formatted creation timestamp")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
