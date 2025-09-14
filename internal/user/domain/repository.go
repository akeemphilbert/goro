package domain

import (
	"context"
	"fmt"
)

// UserFilter represents filtering options for user queries
type UserFilter struct {
	Status       UserStatus `json:"status"`
	EmailPattern string     `json:"email_pattern"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

// Validate validates the user filter
func (f UserFilter) Validate() error {
	if f.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	if f.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// Read-only repository interfaces (for queries)

// UserRepository provides read-only access to users
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByWebID(ctx context.Context, webid string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, filter UserFilter) ([]*User, error)
	Exists(ctx context.Context, id string) (bool, error)
}

// AccountRepository provides read-only access to accounts
type AccountRepository interface {
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByOwner(ctx context.Context, ownerID string) ([]*Account, error)
}

// AccountMemberRepository provides read-only access to account members (projection)
type AccountMemberRepository interface {
	GetByID(ctx context.Context, id string) (*AccountMember, error)
	GetByAccountAndUser(ctx context.Context, accountID, userID string) (*AccountMember, error)
	ListByAccount(ctx context.Context, accountID string) ([]*AccountMember, error)
	ListByUser(ctx context.Context, userID string) ([]*AccountMember, error)
}

// InvitationRepository provides read-only access to invitations
type InvitationRepository interface {
	GetByID(ctx context.Context, id string) (*Invitation, error)
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	ListByAccount(ctx context.Context, accountID string) ([]*Invitation, error)
	ListByEmail(ctx context.Context, email string) ([]*Invitation, error)
}

// RoleRepository provides read-only access to roles
type RoleRepository interface {
	GetByID(ctx context.Context, id string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	GetSystemRoles(ctx context.Context) ([]*Role, error)
}

// Write repository interfaces (for persistence via events)

// UserWriteRepository handles user persistence operations
type UserWriteRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// AccountWriteRepository handles account persistence operations
type AccountWriteRepository interface {
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
}

// AccountMemberWriteRepository handles account member persistence operations
type AccountMemberWriteRepository interface {
	Create(ctx context.Context, member *AccountMember) error
	Update(ctx context.Context, member *AccountMember) error
	Delete(ctx context.Context, id string) error
}

// InvitationWriteRepository handles invitation persistence operations
type InvitationWriteRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id string) error
}

// RoleWriteRepository handles role persistence operations
type RoleWriteRepository interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	SeedSystemRoles(ctx context.Context) error
}

// Service interfaces

// WebIDGenerator defines the interface for WebID generation and validation
type WebIDGenerator interface {
	GenerateWebID(ctx context.Context, userID, email, userName string) (string, error)
	GenerateWebIDDocument(ctx context.Context, webID, email, userName string) (string, error)
	ValidateWebID(ctx context.Context, webID string) error
	IsUniqueWebID(ctx context.Context, webID string) (bool, error)
	GenerateAlternativeWebID(ctx context.Context, baseWebID string) (string, error)
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	StoreUserProfile(ctx context.Context, userID string, profile *User) error
	LoadUserProfile(ctx context.Context, userID string) (*User, error)
	DeleteUserProfile(ctx context.Context, userID string) error
	StoreWebIDDocument(ctx context.Context, userID, webIDDoc string) error
	LoadWebIDDocument(ctx context.Context, userID string) (string, error)
	DeleteWebIDDocument(ctx context.Context, userID string) error
	StoreAccountData(ctx context.Context, accountID string, account *Account) error
	LoadAccountData(ctx context.Context, accountID string) (*Account, error)
	DeleteAccountData(ctx context.Context, accountID string) error
}
