package infrastructure

import (
	"fmt"
	"os"
)

// OAuthConfig represents OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// IsValid checks if the OAuth configuration is complete
func (c *OAuthConfig) IsValid() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURL != ""
}

// GoogleOAuthConfig creates OAuth configuration for Google
func GoogleOAuthConfig() (*OAuthConfig, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")

	if clientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is required")
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for development
	}

	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/auth/oauth/google/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}, nil
}

// GitHubOAuthConfig creates OAuth configuration for GitHub
func GitHubOAuthConfig() (*OAuthConfig, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")

	if clientID == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_SECRET environment variable is required")
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for development
	}

	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/auth/oauth/github/callback",
		Scopes:       []string{"user:email"},
	}, nil
}

// WebIDOIDCConfig represents WebID-OIDC provider configuration
type WebIDOIDCConfig struct {
	Timeout  string
	CacheTTL string
}

// WebIDOIDCConfigFromEnv creates WebID-OIDC configuration from environment variables
func WebIDOIDCConfigFromEnv() *WebIDOIDCConfig {
	timeout := os.Getenv("WEBID_OIDC_TIMEOUT")
	if timeout == "" {
		timeout = "30s"
	}

	cacheTTL := os.Getenv("WEBID_OIDC_CACHE_TTL")
	if cacheTTL == "" {
		cacheTTL = "1h"
	}

	return &WebIDOIDCConfig{
		Timeout:  timeout,
		CacheTTL: cacheTTL,
	}
}
