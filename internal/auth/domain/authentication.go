package domain

import (
	"fmt"
	"strings"
	"time"
)

// AuthenticationMethod represents supported authentication methods
type AuthenticationMethod string

const (
	MethodWebIDOIDC AuthenticationMethod = "webid-oidc"
	MethodPassword  AuthenticationMethod = "password"
	MethodOAuth     AuthenticationMethod = "oauth"
)

// String returns the string representation of the authentication method
func (am AuthenticationMethod) String() string {
	return string(am)
}

// IsValid checks if the authentication method is supported
func (am AuthenticationMethod) IsValid() bool {
	switch am {
	case MethodWebIDOIDC, MethodPassword, MethodOAuth:
		return true
	default:
		return false
	}
}

// ParseAuthenticationMethod parses a string into an AuthenticationMethod
func ParseAuthenticationMethod(method string) (AuthenticationMethod, error) {
	am := AuthenticationMethod(strings.ToLower(method))
	if !am.IsValid() {
		return "", fmt.Errorf("invalid authentication method: %s", method)
	}
	return am, nil
}

// ExternalIdentity represents a linked external identity (OAuth provider)
type ExternalIdentity struct {
	ID         uint
	UserID     string
	Provider   string
	ExternalID string
	CreatedAt  time.Time
}

// IsValid checks if the external identity has required fields
func (ei *ExternalIdentity) IsValid() bool {
	return ei.UserID != "" && ei.Provider != "" && ei.ExternalID != ""
}

// ExternalProfile represents user profile data from external providers
type ExternalProfile struct {
	ID       string
	Email    string
	Name     string
	Username string
	Provider string
}

// IsValid checks if the external profile has required fields
func (ep *ExternalProfile) IsValid() bool {
	return ep.ID != "" && ep.Provider != "" && ep.Email != ""
}
