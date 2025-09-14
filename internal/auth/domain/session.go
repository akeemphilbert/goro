package domain

import (
	"time"
)

// Session represents an authenticated user session
type Session struct {
	ID           string
	UserID       string
	WebID        string
	AccountID    string // Current account context for authorization
	RoleID       string // User's role in the current account
	TokenHash    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	LastActivity time.Time
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

// HasAccountContext checks if the session has account context information
func (s *Session) HasAccountContext() bool {
	return s.AccountID != "" && s.RoleID != ""
}

// SetAccountContext sets the account and role context for the session
func (s *Session) SetAccountContext(accountID, roleID string) {
	s.AccountID = accountID
	s.RoleID = roleID
}

// ClearAccountContext removes account context from the session
func (s *Session) ClearAccountContext() {
	s.AccountID = ""
	s.RoleID = ""
}

// IsValid checks if the session is valid (not expired and has required fields)
func (s *Session) IsValid() bool {
	if s.ID == "" || s.UserID == "" || s.WebID == "" || s.TokenHash == "" {
		return false
	}
	// AccountID and RoleID are optional - user might not be in any account context
	return !s.IsExpired()
}
