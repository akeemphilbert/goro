package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// AccountSettings contains account configuration
type AccountSettings struct {
	AllowInvitations bool   `json:"allow_invitations"`
	DefaultRoleID    string `json:"default_role_id"`
	MaxMembers       int    `json:"max_members"`
}

// Validate validates the account settings
func (s AccountSettings) Validate() error {
	if strings.TrimSpace(s.DefaultRoleID) == "" {
		return fmt.Errorf("default role ID is required")
	}
	if s.MaxMembers < 0 {
		return fmt.Errorf("max members cannot be negative")
	}
	return nil
}

// Account represents an account in the system
type Account struct {
	*pericarpdomain.BasicEntity
	OwnerID     string          `json:"owner_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Settings    AccountSettings `json:"settings"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewAccount creates a new account with validation
func NewAccount(ctx context.Context, id string, owner *User, name, description string) (*Account, error) {
	log.Context(ctx).Debugf("[NewAccount] Starting account creation: id=%s, name=%s, description=%s", id, name, description)

	now := time.Now()
	ownerID := ""
	if owner != nil {
		ownerID = owner.ID()
		log.Context(ctx).Debugf("[NewAccount] Owner provided: ownerID=%s", ownerID)
	} else {
		log.Context(ctx).Debug("[NewAccount] No owner provided")
	}

	log.Context(ctx).Debug("[NewAccount] Creating account entity with default settings")

	account := &Account{
		BasicEntity: pericarpdomain.NewEntity(id),
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		Settings: AccountSettings{
			AllowInvitations: true,
			DefaultRoleID:    "member",
			MaxMembers:       100,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	log.Context(ctx).Debugf("[NewAccount] Account entity created, starting validation")

	if strings.TrimSpace(id) == "" {
		log.Context(ctx).Debug("[NewAccount] Validation failed: account ID is required")
		account.AddError(fmt.Errorf("account ID is required"))
		return account, fmt.Errorf("account ID is required")
	}
	log.Context(ctx).Debugf("[NewAccount] ID validation passed: %s", id)

	if owner == nil {
		log.Context(ctx).Debug("[NewAccount] Validation failed: owner is required")
		log.Context(ctx).Error("Account creation failed: owner is required")
		account.AddError(fmt.Errorf("owner is required"))
		return account, fmt.Errorf("owner is required")
	}
	log.Context(ctx).Debugf("[NewAccount] Owner validation passed: %s", owner.ID())

	if strings.TrimSpace(name) == "" {
		log.Context(ctx).Debug("[NewAccount] Validation failed: account name is required")
		account.AddError(fmt.Errorf("account name is required"))
		return account, fmt.Errorf("account name is required")
	}
	log.Context(ctx).Debugf("[NewAccount] Name validation passed: %s", name)

	log.Context(ctx).Debug("[NewAccount] All validations passed, creating event")

	// Emit account created event
	event := NewAccountCreatedEvent(account, owner)
	account.AddEvent(event)
	log.Context(ctx).Debug("[NewAccount] Account created event added to entity")

	// Log successful account creation
	log.Context(ctx).Infof("Account created successfully: id=%s, owner=%s, name=%s", id, owner.ID(), name)
	log.Context(ctx).Debug("[NewAccount] Account creation completed successfully")

	return account, nil
}

// AddMember adds a member to the account (domain method for business logic)
func (a *Account) AddMember(ctx context.Context, user *User, role *Role, invitation *Invitation, accountMember *AccountMember, addedBy *User) error {
	log.Context(ctx).Debugf("[AddMember] Starting member addition to account: account=%s", a.ID())

	if user == nil {
		log.Context(ctx).Debug("[AddMember] Validation failed: user cannot be nil")
		log.Context(ctx).Error("AddMember failed: user cannot be nil")
		err := fmt.Errorf("user cannot be nil")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[AddMember] User validation passed: userID=%s", user.ID())

	if role == nil {
		log.Context(ctx).Debug("[AddMember] Validation failed: role must be specified")
		err := fmt.Errorf("role must be specified")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[AddMember] Role validation passed: roleID=%s", role.ID())

	if invitation == nil {
		log.Context(ctx).Debug("[AddMember] Validation failed: invitation cannot be nil")
		err := fmt.Errorf("invitation cannot be nil")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[AddMember] Invitation validation passed: invitationID=%s", invitation.ID())

	if accountMember == nil {
		log.Context(ctx).Debug("[AddMember] Validation failed: account member cannot be nil")
		err := fmt.Errorf("account member cannot be nil")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[AddMember] AccountMember validation passed: memberID=%s", accountMember.ID())

	if addedBy == nil {
		log.Context(ctx).Debug("[AddMember] Validation failed: added by user cannot be nil")
		err := fmt.Errorf("added by user cannot be nil")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[AddMember] AddedBy validation passed: addedByID=%s", addedBy.ID())

	log.Context(ctx).Debug("[AddMember] All validations passed, processing member addition")

	// Business logic validation would go here
	// For now, this is a placeholder that validates inputs
	a.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[AddMember] Account updated timestamp set")

	// Log successful member addition
	log.Context(ctx).Infof("Member added to account: account=%s, user=%s, role=%s, addedBy=%s", a.ID(), user.ID(), role.ID(), addedBy.ID())

	log.Context(ctx).Debug("[AddMember] Creating member added event")
	// Emit member added event
	event := NewAccountMemberAddedEvent(a, user, role, invitation, accountMember, addedBy)
	a.AddEvent(event)
	log.Context(ctx).Debug("[AddMember] Member added event created and added to entity")

	log.Context(ctx).Debug("[AddMember] Member addition completed successfully")
	return nil
}

// RemoveMember removes a member from the account
func (a *Account) RemoveMember(ctx context.Context, user *User, accountMember *AccountMember, removedBy *User, reason string) error {
	log.Context(ctx).Debugf("[RemoveMember] Starting member removal from account: account=%s, reason=%s", a.ID(), reason)

	if user == nil {
		log.Context(ctx).Debug("[RemoveMember] Validation failed: user is required")
		err := fmt.Errorf("user is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[RemoveMember] User validation passed: userID=%s", user.ID())

	if accountMember == nil {
		log.Context(ctx).Debug("[RemoveMember] Validation failed: account member is required")
		err := fmt.Errorf("account member is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[RemoveMember] AccountMember validation passed: memberID=%s", accountMember.ID())

	if removedBy == nil {
		log.Context(ctx).Debug("[RemoveMember] Validation failed: removed by user is required")
		err := fmt.Errorf("removed by user is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[RemoveMember] RemovedBy validation passed: removedByID=%s", removedBy.ID())

	log.Context(ctx).Debug("[RemoveMember] All validations passed, processing member removal")

	// Business logic for removing member would go here
	a.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[RemoveMember] Account updated timestamp set")

	log.Context(ctx).Debug("[RemoveMember] Creating member removed event")
	// Emit member removed event
	event := NewAccountMemberRemovedEvent(a, user, accountMember, removedBy, reason)
	a.AddEvent(event)
	log.Context(ctx).Debug("[RemoveMember] Member removed event created and added to entity")

	log.Context(ctx).Infof("Member removed from account: account=%s, user=%s, removedBy=%s, reason=%s", a.ID(), user.ID(), removedBy.ID(), reason)
	log.Context(ctx).Debug("[RemoveMember] Member removal completed successfully")
	return nil
}

// UpdateMemberRole updates a member's role in the account
func (a *Account) UpdateMemberRole(ctx context.Context, user *User, accountMember *AccountMember, oldRole, newRole *Role, updatedBy *User) error {
	log.Context(ctx).Debugf("[UpdateMemberRole] Starting member role update in account: account=%s", a.ID())

	if user == nil {
		log.Context(ctx).Debug("[UpdateMemberRole] Validation failed: user is required")
		err := fmt.Errorf("user is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateMemberRole] User validation passed: userID=%s", user.ID())

	if accountMember == nil {
		log.Context(ctx).Debug("[UpdateMemberRole] Validation failed: account member is required")
		err := fmt.Errorf("account member is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateMemberRole] AccountMember validation passed: memberID=%s", accountMember.ID())

	if oldRole == nil {
		log.Context(ctx).Debug("[UpdateMemberRole] Validation failed: old role is required")
		err := fmt.Errorf("old role is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateMemberRole] OldRole validation passed: oldRoleID=%s", oldRole.ID())

	if newRole == nil {
		log.Context(ctx).Debug("[UpdateMemberRole] Validation failed: new role is required")
		err := fmt.Errorf("new role is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateMemberRole] NewRole validation passed: newRoleID=%s", newRole.ID())

	if updatedBy == nil {
		log.Context(ctx).Debug("[UpdateMemberRole] Validation failed: updated by user is required")
		err := fmt.Errorf("updated by user is required")
		a.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateMemberRole] UpdatedBy validation passed: updatedByID=%s", updatedBy.ID())

	log.Context(ctx).Debug("[UpdateMemberRole] All validations passed, processing role update")

	// Business logic for updating member role would go here
	a.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[UpdateMemberRole] Account updated timestamp set")

	log.Context(ctx).Debug("[UpdateMemberRole] Creating member role updated event")
	// Emit member role updated event
	event := NewAccountMemberRoleUpdatedEvent(a, user, accountMember, oldRole, newRole, updatedBy)
	a.AddEvent(event)
	log.Context(ctx).Debug("[UpdateMemberRole] Member role updated event created and added to entity")

	log.Context(ctx).Infof("Member role updated in account: account=%s, user=%s, oldRole=%s, newRole=%s, updatedBy=%s", a.ID(), user.ID(), oldRole.ID(), newRole.ID(), updatedBy.ID())
	log.Context(ctx).Debug("[UpdateMemberRole] Member role update completed successfully")
	return nil
}

// UpdateSettings updates the account settings
func (a *Account) UpdateSettings(ctx context.Context, settings AccountSettings) error {
	log.Context(ctx).Debugf("[UpdateSettings] Starting account settings update: account=%s", a.ID())
	log.Context(ctx).Debugf("[UpdateSettings] New settings - AllowInvitations=%t, DefaultRoleID=%s, MaxMembers=%d",
		settings.AllowInvitations, settings.DefaultRoleID, settings.MaxMembers)

	log.Context(ctx).Debug("[UpdateSettings] Validating new settings")
	if err := settings.Validate(); err != nil {
		log.Context(ctx).Debugf("[UpdateSettings] Settings validation failed: %v", err)
		validationErr := fmt.Errorf("invalid settings: %w", err)
		a.AddError(validationErr)
		return validationErr
	}
	log.Context(ctx).Debug("[UpdateSettings] Settings validation passed")

	oldSettings := a.Settings
	log.Context(ctx).Debugf("[UpdateSettings] Current settings - AllowInvitations=%t, DefaultRoleID=%s, MaxMembers=%d",
		oldSettings.AllowInvitations, oldSettings.DefaultRoleID, oldSettings.MaxMembers)

	log.Context(ctx).Debug("[UpdateSettings] Applying new settings to account")
	a.Settings = settings
	a.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[UpdateSettings] Account settings and timestamp updated")

	log.Context(ctx).Debug("[UpdateSettings] Creating settings updated event")
	// Emit settings updated event
	event := NewAccountSettingsUpdatedEvent(a, oldSettings, settings)
	a.AddEvent(event)
	log.Context(ctx).Debug("[UpdateSettings] Settings updated event created and added to entity")

	log.Context(ctx).Infof("Account settings updated: account=%s, allowInvitations=%t, defaultRole=%s, maxMembers=%d",
		a.ID(), settings.AllowInvitations, settings.DefaultRoleID, settings.MaxMembers)
	log.Context(ctx).Debug("[UpdateSettings] Account settings update completed successfully")
	return nil
}
