package domain

import (
	"encoding/json"
	"fmt"
)

// EventDataUnmarshaler provides helper functions to unmarshal EntityEvent payloads
// back to strongly typed event data structures
type EventDataUnmarshaler struct{}

// NewEventDataUnmarshaler creates a new event data unmarshaler
func NewEventDataUnmarshaler() *EventDataUnmarshaler {
	return &EventDataUnmarshaler{}
}

// UnmarshalUserCreatedEvent unmarshals a user created event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserCreatedEvent(event *EntityEvent) (*UserCreatedEventData, error) {
	if event.EventType() != EventTypeUserCreated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeUserCreated, event.EventType())
	}

	var data UserCreatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user created event: %w", err)
	}

	return &data, nil
}

// UnmarshalUserRegisteredEvent unmarshals a user registered event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserRegisteredEvent(event *EntityEvent) (*UserRegisteredEventData, error) {
	if event.EventType() != "user.registered" {
		return nil, fmt.Errorf("expected event type %s, got %s", "user.registered", event.EventType())
	}

	var data UserRegisteredEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user registered event: %w", err)
	}

	return &data, nil
}

// UnmarshalUserProfileUpdatedEvent unmarshals a user profile updated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserProfileUpdatedEvent(event *EntityEvent) (*UserProfileUpdatedEventData, error) {
	if event.EventType() != "user.profile.updated" {
		return nil, fmt.Errorf("expected event type %s, got %s", "user.profile.updated", event.EventType())
	}

	var data UserProfileUpdatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile updated event: %w", err)
	}

	return &data, nil
}

// UnmarshalUserSuspendedEvent unmarshals a user suspended event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserSuspendedEvent(event *EntityEvent) (*UserSuspendedEventData, error) {
	if event.EventType() != EventTypeUserSuspended {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeUserSuspended, event.EventType())
	}

	var data UserSuspendedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user suspended event: %w", err)
	}

	return &data, nil
}

// UnmarshalUserActivatedEvent unmarshals a user activated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserActivatedEvent(event *EntityEvent) (*UserActivatedEventData, error) {
	if event.EventType() != EventTypeUserActivated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeUserActivated, event.EventType())
	}

	var data UserActivatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user activated event: %w", err)
	}

	return &data, nil
}

// UnmarshalUserDeletedEvent unmarshals a user deleted event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalUserDeletedEvent(event *EntityEvent) (*UserDeletedEventData, error) {
	if event.EventType() != EventTypeUserDeleted {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeUserDeleted, event.EventType())
	}

	var data UserDeletedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user deleted event: %w", err)
	}

	return &data, nil
}

// UnmarshalWebIDGeneratedEvent unmarshals a WebID generated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalWebIDGeneratedEvent(event *EntityEvent) (*WebIDGeneratedEventData, error) {
	if event.EventType() != EventTypeWebIDGenerated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeWebIDGenerated, event.EventType())
	}

	var data WebIDGeneratedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal WebID generated event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountCreatedEvent unmarshals an account created event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountCreatedEvent(event *EntityEvent) (*AccountCreatedEventData, error) {
	if event.EventType() != EventTypeAccountCreated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountCreated, event.EventType())
	}

	var data AccountCreatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account created event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountUpdatedEvent unmarshals an account updated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountUpdatedEvent(event *EntityEvent) (*AccountUpdatedEventData, error) {
	if event.EventType() != EventTypeAccountUpdated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountUpdated, event.EventType())
	}

	var data AccountUpdatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account updated event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountSettingsUpdatedEvent unmarshals an account settings updated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountSettingsUpdatedEvent(event *EntityEvent) (*AccountSettingsUpdatedEventData, error) {
	if event.EventType() != EventTypeAccountSettingsUpdated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountSettingsUpdated, event.EventType())
	}

	var data AccountSettingsUpdatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account settings updated event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountMemberAddedEvent unmarshals an account member added event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountMemberAddedEvent(event *EntityEvent) (*AccountMemberAddedEventData, error) {
	if event.EventType() != EventTypeAccountMemberAdded {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountMemberAdded, event.EventType())
	}

	var data AccountMemberAddedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account member added event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountMemberRemovedEvent unmarshals an account member removed event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountMemberRemovedEvent(event *EntityEvent) (*AccountMemberRemovedEventData, error) {
	if event.EventType() != EventTypeAccountMemberRemoved {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountMemberRemoved, event.EventType())
	}

	var data AccountMemberRemovedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account member removed event: %w", err)
	}

	return &data, nil
}

