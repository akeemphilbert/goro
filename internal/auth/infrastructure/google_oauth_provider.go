package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
)

// GoogleOAuthProvider implements OAuth2 authentication for Google
type GoogleOAuthProvider struct {
	config     *OAuthConfig
	httpClient *http.Client
}

// SetHTTPClient sets the HTTP client for testing purposes
func (p *GoogleOAuthProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// NewGoogleOAuthProvider creates a new Google OAuth provider
func NewGoogleOAuthProvider(config *OAuthConfig) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderName returns the provider name
func (p *GoogleOAuthProvider) GetProviderName() string {
	return "google"
}

// GetAuthURL returns the Google OAuth authorization URL
func (p *GoogleOAuthProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")

	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

// ExchangeCode exchanges authorization code for access token
func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthToken, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", p.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &domain.OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		Scope:        tokenResp.Scope,
	}, nil
}

// GetUserProfile retrieves user profile from Google
func (p *GoogleOAuthProvider) GetUserProfile(ctx context.Context, token *domain.OAuthToken) (*domain.ExternalProfile, error) {
	if !token.IsValid() {
		return nil, fmt.Errorf("invalid or expired token")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("profile request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var profile struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile response: %w", err)
	}

	// Only accept verified email addresses
	if !profile.VerifiedEmail {
		return nil, fmt.Errorf("email address is not verified")
	}

	return &domain.ExternalProfile{
		ID:       profile.ID,
		Email:    profile.Email,
		Name:     profile.Name,
		Username: profile.Email, // Google doesn't have usernames, use email
		Provider: p.GetProviderName(),
	}, nil
}
