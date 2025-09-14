package infrastructure_test

import (
	"os"
	"testing"

	"github.com/akeemphilbert/goro/internal/auth/infrastructure"
)

func TestOAuthConfig_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		config *infrastructure.OAuthConfig
		want   bool
	}{
		{
			name: "valid config",
			config: &infrastructure.OAuthConfig{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "http://localhost:8080/callback",
				Scopes:       []string{"openid", "profile"},
			},
			want: true,
		},
		{
			name: "missing client ID",
			config: &infrastructure.OAuthConfig{
				ClientID:     "",
				ClientSecret: "client-secret",
				RedirectURL:  "http://localhost:8080/callback",
			},
			want: false,
		},
		{
			name: "missing client secret",
			config: &infrastructure.OAuthConfig{
				ClientID:     "client-id",
				ClientSecret: "",
				RedirectURL:  "http://localhost:8080/callback",
			},
			want: false,
		},
		{
			name: "missing redirect URL",
			config: &infrastructure.OAuthConfig{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURL:  "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsValid(); got != tt.want {
				t.Errorf("OAuthConfig.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleOAuthConfig(t *testing.T) {
	// Save original env vars
	originalClientID := os.Getenv("GOOGLE_CLIENT_ID")
	originalClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	originalBaseURL := os.Getenv("BASE_URL")

	// Clean up after test
	defer func() {
		os.Setenv("GOOGLE_CLIENT_ID", originalClientID)
		os.Setenv("GOOGLE_CLIENT_SECRET", originalClientSecret)
		os.Setenv("BASE_URL", originalBaseURL)
	}()

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		baseURL      string
		wantErr      bool
		wantRedirect string
	}{
		{
			name:         "valid config",
			clientID:     "google-client-id",
			clientSecret: "google-client-secret",
			baseURL:      "https://example.com",
			wantErr:      false,
			wantRedirect: "https://example.com/auth/oauth/google/callback",
		},
		{
			name:         "valid config with default base URL",
			clientID:     "google-client-id",
			clientSecret: "google-client-secret",
			baseURL:      "",
			wantErr:      false,
			wantRedirect: "http://localhost:8080/auth/oauth/google/callback",
		},
		{
			name:         "missing client ID",
			clientID:     "",
			clientSecret: "google-client-secret",
			baseURL:      "https://example.com",
			wantErr:      true,
		},
		{
			name:         "missing client secret",
			clientID:     "google-client-id",
			clientSecret: "",
			baseURL:      "https://example.com",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GOOGLE_CLIENT_ID", tt.clientID)
			os.Setenv("GOOGLE_CLIENT_SECRET", tt.clientSecret)
			os.Setenv("BASE_URL", tt.baseURL)

			config, err := infrastructure.GoogleOAuthConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GoogleOAuthConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GoogleOAuthConfig() unexpected error: %v", err)
				return
			}

			if config.ClientID != tt.clientID {
				t.Errorf("GoogleOAuthConfig() ClientID = %v, want %v", config.ClientID, tt.clientID)
			}

			if config.ClientSecret != tt.clientSecret {
				t.Errorf("GoogleOAuthConfig() ClientSecret = %v, want %v", config.ClientSecret, tt.clientSecret)
			}

			if config.RedirectURL != tt.wantRedirect {
				t.Errorf("GoogleOAuthConfig() RedirectURL = %v, want %v", config.RedirectURL, tt.wantRedirect)
			}

			expectedScopes := []string{"openid", "profile", "email"}
			if len(config.Scopes) != len(expectedScopes) {
				t.Errorf("GoogleOAuthConfig() Scopes length = %v, want %v", len(config.Scopes), len(expectedScopes))
			}
		})
	}
}

