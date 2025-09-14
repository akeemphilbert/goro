package infrastructure_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
	"github.com/akeemphilbert/goro/internal/auth/infrastructure"
)

func TestGoogleOAuthProvider_Integration(t *testing.T) {
	// Mock Google OAuth server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			// Mock token exchange
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"access_token":  "mock-access-token",
				"token_type":    "Bearer",
				"refresh_token": "mock-refresh-token",
				"expires_in":    3600,
				"scope":         "openid profile email",
			}
			json.NewEncoder(w).Encode(response)

		case "/userinfo":
			// Mock user profile
			auth := r.Header.Get("Authorization")
			if auth != "Bearer mock-access-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":             "123456789",
				"email":          "test@example.com",
				"verified_email": true,
				"name":           "Test User",
				"given_name":     "Test",
				"family_name":    "User",
				"picture":        "https://example.com/avatar.jpg",
				"locale":         "en",
			}
			json.NewEncoder(w).Encode(response)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create provider with mock server URLs
	config := &infrastructure.OAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"openid", "profile", "email"},
	}

	provider := infrastructure.NewGoogleOAuthProvider(config)

	// Override the HTTP client to use mock server
	provider.SetHTTPClient(&http.Client{
		Transport: &mockTransport{
			tokenURL:   mockServer.URL + "/token",
			profileURL: mockServer.URL + "/userinfo",
		},
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	t.Run("GetAuthURL", func(t *testing.T) {
		state := "test-state"
		authURL := provider.GetAuthURL(state)

		parsedURL, err := url.Parse(authURL)
		if err != nil {
			t.Fatalf("Failed to parse auth URL: %v", err)
		}

		if parsedURL.Host != "accounts.google.com" {
			t.Errorf("Expected Google auth host, got %s", parsedURL.Host)
		}

		params := parsedURL.Query()
		if params.Get("client_id") != config.ClientID {
			t.Errorf("Expected client_id %s, got %s", config.ClientID, params.Get("client_id"))
		}

		if params.Get("state") != state {
			t.Errorf("Expected state %s, got %s", state, params.Get("state"))
		}
	})

	t.Run("ExchangeCode", func(t *testing.T) {
		token, err := provider.ExchangeCode(ctx, "test-code")
		if err != nil {
			t.Fatalf("ExchangeCode failed: %v", err)
		}

		if token.AccessToken != "mock-access-token" {
			t.Errorf("Expected access token 'mock-access-token', got %s", token.AccessToken)
		}

		if token.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got %s", token.TokenType)
		}

		if !token.IsValid() {
			t.Error("Token should be valid")
		}
	})

	t.Run("GetUserProfile", func(t *testing.T) {
		token := &domain.OAuthToken{
			AccessToken: "mock-access-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(time.Hour),
		}

		profile, err := provider.GetUserProfile(ctx, token)
		if err != nil {
			t.Fatalf("GetUserProfile failed: %v", err)
		}

		if profile.ID != "123456789" {
			t.Errorf("Expected ID '123456789', got %s", profile.ID)
		}

		if profile.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", profile.Email)
		}

		if profile.Name != "Test User" {
			t.Errorf("Expected name 'Test User', got %s", profile.Name)
		}

		if profile.Provider != "google" {
			t.Errorf("Expected provider 'google', got %s", profile.Provider)
		}
	})
}

