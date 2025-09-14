package infrastructure

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/ldp/domain"
)

// ContainerRDFConverter handles conversion of container entities to various RDF formats
type ContainerRDFConverter struct {
	rdfConverter *RDFConverter
}

// NewContainerRDFConverter creates a new container RDF converter
func NewContainerRDFConverter() *ContainerRDFConverter {
	return &ContainerRDFConverter{
		rdfConverter: NewRDFConverter(),
	}
}

// ContainerTriple represents an RDF triple for container serialization
type ContainerTriple struct {
	Subject    string `json:"subject"`
	Predicate  string `json:"predicate"`
	Object     string `json:"object"`
	ObjectType string `json:"objectType"` // "uri", "literal", "blank"
	DataType   string `json:"dataType,omitempty"`
}

// ConvertToTurtle converts a container to Turtle format
func (c *ContainerRDFConverter) ConvertToTurtle(container *domain.Container, baseURI string) ([]byte, error) {
	if container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	// Generate all triples for the container
	triples := c.generateAllTriples(container, baseURI)

	// Build Turtle representation
	var turtle strings.Builder

	// Add namespace prefixes
	turtle.WriteString("@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .\n")
	turtle.WriteString("@prefix ldp: <http://www.w3.org/ns/ldp#> .\n")
	turtle.WriteString("@prefix dcterms: <http://purl.org/dc/terms/> .\n")
	turtle.WriteString("@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .\n")
	turtle.WriteString("\n")

	// Group triples by subject
	subjectTriples := make(map[string][]ContainerTriple)
	for _, triple := range triples {
		subjectTriples[triple.Subject] = append(subjectTriples[triple.Subject], triple)
	}

	// Serialize each subject
	for subject, subjectTripleList := range subjectTriples {
		turtle.WriteString(fmt.Sprintf("<%s>", subject))

		for i, triple := range subjectTripleList {
			if i == 0 {
				turtle.WriteString(" ")
			} else {
				turtle.WriteString(" ;\n    ")
			}

			// Write predicate
			predicate := c.shortenURI(triple.Predicate)
			turtle.WriteString(predicate)
			turtle.WriteString(" ")

			// Write object
			if triple.ObjectType == "uri" {
				turtle.WriteString(fmt.Sprintf("<%s>", triple.Object))
			} else if triple.ObjectType == "literal" {
				if triple.DataType != "" {
					dataType := c.shortenURI(triple.DataType)
					turtle.WriteString(fmt.Sprintf("\"%s\"^^%s", c.escapeLiteral(triple.Object), dataType))
				} else {
					turtle.WriteString(fmt.Sprintf("\"%s\"", c.escapeLiteral(triple.Object)))
				}
			}
		}

		turtle.WriteString(" .\n\n")
	}

	return []byte(turtle.String()), nil
}

// ConvertToJSONLD converts a container to JSON-LD format
func (c *ContainerRDFConverter) ConvertToJSONLD(container *domain.Container, baseURI string) ([]byte, error) {
	if container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	containerURI := baseURI + container.ID()

	// Build JSON-LD structure
	jsonld := map[string]interface{}{
		"@context": map[string]interface{}{
			"rdf":         "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"ldp":         "http://www.w3.org/ns/ldp#",
			"dcterms":     "http://purl.org/dc/terms/",
			"xsd":         "http://www.w3.org/2001/XMLSchema#",
			"type":        "rdf:type",
			"contains":    "ldp:contains",
			"title":       "dcterms:title",
			"description": "dcterms:description",
			"created":     "dcterms:created",
			"modified":    "dcterms:modified",
		},
		"@id":   containerURI,
		"@type": []string{"ldp:" + container.ContainerType.String()},
	}

	// Add title if present
	if title := container.GetTitle(); title != "" {
		jsonld["title"] = title
	}

	// Add description if present
	if description := container.GetDescription(); description != "" {
		jsonld["description"] = description
	}

	// Add timestamps
	metadata := container.GetMetadata()
	if createdAt, exists := metadata["createdAt"]; exists {
		if t, ok := createdAt.(time.Time); ok {
			jsonld["created"] = map[string]interface{}{
				"@type":  "xsd:dateTime",
				"@value": t.Format(time.RFC3339),
			}
		}
	}

	if updatedAt, exists := metadata["updatedAt"]; exists {
		if t, ok := updatedAt.(time.Time); ok {
			jsonld["modified"] = map[string]interface{}{
				"@type":  "xsd:dateTime",
				"@value": t.Format(time.RFC3339),
			}
		}
	}

	// Add membership information
	if len(container.Members) > 0 {
		contains := make([]map[string]interface{}, len(container.Members))
		for i, memberID := range container.Members {
			contains[i] = map[string]interface{}{
				"@id": baseURI + memberID,
			}
		}
		jsonld["contains"] = contains
	}

	// Marshal to JSON
	result, err := json.MarshalIndent(jsonld, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-LD: %w", err)
	}

	return result, nil
}