func TestGitHubOAuthConfig(t *testing.T) {
	// Save original env vars
	originalClientID := os.Getenv("GITHUB_CLIENT_ID")
	originalClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	originalBaseURL := os.Getenv("BASE_URL")

	// Clean up after test
	defer func() {
		os.Setenv("GITHUB_CLIENT_ID", originalClientID)
		os.Setenv("GITHUB_CLIENT_SECRET", originalClientSecret)
		os.Setenv("BASE_URL", originalBaseURL)
	}()

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		baseURL      string
		wantErr      bool
		wantRedirect string
	}{
		{
			name:         "valid config",
			clientID:     "github-client-id",
			clientSecret: "github-client-secret",
			baseURL:      "https://example.com",
			wantErr:      false,
			wantRedirect: "https://example.com/auth/oauth/github/callback",
		},
		{
			name:         "valid config with default base URL",
			clientID:     "github-client-id",
			clientSecret: "github-client-secret",
			baseURL:      "",
			wantErr:      false,
			wantRedirect: "http://localhost:8080/auth/oauth/github/callback",
		},
		{
			name:         "missing client ID",
			clientID:     "",
			clientSecret: "github-client-secret",
			baseURL:      "https://example.com",
			wantErr:      true,
		},
		{
			name:         "missing client secret",
			clientID:     "github-client-id",
			clientSecret: "",
			baseURL:      "https://example.com",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GITHUB_CLIENT_ID", tt.clientID)
			os.Setenv("GITHUB_CLIENT_SECRET", tt.clientSecret)
			os.Setenv("BASE_URL", tt.baseURL)

			config, err := infrastructure.GitHubOAuthConfig()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GitHubOAuthConfig() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GitHubOAuthConfig() unexpected error: %v", err)
				return
			}

			if config.ClientID != tt.clientID {
				t.Errorf("GitHubOAuthConfig() ClientID = %v, want %v", config.ClientID, tt.clientID)
			}

			if config.ClientSecret != tt.clientSecret {
				t.Errorf("GitHubOAuthConfig() ClientSecret = %v, want %v", config.ClientSecret, tt.clientSecret)
			}

			if config.RedirectURL != tt.wantRedirect {
				t.Errorf("GitHubOAuthConfig() RedirectURL = %v, want %v", config.RedirectURL, tt.wantRedirect)
			}

			expectedScopes := []string{"user:email"}
			if len(config.Scopes) != len(expectedScopes) {
				t.Errorf("GitHubOAuthConfig() Scopes length = %v, want %v", len(config.Scopes), len(expectedScopes))
			}
		})
	}
}

func TestWebIDOIDCConfigFromEnv(t *testing.T) {
	// Save original env vars
	originalTimeout := os.Getenv("WEBID_OIDC_TIMEOUT")
	originalCacheTTL := os.Getenv("WEBID_OIDC_CACHE_TTL")

	// Clean up after test
	defer func() {
		os.Setenv("WEBID_OIDC_TIMEOUT", originalTimeout)
		os.Setenv("WEBID_OIDC_CACHE_TTL", originalCacheTTL)
	}()

	tests := []struct {
		name         string
		timeout      string
		cacheTTL     string
		wantTimeout  string
		wantCacheTTL string
	}{
		{
			name:         "custom values",
			timeout:      "60s",
			cacheTTL:     "2h",
			wantTimeout:  "60s",
			wantCacheTTL: "2h",
		},
		{
			name:         "default values",
			timeout:      "",
			cacheTTL:     "",
			wantTimeout:  "30s",
			wantCacheTTL: "1h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("WEBID_OIDC_TIMEOUT", tt.timeout)
			os.Setenv("WEBID_OIDC_CACHE_TTL", tt.cacheTTL)

			config := infrastructure.WebIDOIDCConfigFromEnv()

			if config.Timeout != tt.wantTimeout {
				t.Errorf("WebIDOIDCConfigFromEnv() Timeout = %v, want %v", config.Timeout, tt.wantTimeout)
			}

			if config.CacheTTL != tt.wantCacheTTL {
				t.Errorf("WebIDOIDCConfigFromEnv() CacheTTL = %v, want %v", config.CacheTTL, tt.wantCacheTTL)
			}
		})
	}
}
