package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// AccountMember represents a user's membership in an account (projection)
type AccountMember struct {
	*pericarpdomain.BasicEntity
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	InvitedBy string    `json:"invited_by"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAccountMember creates a new account member with validation
func NewAccountMember(ctx context.Context, id string, account *Account, user *User, role *Role, invitedByUser *User, joinedAt time.Time) (*AccountMember, error) {
	log.Context(ctx).Debugf("[NewAccountMember] Starting account member creation: id=%s, joinedAt=%s",
		id, joinedAt.Format(time.RFC3339))

	now := time.Now()
	log.Context(ctx).Debug("[NewAccountMember] Extracting IDs from entity references")

	// Extract IDs from entities
	accountID := ""
	if account != nil {
		accountID = account.ID()
		log.Context(ctx).Debugf("[NewAccountMember] Account ID extracted: %s", accountID)
	} else {
		log.Context(ctx).Debug("[NewAccountMember] No account provided")
	}

	userID := ""
	if user != nil {
		userID = user.ID()
		log.Context(ctx).Debugf("[NewAccountMember] User ID extracted: %s", userID)
	} else {
		log.Context(ctx).Debug("[NewAccountMember] No user provided")
	}

	roleID := ""
	if role != nil {
		roleID = role.ID()
		log.Context(ctx).Debugf("[NewAccountMember] Role ID extracted: %s", roleID)
	} else {
		log.Context(ctx).Debug("[NewAccountMember] No role provided")
	}

	invitedBy := ""
	if invitedByUser != nil {
		invitedBy = invitedByUser.ID()
		log.Context(ctx).Debugf("[NewAccountMember] InvitedBy ID extracted: %s", invitedBy)
	} else {
		log.Context(ctx).Debug("[NewAccountMember] No invited by user provided")
	}

	log.Context(ctx).Debug("[NewAccountMember] Creating account member entity")
	member := &AccountMember{
		BasicEntity: pericarpdomain.NewEntity(id),
		AccountID:   accountID,
		UserID:      userID,
		RoleID:      roleID,
		InvitedBy:   invitedBy,
		JoinedAt:    joinedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	log.Context(ctx).Debug("[NewAccountMember] Account member entity created, starting validation")

	if strings.TrimSpace(id) == "" {
		log.Context(ctx).Debug("[NewAccountMember] Validation failed: member ID is required")
		member.AddError(fmt.Errorf("member ID is required"))
		return member, fmt.Errorf("member ID is required")
	}
	log.Context(ctx).Debugf("[NewAccountMember] ID validation passed: %s", id)

	if account == nil {
		log.Context(ctx).Debug("[NewAccountMember] Validation failed: account is required")
		member.AddError(fmt.Errorf("account is required"))
		return member, fmt.Errorf("account is required")
	}
	log.Context(ctx).Debugf("[NewAccountMember] Account validation passed: %s", account.ID())

	if user == nil {
		log.Context(ctx).Debug("[NewAccountMember] Validation failed: user is required")
		member.AddError(fmt.Errorf("user is required"))
		return member, fmt.Errorf("user is required")
	}
	log.Context(ctx).Debugf("[NewAccountMember] User validation passed: %s", user.ID())

	if role == nil {
		log.Context(ctx).Debug("[NewAccountMember] Validation failed: role is required")
		member.AddError(fmt.Errorf("role is required"))
		return member, fmt.Errorf("role is required")
	}
	log.Context(ctx).Debugf("[NewAccountMember] Role validation passed: %s", role.ID())

	log.Context(ctx).Debug("[NewAccountMember] All validations passed, creating event")

	log.Context(ctx).Debug("[NewAccountMember] Creating account member created event")
	// Emit account member created event
	event := NewAccountMemberCreatedEvent(member, account, user, role, invitedByUser)
	member.AddEvent(event)
	log.Context(ctx).Debug("[NewAccountMember] Account member created event added to entity")

	log.Context(ctx).Infof("Account member created: id=%s, account=%s, user=%s, role=%s", id, account.ID(), user.ID(), role.ID())
	log.Context(ctx).Debug("[NewAccountMember] Account member creation completed successfully")
	return member, nil
}

// UpdateRole updates the member's role
func (m *AccountMember) UpdateRole(ctx context.Context, account *Account, user *User, oldRole, newRole *Role) error {
	log.Context(ctx).Debugf("[UpdateRole] Starting role update for member: memberID=%s, userID=%s",
		m.ID(), user.ID())
	log.Context(ctx).Debugf("[UpdateRole] Role change: oldRole=%s, newRole=%s", oldRole.ID(), newRole.ID())

	log.Context(ctx).Debug("[UpdateRole] Starting validation")
	if account == nil {
		log.Context(ctx).Debug("[UpdateRole] Validation failed: account is required")
		err := fmt.Errorf("account is required")
		m.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateRole] Account validation passed: %s", account.ID())

	if user == nil {
		log.Context(ctx).Debug("[UpdateRole] Validation failed: user is required")
		err := fmt.Errorf("user is required")
		m.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateRole] User validation passed: %s", user.ID())

	if oldRole == nil {
		log.Context(ctx).Debug("[UpdateRole] Validation failed: old role is required")
		err := fmt.Errorf("old role is required")
		m.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateRole] OldRole validation passed: %s", oldRole.ID())

	if newRole == nil {
		log.Context(ctx).Debug("[UpdateRole] Validation failed: new role is required")
		err := fmt.Errorf("new role is required")
		m.AddError(err)
		return err
	}
	log.Context(ctx).Debugf("[UpdateRole] NewRole validation passed: %s", newRole.ID())

	log.Context(ctx).Debug("[UpdateRole] All validations passed, updating member role")
	m.RoleID = newRole.ID()
	m.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[UpdateRole] Member role and timestamp updated")

	log.Context(ctx).Debug("[UpdateRole] Creating account member updated event")
	// Emit account member updated event
	event := NewAccountMemberUpdatedEvent(m, account, user, oldRole, newRole)
	m.AddEvent(event)
	log.Context(ctx).Debug("[UpdateRole] Account member updated event created and added to entity")

	log.Context(ctx).Infof("Account member role updated: memberID=%s, account=%s, user=%s, oldRole=%s, newRole=%s",
		m.ID(), account.ID(), user.ID(), oldRole.ID(), newRole.ID())
	log.Context(ctx).Debug("[UpdateRole] Member role update completed successfully")
	return nil
}

// IsOwner checks if the member is an owner
func (m *AccountMember) IsOwner() bool {
	return m.RoleID == "owner"
}

// IsAdmin checks if the member is an admin (owner or admin role)
func (m *AccountMember) IsAdmin() bool {
	return m.RoleID == "owner" || m.RoleID == "admin"
}

// CanManageMembers checks if the member can manage other members
func (m *AccountMember) CanManageMembers() bool {
	return m.IsAdmin()
}