// UnmarshalAccountMemberRoleUpdatedEvent unmarshals an account member role updated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalAccountMemberRoleUpdatedEvent(event *EntityEvent) (*AccountMemberRoleUpdatedEventData, error) {
	if event.EventType() != EventTypeAccountMemberRoleUpdated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeAccountMemberRoleUpdated, event.EventType())
	}

	var data AccountMemberRoleUpdatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account member role updated event: %w", err)
	}

	return &data, nil
}

// UnmarshalInvitationCreatedEvent unmarshals an invitation created event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalInvitationCreatedEvent(event *EntityEvent) (*InvitationCreatedEventData, error) {
	if event.EventType() != EventTypeInvitationCreated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeInvitationCreated, event.EventType())
	}

	var data InvitationCreatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation created event: %w", err)
	}

	return &data, nil
}

// UnmarshalInvitationAcceptedEvent unmarshals an invitation accepted event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalInvitationAcceptedEvent(event *EntityEvent) (*InvitationAcceptedEventData, error) {
	if event.EventType() != EventTypeInvitationAccepted {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeInvitationAccepted, event.EventType())
	}

	var data InvitationAcceptedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation accepted event: %w", err)
	}

	return &data, nil
}

// UnmarshalInvitationRevokedEvent unmarshals an invitation revoked event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalInvitationRevokedEvent(event *EntityEvent) (*InvitationRevokedEventData, error) {
	if event.EventType() != EventTypeInvitationRevoked {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeInvitationRevoked, event.EventType())
	}

	var data InvitationRevokedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation revoked event: %w", err)
	}

	return &data, nil
}

// UnmarshalInvitationExpiredEvent unmarshals an invitation expired event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalInvitationExpiredEvent(event *EntityEvent) (*InvitationExpiredEventData, error) {
	if event.EventType() != EventTypeInvitationExpired {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeInvitationExpired, event.EventType())
	}

	var data InvitationExpiredEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invitation expired event: %w", err)
	}

	return &data, nil
}

// UnmarshalMemberInvitedEvent unmarshals a member invited event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalMemberInvitedEvent(event *EntityEvent) (*MemberInvitedEventData, error) {
	if event.EventType() != EventTypeMemberInvited {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeMemberInvited, event.EventType())
	}

	var data MemberInvitedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal member invited event: %w", err)
	}

	return &data, nil
}

// UnmarshalRoleCreatedEvent unmarshals a role created event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalRoleCreatedEvent(event *EntityEvent) (*RoleCreatedEventData, error) {
	if event.EventType() != EventTypeRoleCreated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeRoleCreated, event.EventType())
	}

	var data RoleCreatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role created event: %w", err)
	}

	return &data, nil
}

// UnmarshalRoleUpdatedEvent unmarshals a role updated event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalRoleUpdatedEvent(event *EntityEvent) (*RoleUpdatedEventData, error) {
	if event.EventType() != EventTypeRoleUpdated {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeRoleUpdated, event.EventType())
	}

	var data RoleUpdatedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role updated event: %w", err)
	}

	return &data, nil
}

// UnmarshalRolePermissionAddedEvent unmarshals a role permission added event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalRolePermissionAddedEvent(event *EntityEvent) (*RolePermissionAddedEventData, error) {
	if event.EventType() != EventTypeRolePermissionAdded {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeRolePermissionAdded, event.EventType())
	}

	var data RolePermissionAddedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role permission added event: %w", err)
	}

	return &data, nil
}

