package domain

import (
	"testing"
	"time"
)

func TestSession_IsExpired(t *testing.T) {
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
		{
			name:      "expires now",
			expiresAt: time.Now(),
			want:      true, // Should be expired since time has passed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				ExpiresAt: tt.expiresAt,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("Session.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_UpdateActivity(t *testing.T) {
	s := &Session{
		LastActivity: time.Now().Add(-time.Hour),
	}

	oldActivity := s.LastActivity
	s.UpdateActivity()

	if !s.LastActivity.After(oldActivity) {
		t.Error("UpdateActivity() should update LastActivity to a more recent time")
	}
}

func TestSession_HasAccountContext(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		want    bool
	}{
		{
			name: "has account context",
			session: &Session{
				AccountID: "account-123",
				RoleID:    "role-admin",
			},
			want: true,
		},
		{
			name: "missing account ID",
			session: &Session{
				AccountID: "",
				RoleID:    "role-admin",
			},
			want: false,
		},
		{
			name: "missing role ID",
			session: &Session{
				AccountID: "account-123",
				RoleID:    "",
			},
			want: false,
		},
		{
			name: "missing both",
			session: &Session{
				AccountID: "",
				RoleID:    "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.HasAccountContext(); got != tt.want {
				t.Errorf("Session.HasAccountContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_SetAccountContext(t *testing.T) {
	s := &Session{}

	s.SetAccountContext("account-123", "role-admin")

	if s.AccountID != "account-123" {
		t.Errorf("Expected AccountID to be 'account-123', got '%s'", s.AccountID)
	}
	if s.RoleID != "role-admin" {
		t.Errorf("Expected RoleID to be 'role-admin', got '%s'", s.RoleID)
	}
	if !s.HasAccountContext() {
		t.Error("Expected session to have account context after setting it")
	}
}

func TestSession_ClearAccountContext(t *testing.T) {
	s := &Session{
		AccountID: "account-123",
		RoleID:    "role-admin",
	}

	s.ClearAccountContext()

	if s.AccountID != "" {
		t.Errorf("Expected AccountID to be empty, got '%s'", s.AccountID)
	}
	if s.RoleID != "" {
		t.Errorf("Expected RoleID to be empty, got '%s'", s.RoleID)
	}
	if s.HasAccountContext() {
		t.Error("Expected session to not have account context after clearing it")
	}
}

func TestSession_IsValid(t *testing.T) {
	validSession := &Session{
		ID:           "session-123",
		UserID:       "user-456",
		WebID:        "https://example.com/user#me",
		AccountID:    "account-789",
		RoleID:       "role-admin",
		TokenHash:    "hash123",
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	tests := []struct {
		name    string
		session *Session
		want    bool
	}{
		{
			name:    "valid session",
			session: validSession,
			want:    true,
		},
		{
			name: "empty ID",
			session: &Session{
				ID:        "",
				UserID:    "user-456",
				WebID:     "https://example.com/user#me",
				TokenHash: "hash123",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			want: false,
		},
		{
			name: "empty UserID",
			session: &Session{
				ID:        "session-123",
				UserID:    "",
				WebID:     "https://example.com/user#me",
				TokenHash: "hash123",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			want: false,
		},
		{
			name: "empty WebID",
			session: &Session{
				ID:        "session-123",
				UserID:    "user-456",
				WebID:     "",
				TokenHash: "hash123",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			want: false,
		},
		{
			name: "empty TokenHash",
			session: &Session{
				ID:        "session-123",
				UserID:    "user-456",
				WebID:     "https://example.com/user#me",
				TokenHash: "",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			want: false,
		},
		{
			name: "expired session",
			session: &Session{
				ID:        "session-123",
				UserID:    "user-456",
				WebID:     "https://example.com/user#me",
				TokenHash: "hash123",
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.session.IsValid(); got != tt.want {
				t.Errorf("Session.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
