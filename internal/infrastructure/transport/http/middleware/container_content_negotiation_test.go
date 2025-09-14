package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContainerContentNegotiation_TurtleAccept(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		if negotiatedFormat != "text/turtle" {
			t.Errorf("Expected negotiated format 'text/turtle', got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with Turtle Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "text/turtle")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_JSONLDAccept(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		if negotiatedFormat != "application/ld+json" {
			t.Errorf("Expected negotiated format 'application/ld+json', got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with JSON-LD Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "application/ld+json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_RDFXMLAccept(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		if negotiatedFormat != "application/rdf+xml" {
			t.Errorf("Expected negotiated format 'application/rdf+xml', got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with RDF/XML Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "application/rdf+xml")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_QualityValues(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// Should prefer text/turtle due to higher quality value
		if negotiatedFormat != "text/turtle" {
			t.Errorf("Expected negotiated format 'text/turtle', got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with multiple Accept headers with quality values
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "application/ld+json;q=0.8, text/turtle;q=0.9, application/rdf+xml;q=0.7")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_WildcardAccept(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// Should default to JSON-LD for wildcard
		if negotiatedFormat != "application/ld+json" {
			t.Errorf("Expected negotiated format 'application/ld+json', got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with wildcard Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "*/*")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_NoAcceptHeader(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// Should be empty when no Accept header is provided
		if negotiatedFormat != "" {
			t.Errorf("Expected empty negotiated format, got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request without Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_UnsupportedFormat(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// Should be empty for unsupported format
		if negotiatedFormat != "" {
			t.Errorf("Expected empty negotiated format for unsupported format, got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with unsupported Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "text/html")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_POSTMethod(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// Should be empty for POST method (content negotiation only for GET/HEAD)
		if negotiatedFormat != "" {
			t.Errorf("Expected empty negotiated format for POST method, got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a POST request with Accept header
	req := httptest.NewRequest("POST", "/containers/test", nil)
	req.Header.Set("Accept", "text/turtle")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_HEADMethod(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		if negotiatedFormat != "text/turtle" {
			t.Errorf("Expected negotiated format 'text/turtle' for HEAD method, got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a HEAD request with Accept header
	req := httptest.NewRequest("HEAD", "/containers/test", nil)
	req.Header.Set("Accept", "text/turtle")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestContainerContentNegotiation_ApplicationJSONAlias(t *testing.T) {
	middleware := ContainerContentNegotiation()

	// Create a test handler that checks the negotiated format
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		negotiatedFormat := r.Header.Get("X-Negotiated-Format")
		// application/json should be treated as JSON-LD
		if negotiatedFormat != "application/ld+json" {
			t.Errorf("Expected negotiated format 'application/ld+json' for application/json, got '%s'", negotiatedFormat)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	// Create a request with application/json Accept header
	req := httptest.NewRequest("GET", "/containers/test", nil)
	req.Header.Set("Accept", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rr, req)

	// Check the response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}
