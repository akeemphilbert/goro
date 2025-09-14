package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// UserProfile contains user profile information
type UserProfile struct {
	Name        string                 `json:"name"`
	Bio         string                 `json:"bio"`
	Avatar      string                 `json:"avatar"`
	Preferences map[string]interface{} `json:"preferences"`
}

// Validate validates the user profile
func (p UserProfile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 255 {
		return fmt.Errorf("name must be less than 256 characters")
	}
	return nil
}

// User represents a user in the system
type User struct {
	*pericarpdomain.BasicEntity
	WebID     string      `json:"webid"`
	Email     string      `json:"email"`
	Profile   UserProfile `json:"profile"`
	Status    UserStatus  `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// NewUser creates a new user with validation
func NewUser(ctx context.Context, id, webID, email string, profile UserProfile) (*User, error) {
	log.Context(ctx).Debugf("[NewUser] Starting user creation: id=%s, webID=%s, email=%s, profileName=%s", id, webID, email, profile.Name)

	now := time.Now()
	log.Context(ctx).Debug("[NewUser] Creating user entity with default status active")

	user := &User{
		BasicEntity: pericarpdomain.NewEntity(id),
		WebID:       webID,
		Email:       email,
		Profile:     profile,
		Status:      UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	log.Context(ctx).Debug("[NewUser] User entity created, starting validation")

	if strings.TrimSpace(id) == "" {
		log.Context(ctx).Debug("[NewUser] Validation failed: user ID is required")
		user.AddError(fmt.Errorf("user ID is required"))
		return user, fmt.Errorf("user ID is required")
	}
	log.Context(ctx).Debugf("[NewUser] ID validation passed: %s", id)

	if strings.TrimSpace(webID) == "" {
		log.Context(ctx).Debug("[NewUser] Validation failed: WebID is required")
		user.AddError(fmt.Errorf("WebID is required"))
		return user, fmt.Errorf("WebID is required")
	}
	log.Context(ctx).Debugf("[NewUser] WebID validation passed: %s", webID)

	log.Context(ctx).Debug("[NewUser] Validating email format")
	if !isValidEmail(email) {
		log.Context(ctx).Debugf("[NewUser] Email validation failed: %s", email)
		user.AddError(fmt.Errorf("invalid email format"))
		return user, fmt.Errorf("invalid email format")
	}
	log.Context(ctx).Debugf("[NewUser] Email validation passed: %s", email)

	log.Context(ctx).Debug("[NewUser] Validating user profile")
	if err := profile.Validate(); err != nil {
		log.Context(ctx).Debugf("[NewUser] Profile validation failed: %v", err)
		user.AddError(fmt.Errorf("invalid profile: %w", err))
		return user, fmt.Errorf("invalid profile: %w", err)
	}
	log.Context(ctx).Debugf("[NewUser] Profile validation passed: name=%s", profile.Name)

	log.Context(ctx).Debug("[NewUser] All validations passed, creating event")

	// Log successful user creation
	log.Context(ctx).Infof("User created successfully: id=%s, email=%s, webID=%s", id, email, webID)

	log.Context(ctx).Debug("[NewUser] Creating user created event")
	// Emit user created event
	event := NewUserCreatedEvent(user, webID, string(UserStatusActive))
	user.AddEvent(event)
	log.Context(ctx).Debug("[NewUser] User created event added to entity")

	log.Context(ctx).Debug("[NewUser] User creation completed successfully")
	return user, nil
}

// UpdateProfile updates the user's profile
func (u *User) UpdateProfile(ctx context.Context, profile UserProfile) error {
	log.Context(ctx).Debugf("[UpdateProfile] Starting profile update for user: id=%s, currentName=%s, newName=%s",
		u.ID(), u.Profile.Name, profile.Name)

	log.Context(ctx).Debugf("[UpdateProfile] Current user status: %s", u.Status)
	if u.Status == UserStatusDeleted {
		log.Context(ctx).Debug("[UpdateProfile] Validation failed: user is deleted")
		log.Context(ctx).Warnf("Attempt to update profile of deleted user: id=%s", u.ID())
		err := fmt.Errorf("cannot update profile of deleted user")
		u.AddError(err)
		return err
	}
	log.Context(ctx).Debug("[UpdateProfile] User status validation passed")

	log.Context(ctx).Debug("[UpdateProfile] Validating new profile")
	if err := profile.Validate(); err != nil {
		log.Context(ctx).Debugf("[UpdateProfile] Profile validation failed: %v", err)
		validationErr := fmt.Errorf("invalid profile: %w", err)
		u.AddError(validationErr)
		return validationErr
	}
	log.Context(ctx).Debug("[UpdateProfile] Profile validation passed")

	log.Context(ctx).Debug("[UpdateProfile] Storing old profile for event")
	oldProfile := u.Profile
	log.Context(ctx).Debugf("[UpdateProfile] Old profile - Name=%s, Bio=%s", oldProfile.Name, oldProfile.Bio)

	log.Context(ctx).Debug("[UpdateProfile] Applying new profile to user")
	u.Profile = profile
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[UpdateProfile] User profile and timestamp updated")

	// Log successful profile update
	log.Context(ctx).Infof("User profile updated: id=%s, name=%s", u.ID(), profile.Name)

	log.Context(ctx).Debug("[UpdateProfile] Creating profile updated event")
	// Emit user profile updated event
	event := NewUserProfileUpdatedEvent(u, oldProfile, profile)
	u.AddEvent(event)
	log.Context(ctx).Debug("[UpdateProfile] Profile updated event created and added to entity")

	log.Context(ctx).Debug("[UpdateProfile] Profile update completed successfully")
	return nil
}

// Suspend suspends the user account
func (u *User) Suspend(ctx context.Context) error {
	log.Context(ctx).Debugf("[Suspend] Starting user suspension: id=%s, currentStatus=%s", u.ID(), u.Status)

	log.Context(ctx).Debug("[Suspend] Checking if user is deleted")
	if u.Status == UserStatusDeleted {
		log.Context(ctx).Debug("[Suspend] Suspension failed: user is deleted")
		err := fmt.Errorf("cannot suspend deleted user")
		u.AddError(err)
		return err
	}
	log.Context(ctx).Debug("[Suspend] User deletion check passed")

	log.Context(ctx).Debug("[Suspend] Setting user status to suspended")
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[Suspend] User status and timestamp updated")

	// Log user suspension
	log.Context(ctx).Warnf("User suspended: id=%s", u.ID())

	log.Context(ctx).Debug("[Suspend] Creating user suspended event")
	// Emit user suspended event
	event := NewUserSuspendedEvent(u, "Account suspended by administrator")
	u.AddEvent(event)
	log.Context(ctx).Debug("[Suspend] User suspended event created and added to entity")

	log.Context(ctx).Debug("[Suspend] User suspension completed successfully")
	return nil
}

// Activate activates the user account
func (u *User) Activate(ctx context.Context) error {
	log.Context(ctx).Debugf("[Activate] Starting user activation: id=%s, currentStatus=%s", u.ID(), u.Status)

	log.Context(ctx).Debug("[Activate] Checking if user is deleted")
	if u.Status == UserStatusDeleted {
		log.Context(ctx).Debug("[Activate] Activation failed: user is deleted")
		err := fmt.Errorf("cannot activate deleted user")
		u.AddError(err)
		return err
	}
	log.Context(ctx).Debug("[Activate] User deletion check passed")

	log.Context(ctx).Debug("[Activate] Setting user status to active")
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[Activate] User status and timestamp updated")

	log.Context(ctx).Debug("[Activate] Creating user activated event")
	// Emit user activated event
	event := NewUserActivatedEvent(u)
	u.AddEvent(event)
	log.Context(ctx).Debug("[Activate] User activated event created and added to entity")

	log.Context(ctx).Infof("User activated: id=%s", u.ID())
	log.Context(ctx).Debug("[Activate] User activation completed successfully")
	return nil
}

// Delete marks the user as deleted
func (u *User) Delete(ctx context.Context) error {
	log.Context(ctx).Debugf("[Delete] Starting user deletion: id=%s, currentStatus=%s", u.ID(), u.Status)

	log.Context(ctx).Debug("[Delete] Setting user status to deleted")
	u.Status = UserStatusDeleted
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debug("[Delete] User status and timestamp updated")

	log.Context(ctx).Debug("[Delete] Creating user deleted event")
	// Emit user deleted event
	event := NewUserDeletedEvent(u, "Account deleted by user or administrator")
	u.AddEvent(event)
	log.Context(ctx).Debug("[Delete] User deleted event created and added to entity")

	log.Context(ctx).Warnf("User deleted: id=%s", u.ID())
	log.Context(ctx).Debug("[Delete] User deletion completed successfully")
	return nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
