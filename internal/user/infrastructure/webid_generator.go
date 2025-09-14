package infrastructure

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	// ErrInvalidWebID is returned when a WebID format is invalid
	ErrInvalidWebID = fmt.Errorf("invalid WebID format")

	// ErrWebIDNotUnique is returned when a WebID is not unique
	ErrWebIDNotUnique = fmt.Errorf("WebID is not unique")
)

// WebIDUniquenessChecker defines the interface for checking WebID uniqueness
type WebIDUniquenessChecker interface {
	WebIDExists(ctx context.Context, webID string) (bool, error)
}

// WebIDGenerator defines the interface for WebID generation and validation
type WebIDGenerator interface {
	GenerateWebID(ctx context.Context, userID, email, userName string) (string, error)
	GenerateWebIDDocument(ctx context.Context, webID, email, userName string) (string, error)
	ValidateWebID(ctx context.Context, webID string) error
	IsUniqueWebID(ctx context.Context, webID string) (bool, error)
	GenerateAlternativeWebID(ctx context.Context, baseWebID string) (string, error)
	SetUniquenessChecker(checker WebIDUniquenessChecker)
}

// webIDGenerator implements the WebIDGenerator interface
type webIDGenerator struct {
	baseURL           string
	uniquenessChecker WebIDUniquenessChecker
}

// NewWebIDGenerator creates a new WebID generator with the given base URL
func NewWebIDGenerator(baseURL string) WebIDGenerator {
	return &webIDGenerator{
		baseURL: baseURL,
	}
}

// NewWebIDGeneratorWithValidation creates a new WebID generator with validation
func NewWebIDGeneratorWithValidation(baseURL string) (WebIDGenerator, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL format: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("base URL must use HTTPS scheme")
	}

	return &webIDGenerator{
		baseURL: baseURL,
	}, nil
}

// SetUniquenessChecker sets the uniqueness checker for the generator
func (g *webIDGenerator) SetUniquenessChecker(checker WebIDUniquenessChecker) {
	g.uniquenessChecker = checker
}

// GenerateWebID generates a WebID URI for a user
func (g *webIDGenerator) GenerateWebID(ctx context.Context, userID, email, userName string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("user ID is required")
	}

	if strings.TrimSpace(email) == "" {
		return "", fmt.Errorf("email is required")
	}

	if strings.TrimSpace(userName) == "" {
		return "", fmt.Errorf("user name is required")
	}

	// Sanitize user ID for URL usage
	sanitizedUserID := sanitizeForURL(userID)

	// Construct WebID URI
	webID := fmt.Sprintf("%s/users/%s#me", g.baseURL, sanitizedUserID)

	return webID, nil
}

// GenerateWebIDDocument generates a Turtle format WebID document
func (g *webIDGenerator) GenerateWebIDDocument(ctx context.Context, webID, email, userName string) (string, error) {
	if strings.TrimSpace(webID) == "" {
		return "", fmt.Errorf("WebID is required")
	}

	if strings.TrimSpace(email) == "" {
		return "", fmt.Errorf("email is required")
	}

	if strings.TrimSpace(userName) == "" {
		return "", fmt.Errorf("user name is required")
	}

	// Escape special characters in user name for Turtle format
	escapedUserName := escapeTurtleString(userName)

	// Generate Turtle document
	document := fmt.Sprintf(`@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix solid: <http://www.w3.org/ns/solid/terms#> .
@prefix ldp: <http://www.w3.org/ns/ldp#> .

<%s> a foaf:Person ;
    foaf:name "%s" ;
    foaf:mbox <mailto:%s> ;
    solid:account <%s/account> ;
    solid:privateTypeIndex <%s/private/index.ttl> ;
    solid:publicTypeIndex <%s/public/index.ttl> .
`, webID, escapedUserName, email,
		strings.TrimSuffix(webID, "#me"),
		strings.TrimSuffix(webID, "#me"),
		strings.TrimSuffix(webID, "#me"))

	return document, nil
}

// ValidateWebID validates the format of a WebID
func (g *webIDGenerator) ValidateWebID(ctx context.Context, webID string) error {
	if strings.TrimSpace(webID) == "" {
		return ErrInvalidWebID
	}

	parsedURL, err := url.Parse(webID)
	if err != nil {
		return fmt.Errorf("%w: invalid URL format", ErrInvalidWebID)
	}

	// WebID must use HTTPS
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: WebID must use HTTPS scheme", ErrInvalidWebID)
	}

	// WebID must have a fragment
	if parsedURL.Fragment == "" {
		return fmt.Errorf("%w: WebID must have a fragment identifier", ErrInvalidWebID)
	}

	return nil
}

// IsUniqueWebID checks if a WebID is unique
func (g *webIDGenerator) IsUniqueWebID(ctx context.Context, webID string) (bool, error) {
	if strings.TrimSpace(webID) == "" {
		return false, ErrInvalidWebID
	}

	if g.uniquenessChecker == nil {
		// If no checker is set, assume it's unique (for testing)
		return true, nil
	}

	exists, err := g.uniquenessChecker.WebIDExists(ctx, webID)
	if err != nil {
		return false, fmt.Errorf("failed to check WebID uniqueness: %w", err)
	}

	return !exists, nil
}

// GenerateAlternativeWebID generates an alternative WebID when the base is not unique
func (g *webIDGenerator) GenerateAlternativeWebID(ctx context.Context, baseWebID string) (string, error) {
	if strings.TrimSpace(baseWebID) == "" {
		return "", fmt.Errorf("base WebID is required")
	}

	// Parse the base WebID to extract components
	parsedURL, err := url.Parse(baseWebID)
	if err != nil {
		return "", fmt.Errorf("invalid base WebID format: %w", err)
	}

	// Extract the path without the fragment
	basePath := parsedURL.Path
	fragment := parsedURL.Fragment

	// Try adding numeric suffixes until we find a unique one
	for i := 1; i <= 100; i++ { // Limit attempts to prevent infinite loops
		// Insert suffix before the fragment
		alternativePath := strings.TrimSuffix(basePath, "/") + "-" + strconv.Itoa(i)
		alternativeWebID := fmt.Sprintf("%s://%s%s#%s", parsedURL.Scheme, parsedURL.Host, alternativePath, fragment)

		isUnique, err := g.IsUniqueWebID(ctx, alternativeWebID)
		if err != nil {
			return "", fmt.Errorf("failed to check uniqueness for alternative WebID: %w", err)
		}

		if isUnique {
			return alternativeWebID, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique alternative WebID after 100 attempts")
}

// sanitizeForURL removes or replaces characters that are not safe for URLs
func sanitizeForURL(input string) string {
	// Replace common problematic characters
	sanitized := strings.ReplaceAll(input, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "+", "-plus-")

	// Remove any remaining non-alphanumeric characters except hyphens and underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	sanitized = reg.ReplaceAllString(sanitized, "")

	// Ensure it doesn't start or end with hyphens
	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}

// escapeTurtleString escapes special characters for Turtle format
func escapeTurtleString(input string) string {
	// Escape backslashes first
	escaped := strings.ReplaceAll(input, "\\", "\\\\")

	// Escape double quotes
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

	// Escape newlines and tabs
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")

	return escaped
}
