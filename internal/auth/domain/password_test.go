package domain

import (
	"testing"
	"time"
)

func TestPasswordCredential_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		credential *PasswordCredential
		want       bool
	}{
		{
			name: "valid credential",
			credential: &PasswordCredential{
				UserID:       "user-123",
				PasswordHash: "hash123",
				Salt:         "salt123",
			},
			want: true,
		},
		{
			name: "empty UserID",
			credential: &PasswordCredential{
				UserID:       "",
				PasswordHash: "hash123",
				Salt:         "salt123",
			},
			want: false,
		},
		{
			name: "empty PasswordHash",
			credential: &PasswordCredential{
				UserID:       "user-123",
				PasswordHash: "",
				Salt:         "salt123",
			},
			want: false,
		},
		{
			name: "empty Salt",
			credential: &PasswordCredential{
				UserID:       "user-123",
				PasswordHash: "hash123",
				Salt:         "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.credential.IsValid(); got != tt.want {
				t.Errorf("PasswordCredential.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordCredential_UpdatePassword(t *testing.T) {
	pc := &PasswordCredential{
		UserID:       "user-123",
		PasswordHash: "oldHash",
		Salt:         "oldSalt",
		UpdatedAt:    time.Now().Add(-time.Hour),
	}

	oldUpdatedAt := pc.UpdatedAt
	newHash := "newHash"
	newSalt := "newSalt"

	pc.UpdatePassword(newHash, newSalt)

	if pc.PasswordHash != newHash {
		t.Errorf("UpdatePassword() PasswordHash = %v, want %v", pc.PasswordHash, newHash)
	}
	if pc.Salt != newSalt {
		t.Errorf("UpdatePassword() Salt = %v, want %v", pc.Salt, newSalt)
	}
	if !pc.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatePassword() should update UpdatedAt to a more recent time")
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prt := &PasswordResetToken{
				ExpiresAt: tt.expiresAt,
			}
			if got := prt.IsExpired(); got != tt.want {
				t.Errorf("PasswordResetToken.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	validToken := &PasswordResetToken{
		Token:     "token123",
		UserID:    "user-456",
		Email:     "user@example.com",
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}

	tests := []struct {
		name  string
		token *PasswordResetToken
		want  bool
	}{
		{
			name:  "valid token",
			token: validToken,
			want:  true,
		},
		{
			name: "empty Token",
			token: &PasswordResetToken{
				Token:     "",
				UserID:    "user-456",
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			want: false,
		},
		{
			name: "empty UserID",
			token: &PasswordResetToken{
				Token:     "token123",
				UserID:    "",
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			want: false,
		},
		{
			name: "empty Email",
			token: &PasswordResetToken{
				Token:     "token123",
				UserID:    "user-456",
				Email:     "",
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			},
			want: false,
		},
		{
			name: "expired token",
			token: &PasswordResetToken{
				Token:     "token123",
				UserID:    "user-456",
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(-time.Hour),
				Used:      false,
			},
			want: false,
		},
		{
			name: "used token",
			token: &PasswordResetToken{
				Token:     "token123",
				UserID:    "user-456",
				Email:     "user@example.com",
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("PasswordResetToken.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_MarkAsUsed(t *testing.T) {
	prt := &PasswordResetToken{
		Used: false,
	}

	prt.MarkAsUsed()

	if !prt.Used {
		t.Error("MarkAsUsed() should set Used to true")
	}
}