// UnmarshalRolePermissionRemovedEvent unmarshals a role permission removed event from EntityEvent payload
func (u *EventDataUnmarshaler) UnmarshalRolePermissionRemovedEvent(event *EntityEvent) (*RolePermissionRemovedEventData, error) {
	if event.EventType() != EventTypeRolePermissionRemoved {
		return nil, fmt.Errorf("expected event type %s, got %s", EventTypeRolePermissionRemoved, event.EventType())
	}

	var data RolePermissionRemovedEventData
	if err := json.Unmarshal(event.Payload(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role permission removed event: %w", err)
	}

	return &data, nil
}

// EventTypeDispatcher provides a way to handle events in a type-safe manner
// by dispatching them to appropriate handlers based on event type
type EventTypeDispatcher struct {
	unmarshaler *EventDataUnmarshaler
}

// NewEventTypeDispatcher creates a new event type dispatcher
func NewEventTypeDispatcher() *EventTypeDispatcher {
	return &EventTypeDispatcher{
		unmarshaler: NewEventDataUnmarshaler(),
	}
}

// DispatchEvent dispatches an EntityEvent to the appropriate handler based on event type
func (d *EventTypeDispatcher) DispatchEvent(event *EntityEvent, handlers EventHandlers) error {
	// Handle events based on entity type and event type
	switch event.EntityType {
	case "user":
		return d.dispatchUserEvent(event, handlers)
	case "account":
		return d.dispatchAccountEvent(event, handlers)
	case "invitation":
		return d.dispatchInvitationEvent(event, handlers)
	case "role":
		return d.dispatchRoleEvent(event, handlers)
	case "account_member":
		return d.dispatchAccountMemberEvent(event, handlers)
	default:
		// Unknown entity type, ignore
		return nil
	}
}

// dispatchUserEvent handles user-related events
func (d *EventTypeDispatcher) dispatchUserEvent(event *EntityEvent, handlers EventHandlers) error {
	switch event.Type {
	case EventTypeUserCreated:
		if handlers.UserCreated != nil {
			data, err := d.unmarshaler.UnmarshalUserCreatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserCreated(data)
		}
	case EventTypeUserRegistered:
		if handlers.UserRegistered != nil {
			data, err := d.unmarshaler.UnmarshalUserRegisteredEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserRegistered(data)
		}
	case EventTypeUserProfileUpdated:
		if handlers.UserProfileUpdated != nil {
			data, err := d.unmarshaler.UnmarshalUserProfileUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserProfileUpdated(data)
		}
	case EventTypeUserSuspended:
		if handlers.UserSuspended != nil {
			data, err := d.unmarshaler.UnmarshalUserSuspendedEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserSuspended(data)
		}
	case EventTypeUserActivated:
		if handlers.UserActivated != nil {
			data, err := d.unmarshaler.UnmarshalUserActivatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserActivated(data)
		}
	case EventTypeUserDeleted:
		if handlers.UserDeleted != nil {
			data, err := d.unmarshaler.UnmarshalUserDeletedEvent(event)
			if err != nil {
				return err
			}
			return handlers.UserDeleted(data)
		}
	case EventTypeWebIDGenerated:
		if handlers.WebIDGenerated != nil {
			data, err := d.unmarshaler.UnmarshalWebIDGeneratedEvent(event)
			if err != nil {
				return err
			}
			return handlers.WebIDGenerated(data)
		}
	}
	return nil
}

// dispatchAccountEvent handles account-related events
func (d *EventTypeDispatcher) dispatchAccountEvent(event *EntityEvent, handlers EventHandlers) error {
	switch event.Type {
	case EventTypeAccountCreated:
		if handlers.AccountCreated != nil {
			data, err := d.unmarshaler.UnmarshalAccountCreatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountCreated(data)
		}
	case EventTypeAccountUpdated:
		if handlers.AccountUpdated != nil {
			data, err := d.unmarshaler.UnmarshalAccountUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountUpdated(data)
		}
	case EventTypeAccountSettingsUpdated:
		if handlers.AccountSettingsUpdated != nil {
			data, err := d.unmarshaler.UnmarshalAccountSettingsUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountSettingsUpdated(data)
		}
	case EventTypeAccountMemberAdded:
		if handlers.AccountMemberAdded != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberAddedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberAdded(data)
		}
	case EventTypeAccountMemberRemoved:
		if handlers.AccountMemberRemoved != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberRemovedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberRemoved(data)
		}
	case EventTypeAccountMemberRoleUpdated:
		if handlers.AccountMemberRoleUpdated != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberRoleUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberRoleUpdated(data)
		}
	}
	return nil
}

