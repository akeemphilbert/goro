package middleware

import (
	"net/http"
	"strconv"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ContentNegotiationConfig holds configuration for content negotiation middleware
type ContentNegotiationConfig struct {
	SupportedFormats []string
	DefaultFormat    string
}

// DefaultContentNegotiationConfig returns a default content negotiation configuration for RDF formats
func DefaultContentNegotiationConfig() ContentNegotiationConfig {
	return ContentNegotiationConfig{
		SupportedFormats: []string{
			"application/ld+json",
			"text/turtle",
			"application/rdf+xml",
		},
		DefaultFormat: "application/ld+json",
	}
}

// ContentNegotiation returns a content negotiation filter with default configuration
func ContentNegotiation() khttp.FilterFunc {
	return ContentNegotiationWithConfig(DefaultContentNegotiationConfig())
}

// ContentNegotiationWithConfig returns a content negotiation filter with custom configuration
func ContentNegotiationWithConfig(config ContentNegotiationConfig) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only process for GET and HEAD requests (content negotiation for responses)
			if r.Method == "GET" || r.Method == "HEAD" {
				acceptHeader := r.Header.Get("Accept")
				if acceptHeader != "" {
					negotiatedFormat := negotiateFormat(acceptHeader, config.SupportedFormats, config.DefaultFormat)

					// Store negotiated format in request context for handlers to use
					r.Header.Set("X-Negotiated-Format", negotiatedFormat)
				}
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// negotiateFormat performs content negotiation based on Accept header
func negotiateFormat(acceptHeader string, supportedFormats []string, defaultFormat string) string {
	if acceptHeader == "" {
		return defaultFormat
	}

	// Parse Accept header
	acceptTypes := parseAcceptHeader(acceptHeader)

	// Find the best match
	for _, acceptType := range acceptTypes {
		for _, format := range supportedFormats {
			if matchesMediaType(acceptType.MediaType, format) {
				return format
			}
		}
	}

	// Check for wildcard acceptance
	for _, acceptType := range acceptTypes {
		if acceptType.MediaType == "*/*" || acceptType.MediaType == "application/*" {
			return defaultFormat
		}
	}

	// No match found, return empty string to let handler decide
	return ""
}

// AcceptType represents a parsed Accept header entry
type AcceptType struct {
	MediaType string
	Quality   float64
}

// parseAcceptHeader parses the Accept header into media types with quality values
func parseAcceptHeader(acceptHeader string) []AcceptType {
	var acceptTypes []AcceptType

	parts := strings.Split(acceptHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split media type and parameters
		segments := strings.Split(part, ";")
		mediaType := strings.TrimSpace(segments[0])

		quality := 1.0 // Default quality

		// Parse quality parameter if present
		for i := 1; i < len(segments); i++ {
			param := strings.TrimSpace(segments[i])
			if strings.HasPrefix(param, "q=") {
				if q, err := strconv.ParseFloat(param[2:], 64); err == nil {
					quality = q
				}
			}
		}

		acceptTypes = append(acceptTypes, AcceptType{
			MediaType: mediaType,
			Quality:   quality,
		})
	}

	// Sort by quality (highest first)
	for i := 0; i < len(acceptTypes)-1; i++ {
		for j := i + 1; j < len(acceptTypes); j++ {
			if acceptTypes[j].Quality > acceptTypes[i].Quality {
				acceptTypes[i], acceptTypes[j] = acceptTypes[j], acceptTypes[i]
			}
		}
	}

	return acceptTypes
}

// matchesMediaType checks if an accept type matches a supported format
func matchesMediaType(acceptType, format string) bool {
	// Exact match
	if acceptType == format {
		return true
	}

	// Handle common aliases and variations
	switch acceptType {
	case "application/json":
		return format == "application/ld+json"
	case "text/plain":
		return format == "text/turtle"
	case "application/xml":
		return format == "application/rdf+xml"
	case "application/turtle":
		return format == "text/turtle"
	case "text/rdf":
		return format == "text/turtle"
	}

	// Handle wildcard patterns
	if strings.Contains(acceptType, "*") {
		parts := strings.Split(acceptType, "/")
		formatParts := strings.Split(format, "/")

		if len(parts) == 2 && len(formatParts) == 2 {
			// Check type/* patterns
			if parts[0] == formatParts[0] && parts[1] == "*" {
				return true
			}
		}
	}

	return false
}

// ValidateContentType validates if a content type is supported for RDF resources
func ValidateContentType(contentType string, supportedFormats []string) bool {
	// Normalize content type (remove charset and other parameters)
	normalizedType := strings.Split(strings.ToLower(strings.TrimSpace(contentType)), ";")[0]

	for _, format := range supportedFormats {
		if normalizedType == format {
			return true
		}
	}

	// Check common aliases
	switch normalizedType {
	case "application/json":
		return containsString(supportedFormats, "application/ld+json")
	case "text/plain":
		return containsString(supportedFormats, "text/turtle")
	case "application/xml":
		return containsString(supportedFormats, "application/rdf+xml")
	}

	return false
}

// containsString checks if a slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
