package domain

import (
	"testing"
)

func TestAuthenticationMethod_String(t *testing.T) {
	tests := []struct {
		name   string
		method AuthenticationMethod
		want   string
	}{
		{
			name:   "WebID-OIDC method",
			method: MethodWebIDOIDC,
			want:   "webid-oidc",
		},
		{
			name:   "Password method",
			method: MethodPassword,
			want:   "password",
		},
		{
			name:   "OAuth method",
			method: MethodOAuth,
			want:   "oauth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.String(); got != tt.want {
				t.Errorf("AuthenticationMethod.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthenticationMethod_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		method AuthenticationMethod
		want   bool
	}{
		{
			name:   "valid WebID-OIDC",
			method: MethodWebIDOIDC,
			want:   true,
		},
		{
			name:   "valid Password",
			method: MethodPassword,
			want:   true,
		},
		{
			name:   "valid OAuth",
			method: MethodOAuth,
			want:   true,
		},
		{
			name:   "invalid method",
			method: AuthenticationMethod("invalid"),
			want:   false,
		},
		{
			name:   "empty method",
			method: AuthenticationMethod(""),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method.IsValid(); got != tt.want {
				t.Errorf("AuthenticationMethod.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAuthenticationMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		want    AuthenticationMethod
		wantErr bool
	}{
		{
			name:    "valid webid-oidc",
			method:  "webid-oidc",
			want:    MethodWebIDOIDC,
			wantErr: false,
		},
		{
			name:    "valid password",
			method:  "password",
			want:    MethodPassword,
			wantErr: false,
		},
		{
			name:    "valid oauth",
			method:  "oauth",
			want:    MethodOAuth,
			wantErr: false,
		},
		{
			name:    "case insensitive WebID-OIDC",
			method:  "WEBID-OIDC",
			want:    MethodWebIDOIDC,
			wantErr: false,
		},
		{
			name:    "case insensitive Password",
			method:  "PASSWORD",
			want:    MethodPassword,
			wantErr: false,
		},
		{
			name:    "invalid method",
			method:  "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty method",
			method:  "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAuthenticationMethod(tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAuthenticationMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAuthenticationMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExternalIdentity_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		identity *ExternalIdentity
		want     bool
	}{
		{
			name: "valid identity",
			identity: &ExternalIdentity{
				UserID:     "user-123",
				Provider:   "google",
				ExternalID: "google-456",
			},
			want: true,
		},
		{
			name: "empty UserID",
			identity: &ExternalIdentity{
				UserID:     "",
				Provider:   "google",
				ExternalID: "google-456",
			},
			want: false,
		},
		{
			name: "empty Provider",
			identity: &ExternalIdentity{
				UserID:     "user-123",
				Provider:   "",
				ExternalID: "google-456",
			},
			want: false,
		},
		{
			name: "empty ExternalID",
			identity: &ExternalIdentity{
				UserID:     "user-123",
				Provider:   "google",
				ExternalID: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.IsValid(); got != tt.want {
				t.Errorf("ExternalIdentity.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExternalProfile_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		profile *ExternalProfile
		want    bool
	}{
		{
			name: "valid profile",
			profile: &ExternalProfile{
				ID:       "google-123",
				Email:    "user@example.com",
				Name:     "John Doe",
				Provider: "google",
			},
			want: true,
		},
		{
			name: "empty ID",
			profile: &ExternalProfile{
				ID:       "",
				Email:    "user@example.com",
				Name:     "John Doe",
				Provider: "google",
			},
			want: false,
		},
		{
			name: "empty Email",
			profile: &ExternalProfile{
				ID:       "google-123",
				Email:    "",
				Name:     "John Doe",
				Provider: "google",
			},
			want: false,
		},
		{
			name: "empty Provider",
			profile: &ExternalProfile{
				ID:       "google-123",
				Email:    "user@example.com",
				Name:     "John Doe",
				Provider: "",
			},
			want: false,
		},
		{
			name: "missing Name and Username (should still be valid)",
			profile: &ExternalProfile{
				ID:       "google-123",
				Email:    "user@example.com",
				Provider: "google",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsValid(); got != tt.want {
				t.Errorf("ExternalProfile.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
