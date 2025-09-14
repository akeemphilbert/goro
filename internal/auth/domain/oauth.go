package domain

import (
	"context"
	"time"
)

// OAuthProvider interface for OAuth2 authentication providers
type OAuthProvider interface {
	// GetAuthURL returns the OAuth authorization URL with the given state
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for an access token
	ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)

	// GetUserProfile retrieves user profile information using the access token
	GetUserProfile(ctx context.Context, token *OAuthToken) (*ExternalProfile, error)

	// GetProviderName returns the name of the OAuth provider
	GetProviderName() string
}

// OAuthToken represents an OAuth access token with metadata
type OAuthToken struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	ExpiresAt    time.Time
	Scope        string
}

// IsValid checks if the OAuth token is valid and not expired
func (t *OAuthToken) IsValid() bool {
	return t.AccessToken != "" && time.Now().Before(t.ExpiresAt)
}

// IsExpired checks if the OAuth token has expired
func (t *OAuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// WebIDOIDCProvider interface for WebID-OIDC authentication
type WebIDOIDCProvider interface {
	// ValidateWebIDToken validates a WebID-OIDC JWT token
	ValidateWebIDToken(ctx context.Context, token string) (*WebIDClaims, error)

	// DiscoverProvider discovers OIDC configuration for a WebID
	DiscoverProvider(ctx context.Context, webID string) (*OIDCConfiguration, error)

	// ValidateWebIDDocument validates a WebID document
	ValidateWebIDDocument(ctx context.Context, webID string) error
}

// WebIDClaims represents claims from a WebID-OIDC JWT token
type WebIDClaims struct {
	Subject   string    `json:"sub"`
	WebID     string    `json:"webid"`
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// IsValid checks if the WebID claims are valid
func (c *WebIDClaims) IsValid() bool {
	return c.Subject != "" && c.WebID != "" && c.Issuer != "" && time.Now().Before(c.ExpiresAt)
}

// OIDCConfiguration represents OIDC provider configuration
type OIDCConfiguration struct {
	Issuer                 string   `json:"issuer"`
	AuthorizationEndpoint  string   `json:"authorization_endpoint"`
	TokenEndpoint          string   `json:"token_endpoint"`
	JWKSEndpoint           string   `json:"jwks_uri"`
	SupportedScopes        []string `json:"scopes_supported"`
	SupportedResponseTypes []string `json:"response_types_supported"`
}

// IsValid checks if the OIDC configuration is valid
func (c *OIDCConfiguration) IsValid() bool {
	return c.Issuer != "" && c.AuthorizationEndpoint != "" && c.TokenEndpoint != "" && c.JWKSEndpoint != ""
}
