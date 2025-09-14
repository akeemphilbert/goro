package infrastructure

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/golang-jwt/jwt/v5"
)

// WebIDOIDCProvider implements WebID-OIDC authentication
type WebIDOIDCProvider struct {
	httpClient *http.Client
	cache      map[string]*cachedOIDCConfig
	cacheTTL   time.Duration
}

// cachedOIDCConfig represents a cached OIDC configuration
type cachedOIDCConfig struct {
	config    *domain.OIDCConfiguration
	expiresAt time.Time
}

// NewWebIDOIDCProvider creates a new WebID-OIDC provider
func NewWebIDOIDCProvider(config *WebIDOIDCConfig) *WebIDOIDCProvider {
	timeout, _ := time.ParseDuration(config.Timeout)
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	cacheTTL, _ := time.ParseDuration(config.CacheTTL)
	if cacheTTL == 0 {
		cacheTTL = time.Hour
	}

	return &WebIDOIDCProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cache:    make(map[string]*cachedOIDCConfig),
		cacheTTL: cacheTTL,
	}
}

// WebIDClaims represents WebID-OIDC JWT claims
type WebIDClaims struct {
	WebID string `json:"webid"`
	jwt.RegisteredClaims
}

// ValidateWebIDToken validates a WebID-OIDC JWT token
func (p *WebIDOIDCProvider) ValidateWebIDToken(ctx context.Context, tokenString string) (*domain.WebIDClaims, error) {
	// Parse the JWT token to extract claims without verification first
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &WebIDClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	claims, ok := token.Claims.(*WebIDClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Extract WebID from claims
	webID := claims.WebID
	if webID == "" {
		return nil, fmt.Errorf("WebID claim not found in token")
	}

	// Discover OIDC configuration for the WebID
	oidcConfig, err := p.DiscoverProvider(ctx, webID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
	}

	// Validate the WebID document
	if err := p.ValidateWebIDDocument(ctx, webID); err != nil {
		return nil, fmt.Errorf("WebID document validation failed: %w", err)
	}

	// Get the public key from JWKS endpoint
	publicKey, err := p.getPublicKey(ctx, oidcConfig.JWKSEndpoint, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and verify the JWT token with proper signature verification
	verifiedToken, err := jwt.ParseWithClaims(tokenString, &WebIDClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("JWT signature verification failed: %w", err)
	}

	verifiedClaims, ok := verifiedToken.Claims.(*WebIDClaims)
	if !ok || !verifiedToken.Valid {
		return nil, fmt.Errorf("invalid JWT token claims")
	}

	// Convert to domain WebIDClaims
	webIDClaims := &domain.WebIDClaims{
		Subject:   verifiedClaims.Subject,
		WebID:     verifiedClaims.WebID,
		Issuer:    verifiedClaims.Issuer,
		Audience:  strings.Join(verifiedClaims.Audience, ","),
		ExpiresAt: verifiedClaims.ExpiresAt.Time,
		IssuedAt:  verifiedClaims.IssuedAt.Time,
	}

	// Validate claims
	if !webIDClaims.IsValid() {
		return nil, fmt.Errorf("invalid WebID claims")
	}

	return webIDClaims, nil
}

// DiscoverProvider discovers OIDC configuration for a WebID
func (p *WebIDOIDCProvider) DiscoverProvider(ctx context.Context, webID string) (*domain.OIDCConfiguration, error) {
	// Check cache first
	if cached, exists := p.cache[webID]; exists && time.Now().Before(cached.expiresAt) {
		return cached.config, nil
	}

	// Parse WebID to extract the issuer
	parsedURL, err := url.Parse(webID)
	if err != nil {
		return nil, fmt.Errorf("invalid WebID URL: %w", err)
	}

	// Construct OIDC discovery URL
	issuer := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	discoveryURL := issuer + "/.well-known/openid_configuration"

	// Fetch OIDC configuration
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery failed with status %d", resp.StatusCode)
	}

	var config domain.OIDCConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC configuration: %w", err)
	}

	// Validate configuration
	if !config.IsValid() {
		return nil, fmt.Errorf("invalid OIDC configuration")
	}

	// Cache the configuration
	p.cache[webID] = &cachedOIDCConfig{
		config:    &config,
		expiresAt: time.Now().Add(p.cacheTTL),
	}

	return &config, nil
}

// ValidateWebIDDocument validates a WebID document
func (p *WebIDOIDCProvider) ValidateWebIDDocument(ctx context.Context, webID string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", webID, nil)
	if err != nil {
		return fmt.Errorf("failed to create WebID request: %w", err)
	}

	// Accept RDF formats
	req.Header.Set("Accept", "text/turtle, application/rdf+xml, application/ld+json, */*")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch WebID document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WebID document fetch failed with status %d", resp.StatusCode)
	}

	// Read the document content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read WebID document: %w", err)
	}

	// Basic validation - check if document contains the WebID
	contentStr := string(content)
	if !strings.Contains(contentStr, webID) {
		return fmt.Errorf("WebID document does not contain the WebID")
	}

	// Additional validation could be added here to parse RDF and validate structure
	// For now, we do basic content validation

	return nil
}

// getPublicKey retrieves the public key from JWKS endpoint
func (p *WebIDOIDCProvider) getPublicKey(ctx context.Context, jwksURL, tokenString string) (*rsa.PublicKey, error) {
	// Parse token to get key ID from header
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token header: %w", err)
	}

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("key ID not found in JWT header")
	}

	// Fetch JWKS
	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS fetch failed with status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []map[string]interface{} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Find the key with matching key ID
	for _, key := range jwks.Keys {
		if kid, ok := key["kid"].(string); ok && kid == keyID {
			return parseRSAPublicKeyFromJWK(key)
		}
	}

	return nil, fmt.Errorf("public key not found for key ID: %s", keyID)
}

// parseRSAPublicKeyFromJWK parses an RSA public key from JWK format
func parseRSAPublicKeyFromJWK(jwk map[string]interface{}) (*rsa.PublicKey, error) {
	// Check key type
	kty, ok := jwk["kty"].(string)
	if !ok || kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", kty)
	}

	// Get modulus (n) and exponent (e)
	nStr, ok := jwk["n"].(string)
	if !ok {
		return nil, fmt.Errorf("modulus (n) not found in JWK")
	}

	eStr, ok := jwk["e"].(string)
	if !ok {
		return nil, fmt.Errorf("exponent (e) not found in JWK")
	}

	// Decode base64url encoded values
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// getStringClaim safely extracts a string claim from JWT claims
func getStringClaim(claims map[string]interface{}, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}
