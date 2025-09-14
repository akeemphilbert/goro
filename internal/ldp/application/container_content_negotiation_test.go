package application

import (
	"context"
	"testing"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
	"github.com/akeemphilbert/goro/internal/ldp/infrastructure"
	"github.com/stretchr/testify/mock"
)

func TestContainerService_GetContainerWithFormat_Turtle(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container with members
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container")

	resource := domain.NewResource(context.Background(), "resource-1", "text/plain", []byte("Hello, World!"))
	resource2 := domain.NewResource(context.Background(), "resource-2", "text/plain", []byte("Hello, World!"))
	container.AddMember(context.Background(), resource)
	container.AddMember(context.Background(), resource2)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// Get container in Turtle format
	result, err := service.GetContainerWithFormat(context.Background(), container.ID(), "text/turtle", "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to get container with Turtle format: %v", err)
	}

	// Verify the result contains Turtle-specific elements
	turtleStr := string(result)
	if !contains(turtleStr, "@prefix") {
		t.Error("Turtle format should contain namespace prefixes")
	}
	if !contains(turtleStr, "BasicContainer") {
		t.Error("Turtle format should contain container type")
	}
	if !contains(turtleStr, "Test Container") {
		t.Error("Turtle format should contain container title")
	}
	if !contains(turtleStr, "ldp:contains") {
		t.Error("Turtle format should contain membership triples")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetContainerWithFormat_JSONLD(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container with members
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container")
	resource := domain.NewResource(context.Background(), "resource-1", "text/plain", []byte("Hello, World!"))
	resource2 := domain.NewResource(context.Background(), "resource-2", "text/plain", []byte("Hello, World!"))
	container.AddMember(context.Background(), resource)
	container.AddMember(context.Background(), resource2)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// Get container in JSON-LD format
	result, err := service.GetContainerWithFormat(context.Background(), container.ID(), "application/ld+json", "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to get container with JSON-LD format: %v", err)
	}

	// Verify the result contains JSON-LD-specific elements
	jsonldStr := string(result)
	if !contains(jsonldStr, "@context") {
		t.Error("JSON-LD format should contain @context")
	}
	if !contains(jsonldStr, "@id") {
		t.Error("JSON-LD format should contain @id")
	}
	if !contains(jsonldStr, "BasicContainer") {
		t.Error("JSON-LD format should contain container type")
	}
	if !contains(jsonldStr, "Test Container") {
		t.Error("JSON-LD format should contain container title")
	}
	if !contains(jsonldStr, "contains") {
		t.Error("JSON-LD format should contain membership information")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetContainerWithFormat_RDFXML(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container with members
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.SetDescription("A test container")

	resource := domain.NewResource(context.Background(), "resource-1", "text/plain", []byte("Hello, World!"))
	container.AddMember(context.Background(), resource)
	container.AddMember(context.Background(), resource)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// Get container in RDF/XML format
	result, err := service.GetContainerWithFormat(context.Background(), container.ID(), "application/rdf+xml", "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to get container with RDF/XML format: %v", err)
	}

	// Verify the result contains RDF/XML-specific elements
	rdfxmlStr := string(result)
	if !contains(rdfxmlStr, "<?xml") {
		t.Error("RDF/XML format should contain XML declaration")
	}
	if !contains(rdfxmlStr, "xmlns:rdf") {
		t.Error("RDF/XML format should contain RDF namespace")
	}
	if !contains(rdfxmlStr, "xmlns:ldp") {
		t.Error("RDF/XML format should contain LDP namespace")
	}
	if !contains(rdfxmlStr, "BasicContainer") {
		t.Error("RDF/XML format should contain container type")
	}
	if !contains(rdfxmlStr, "Test Container") {
		t.Error("RDF/XML format should contain container title")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetContainerWithFormat_UnsupportedFormat(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// Try to get container in unsupported format
	_, err := service.GetContainerWithFormat(context.Background(), container.ID(), "text/html", "http://example.org/")
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_GetContainerWithFormat_EmptyContainer(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create an empty test container
	container := domain.NewContainer(context.Background(), "empty-container", "", domain.BasicContainer)
	container.SetTitle("Empty Container")

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// Get container in Turtle format
	result, err := service.GetContainerWithFormat(context.Background(), container.ID(), "text/turtle", "http://example.org/")
	if err != nil {
		t.Fatalf("Failed to get empty container with Turtle format: %v", err)
	}

	// Verify the result contains container metadata but no membership triples
	turtleStr := string(result)
	if !contains(turtleStr, "Empty Container") {
		t.Error("Empty container should still contain title")
	}
	if !contains(turtleStr, "BasicContainer") {
		t.Error("Empty container should contain container type")
	}
	// Should not contain membership triples for empty container
	if contains(turtleStr, "ldp:contains") {
		t.Error("Empty container should not contain membership triples")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_ListContainerMembersWithFormat_Turtle(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container with multiple members
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)

	resource := domain.NewResource(context.Background(), "resource-1", "text/plain", []byte("Hello, World!"))
	resource2 := domain.NewResource(context.Background(), "resource-2", "text/plain", []byte("Hello, World!"))
	subContainer := domain.NewContainer(context.Background(), "sub-container-1", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.AddMember(context.Background(), resource)
	container.AddMember(context.Background(), resource2)
	container.AddMember(context.Background(), subContainer)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// List members in Turtle format
	result, err := service.ListContainerMembersWithFormat(context.Background(), container.ID(), "text/turtle", "http://example.org/", domain.GetDefaultPagination())
	if err != nil {
		t.Fatalf("Failed to list container members with Turtle format: %v", err)
	}

	// Verify the result contains member listing in Turtle format
	turtleStr := string(result)
	if !contains(turtleStr, "ldp:contains") {
		t.Error("Member listing should contain membership triples")
	}
	if !contains(turtleStr, "resource-1") {
		t.Error("Member listing should contain resource-1")
	}
	if !contains(turtleStr, "resource-2") {
		t.Error("Member listing should contain resource-2")
	}
	if !contains(turtleStr, "sub-container-1") {
		t.Error("Member listing should contain sub-container-1")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestContainerService_ListContainerMembersWithFormat_JSONLD(t *testing.T) {
	// Create a mock container repository
	mockRepo := &MockContainerRepository{}

	// Create a container service with RDF converter
	rdfConverter := infrastructure.NewContainerRDFConverter()
	service := &ContainerService{
		containerRepo: mockRepo,
		rdfConverter:  rdfConverter,
	}

	// Create a test container with multiple members
	container := domain.NewContainer(context.Background(), "test-container", "", domain.BasicContainer)

	resource := domain.NewResource(context.Background(), "resource-1", "text/plain", []byte("Hello, World!"))
	resource2 := domain.NewResource(context.Background(), "resource-2", "text/plain", []byte("Hello, World!"))
	subContainer := domain.NewContainer(context.Background(), "sub-container-1", "", domain.BasicContainer)
	container.SetTitle("Test Container")
	container.AddMember(context.Background(), resource)
	container.AddMember(context.Background(), resource2)
	container.AddMember(context.Background(), subContainer)

	// Set up mock expectations
	mockRepo.On("GetContainer", mock.Anything, container.ID()).Return(container, nil)

	// List members in JSON-LD format
	result, err := service.ListContainerMembersWithFormat(context.Background(), container.ID(), "application/ld+json", "http://example.org/", domain.GetDefaultPagination())
	if err != nil {
		t.Fatalf("Failed to list container members with JSON-LD format: %v", err)
	}

	// Verify the result contains member listing in JSON-LD format
	jsonldStr := string(result)
	if !contains(jsonldStr, "contains") {
		t.Error("Member listing should contain contains property")
	}
	if !contains(jsonldStr, "resource-1") {
		t.Error("Member listing should contain resource-1")
	}
	if !contains(jsonldStr, "resource-2") {
		t.Error("Member listing should contain resource-2")
	}
	if !contains(jsonldStr, "sub-container-1") {
		t.Error("Member listing should contain sub-container-1")
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
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
