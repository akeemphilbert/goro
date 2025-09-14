package domain_test

import (
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/auth/domain"
)

func TestOAuthToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token *domain.OAuthToken
		want  bool
	}{
		{
			name: "valid token",
			token: &domain.OAuthToken{
				AccessToken: "valid-token",
				TokenType:   "Bearer",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			want: true,
		},
		{
			name: "expired token",
			token: &domain.OAuthToken{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				ExpiresAt:   time.Now().Add(-time.Hour),
			},
			want: false,
		},
		{
			name: "empty access token",
			token: &domain.OAuthToken{
				AccessToken: "",
				TokenType:   "Bearer",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("OAuthToken.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOAuthToken_IsExpired(t *testing.T) {
	tests := []struct {
		name  string
		token *domain.OAuthToken
		want  bool
	}{
		{
			name: "not expired",
			token: &domain.OAuthToken{
				ExpiresAt: time.Now().Add(time.Hour),
			},
			want: false,
		},
		{
			name: "expired",
			token: &domain.OAuthToken{
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.want {
				t.Errorf("OAuthToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebIDClaims_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		claims *domain.WebIDClaims
		want   bool
	}{
		{
			name: "valid claims",
			claims: &domain.WebIDClaims{
				Subject:   "user123",
				WebID:     "https://example.com/profile#me",
				Issuer:    "https://provider.com",
				Audience:  "client-id",
				ExpiresAt: time.Now().Add(time.Hour),
				IssuedAt:  time.Now(),
			},
			want: true,
		},
		{
			name: "expired claims",
			claims: &domain.WebIDClaims{
				Subject:   "user123",
				WebID:     "https://example.com/profile#me",
				Issuer:    "https://provider.com",
				Audience:  "client-id",
				ExpiresAt: time.Now().Add(-time.Hour),
				IssuedAt:  time.Now().Add(-2 * time.Hour),
			},
			want: false,
		},
		{
			name: "missing subject",
			claims: &domain.WebIDClaims{
				Subject:   "",
				WebID:     "https://example.com/profile#me",
				Issuer:    "https://provider.com",
				Audience:  "client-id",
				ExpiresAt: time.Now().Add(time.Hour),
				IssuedAt:  time.Now(),
			},
			want: false,
		},
		{
			name: "missing WebID",
			claims: &domain.WebIDClaims{
				Subject:   "user123",
				WebID:     "",
				Issuer:    "https://provider.com",
				Audience:  "client-id",
				ExpiresAt: time.Now().Add(time.Hour),
				IssuedAt:  time.Now(),
			},
			want: false,
		},
		{
			name: "missing issuer",
			claims: &domain.WebIDClaims{
				Subject:   "user123",
				WebID:     "https://example.com/profile#me",
				Issuer:    "",
				Audience:  "client-id",
				ExpiresAt: time.Now().Add(time.Hour),
				IssuedAt:  time.Now(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.claims.IsValid(); got != tt.want {
				t.Errorf("WebIDClaims.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOIDCConfiguration_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		config *domain.OIDCConfiguration
		want   bool
	}{
		{
			name: "valid configuration",
			config: &domain.OIDCConfiguration{
				Issuer:                 "https://provider.com",
				AuthorizationEndpoint:  "https://provider.com/auth",
				TokenEndpoint:          "https://provider.com/token",
				JWKSEndpoint:           "https://provider.com/jwks",
				SupportedScopes:        []string{"openid", "profile"},
				SupportedResponseTypes: []string{"code"},
			},
			want: true,
		},
		{
			name: "missing issuer",
			config: &domain.OIDCConfiguration{
				Issuer:                "",
				AuthorizationEndpoint: "https://provider.com/auth",
				TokenEndpoint:         "https://provider.com/token",
				JWKSEndpoint:          "https://provider.com/jwks",
			},
			want: false,
		},
		{
			name: "missing authorization endpoint",
			config: &domain.OIDCConfiguration{
				Issuer:                "https://provider.com",
				AuthorizationEndpoint: "",
				TokenEndpoint:         "https://provider.com/token",
				JWKSEndpoint:          "https://provider.com/jwks",
			},
			want: false,
		},
		{
			name: "missing token endpoint",
			config: &domain.OIDCConfiguration{
				Issuer:                "https://provider.com",
				AuthorizationEndpoint: "https://provider.com/auth",
				TokenEndpoint:         "",
				JWKSEndpoint:          "https://provider.com/jwks",
			},
			want: false,
		},
		{
			name: "missing JWKS endpoint",
			config: &domain.OIDCConfiguration{
				Issuer:                "https://provider.com",
				AuthorizationEndpoint: "https://provider.com/auth",
				TokenEndpoint:         "https://provider.com/token",
				JWKSEndpoint:          "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsValid(); got != tt.want {
				t.Errorf("OIDCConfiguration.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
