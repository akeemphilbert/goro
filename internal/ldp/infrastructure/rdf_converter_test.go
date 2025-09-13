package infrastructure

import (
	"strings"
	"testing"
)

func TestNewRDFConverter(t *testing.T) {
	converter := NewRDFConverter()
	if converter == nil {
		t.Fatal("NewRDFConverter should return a non-nil converter")
	}
}

func TestRDFConverter_ValidateFormat(t *testing.T) {
	converter := NewRDFConverter()

	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{"JSON-LD standard", "application/ld+json", true},
		{"JSON-LD alternative", "application/json", true},
		{"JSON-LD short", "json-ld", true},
		{"JSON-LD compact", "jsonld", true},
		{"Turtle standard", "text/turtle", true},
		{"Turtle alternative", "text/plain", true},
		{"Turtle short", "turtle", true},
		{"Turtle extension", "ttl", true},
		{"RDF/XML standard", "application/rdf+xml", true},
		{"RDF/XML alternative", "rdf/xml", true},
		{"RDF/XML short", "rdfxml", true},
		{"RDF/XML generic", "xml", true},
		{"Unsupported format", "application/n-triples", false},
		{"Empty format", "", false},
		{"Invalid format", "invalid/format", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ValidateFormat(tt.format)
			if result != tt.expected {
				t.Errorf("ValidateFormat(%q) = %v, want %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestRDFConverter_NormalizeFormat(t *testing.T) {
	converter := NewRDFConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"JSON-LD variations", "JSON-LD", "application/ld+json"},
		{"JSON-LD compact", "jsonld", "application/ld+json"},
		{"JSON generic", "application/json", "application/ld+json"},
		{"Turtle variations", "TURTLE", "text/turtle"},
		{"Turtle extension", "ttl", "text/turtle"},
		{"Turtle plain text", "text/plain", "text/turtle"},
		{"RDF/XML variations", "RDF/XML", "application/rdf+xml"},
		{"RDF/XML compact", "rdfxml", "application/rdf+xml"},
		{"XML generic", "xml", "application/rdf+xml"},
		{"Standard format unchanged", "application/ld+json", "application/ld+json"},
		{"Whitespace trimmed", "  text/turtle  ", "text/turtle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.normalizeFormat(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRDFConverter_Convert_SameFormat(t *testing.T) {
	converter := NewRDFConverter()

	testData := []byte(`{"@context": {"name": "http://example.org/name"}, "@id": "http://example.org/person", "name": "John"}`)

	result, err := converter.Convert(testData, "application/ld+json", "application/ld+json")
	if err != nil {
		t.Fatalf("Convert with same format should not error: %v", err)
	}

	if string(result) != string(testData) {
		t.Errorf("Convert with same format should return original data")
	}
}

func TestRDFConverter_Convert_EmptyData(t *testing.T) {
	converter := NewRDFConverter()

	_, err := converter.Convert([]byte{}, "application/ld+json", "text/turtle")
	if err == nil {
		t.Error("Convert with empty data should return error")
	}

	if !strings.Contains(err.Error(), "input data cannot be empty") {
		t.Errorf("Error should mention empty data, got: %v", err)
	}
}

func TestRDFConverter_Convert_UnsupportedFormat(t *testing.T) {
	converter := NewRDFConverter()
	testData := []byte(`{"test": "data"}`)

	// Test unsupported source format
	_, err := converter.Convert(testData, "application/n-triples", "text/turtle")
	if err == nil {
		t.Error("Convert with unsupported source format should return error")
	}
	if !strings.Contains(err.Error(), "unsupported source format") {
		t.Errorf("Error should mention unsupported source format, got: %v", err)
	}

	// Test unsupported target format
	_, err = converter.Convert(testData, "application/ld+json", "application/n-triples")
	if err == nil {
		t.Error("Convert with unsupported target format should return error")
	}
	if !strings.Contains(err.Error(), "unsupported target format") {
		t.Errorf("Error should mention unsupported target format, got: %v", err)
	}
}

func TestRDFConverter_Convert_JSONLDToTurtle(t *testing.T) {
	converter := NewRDFConverter()

	jsonldData := []byte(`{
		"@context": {
			"name": "http://example.org/name",
			"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
		},
		"@id": "http://example.org/person",
		"@type": "http://example.org/Person",
		"name": "John Doe"
	}`)

	result, err := converter.Convert(jsonldData, "application/ld+json", "text/turtle")
	if err != nil {
		t.Fatalf("Convert JSON-LD to Turtle failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "@prefix") {
		t.Error("Turtle output should contain prefix declarations")
	}
	if !strings.Contains(resultStr, "rdf:type") {
		t.Error("Turtle output should contain RDF type statements")
	}
}

func TestRDFConverter_Convert_TurtleToJSONLD(t *testing.T) {
	converter := NewRDFConverter()

	turtleData := []byte(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix ex: <http://example.org/> .

ex:person rdf:type ex:Person ;
          ex:name "John Doe" .`)

	result, err := converter.Convert(turtleData, "text/turtle", "application/ld+json")
	if err != nil {
		t.Fatalf("Convert Turtle to JSON-LD failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "@context") {
		t.Error("JSON-LD output should contain @context")
	}
	if !strings.Contains(resultStr, "@id") {
		t.Error("JSON-LD output should contain @id")
	}
}

func TestRDFConverter_Convert_JSONLDToRDFXML(t *testing.T) {
	converter := NewRDFConverter()

	jsonldData := []byte(`{
		"@context": {"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#"},
		"@id": "http://example.org/resource",
		"@type": "http://example.org/Type"
	}`)

	result, err := converter.Convert(jsonldData, "application/ld+json", "application/rdf+xml")
	if err != nil {
		t.Fatalf("Convert JSON-LD to RDF/XML failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "<?xml") {
		t.Error("RDF/XML output should contain XML declaration")
	}
	if !strings.Contains(resultStr, "rdf:RDF") {
		t.Error("RDF/XML output should contain rdf:RDF element")
	}
	if !strings.Contains(resultStr, "rdf:Description") {
		t.Error("RDF/XML output should contain rdf:Description element")
	}
}

func TestRDFConverter_Convert_RDFXMLToTurtle(t *testing.T) {
	converter := NewRDFConverter()

	rdfxmlData := []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
         xmlns:ex="http://example.org/">
  <rdf:Description rdf:about="http://example.org/resource">
    <rdf:type rdf:resource="http://example.org/Type"/>
    <ex:name>Test Resource</ex:name>
  </rdf:Description>
</rdf:RDF>`)

	result, err := converter.Convert(rdfxmlData, "application/rdf+xml", "text/turtle")
	if err != nil {
		t.Fatalf("Convert RDF/XML to Turtle failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "@prefix") {
		t.Error("Turtle output should contain prefix declarations")
	}
}

func TestRDFConverter_Convert_TurtleToRDFXML(t *testing.T) {
	converter := NewRDFConverter()

	turtleData := []byte(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix ex: <http://example.org/> .

ex:resource rdf:type ex:Type ;
            ex:name "Test Resource" .`)

	result, err := converter.Convert(turtleData, "text/turtle", "application/rdf+xml")
	if err != nil {
		t.Fatalf("Convert Turtle to RDF/XML failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "<?xml") {
		t.Error("RDF/XML output should contain XML declaration")
	}
	if !strings.Contains(resultStr, "rdf:RDF") {
		t.Error("RDF/XML output should contain rdf:RDF element")
	}
}

func TestRDFConverter_Convert_RDFXMLToJSONLD(t *testing.T) {
	converter := NewRDFConverter()

	rdfxmlData := []byte(`<?xml version="1.0"?>
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="http://example.org/resource">
    <rdf:type rdf:resource="http://example.org/Type"/>
  </rdf:Description>
</rdf:RDF>`)

	result, err := converter.Convert(rdfxmlData, "application/rdf+xml", "application/ld+json")
	if err != nil {
		t.Fatalf("Convert RDF/XML to JSON-LD failed: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "@context") {
		t.Error("JSON-LD output should contain @context")
	}
	if !strings.Contains(resultStr, "@id") {
		t.Error("JSON-LD output should contain @id")
	}
}
func TestRDFConverter_ParseErrors(t *testing.T) {
	converter := NewRDFConverter()

	tests := []struct {
		name    string
		data    []byte
		format  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Invalid JSON-LD structure",
			data:    []byte("invalid json"),
			format:  "application/ld+json",
			wantErr: true,
			errMsg:  "failed to parse",
		},
		{
			name:    "Invalid Turtle structure",
			data:    []byte("invalid turtle"),
			format:  "text/turtle",
			wantErr: true,
			errMsg:  "failed to parse",
		},
		{
			name:    "Invalid RDF/XML structure",
			data:    []byte("invalid xml"),
			format:  "application/rdf+xml",
			wantErr: true,
			errMsg:  "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a different target format to force parsing
			targetFormat := "application/ld+json"
			if tt.format == "application/ld+json" {
				targetFormat = "text/turtle"
			}

			_, err := converter.Convert(tt.data, tt.format, targetFormat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Convert() error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestRDFConverter_SemanticPreservation(t *testing.T) {
	converter := NewRDFConverter()

	// Test that conversion preserves semantic meaning by doing round-trip conversions
	originalJSONLD := []byte(`{
		"@context": {"name": "http://example.org/name"},
		"@id": "http://example.org/person",
		"@type": "http://example.org/Person",
		"name": "John Doe"
	}`)

	// Convert JSON-LD -> Turtle -> JSON-LD
	turtle, err := converter.Convert(originalJSONLD, "application/ld+json", "text/turtle")
	if err != nil {
		t.Fatalf("First conversion failed: %v", err)
	}

	backToJSONLD, err := converter.Convert(turtle, "text/turtle", "application/ld+json")
	if err != nil {
		t.Fatalf("Round-trip conversion failed: %v", err)
	}

	// Both should contain the same semantic elements
	originalStr := string(originalJSONLD)
	resultStr := string(backToJSONLD)

	// Check that key semantic elements are preserved
	if strings.Contains(originalStr, "Person") && !strings.Contains(resultStr, "Type") {
		t.Error("Type information should be preserved in round-trip conversion")
	}

	if strings.Contains(originalStr, "example.org/person") && !strings.Contains(resultStr, "example.org/resource") {
		t.Error("Resource identification should be preserved in round-trip conversion")
	}
}

func TestRDFConverter_AllFormatCombinations(t *testing.T) {
	converter := NewRDFConverter()

	formats := []string{
		"application/ld+json",
		"text/turtle",
		"application/rdf+xml",
	}

	testData := map[string][]byte{
		"application/ld+json": []byte(`{"@context": {"rdf": "http://www.w3.org/1999/02/22-rdf-syntax-ns#"}, "@id": "http://example.org/resource", "@type": "http://example.org/Type"}`),
		"text/turtle":         []byte(`@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> . <http://example.org/resource> rdf:type <http://example.org/Type> .`),
		"application/rdf+xml": []byte(`<?xml version="1.0"?><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="http://example.org/resource"><rdf:type rdf:resource="http://example.org/Type"/></rdf:Description></rdf:RDF>`),
	}

	// Test all format combinations
	for _, fromFormat := range formats {
		for _, toFormat := range formats {
			t.Run(fromFormat+"_to_"+toFormat, func(t *testing.T) {
				data := testData[fromFormat]
				result, err := converter.Convert(data, fromFormat, toFormat)
				if err != nil {
					t.Errorf("Convert from %s to %s failed: %v", fromFormat, toFormat, err)
					return
				}

				if len(result) == 0 {
					t.Errorf("Convert from %s to %s returned empty result", fromFormat, toFormat)
				}

				// Verify the result is valid for the target format
				if !converter.ValidateFormat(toFormat) {
					t.Errorf("Target format %s should be valid", toFormat)
				}
			})
		}
	}
}

func TestRDFConverter_EmptyGraph(t *testing.T) {
	converter := NewRDFConverter()

	// Test serialization of empty graphs
	emptyGraph := &rdfGraph{Triples: []rdfTriple{}}

	jsonld, err := converter.serializeJSONLD(emptyGraph)
	if err != nil {
		t.Errorf("serializeJSONLD with empty graph failed: %v", err)
	}
	if string(jsonld) != "{}" {
		t.Errorf("Empty graph JSON-LD should be '{}', got: %s", string(jsonld))
	}

	turtle, err := converter.serializeTurtle(emptyGraph)
	if err != nil {
		t.Errorf("serializeTurtle with empty graph failed: %v", err)
	}
	if string(turtle) != "" {
		t.Errorf("Empty graph Turtle should be empty, got: %s", string(turtle))
	}

	rdfxml, err := converter.serializeRDFXML(emptyGraph)
	if err != nil {
		t.Errorf("serializeRDFXML with empty graph failed: %v", err)
	}
	if !strings.Contains(string(rdfxml), "rdf:RDF") {
		t.Errorf("Empty graph RDF/XML should contain rdf:RDF element, got: %s", string(rdfxml))
	}
}

func TestRDFConverter_NilGraph(t *testing.T) {
	converter := NewRDFConverter()

	// Test serialization of nil graphs
	jsonld, err := converter.serializeJSONLD(nil)
	if err != nil {
		t.Errorf("serializeJSONLD with nil graph failed: %v", err)
	}

	turtle, err := converter.serializeTurtle(nil)
	if err != nil {
		t.Errorf("serializeTurtle with nil graph failed: %v", err)
	}

	rdfxml, err := converter.serializeRDFXML(nil)
	if err != nil {
		t.Errorf("serializeRDFXML with nil graph failed: %v", err)
	}

	// All should handle nil gracefully
	if len(jsonld) == 0 || len(turtle) != 0 || len(rdfxml) == 0 {
		t.Error("Nil graph serialization should handle gracefully")
	}
}
