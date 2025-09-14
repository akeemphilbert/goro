package domain

import (
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
)

// Re-export pericarp types for convenience
type EntityEvent = pericarpdomain.EntityEvent
type EventDispatcher = pericarpdomain.EventDispatcher
type EventHandler = pericarpdomain.EventHandler

// Event types for user operations
const (
	EventTypeUserCreated        = "user.created"
	EventTypeUserRegistered     = "user.registered"
	EventTypeUserProfileUpdated = "user.profile.updated"
	EventTypeUserSuspended      = "user.suspended"
	EventTypeUserActivated      = "user.activated"
	EventTypeUserDeleted        = "user.deleted"
	EventTypeWebIDGenerated     = "user.webid.generated"
)

// Event types for account operations
const (
	EventTypeAccountCreated           = "account.created"
	EventTypeAccountUpdated           = "account.updated"
	EventTypeAccountSettingsUpdated   = "account.settings.updated"
	EventTypeAccountMemberAdded       = "account.member.added"
	EventTypeAccountMemberRemoved     = "account.member.removed"
	EventTypeAccountMemberRoleUpdated = "account.member.role.updated"
)

// Event types for invitation operations
const (
	EventTypeInvitationCreated  = "invitation.created"
	EventTypeInvitationAccepted = "invitation.accepted"
	EventTypeInvitationRevoked  = "invitation.revoked"
	EventTypeInvitationExpired  = "invitation.expired"
	EventTypeMemberInvited      = "member.invited"
)

// Event types for membership operations
const (
	EventTypeMemberAdded       = "member.added"
	EventTypeMemberRemoved     = "member.removed"
	EventTypeMemberRoleUpdated = "member.role.updated"
)

// Event types for role operations
const (
	EventTypeRoleCreated           = "role.created"
	EventTypeRoleUpdated           = "role.updated"
	EventTypeRolePermissionAdded   = "role.permission.added"
	EventTypeRolePermissionRemoved = "role.permission.removed"
)

// User event constructors
func NewUserCreatedEvent(user *User, webID string, status string) *EntityEvent {
	data := UserCreatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		WebID:         webID,
		Status:        status,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserCreated, user.ID(), "", "", data)
}

func NewUserRegisteredEvent(user *User) *EntityEvent {
	data := UserRegisteredEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserRegistered, user.ID(), "", "", data)
}

func NewUserProfileUpdatedEvent(user *User, oldProfile, newProfile UserProfile) *EntityEvent {
	data := UserProfileUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		OldProfile:    oldProfile,
		NewProfile:    newProfile,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserProfileUpdated, user.ID(), "", "", data)
}

func NewUserSuspendedEvent(user *User, reason string) *EntityEvent {
	data := UserSuspendedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		Reason:        reason,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserSuspended, user.ID(), "", "", data)
}

func NewUserActivatedEvent(user *User) *EntityEvent {
	data := UserActivatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserActivated, user.ID(), "", "", data)
}

func NewUserDeletedEvent(user *User, reason string) *EntityEvent {
	data := UserDeletedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		Reason:        reason,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeUserDeleted, user.ID(), "", "", data)
}

func NewWebIDGeneratedEvent(user *User, webID string) *EntityEvent {
	data := WebIDGeneratedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		UserID:        user.ID(),
		User:          user,
		WebID:         webID,
	}
	return pericarpdomain.NewEntityEvent("user", EventTypeWebIDGenerated, user.ID(), "", "", data)
}

// Account event constructors
func NewAccountCreatedEvent(account *Account, owner *User) *EntityEvent {
	data := AccountCreatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
		Owner:         owner,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountCreated, account.ID(), "", "", data)
}

func NewAccountUpdatedEvent(account *Account) *EntityEvent {
	data := AccountUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountUpdated, account.ID(), "", "", data)
}

func NewAccountSettingsUpdatedEvent(account *Account, oldSettings, newSettings AccountSettings) *EntityEvent {
	data := AccountSettingsUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
		OldSettings:   oldSettings,
		NewSettings:   newSettings,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountSettingsUpdated, account.ID(), "", "", data)
}

func NewAccountMemberAddedEvent(account *Account, user *User, role *Role, invitation *Invitation, accountMember *AccountMember, addedBy *User) *EntityEvent {
	data := AccountMemberAddedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
		User:          user,
		Role:          role,
		Invitation:    invitation,
		AccountMember: accountMember,
		AddedBy:       addedBy,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountMemberAdded, account.ID(), "", "", data)
}

func NewAccountMemberRemovedEvent(account *Account, user *User, accountMember *AccountMember, removedBy *User, reason string) *EntityEvent {
	data := AccountMemberRemovedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
		User:          user,
		AccountMember: accountMember,
		RemovedBy:     removedBy,
		Reason:        reason,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountMemberRemoved, account.ID(), "", "", data)
}

func NewAccountMemberRoleUpdatedEvent(account *Account, user *User, accountMember *AccountMember, oldRole, newRole *Role, updatedBy *User) *EntityEvent {
	data := AccountMemberRoleUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Account:       account,
		User:          user,
		AccountMember: accountMember,
		OldRole:       oldRole,
		NewRole:       newRole,
		UpdatedBy:     updatedBy,
	}
	return pericarpdomain.NewEntityEvent("account", EventTypeAccountMemberRoleUpdated, account.ID(), "", "", data)
}

// Invitation event constructors
func NewInvitationCreatedEvent(invitation *Invitation, account *Account, role *Role, invitedBy *User) *EntityEvent {
	data := InvitationCreatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Invitation:    invitation,
		Account:       account,
		Role:          role,
		InvitedBy:     invitedBy,
	}
	return pericarpdomain.NewEntityEvent("invitation", EventTypeInvitationCreated, invitation.ID(), "", "", data)
}

