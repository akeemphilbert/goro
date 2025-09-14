package domain

import (
	"context"
)

// SessionRepository interface for session management
type SessionRepository interface {
	// Save creates or updates a session
	Save(ctx context.Context, session *Session) error

	// FindByID retrieves a session by its ID
	FindByID(ctx context.Context, id string) (*Session, error)

	// FindByUserID retrieves all sessions for a user
	FindByUserID(ctx context.Context, userID string) ([]*Session, error)

	// Delete removes a session by ID
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired removes all expired sessions
	DeleteExpired(ctx context.Context) error

	// UpdateActivity updates the last activity timestamp for a session
	UpdateActivity(ctx context.Context, sessionID string) error

	// FindByAccountID retrieves all sessions for users in a specific account
	FindByAccountID(ctx context.Context, accountID string) ([]*Session, error)

	// FindByUserIDAndAccountID retrieves sessions for a user in a specific account
	FindByUserIDAndAccountID(ctx context.Context, userID, accountID string) ([]*Session, error)
}

// PasswordRepository interface for credential storage
type PasswordRepository interface {
	// Save creates or updates password credentials for a user
	Save(ctx context.Context, credential *PasswordCredential) error

	// FindByUserID retrieves password credentials for a user
	FindByUserID(ctx context.Context, userID string) (*PasswordCredential, error)

	// Update modifies existing password credentials
	Update(ctx context.Context, credential *PasswordCredential) error

	// Delete removes password credentials for a user
	Delete(ctx context.Context, userID string) error

	// Exists checks if password credentials exist for a user
	Exists(ctx context.Context, userID string) (bool, error)
}

// PasswordResetRepository interface for reset token management
type PasswordResetRepository interface {
	// Save creates a new password reset token
	Save(ctx context.Context, token *PasswordResetToken) error

	// FindByToken retrieves a password reset token by its token value
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)

	// FindByUserID retrieves all password reset tokens for a user
	FindByUserID(ctx context.Context, userID string) ([]*PasswordResetToken, error)

	// MarkAsUsed marks a password reset token as used
	MarkAsUsed(ctx context.Context, token string) error

	// Delete removes a password reset token
	Delete(ctx context.Context, token string) error

	// DeleteByUserID removes all password reset tokens for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired removes all expired password reset tokens
	DeleteExpired(ctx context.Context) error
}

// ExternalIdentityRepository interface for OAuth identity linking
type ExternalIdentityRepository interface {
	// LinkIdentity creates a link between a user and an external identity
	LinkIdentity(ctx context.Context, userID, provider, externalID string) error

	// FindByExternalID finds a user ID by external provider and ID
	FindByExternalID(ctx context.Context, provider, externalID string) (string, error)

	// GetLinkedIdentities retrieves all external identities linked to a user
	GetLinkedIdentities(ctx context.Context, userID string) ([]*ExternalIdentity, error)

	// UnlinkIdentity removes the link between a user and an external identity
	UnlinkIdentity(ctx context.Context, userID, provider, externalID string) error

	// UnlinkAllIdentities removes all external identity links for a user
	UnlinkAllIdentities(ctx context.Context, userID string) error

	// IsLinked checks if an external identity is already linked to any user
	IsLinked(ctx context.Context, provider, externalID string) (bool, error)

	// GetByProvider retrieves all external identities for a specific provider
	GetByProvider(ctx context.Context, provider string) ([]*ExternalIdentity, error)
}
