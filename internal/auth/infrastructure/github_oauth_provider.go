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

// GitHubOAuthProvider implements OAuth2 authentication for GitHub
type GitHubOAuthProvider struct {
	config     *OAuthConfig
	httpClient *http.Client
}

// SetHTTPClient sets the HTTP client for testing purposes
func (p *GitHubOAuthProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// NewGitHubOAuthProvider creates a new GitHub OAuth provider
func NewGitHubOAuthProvider(config *OAuthConfig) *GitHubOAuthProvider {
	return &GitHubOAuthProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderName returns the provider name
func (p *GitHubOAuthProvider) GetProviderName() string {
	return "github"
}

// GetAuthURL returns the GitHub OAuth authorization URL
func (p *GitHubOAuthProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)
	params.Set("allow_signup", "true")

	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// ExchangeCode exchanges authorization code for access token
func (p *GitHubOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthToken, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
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
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// GitHub tokens don't expire by default, set a reasonable expiry
	expiresAt := time.Now().Add(24 * time.Hour)

	return &domain.OAuthToken{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   expiresAt,
		Scope:       tokenResp.Scope,
	}, nil
}

// GetUserProfile retrieves user profile from GitHub
func (p *GitHubOAuthProvider) GetUserProfile(ctx context.Context, token *domain.OAuthToken) (*domain.ExternalProfile, error) {
	if !token.IsValid() {
		return nil, fmt.Errorf("invalid or expired token")
	}

	// Get user profile
	profile, err := p.getUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Get primary email if not public
	if profile.Email == "" {
		email, err := p.getPrimaryEmail(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to get user email: %w", err)
		}
		profile.Email = email
	}

	return profile, nil
}

// getUserInfo retrieves basic user information from GitHub
func (p *GitHubOAuthProvider) getUserInfo(ctx context.Context, token *domain.OAuthToken) (*domain.ExternalProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var user struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &domain.ExternalProfile{
		ID:       fmt.Sprintf("%d", user.ID),
		Email:    user.Email, // May be empty if private
		Name:     user.Name,
		Username: user.Login,
		Provider: p.GetProviderName(),
	}, nil
}

// getPrimaryEmail retrieves the user's primary email from GitHub
func (p *GitHubOAuthProvider) getPrimaryEmail(ctx context.Context, token *domain.OAuthToken) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get user emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("email request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode email response: %w", err)
	}

	// Find primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// Fallback to any verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email address found")
}
