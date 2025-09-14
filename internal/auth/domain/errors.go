package domain

import "errors"

// Authentication domain errors
var (
	// Session errors
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionExpired    = errors.New("session expired")
	ErrSessionInvalid    = errors.New("session invalid")
	ErrSessionTokenEmpty = errors.New("session token is empty")

	// Authentication errors
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrUnsupportedAuthMethod = errors.New("unsupported authentication method")
	ErrWebIDValidationFailed = errors.New("WebID validation failed")
	ErrExternalAuthFailed    = errors.New("external authentication failed")

	// Password errors
	ErrPasswordTooWeak        = errors.New("password does not meet security requirements")
	ErrPasswordHashInvalid    = errors.New("password hash is invalid")
	ErrCurrentPasswordInvalid = errors.New("current password is incorrect")

	// Password reset errors
	ErrPasswordResetNotFound = errors.New("password reset token not found")
	ErrPasswordResetExpired  = errors.New("password reset token expired")
	ErrPasswordResetInvalid  = errors.New("invalid password reset token")
	ErrPasswordResetUsed     = errors.New("password reset token already used")

	// External identity errors
	ErrExternalIdentityNotFound      = errors.New("external identity not found")
	ErrExternalIdentityExists        = errors.New("external identity already exists")
	ErrExternalIdentityAlreadyLinked = errors.New("external identity already linked to another user")
	ErrExternalIdentityInvalid       = errors.New("external identity is invalid")
	ErrExternalProviderUnsupported   = errors.New("external provider not supported")

	// Repository errors
	ErrPasswordCredentialNotFound = errors.New("password credential not found")
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")

	// Validation errors
	ErrUserIDEmpty     = errors.New("user ID is empty")
	ErrWebIDEmpty      = errors.New("WebID is empty")
	ErrEmailEmpty      = errors.New("email is empty")
	ErrTokenEmpty      = errors.New("token is empty")
	ErrProviderEmpty   = errors.New("provider is empty")
	ErrExternalIDEmpty = errors.New("external ID is empty")
)
