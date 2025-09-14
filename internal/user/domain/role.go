package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	pericarpdomain "github.com/akeemphilbert/pericarp/pkg/domain"
	"github.com/go-kratos/kratos/v2/log"
)

// Permission represents a permission for a role
type Permission struct {
	Resource string `json:"resource"` // "user", "account", "resource", "*"
	Action   string `json:"action"`   // "create", "read", "update", "delete", "*"
	Scope    string `json:"scope"`    // "own", "account", "global"
}

// Validate validates the permission
func (p Permission) Validate() error {
	if strings.TrimSpace(p.Resource) == "" {
		return fmt.Errorf("resource is required")
	}
	if strings.TrimSpace(p.Action) == "" {
		return fmt.Errorf("action is required")
	}
	if strings.TrimSpace(p.Scope) == "" {
		return fmt.Errorf("scope is required")
	}
	return nil
}

// Role represents a role with permissions
type Role struct {
	*pericarpdomain.BasicEntity
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewRole creates a new role with validation
func NewRole(ctx context.Context, id, name, description string, permissions []Permission) (*Role, error) {
	log.Context(ctx).Debugf("[NewRole] Starting role creation: id=%s, name=%s, permissions=%d",
		id, name, len(permissions))

	now := time.Now()
	log.Context(ctx).Debug("[NewRole] Creating role entity")

	role := &Role{
		BasicEntity: pericarpdomain.NewEntity(id),
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	log.Context(ctx).Debug("[NewRole] Role entity created, starting validation")

	if strings.TrimSpace(id) == "" {
		log.Context(ctx).Debug("[NewRole] Validation failed: role ID is required")
		role.AddError(fmt.Errorf("role ID is required"))
		return role, fmt.Errorf("role ID is required")
	}
	log.Context(ctx).Debugf("[NewRole] ID validation passed: %s", id)

	if strings.TrimSpace(name) == "" {
		log.Context(ctx).Debug("[NewRole] Validation failed: role name is required")
		role.AddError(fmt.Errorf("role name is required"))
		return role, fmt.Errorf("role name is required")
	}
	log.Context(ctx).Debugf("[NewRole] Name validation passed: %s", name)

	log.Context(ctx).Debugf("[NewRole] Validating %d permissions", len(permissions))
	// Validate all permissions
	for i, perm := range permissions {
		log.Context(ctx).Debugf("[NewRole] Validating permission %d: resource=%s, action=%s, scope=%s",
			i+1, perm.Resource, perm.Action, perm.Scope)
		if err := perm.Validate(); err != nil {
			log.Context(ctx).Debugf("[NewRole] Permission %d validation failed: %v", i+1, err)
			validationErr := fmt.Errorf("invalid permission: %w", err)
			role.AddError(validationErr)
			return role, validationErr
		}
		log.Context(ctx).Debugf("[NewRole] Permission %d validation passed", i+1)
	}
	log.Context(ctx).Debug("[NewRole] All permission validations passed")

	log.Context(ctx).Debug("[NewRole] All validations passed, creating event")

	// Log successful role creation
	log.Context(ctx).Infof("Role created successfully: id=%s, name=%s, permissions=%d", id, name, len(permissions))

	log.Context(ctx).Debug("[NewRole] Creating role created event")
	// Emit role created event
	event := NewRoleCreatedEvent(role, permissions)
	role.AddEvent(event)
	log.Context(ctx).Debug("[NewRole] Role created event added to entity")

	log.Context(ctx).Debug("[NewRole] Role creation completed successfully")
	return role, nil
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(resource, action, scope string) bool {
	// First check for exact matches
	for _, perm := range r.Permissions {
		if perm.Resource == resource && perm.Action == action && perm.Scope == scope {
			return true
		}
	}

	// Then check for wildcard matches
	for _, perm := range r.Permissions {
		if (perm.Resource == "*" || perm.Resource == resource) &&
			(perm.Action == "*" || perm.Action == action) &&
			perm.Scope == scope {
			return true
		}
	}
	return false
}

// AddPermission adds a permission to the role
func (r *Role) AddPermission(ctx context.Context, permission Permission) error {
	log.Context(ctx).Debugf("[AddPermission] Starting permission addition to role: roleID=%s, resource=%s, action=%s, scope=%s",
		r.ID(), permission.Resource, permission.Action, permission.Scope)

	log.Context(ctx).Debug("[AddPermission] Validating permission")
	if err := permission.Validate(); err != nil {
		log.Context(ctx).Debugf("[AddPermission] Permission validation failed: %v", err)
		validationErr := fmt.Errorf("invalid permission: %w", err)
		r.AddError(validationErr)
		return validationErr
	}
	log.Context(ctx).Debug("[AddPermission] Permission validation passed")

	log.Context(ctx).Debug("[AddPermission] Checking if permission already exists")
	// Check if permission already exists
	for i, perm := range r.Permissions {
		if perm.Resource == permission.Resource &&
			perm.Action == permission.Action &&
			perm.Scope == permission.Scope {
			log.Context(ctx).Debugf("[AddPermission] Permission already exists at index %d, skipping addition", i)
			return nil // Permission already exists, no need to add
		}
	}
	log.Context(ctx).Debug("[AddPermission] Permission does not exist, proceeding with addition")

	log.Context(ctx).Debug("[AddPermission] Adding permission to role")
	r.Permissions = append(r.Permissions, permission)
	r.UpdatedAt = time.Now()
	log.Context(ctx).Debugf("[AddPermission] Permission added, role now has %d permissions", len(r.Permissions))

	// Log successful permission addition
	log.Context(ctx).Infof("Permission added to role: role=%s, resource=%s, action=%s, scope=%s", r.ID(), permission.Resource, permission.Action, permission.Scope)

	log.Context(ctx).Debug("[AddPermission] Creating permission added event")
	// Emit permission added event
	event := NewRolePermissionAddedEvent(r, permission)
	r.AddEvent(event)
	log.Context(ctx).Debug("[AddPermission] Permission added event created and added to entity")

	log.Context(ctx).Debug("[AddPermission] Permission addition completed successfully")
	return nil
}

// RemovePermission removes a permission from the role
func (r *Role) RemovePermission(ctx context.Context, permission Permission) error {
	log.Context(ctx).Debugf("[RemovePermission] Starting permission removal from role: roleID=%s, resource=%s, action=%s, scope=%s",
		r.ID(), permission.Resource, permission.Action, permission.Scope)

	log.Context(ctx).Debugf("[RemovePermission] Searching through %d permissions", len(r.Permissions))

	for i, perm := range r.Permissions {
		log.Context(ctx).Debugf("[RemovePermission] Checking permission %d: resource=%s, action=%s, scope=%s",
			i, perm.Resource, perm.Action, perm.Scope)
		if perm.Resource == permission.Resource &&
			perm.Action == permission.Action &&
			perm.Scope == permission.Scope {
			log.Context(ctx).Debugf("[RemovePermission] Found matching permission at index %d", i)

			log.Context(ctx).Debug("[RemovePermission] Removing permission from role")
			r.Permissions = append(r.Permissions[:i], r.Permissions[i+1:]...)
			r.UpdatedAt = time.Now()
			log.Context(ctx).Debugf("[RemovePermission] Permission removed, role now has %d permissions", len(r.Permissions))

			log.Context(ctx).Debug("[RemovePermission] Creating permission removed event")
			// Emit permission removed event
			event := NewRolePermissionRemovedEvent(r, permission)
			r.AddEvent(event)
			log.Context(ctx).Debug("[RemovePermission] Permission removed event created and added to entity")

			log.Context(ctx).Infof("Permission removed from role: role=%s, resource=%s, action=%s, scope=%s",
				r.ID(), permission.Resource, permission.Action, permission.Scope)
			log.Context(ctx).Debug("[RemovePermission] Permission removal completed successfully")
			break
		}
	}

	log.Context(ctx).Debug("[RemovePermission] Permission removal process completed")
	return nil
}

// System roles - predefined roles with specific permissions
var (
	RoleOwner  *Role
	RoleAdmin  *Role
	RoleMember *Role
	RoleViewer *Role
)

// InitializeSystemRoles initializes the system roles
func InitializeSystemRoles(ctx context.Context) {
	RoleOwner, _ = NewRole(ctx, "owner", "Owner", "Full access to account and all resources", []Permission{
		{Resource: "*", Action: "*", Scope: "account"},
	})

	RoleAdmin, _ = NewRole(ctx, "admin", "Administrator", "Administrative access to account management", []Permission{
		{Resource: "user", Action: "*", Scope: "account"},
		{Resource: "account", Action: "read", Scope: "account"},
		{Resource: "account", Action: "update", Scope: "account"},
		{Resource: "resource", Action: "*", Scope: "account"},
	})

	RoleMember, _ = NewRole(ctx, "member", "Member", "Standard member access", []Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "create", Scope: "own"},
		{Resource: "resource", Action: "read", Scope: "own"},
		{Resource: "resource", Action: "update", Scope: "own"},
	})

	RoleViewer, _ = NewRole(ctx, "viewer", "Viewer", "Read-only access", []Permission{
		{Resource: "user", Action: "read", Scope: "account"},
		{Resource: "resource", Action: "read", Scope: "account"},
	})
}

// GetSystemRoles returns all system roles
func GetSystemRoles(ctx context.Context) []*Role {
	if RoleOwner == nil {
		InitializeSystemRoles(ctx)
	}
	return []*Role{RoleOwner, RoleAdmin, RoleMember, RoleViewer}
}
