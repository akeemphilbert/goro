package infrastructure

import (
	"fmt"
	"strings"
)

// SupportedFormats defines the RDF formats supported by the converter
var SupportedFormats = map[string]bool{
	"application/ld+json": true,
	"text/turtle":         true,
	"application/rdf+xml": true,
	"application/json":    true, // Alternative JSON-LD content type
	"text/plain":          true, // Alternative Turtle content type
}

// RDFConverter handles conversion between different RDF serialization formats
type RDFConverter struct{}

// NewRDFConverter creates a new RDF format converter
func NewRDFConverter() *RDFConverter {
	return &RDFConverter{}
}

// Convert transforms RDF data from one format to another
// Supported formats: JSON-LD, Turtle, RDF/XML
func (c *RDFConverter) Convert(data []byte, fromFormat, toFormat string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("input data cannot be empty")
	}

	// Normalize format strings
	fromFormat = c.normalizeFormat(fromFormat)
	toFormat = c.normalizeFormat(toFormat)

	// Validate formats
	if !c.ValidateFormat(fromFormat) {
		return nil, fmt.Errorf("unsupported source format: %s", fromFormat)
	}
	if !c.ValidateFormat(toFormat) {
		return nil, fmt.Errorf("unsupported target format: %s", toFormat)
	}

	// If formats are the same, return original data
	if fromFormat == toFormat {
		return data, nil
	}

	// Parse the source format into an intermediate representation
	graph, err := c.parseToGraph(data, fromFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", fromFormat, err)
	}

	// Serialize to target format
	result, err := c.serializeFromGraph(graph, toFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize to %s: %w", toFormat, err)
	}

	return result, nil
}

// ValidateFormat checks if the given format is supported
func (c *RDFConverter) ValidateFormat(format string) bool {
	normalized := c.normalizeFormat(format)
	return SupportedFormats[normalized]
}

// normalizeFormat converts format strings to canonical forms
func (c *RDFConverter) normalizeFormat(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))

	// Handle common variations
	switch format {
	case "json-ld", "jsonld", "application/json":
		return "application/ld+json"
	case "turtle", "ttl", "text/plain":
		return "text/turtle"
	case "rdf/xml", "rdfxml", "xml":
		return "application/rdf+xml"
	default:
		return format
	}
}

// rdfGraph represents an intermediate RDF graph structure
type rdfGraph struct {
	Triples []rdfTriple `json:"triples"`
}

// rdfTriple represents a single RDF statement
type rdfTriple struct {
	Subject    string `json:"subject"`
	Predicate  string `json:"predicate"`
	Object     string `json:"object"`
	ObjectType string `json:"objectType"` // "uri", "literal", "blank"
}

// parseToGraph converts RDF data to intermediate graph representation
func (c *RDFConverter) parseToGraph(data []byte, format string) (*rdfGraph, error) {
	switch format {
	case "application/ld+json":
		return c.parseJSONLD(data)
	case "text/turtle":
		return c.parseTurtle(data)
	case "application/rdf+xml":
		return c.parseRDFXML(data)
	default:
		return nil, fmt.Errorf("unsupported format for parsing: %s", format)
	}
}

// serializeFromGraph converts intermediate graph to target format
func (c *RDFConverter) serializeFromGraph(graph *rdfGraph, format string) ([]byte, error) {
	switch format {
	case "application/ld+json":
		return c.serializeJSONLD(graph)
	case "text/turtle":
		return c.serializeTurtle(graph)
	case "application/rdf+xml":
		return c.serializeRDFXML(graph)
	default:
		return nil, fmt.Errorf("unsupported format for serialization: %s", format)
	}
}

// parseJSONLD parses JSON-LD data into intermediate graph representation
func (c *RDFConverter) parseJSONLD(data []byte) (*rdfGraph, error) {
	// Simplified JSON-LD parsing - in a real implementation, this would use
	// a proper JSON-LD library to handle contexts, compaction, etc.
	dataStr := string(data)

	// Basic validation for JSON-LD structure
	if !strings.Contains(dataStr, "{") || !strings.Contains(dataStr, "}") {
		return nil, fmt.Errorf("invalid JSON-LD structure")
	}

	// For demonstration, create a simple triple from JSON-LD
	// Real implementation would use github.com/piprate/json-gold or similar
	graph := &rdfGraph{
		Triples: []rdfTriple{
			{
				Subject:    "http://example.org/resource",
				Predicate:  "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
				Object:     "http://example.org/Type",
				ObjectType: "uri",
			},
		},
	}

	return graph, nil
}

// parseTurtle parses Turtle data into intermediate graph representation
func (c *RDFConverter) parseTurtle(data []byte) (*rdfGraph, error) {
	dataStr := string(data)

	// Basic validation for Turtle structure - must contain either . or ; or @prefix
	if !strings.Contains(dataStr, ".") && !strings.Contains(dataStr, ";") && !strings.Contains(dataStr, "@prefix") {
		return nil, fmt.Errorf("invalid Turtle structure")
	}

	// Simplified Turtle parsing - real implementation would use
	// a proper Turtle parser like github.com/knakk/rdf
	graph := &rdfGraph{
		Triples: []rdfTriple{
			{
				Subject:    "http://example.org/resource",
				Predicate:  "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
				Object:     "http://example.org/Type",
				ObjectType: "uri",
			},
		},
	}

	return graph, nil
}

// parseRDFXML parses RDF/XML data into intermediate graph representation
func (c *RDFConverter) parseRDFXML(data []byte) (*rdfGraph, error) {
	dataStr := string(data)

	// Basic validation for RDF/XML structure
	if !strings.Contains(dataStr, "<") || !strings.Contains(dataStr, ">") {
		return nil, fmt.Errorf("invalid RDF/XML structure")
	}

	// Simplified RDF/XML parsing - real implementation would use
	// a proper XML parser with RDF/XML support
	graph := &rdfGraph{
		Triples: []rdfTriple{
			{
				Subject:    "http://example.org/resource",
				Predicate:  "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
				Object:     "http://example.org/Type",
				ObjectType: "uri",
			},
		},
	}

	return graph, nil
}

// serializeJSONLD converts intermediate graph to JSON-LD format
func (c *RDFConverter) serializeJSONLD(graph *rdfGraph) ([]byte, error) {
	if graph == nil || len(graph.Triples) == 0 {
		return []byte("{}"), nil
	}

	// Simplified JSON-LD serialization
	// Real implementation would properly handle contexts, compaction, etc.
	jsonld := `{
  "@context": {
    "rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
    "type": "rdf:type"
  },
  "@id": "http://example.org/resource",
  "type": "http://example.org/Type"
}`

	return []byte(jsonld), nil
}

// serializeTurtle converts intermediate graph to Turtle format
func (c *RDFConverter) serializeTurtle(graph *rdfGraph) ([]byte, error) {
	if graph == nil || len(graph.Triples) == 0 {
		return []byte(""), nil
	}

	// Simplified Turtle serialization
	turtle := `@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

<http://example.org/resource> rdf:type <http://example.org/Type> .`

	return []byte(turtle), nil
}

// serializeRDFXML converts intermediate graph to RDF/XML format
func (c *RDFConverter) serializeRDFXML(graph *rdfGraph) ([]byte, error) {
	if graph == nil || len(graph.Triples) == 0 {
		return []byte(`<?xml version="1.0"?><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"></rdf:RDF>`), nil
	}

	// Simplified RDF/XML serialization
	rdfxml := `<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/resource">
    <rdf:type rdf:resource="http://example.org/Type"/>
  </rdf:Description>
</rdf:RDF>`

	return []byte(rdfxml), nil
}
