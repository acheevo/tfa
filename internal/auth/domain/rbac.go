package domain

import (
	"fmt"
	"strings"
)

// Permission represents a specific permission in the system
type Permission string

// Resource represents a system resource that can be accessed
type Resource string

// Action represents an action that can be performed on a resource
type Action string

// Define system resources
const (
	ResourceUser    Resource = "user"
	ResourceAdmin   Resource = "admin"
	ResourceProfile Resource = "profile"
	ResourceAuth    Resource = "auth"
	ResourceAudit   Resource = "audit"
	ResourceSystem  Resource = "system"
)

// Define actions
const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionManage Action = "manage"
)

// System permissions
const (
	// User permissions
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserCreate Permission = "user:create"
	PermissionUserUpdate Permission = "user:update"
	PermissionUserDelete Permission = "user:delete"
	PermissionUserManage Permission = "user:manage"

	// Profile permissions (own profile)
	PermissionProfileRead   Permission = "profile:read"
	PermissionProfileUpdate Permission = "profile:update"

	// Admin permissions
	PermissionAdminRead   Permission = "admin:read"
	PermissionAdminWrite  Permission = "admin:write"
	PermissionAdminManage Permission = "admin:manage"

	// Auth permissions
	PermissionAuthRead   Permission = "auth:read"
	PermissionAuthWrite  Permission = "auth:write"
	PermissionAuthManage Permission = "auth:manage"

	// Audit permissions
	PermissionAuditRead   Permission = "audit:read"
	PermissionAuditWrite  Permission = "audit:write"
	PermissionAuditManage Permission = "audit:manage"

	// System permissions
	PermissionSystemRead   Permission = "system:read"
	PermissionSystemWrite  Permission = "system:write"
	PermissionSystemManage Permission = "system:manage"
)

// RolePermissions defines permissions for each role
var RolePermissions = map[UserRole][]Permission{
	RoleUser: {
		// Users can read and update their own profile
		PermissionProfileRead,
		PermissionProfileUpdate,
		// Users can manage their own auth (password change, etc.)
		PermissionAuthRead,
		PermissionAuthWrite,
	},
	RoleAdmin: {
		// Admins have all user permissions
		PermissionProfileRead,
		PermissionProfileUpdate,
		PermissionAuthRead,
		PermissionAuthWrite,
		// Plus admin-specific permissions
		PermissionUserRead,
		PermissionUserWrite,
		PermissionUserCreate,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserManage,
		PermissionAdminRead,
		PermissionAdminWrite,
		PermissionAdminManage,
		PermissionAuditRead,
		PermissionAuditWrite,
		PermissionSystemRead,
	},
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID     uint                   `json:"user_id"`
	UserRole   UserRole               `json:"user_role"`
	Resource   Resource               `json:"resource"`
	Action     Action                 `json:"action"`
	Permission Permission             `json:"permission"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// RBAC authorization functions

// HasPermission checks if a role has a specific permission
func HasPermission(role UserRole, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func HasAnyPermission(role UserRole, permissions []Permission) bool {
	for _, permission := range permissions {
		if HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func HasAllPermissions(role UserRole, permissions []Permission) bool {
	for _, permission := range permissions {
		if !HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// CanAccessResource checks if a role can perform an action on a resource
func CanAccessResource(role UserRole, resource Resource, action Action) bool {
	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return HasPermission(role, permission)
}

// CanManageUser checks if a user can manage another user based on roles
func CanManageUser(adminRole UserRole, adminID uint, targetRole UserRole, targetID uint) bool {
	// Check basic admin permissions
	if !HasPermission(adminRole, PermissionUserManage) {
		return false
	}

	// Users cannot manage themselves through admin interface
	if adminID == targetID {
		return false
	}

	// For now, any admin can manage any user
	// Future enhancement: role hierarchy (super admin > admin > user)
	return true
}

// GetRolePermissions returns all permissions for a role
func GetRolePermissions(role UserRole) []Permission {
	permissions, exists := RolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// IsValidRole checks if a role is valid
func IsValidRole(role UserRole) bool {
	_, exists := RolePermissions[role]
	return exists
}

// GetHigherRoles returns roles that are higher than the given role
func GetHigherRoles(role UserRole) []UserRole {
	switch role {
	case RoleUser:
		return []UserRole{RoleAdmin}
	case RoleAdmin:
		return []UserRole{} // No higher role currently
	default:
		return []UserRole{}
	}
}

// GetLowerRoles returns roles that are lower than the given role
func GetLowerRoles(role UserRole) []UserRole {
	switch role {
	case RoleAdmin:
		return []UserRole{RoleUser}
	case RoleUser:
		return []UserRole{} // No lower role currently
	default:
		return []UserRole{}
	}
}

// IsRoleHigherThan checks if role1 is higher than role2
func IsRoleHigherThan(role1, role2 UserRole) bool {
	higherRoles := GetHigherRoles(role2)
	for _, r := range higherRoles {
		if r == role1 {
			return true
		}
	}
	return false
}

// Permission validation helpers

// ValidatePermissionString validates a permission string format
func ValidatePermissionString(permission string) error {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid permission format: %s (expected resource:action)", permission)
	}
	return nil
}

// ParsePermission parses a permission string into resource and action
func ParsePermission(permission Permission) (Resource, Action, error) {
	parts := strings.Split(string(permission), ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid permission format: %s", permission)
	}
	return Resource(parts[0]), Action(parts[1]), nil
}

// BuildPermission builds a permission from resource and action
func BuildPermission(resource Resource, action Action) Permission {
	return Permission(fmt.Sprintf("%s:%s", resource, action))
}

// Context-aware permission checking

// PermissionContext represents additional context for permission checks
type PermissionContext struct {
	UserID   uint                   `json:"user_id"`
	TargetID uint                   `json:"target_id,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// HasPermissionWithContext checks permission with additional context
func HasPermissionWithContext(role UserRole, permission Permission, ctx *PermissionContext) bool {
	// First check basic permission
	if !HasPermission(role, permission) {
		return false
	}

	// Add context-aware logic here
	// For example: users can only update their own profile
	if permission == PermissionProfileUpdate {
		return ctx != nil && ctx.UserID == ctx.TargetID
	}

	return true
}

// Audit helpers for RBAC

// RBACEvent represents an RBAC-related event for auditing
type RBACEvent struct {
	Type       string                 `json:"type"`
	UserID     uint                   `json:"user_id"`
	UserRole   UserRole               `json:"user_role"`
	Resource   Resource               `json:"resource"`
	Action     Action                 `json:"action"`
	Permission Permission             `json:"permission"`
	Allowed    bool                   `json:"allowed"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
}

// CreateRBACEvent creates an RBAC event for auditing
func CreateRBACEvent(
	userID uint,
	userRole UserRole,
	resource Resource,
	action Action,
	allowed bool,
	reason string,
) *RBACEvent {
	return &RBACEvent{
		Type:       "rbac_check",
		UserID:     userID,
		UserRole:   userRole,
		Resource:   resource,
		Action:     action,
		Permission: BuildPermission(resource, action),
		Allowed:    allowed,
		Reason:     reason,
	}
}