func TestGitHubOAuthProvider_Integration(t *testing.T) {
	// Mock GitHub OAuth server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			// Mock token exchange
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"access_token": "mock-github-token",
				"token_type":   "bearer",
				"scope":        "user:email",
			}
			json.NewEncoder(w).Encode(response)

		case "/user":
			// Mock user profile
			auth := r.Header.Get("Authorization")
			if auth != "Bearer mock-github-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"id":         12345,
				"login":      "testuser",
				"name":       "Test User",
				"email":      "test@example.com",
				"avatar_url": "https://github.com/avatar.jpg",
			}
			json.NewEncoder(w).Encode(response)

		case "/user/emails":
			// Mock user emails
			auth := r.Header.Get("Authorization")
			if auth != "Bearer mock-github-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := []map[string]interface{}{
				{
					"email":    "test@example.com",
					"primary":  true,
					"verified": true,
				},
			}
			json.NewEncoder(w).Encode(response)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create provider with mock server URLs
	config := &infrastructure.OAuthConfig{
		ClientID:     "test-github-client-id",
		ClientSecret: "test-github-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"user:email"},
	}

	provider := infrastructure.NewGitHubOAuthProvider(config)

	// Override the HTTP client to use mock server
	provider.SetHTTPClient(&http.Client{
		Transport: &mockGitHubTransport{
			tokenURL:  mockServer.URL + "/login/oauth/access_token",
			userURL:   mockServer.URL + "/user",
			emailsURL: mockServer.URL + "/user/emails",
		},
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	t.Run("GetAuthURL", func(t *testing.T) {
		state := "test-state"
		authURL := provider.GetAuthURL(state)

		parsedURL, err := url.Parse(authURL)
		if err != nil {
			t.Fatalf("Failed to parse auth URL: %v", err)
		}

		if parsedURL.Host != "github.com" {
			t.Errorf("Expected GitHub auth host, got %s", parsedURL.Host)
		}

		params := parsedURL.Query()
		if params.Get("client_id") != config.ClientID {
			t.Errorf("Expected client_id %s, got %s", config.ClientID, params.Get("client_id"))
		}

		if params.Get("state") != state {
			t.Errorf("Expected state %s, got %s", state, params.Get("state"))
		}
	})

	t.Run("ExchangeCode", func(t *testing.T) {
		token, err := provider.ExchangeCode(ctx, "test-code")
		if err != nil {
			t.Fatalf("ExchangeCode failed: %v", err)
		}

		if token.AccessToken != "mock-github-token" {
			t.Errorf("Expected access token 'mock-github-token', got %s", token.AccessToken)
		}

		if token.TokenType != "bearer" {
			t.Errorf("Expected token type 'bearer', got %s", token.TokenType)
		}

		if !token.IsValid() {
			t.Error("Token should be valid")
		}
	})

	t.Run("GetUserProfile", func(t *testing.T) {
		token := &domain.OAuthToken{
			AccessToken: "mock-github-token",
			TokenType:   "bearer",
			ExpiresAt:   time.Now().Add(time.Hour),
		}

		profile, err := provider.GetUserProfile(ctx, token)
		if err != nil {
			t.Fatalf("GetUserProfile failed: %v", err)
		}

		if profile.ID != "12345" {
			t.Errorf("Expected ID '12345', got %s", profile.ID)
		}

		if profile.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", profile.Email)
		}

		if profile.Name != "Test User" {
			t.Errorf("Expected name 'Test User', got %s", profile.Name)
		}

		if profile.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got %s", profile.Username)
		}

		if profile.Provider != "github" {
			t.Errorf("Expected provider 'github', got %s", profile.Provider)
		}
	})
}

// mockTransport redirects requests to mock server URLs
type mockTransport struct {
	tokenURL   string
	profileURL string
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch {
	case strings.Contains(req.URL.String(), "oauth2.googleapis.com/token"):
		req.URL, _ = url.Parse(t.tokenURL)
	case strings.Contains(req.URL.String(), "www.googleapis.com/oauth2/v2/userinfo"):
		req.URL, _ = url.Parse(t.profileURL)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// mockGitHubTransport redirects GitHub requests to mock server URLs
type mockGitHubTransport struct {
	tokenURL  string
	userURL   string
	emailsURL string
}

func (t *mockGitHubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch {
	case strings.Contains(req.URL.String(), "github.com/login/oauth/access_token"):
		req.URL, _ = url.Parse(t.tokenURL)
	case strings.Contains(req.URL.String(), "api.github.com/user/emails"):
		req.URL, _ = url.Parse(t.emailsURL)
	case strings.Contains(req.URL.String(), "api.github.com/user"):
		req.URL, _ = url.Parse(t.userURL)
	}
	return http.DefaultTransport.RoundTrip(req)
}