// ConvertToRDFXML converts a container to RDF/XML format
func (c *ContainerRDFConverter) ConvertToRDFXML(container *domain.Container, baseURI string) ([]byte, error) {
	if container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}

	containerURI := baseURI + container.ID()

	var rdfxml strings.Builder

	// XML declaration and RDF root element
	rdfxml.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	rdfxml.WriteString("<rdf:RDF\n")
	rdfxml.WriteString("    xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\"\n")
	rdfxml.WriteString("    xmlns:ldp=\"http://www.w3.org/ns/ldp#\"\n")
	rdfxml.WriteString("    xmlns:dcterms=\"http://purl.org/dc/terms/\"\n")
	rdfxml.WriteString("    xmlns:xsd=\"http://www.w3.org/2001/XMLSchema#\">\n\n")

	// Container description
	rdfxml.WriteString(fmt.Sprintf("  <rdf:Description rdf:about=\"%s\">\n", containerURI))
	rdfxml.WriteString(fmt.Sprintf("    <rdf:type rdf:resource=\"http://www.w3.org/ns/ldp#%s\"/>\n", container.ContainerType.String()))

	// Add title if present
	if title := container.GetTitle(); title != "" {
		rdfxml.WriteString(fmt.Sprintf("    <dcterms:title>%s</dcterms:title>\n", c.escapeXML(title)))
	}

	// Add description if present
	if description := container.GetDescription(); description != "" {
		rdfxml.WriteString(fmt.Sprintf("    <dcterms:description>%s</dcterms:description>\n", c.escapeXML(description)))
	}

	// Add timestamps
	metadata := container.GetMetadata()
	if createdAt, exists := metadata["createdAt"]; exists {
		if t, ok := createdAt.(time.Time); ok {
			rdfxml.WriteString(fmt.Sprintf("    <dcterms:created rdf:datatype=\"http://www.w3.org/2001/XMLSchema#dateTime\">%s</dcterms:created>\n", t.Format(time.RFC3339)))
		}
	}

	if updatedAt, exists := metadata["updatedAt"]; exists {
		if t, ok := updatedAt.(time.Time); ok {
			rdfxml.WriteString(fmt.Sprintf("    <dcterms:modified rdf:datatype=\"http://www.w3.org/2001/XMLSchema#dateTime\">%s</dcterms:modified>\n", t.Format(time.RFC3339)))
		}
	}

	// Add membership triples
	for _, memberID := range container.Members {
		memberURI := baseURI + memberID
		rdfxml.WriteString(fmt.Sprintf("    <ldp:contains rdf:resource=\"%s\"/>\n", memberURI))
	}

	rdfxml.WriteString("  </rdf:Description>\n")
	rdfxml.WriteString("</rdf:RDF>\n")

	return []byte(rdfxml.String()), nil
}

