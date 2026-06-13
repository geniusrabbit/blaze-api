package models

import (
	"context"

	"github.com/demdxx/rbac"
)

// CtxPermissionCheckAccount is the context key used to store the account for permission checks.
var CtxPermissionCheckAccount = &struct{ s string }{s: "permc:account"}

// PermissionChecker describes the interface for checking permissions for account members.
type PermissionChecker interface {
	// CheckPermissions verifies if the given patterns are permitted for a resource in the context.
	CheckPermissions(ctx context.Context, resource any, patterns ...string) bool
	// CheckedPermissions returns the permission object if patterns are permitted, nil otherwise.
	CheckedPermissions(ctx context.Context, resource any, patterns ...string) rbac.Permission
	// ChildRoles returns all child roles associated with this permission checker.
	ChildRoles() []rbac.Role
	// ChildPermissions returns all child permissions associated with this permission checker.
	ChildPermissions() []rbac.Permission
	// Permissions returns all permissions matching the given patterns.
	Permissions(patterns ...string) []rbac.Permission
	// HasPermission checks if any of the given permission patterns exist.
	HasPermission(patterns ...string) bool
}

// groupPermissionChecker is a slice of PermissionChecker that aggregates multiple permission checkers.
type groupPermissionChecker []PermissionChecker

// CheckPermissions checks if any group permits the resource with the given patterns.
func (groups groupPermissionChecker) CheckPermissions(ctx context.Context, resource any, patterns ...string) bool {
	for _, group := range groups {
		if group.CheckPermissions(ctx, resource, patterns...) {
			return true
		}
	}
	return false
}

// CheckedPermissions returns the first permission found across groups, or nil.
func (groups groupPermissionChecker) CheckedPermissions(ctx context.Context, resource any, patterns ...string) rbac.Permission {
	for _, group := range groups {
		if perm := group.CheckedPermissions(ctx, resource, patterns...); perm != nil {
			return perm
		}
	}
	return nil
}

// ChildRoles aggregates child roles from all groups.
func (groups groupPermissionChecker) ChildRoles() []rbac.Role {
	var roles []rbac.Role
	for _, group := range groups {
		if role, ok := group.(rbac.Role); ok {
			roles = append(roles, role)
		}
		roles = append(roles, group.ChildRoles()...)
	}
	return roles
}

// ChildPermissions aggregates child permissions from all groups.
func (groups groupPermissionChecker) ChildPermissions() []rbac.Permission {
	var perms []rbac.Permission
	for _, group := range groups {
		if perm, ok := group.(rbac.Permission); ok {
			perms = append(perms, perm)
		}
		perms = append(perms, group.ChildPermissions()...)
	}
	return perms
}

// Permissions aggregates permissions matching the patterns across all groups.
func (groups groupPermissionChecker) Permissions(patterns ...string) []rbac.Permission {
	var perms []rbac.Permission
	for _, group := range groups {
		perms = append(perms, group.Permissions(patterns...)...)
	}
	return perms
}

// HasPermission checks if any group has the given permission patterns.
func (groups groupPermissionChecker) HasPermission(patterns ...string) bool {
	for _, group := range groups {
		if group.HasPermission(patterns...) {
			return true
		}
	}
	return false
}

var _ PermissionChecker = groupPermissionChecker{}
