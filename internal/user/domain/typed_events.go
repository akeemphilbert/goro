package domain

import (
	"time"
)

// BaseEventData contains common fields for all events
type BaseEventData struct {
	OccurredAt time.Time `json:"occurred_at"`
}

// UserCreatedEventData represents data for when a user is created
type UserCreatedEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	WebID  string `json:"webid"`
	Status string `json:"status"`
}

// UserRegisteredEventData represents data for when a user registers
type UserRegisteredEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
}

// UserProfileUpdatedEventData represents data for when a user's profile is updated
type UserProfileUpdatedEventData struct {
	BaseEventData
	UserID     string      `json:"user_id"`
	User       *User       `json:"user"`
	OldProfile UserProfile `json:"old_profile"`
	NewProfile UserProfile `json:"new_profile"`
}

// UserSuspendedEventData represents data for when a user is suspended
type UserSuspendedEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	Reason string `json:"reason"`
}

// UserActivatedEventData represents data for when a user is activated
type UserActivatedEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
}

// UserDeletedEventData represents data for when a user is deleted
type UserDeletedEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	Reason string `json:"reason"`
}

// WebIDGeneratedEventData represents data for when a WebID is generated for a user
type WebIDGeneratedEventData struct {
	BaseEventData
	UserID string `json:"user_id"`
	User   *User  `json:"user"`
	WebID  string `json:"webid"`
}

// AccountCreatedEventData represents data for when an account is created
type AccountCreatedEventData struct {
	BaseEventData
	Account *Account `json:"account"`
	Owner   *User    `json:"owner"`
}

// AccountUpdatedEventData represents data for when an account is updated
type AccountUpdatedEventData struct {
	BaseEventData
	Account *Account `json:"account"`
}

// AccountSettingsUpdatedEventData represents data for when account settings are updated
type AccountSettingsUpdatedEventData struct {
	BaseEventData
	Account     *Account        `json:"account"`
	OldSettings AccountSettings `json:"old_settings"`
	NewSettings AccountSettings `json:"new_settings"`
}

// AccountMemberAddedEventData represents data for when a member is added to an account
type AccountMemberAddedEventData struct {
	BaseEventData
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	Role          *Role          `json:"role"`
	Invitation    *Invitation    `json:"invitation"`
	AccountMember *AccountMember `json:"account_member"`
	AddedBy       *User          `json:"added_by"`
}

// AccountMemberRemovedEventData represents data for when a member is removed from an account
type AccountMemberRemovedEventData struct {
	BaseEventData
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	AccountMember *AccountMember `json:"account_member"`
	RemovedBy     *User          `json:"removed_by"`
	Reason        string         `json:"reason"`
}

// AccountMemberRoleUpdatedEventData represents data for when a member's role is updated
type AccountMemberRoleUpdatedEventData struct {
	BaseEventData
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	AccountMember *AccountMember `json:"account_member"`
	OldRole       *Role          `json:"old_role"`
	NewRole       *Role          `json:"new_role"`
	UpdatedBy     *User          `json:"updated_by"`
}

// InvitationCreatedEventData represents data for when an invitation is created
type InvitationCreatedEventData struct {
	BaseEventData
	Invitation *Invitation `json:"invitation"`
	Account    *Account    `json:"account"`
	Role       *Role       `json:"role"`
	InvitedBy  *User       `json:"invited_by"`
}

// InvitationAcceptedEventData represents data for when an invitation is accepted
type InvitationAcceptedEventData struct {
	BaseEventData
	Invitation    *Invitation    `json:"invitation"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	Role          *Role          `json:"role"`
	AccountMember *AccountMember `json:"account_member"`
}

// InvitationRevokedEventData represents data for when an invitation is revoked
type InvitationRevokedEventData struct {
	BaseEventData
	Invitation *Invitation `json:"invitation"`
	Account    *Account    `json:"account"`
	RevokedBy  *User       `json:"revoked_by"`
	Reason     string      `json:"reason"`
}

// InvitationExpiredEventData represents data for when an invitation expires
type InvitationExpiredEventData struct {
	BaseEventData
	Invitation *Invitation `json:"invitation"`
	Account    *Account    `json:"account"`
}

// MemberInvitedEventData represents data for when a member is invited (broader than just invitation created)
type MemberInvitedEventData struct {
	BaseEventData
	Invitation *Invitation `json:"invitation"`
	Account    *Account    `json:"account"`
	Role       *Role       `json:"role"`
	InvitedBy  *User       `json:"invited_by"`
	Email      string      `json:"email"`
}

// RoleCreatedEventData represents data for when a role is created
type RoleCreatedEventData struct {
	BaseEventData
	Role        *Role        `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// RoleUpdatedEventData represents data for when a role is updated
type RoleUpdatedEventData struct {
	BaseEventData
	Role           *Role        `json:"role"`
	OldPermissions []Permission `json:"old_permissions"`
	NewPermissions []Permission `json:"new_permissions"`
}

// RolePermissionAddedEventData represents data for when a permission is added to a role
type RolePermissionAddedEventData struct {
	BaseEventData
	Role       *Role      `json:"role"`
	Permission Permission `json:"permission"`
}

// RolePermissionRemovedEventData represents data for when a permission is removed from a role
type RolePermissionRemovedEventData struct {
	BaseEventData
	Role       *Role      `json:"role"`
	Permission Permission `json:"permission"`
}

// AccountMemberCreatedEventData represents data for when an account member entity is created
type AccountMemberCreatedEventData struct {
	BaseEventData
	AccountMember *AccountMember `json:"account_member"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	Role          *Role          `json:"role"`
	InvitedBy     *User          `json:"invited_by"`
}

// AccountMemberUpdatedEventData represents data for when an account member entity is updated
type AccountMemberUpdatedEventData struct {
	BaseEventData
	AccountMember *AccountMember `json:"account_member"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	OldRole       *Role          `json:"old_role"`
	NewRole       *Role          `json:"new_role"`
}

// MemberAddedEventData represents data for when a member is added (lifecycle event)
type MemberAddedEventData struct {
	BaseEventData
	AccountMember *AccountMember `json:"account_member"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	Role          *Role          `json:"role"`
	JoinedVia     string         `json:"joined_via"` // "invitation", "direct", "transfer"
}

// MemberRemovedEventData represents data for when a member is removed (lifecycle event)
type MemberRemovedEventData struct {
	BaseEventData
	AccountMember *AccountMember `json:"account_member"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	RemovedBy     *User          `json:"removed_by"`
	Reason        string         `json:"reason"`
}

// MemberRoleUpdatedEventData represents data for when a member's role is updated (lifecycle event)
type MemberRoleUpdatedEventData struct {
	BaseEventData
	AccountMember *AccountMember `json:"account_member"`
	Account       *Account       `json:"account"`
	User          *User          `json:"user"`
	OldRole       *Role          `json:"old_role"`
	NewRole       *Role          `json:"new_role"`
	UpdatedBy     *User          `json:"updated_by"`
}