// GenerateMembershipTriples generates LDP membership triples for a container
func (c *ContainerRDFConverter) GenerateMembershipTriples(container *domain.Container, baseURI string) []ContainerTriple {
	var triples []ContainerTriple

	containerURI := baseURI + container.ID()

	// Generate ldp:contains triples for each member
	for _, memberID := range container.Members {
		memberURI := baseURI + memberID
		triple := ContainerTriple{
			Subject:    containerURI,
			Predicate:  "http://www.w3.org/ns/ldp#contains",
			Object:     memberURI,
			ObjectType: "uri",
		}
		triples = append(triples, triple)
	}

	return triples
}

// generateAllTriples generates all RDF triples for a container
func (c *ContainerRDFConverter) generateAllTriples(container *domain.Container, baseURI string) []ContainerTriple {
	var triples []ContainerTriple

	containerURI := baseURI + container.ID()

	// Container type triple
	triples = append(triples, ContainerTriple{
		Subject:    containerURI,
		Predicate:  "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		Object:     "http://www.w3.org/ns/ldp#" + container.ContainerType.String(),
		ObjectType: "uri",
	})

	// Title triple
	if title := container.GetTitle(); title != "" {
		triples = append(triples, ContainerTriple{
			Subject:    containerURI,
			Predicate:  "http://purl.org/dc/terms/title",
			Object:     title,
			ObjectType: "literal",
		})
	}

	// Description triple
	if description := container.GetDescription(); description != "" {
		triples = append(triples, ContainerTriple{
			Subject:    containerURI,
			Predicate:  "http://purl.org/dc/terms/description",
			Object:     description,
			ObjectType: "literal",
		})
	}

	// Timestamp triples
	metadata := container.GetMetadata()
	if createdAt, exists := metadata["createdAt"]; exists {
		if t, ok := createdAt.(time.Time); ok {
			triples = append(triples, ContainerTriple{
				Subject:    containerURI,
				Predicate:  "http://purl.org/dc/terms/created",
				Object:     t.Format(time.RFC3339),
				ObjectType: "literal",
				DataType:   "http://www.w3.org/2001/XMLSchema#dateTime",
			})
		}
	}

	if updatedAt, exists := metadata["updatedAt"]; exists {
		if t, ok := updatedAt.(time.Time); ok {
			triples = append(triples, ContainerTriple{
				Subject:    containerURI,
				Predicate:  "http://purl.org/dc/terms/modified",
				Object:     t.Format(time.RFC3339),
				ObjectType: "literal",
				DataType:   "http://www.w3.org/2001/XMLSchema#dateTime",
			})
		}
	}

	// Add membership triples
	membershipTriples := c.GenerateMembershipTriples(container, baseURI)
	triples = append(triples, membershipTriples...)

	return triples
}

// shortenURI shortens common URIs using prefixes
func (c *ContainerRDFConverter) shortenURI(uri string) string {
	prefixes := map[string]string{
		"http://www.w3.org/1999/02/22-rdf-syntax-ns#": "rdf:",
		"http://www.w3.org/ns/ldp#":                   "ldp:",
		"http://purl.org/dc/terms/":                   "dcterms:",
		"http://www.w3.org/2001/XMLSchema#":           "xsd:",
	}

	for namespace, prefix := range prefixes {
		if strings.HasPrefix(uri, namespace) {
			return prefix + uri[len(namespace):]
		}
	}

	return "<" + uri + ">"
}

// escapeLiteral escapes special characters in RDF literals
func (c *ContainerRDFConverter) escapeLiteral(literal string) string {
	literal = strings.ReplaceAll(literal, "\\", "\\\\")
	literal = strings.ReplaceAll(literal, "\"", "\\\"")
	literal = strings.ReplaceAll(literal, "\n", "\\n")
	literal = strings.ReplaceAll(literal, "\r", "\\r")
	literal = strings.ReplaceAll(literal, "\t", "\\t")
	return literal
}

// escapeXML escapes special characters for XML
func (c *ContainerRDFConverter) escapeXML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}
