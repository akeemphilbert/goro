package domain

import (
	"time"
)

// PasswordCredential represents user password credentials
type PasswordCredential struct {
	UserID       string
	PasswordHash string
	Salt         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsValid checks if the password credential has required fields
func (pc *PasswordCredential) IsValid() bool {
	return pc.UserID != "" && pc.PasswordHash != "" && pc.Salt != ""
}

// UpdatePassword updates the password hash and salt with current timestamp
func (pc *PasswordCredential) UpdatePassword(hash, salt string) {
	pc.PasswordHash = hash
	pc.Salt = salt
	pc.UpdatedAt = time.Now()
}

// PasswordResetToken represents password reset tokens
type PasswordResetToken struct {
	Token     string
	UserID    string
	Email     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Used      bool
}

// IsExpired checks if the reset token has expired
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt)
}

// IsValid checks if the reset token is valid (not expired, not used, has required fields)
func (prt *PasswordResetToken) IsValid() bool {
	if prt.Token == "" || prt.UserID == "" || prt.Email == "" {
		return false
	}
	return !prt.IsExpired() && !prt.Used
}

// MarkAsUsed marks the token as used
func (prt *PasswordResetToken) MarkAsUsed() {
	prt.Used = true
}