// dispatchInvitationEvent handles invitation-related events
func (d *EventTypeDispatcher) dispatchInvitationEvent(event *EntityEvent, handlers EventHandlers) error {
	switch event.Type {
	case EventTypeInvitationCreated:
		if handlers.InvitationCreated != nil {
			data, err := d.unmarshaler.UnmarshalInvitationCreatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.InvitationCreated(data)
		}
	case EventTypeInvitationAccepted:
		if handlers.InvitationAccepted != nil {
			data, err := d.unmarshaler.UnmarshalInvitationAcceptedEvent(event)
			if err != nil {
				return err
			}
			return handlers.InvitationAccepted(data)
		}
	case EventTypeInvitationRevoked:
		if handlers.InvitationRevoked != nil {
			data, err := d.unmarshaler.UnmarshalInvitationRevokedEvent(event)
			if err != nil {
				return err
			}
			return handlers.InvitationRevoked(data)
		}
	case EventTypeInvitationExpired:
		if handlers.InvitationExpired != nil {
			data, err := d.unmarshaler.UnmarshalInvitationExpiredEvent(event)
			if err != nil {
				return err
			}
			return handlers.InvitationExpired(data)
		}
	case EventTypeMemberInvited:
		if handlers.MemberInvited != nil {
			data, err := d.unmarshaler.UnmarshalMemberInvitedEvent(event)
			if err != nil {
				return err
			}
			return handlers.MemberInvited(data)
		}
	}
	return nil
}

// dispatchRoleEvent handles role-related events
func (d *EventTypeDispatcher) dispatchRoleEvent(event *EntityEvent, handlers EventHandlers) error {
	switch event.Type {
	case EventTypeRoleCreated:
		if handlers.RoleCreated != nil {
			data, err := d.unmarshaler.UnmarshalRoleCreatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.RoleCreated(data)
		}
	case EventTypeRoleUpdated:
		if handlers.RoleUpdated != nil {
			data, err := d.unmarshaler.UnmarshalRoleUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.RoleUpdated(data)
		}
	}
	return nil
}

// dispatchAccountMemberEvent handles account member-related events
func (d *EventTypeDispatcher) dispatchAccountMemberEvent(event *EntityEvent, handlers EventHandlers) error {
	switch event.Type {
	case EventTypeMemberAdded:
		if handlers.AccountMemberAdded != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberAddedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberAdded(data)
		}
	case EventTypeMemberRemoved:
		if handlers.AccountMemberRemoved != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberRemovedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberRemoved(data)
		}
	case EventTypeMemberRoleUpdated:
		if handlers.AccountMemberRoleUpdated != nil {
			data, err := d.unmarshaler.UnmarshalAccountMemberRoleUpdatedEvent(event)
			if err != nil {
				return err
			}
			return handlers.AccountMemberRoleUpdated(data)
		}
	}
	return nil
}

// EventHandlers contains all possible event handler functions
type EventHandlers struct {
	// User event handlers
	UserCreated        func(*UserCreatedEventData) error
	UserRegistered     func(*UserRegisteredEventData) error
	UserProfileUpdated func(*UserProfileUpdatedEventData) error
	UserSuspended      func(*UserSuspendedEventData) error
	UserActivated      func(*UserActivatedEventData) error
	UserDeleted        func(*UserDeletedEventData) error
	WebIDGenerated     func(*WebIDGeneratedEventData) error

	// Account event handlers
	AccountCreated           func(*AccountCreatedEventData) error
	AccountUpdated           func(*AccountUpdatedEventData) error
	AccountSettingsUpdated   func(*AccountSettingsUpdatedEventData) error
	AccountMemberAdded       func(*AccountMemberAddedEventData) error
	AccountMemberRemoved     func(*AccountMemberRemovedEventData) error
	AccountMemberRoleUpdated func(*AccountMemberRoleUpdatedEventData) error

	// Invitation event handlers
	InvitationCreated  func(*InvitationCreatedEventData) error
	InvitationAccepted func(*InvitationAcceptedEventData) error
	InvitationRevoked  func(*InvitationRevokedEventData) error
	InvitationExpired  func(*InvitationExpiredEventData) error
	MemberInvited      func(*MemberInvitedEventData) error

	// Role event handlers
	RoleCreated           func(*RoleCreatedEventData) error
	RoleUpdated           func(*RoleUpdatedEventData) error
	RolePermissionAdded   func(*RolePermissionAddedEventData) error
	RolePermissionRemoved func(*RolePermissionRemovedEventData) error
}