func NewInvitationAcceptedEvent(invitation *Invitation, account *Account, user *User, role *Role, accountMember *AccountMember) *EntityEvent {
	data := InvitationAcceptedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Invitation:    invitation,
		Account:       account,
		User:          user,
		Role:          role,
		AccountMember: accountMember,
	}
	return pericarpdomain.NewEntityEvent("invitation", EventTypeInvitationAccepted, invitation.ID(), "", "", data)
}

func NewInvitationRevokedEvent(invitation *Invitation, account *Account, revokedBy *User, reason string) *EntityEvent {
	data := InvitationRevokedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Invitation:    invitation,
		Account:       account,
		RevokedBy:     revokedBy,
		Reason:        reason,
	}
	return pericarpdomain.NewEntityEvent("invitation", EventTypeInvitationRevoked, invitation.ID(), "", "", data)
}

func NewInvitationExpiredEvent(invitation *Invitation, account *Account) *EntityEvent {
	data := InvitationExpiredEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Invitation:    invitation,
		Account:       account,
	}
	return pericarpdomain.NewEntityEvent("invitation", EventTypeInvitationExpired, invitation.ID(), "", "", data)
}

// Role event constructors
func NewRoleCreatedEvent(role *Role, permissions []Permission) *EntityEvent {
	data := RoleCreatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Role:          role,
		Permissions:   permissions,
	}
	return pericarpdomain.NewEntityEvent("role", EventTypeRoleCreated, role.ID(), "", "", data)
}

func NewRoleUpdatedEvent(role *Role, oldPermissions, newPermissions []Permission) *EntityEvent {
	data := RoleUpdatedEventData{
		BaseEventData:  BaseEventData{OccurredAt: time.Now()},
		Role:           role,
		OldPermissions: oldPermissions,
		NewPermissions: newPermissions,
	}
	return pericarpdomain.NewEntityEvent("role", EventTypeRoleUpdated, role.ID(), "", "", data)
}

func NewRolePermissionAddedEvent(role *Role, permission Permission) *EntityEvent {
	data := RolePermissionAddedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Role:          role,
		Permission:    permission,
	}
	return pericarpdomain.NewEntityEvent("role", EventTypeRolePermissionAdded, role.ID(), "", "", data)
}

func NewRolePermissionRemovedEvent(role *Role, permission Permission) *EntityEvent {
	data := RolePermissionRemovedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Role:          role,
		Permission:    permission,
	}
	return pericarpdomain.NewEntityEvent("role", EventTypeRolePermissionRemoved, role.ID(), "", "", data)
}

// Account member event constructors
func NewAccountMemberCreatedEvent(accountMember *AccountMember, account *Account, user *User, role *Role, invitedBy *User) *EntityEvent {
	data := AccountMemberCreatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		AccountMember: accountMember,
		Account:       account,
		User:          user,
		Role:          role,
		InvitedBy:     invitedBy,
	}
	return pericarpdomain.NewEntityEvent("account_member", "account_member.created", accountMember.ID(), "", "", data)
}

func NewAccountMemberUpdatedEvent(accountMember *AccountMember, account *Account, user *User, oldRole, newRole *Role) *EntityEvent {
	data := AccountMemberUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		AccountMember: accountMember,
		Account:       account,
		User:          user,
		OldRole:       oldRole,
		NewRole:       newRole,
	}
	return pericarpdomain.NewEntityEvent("account_member", "account_member.updated", accountMember.ID(), "", "", data)
}

// Member invitation and lifecycle event constructors
func NewMemberInvitedEvent(invitation *Invitation, account *Account, role *Role, invitedBy *User, email string) *EntityEvent {
	data := MemberInvitedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		Invitation:    invitation,
		Account:       account,
		Role:          role,
		InvitedBy:     invitedBy,
		Email:         email,
	}
	return pericarpdomain.NewEntityEvent("invitation", EventTypeMemberInvited, invitation.ID(), "", "", data)
}

func NewMemberAddedEvent(accountMember *AccountMember, account *Account, user *User, role *Role, joinedVia string) *EntityEvent {
	data := MemberAddedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		AccountMember: accountMember,
		Account:       account,
		User:          user,
		Role:          role,
		JoinedVia:     joinedVia,
	}
	return pericarpdomain.NewEntityEvent("account_member", EventTypeMemberAdded, accountMember.ID(), "", "", data)
}

func NewMemberRemovedEvent(accountMember *AccountMember, account *Account, user *User, removedBy *User, reason string) *EntityEvent {
	data := MemberRemovedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		AccountMember: accountMember,
		Account:       account,
		User:          user,
		RemovedBy:     removedBy,
		Reason:        reason,
	}
	return pericarpdomain.NewEntityEvent("account_member", EventTypeMemberRemoved, accountMember.ID(), "", "", data)
}

func NewMemberRoleUpdatedEvent(accountMember *AccountMember, account *Account, user *User, oldRole, newRole *Role, updatedBy *User) *EntityEvent {
	data := MemberRoleUpdatedEventData{
		BaseEventData: BaseEventData{OccurredAt: time.Now()},
		AccountMember: accountMember,
		Account:       account,
		User:          user,
		OldRole:       oldRole,
		NewRole:       newRole,
		UpdatedBy:     updatedBy,
	}
	return pericarpdomain.NewEntityEvent("account_member", EventTypeMemberRoleUpdated, accountMember.ID(), "", "", data)
}
