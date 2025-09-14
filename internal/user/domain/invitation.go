package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

// Invitation represents an invitation to join an account
type Invitation struct {
	*pericarpdomain.BasicEntity
	AccountID string           `json:"account_id"`
	Email     string           `json:"email"`
	RoleID    string           `json:"role_id"`
	Token     string           `json:"token"`
	InvitedBy string           `json:"invited_by"`
	Status    InvitationStatus `json:"status"`
	ExpiresAt time.Time        `json:"expires_at"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewInvitation creates a new invitation with validation
func NewInvitation(ctx context.Context, id, token string, account *Account, email string, role *Role, invitedByUser *User, expiresAt time.Time) (*Invitation, error) {
	log.Context(ctx).Debugf("[NewInvitation] Starting invitation creation: id=%s, email=%s, expiresAt=%s", id, email, expiresAt.Format(time.RFC3339))

	now := time.Now()
	log.Context(ctx).Debug("[NewInvitation] Extracting IDs from entity references")

	// Extract IDs from entities
	accountID := ""
	if account != nil {
		accountID = account.ID()
		log.Context(ctx).Debugf("[NewInvitation] Account ID extracted: %s", accountID)
	} else {
		log.Context(ctx).Debug("[NewInvitation] No account provided")
	}

	roleID := ""
	if role != nil {
		roleID = role.ID()
		log.Context(ctx).Debugf("[NewInvitation] Role ID extracted: %s", roleID)
	} else {
		log.Context(ctx).Debug("[NewInvitation] No role provided")
	}

	invitedBy := ""
	if invitedByUser != nil {
		invitedBy = invitedByUser.ID()
		log.Context(ctx).Debugf("[NewInvitation] InvitedBy ID extracted: %s", invitedBy)
	} else {
		log.Context(ctx).Debug("[NewInvitation] No invited by user provided")
	}

	log.Context(ctx).Debug("[NewInvitation] Creating invitation entity with pending status")
	invitation := &Invitation{
		BasicEntity: pericarpdomain.NewEntity(id),
		AccountID:   accountID,
		Email:       email,
		RoleID:      roleID,
		Token:       token,
		InvitedBy:   invitedBy,
		Status:      InvitationStatusPending,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	log.Context(ctx).Debug("[NewInvitation] Invitation entity created, starting validation")

	if strings.TrimSpace(id) == "" {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: invitation ID is required")
		invitation.AddError(fmt.Errorf("invitation ID is required"))
		return invitation, fmt.Errorf("invitation ID is required")
	}
	log.Context(ctx).Debugf("[NewInvitation] ID validation passed: %s", id)

	if account == nil {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: account is required")
		invitation.AddError(fmt.Errorf("account is required"))
		return invitation, fmt.Errorf("account is required")
	}
	log.Context(ctx).Debugf("[NewInvitation] Account validation passed: %s", account.ID())

	log.Context(ctx).Debug("[NewInvitation] Validating email format")
	if !isValidEmail(email) {
		log.Context(ctx).Debugf("[NewInvitation] Email validation failed: %s", email)
		invitation.AddError(fmt.Errorf("invalid email format"))
		return invitation, fmt.Errorf("invalid email format")
	}
	log.Context(ctx).Debugf("[NewInvitation] Email validation passed: %s", email)

	if role == nil {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: role is required")
		invitation.AddError(fmt.Errorf("role is required"))
		return invitation, fmt.Errorf("role is required")
	}
	log.Context(ctx).Debugf("[NewInvitation] Role validation passed: %s", role.ID())

	if strings.TrimSpace(token) == "" {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: token is required")
		invitation.AddError(fmt.Errorf("token is required"))
		return invitation, fmt.Errorf("token is required")
	}
	log.Context(ctx).Debug("[NewInvitation] Token validation passed")

	if invitedByUser == nil {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: invited by user is required")
		invitation.AddError(fmt.Errorf("invited by user is required"))
		return invitation, fmt.Errorf("invited by user is required")
	}
	log.Context(ctx).Debugf("[NewInvitation] InvitedBy validation passed: %s", invitedByUser.ID())

	log.Context(ctx).Debug("[NewInvitation] Validating expiration time")
	if expiresAt.Before(time.Now()) {
		log.Context(ctx).Debug("[NewInvitation] Validation failed: expiration time must be in the future")
		invitation.AddError(fmt.Errorf("expiration time must be in the future"))
		return invitation, fmt.Errorf("expiration time must be in the future")
	}
	log.Context(ctx).Debugf("[NewInvitation] Expiration validation passed: %s", expiresAt.Format(time.RFC3339))

	log.Context(ctx).Debug("[NewInvitation] All validations passed, creating event")

	// Log successful invitation creation
	log.Context(ctx).Infof("Invitation created: id=%s, email=%s, account=%s, role=%s, invitedBy=%s", id, email, account.ID(), role.ID(), invitedByUser.ID())

	log.Context(ctx).Debug("[NewInvitation] Creating invitation created event")
	// Emit invitation created event
	event := NewInvitationCreatedEvent(invitation, account, role, invitedByUser)
	invitation.AddEvent(event)
	log.Context(ctx).Debug("[NewInvitation] Invitation created event added to entity")

	log.Context(ctx).Debug("[NewInvitation] Invitation creation completed successfully")
	return invitation, nil
}

// Accept accepts the invitation
func (i *Invitation) Accept(ctx context.Context, user *User, account *Account, role *Role, accountMember *AccountMember) error {
	log.Context(ctx).Debugf("[Accept] Starting invitation acceptance: id=%s, userID=%s, currentStatus=%s",
		i.ID(), user.ID(), i.Status)

	log.Context(ctx).Debugf("[Accept] Checking if invitation can be accepted: status=%s, expired=%t",
		i.Status, i.IsExpired())
	if !i.CanAccept() {
		log.Context(ctx).Debug("[Accept] Invitation cannot be accepted")
		log.Context(ctx).Warnf("Attempt to accept invalid invitation: id=%s, status=%s, expired=%t", i.ID(), i.Status, i.IsExpired())
		err := fmt.Errorf("invitation cannot be accepted")
		i.AddError(err)
		return err
	}
	log.Context(ctx).Debug("[Accept] Invitation acceptance validation passed")

	log.Context(ctx).Debug("[Accept] Setting invitation status to accepted")
	i.Status = InvitationStatusAccepted
	i.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[Accept] Invitation status and timestamp updated")

	// Log successful invitation acceptance
	log.Context(ctx).Infof("Invitation accepted: id=%s, user=%s, account=%s", i.ID(), user.ID(), account.ID())

	log.Context(ctx).Debug("[Accept] Creating invitation accepted event")
	// Emit invitation accepted event
	event := NewInvitationAcceptedEvent(i, account, user, role, accountMember)
	i.AddEvent(event)
	log.Context(ctx).Debug("[Accept] Invitation accepted event created and added to entity")

	log.Context(ctx).Debug("[Accept] Invitation acceptance completed successfully")
	return nil
}

// Revoke revokes the invitation
func (i *Invitation) Revoke(ctx context.Context, account *Account, revokedBy *User, reason string) error {
	log.Context(ctx).Debugf("[Revoke] Starting invitation revocation: id=%s, revokedBy=%s, reason=%s, currentStatus=%s",
		i.ID(), revokedBy.ID(), reason, i.Status)

	log.Context(ctx).Debug("[Revoke] Checking if invitation can be revoked")
	if i.Status == InvitationStatusAccepted {
		log.Context(ctx).Debug("[Revoke] Revocation failed: invitation is already accepted")
		log.Context(ctx).Warnf("Attempt to revoke accepted invitation: id=%s", i.ID())
		err := fmt.Errorf("cannot revoke accepted invitation")
		i.AddError(err)
		return err
	}
	log.Context(ctx).Debug("[Revoke] Invitation revocation validation passed")

	log.Context(ctx).Debug("[Revoke] Setting invitation status to revoked")
	i.Status = InvitationStatusRevoked
	i.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[Revoke] Invitation status and timestamp updated")

	log.Context(ctx).Debug("[Revoke] Creating invitation revoked event")
	// Emit invitation revoked event
	event := NewInvitationRevokedEvent(i, account, revokedBy, reason)
	i.AddEvent(event)
	log.Context(ctx).Debug("[Revoke] Invitation revoked event created and added to entity")

	log.Context(ctx).Infof("Invitation revoked: id=%s, revokedBy=%s, reason=%s", i.ID(), revokedBy.ID(), reason)
	log.Context(ctx).Debug("[Revoke] Invitation revocation completed successfully")
	return nil
}

// IsExpired checks if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// CanAccept checks if the invitation can be accepted
func (i *Invitation) CanAccept() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}
