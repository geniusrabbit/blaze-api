package model

import (
	"context"

	"github.com/demdxx/rbac"
)

var ctxPermissionCheckAccount = &struct{ s string }{s: "permc:account"}

// PermissionCheckAccountFromContext returns the original account for check
func PermissionCheckAccountFromContext(ctx context.Context) *Account {
	switch acc := ctx.Value(ctxPermissionCheckAccount).(type) {
	case nil:
	case *Account:
		return acc
	}
	return nil
}

type permissionChecker interface {
	CheckPermissions(ctx context.Context, resource any, names ...string) bool
	CheckedPermissions(ctx context.Context, resource any, names ...string) rbac.Permission
	ChildRoles() []rbac.Role
	ChildPermissions() []rbac.Permission
}

type groupPermissionChecker []permissionChecker

func (groups groupPermissionChecker) CheckPermissions(ctx context.Context, resource any, names ...string) bool {
	for _, group := range groups {
		if group.CheckPermissions(ctx, resource, names...) {
			return true
		}
	}
	return false
}

func (groups groupPermissionChecker) CheckedPermissions(ctx context.Context, resource any, names ...string) rbac.Permission {
	for _, group := range groups {
		if perm := group.CheckedPermissions(ctx, resource, names...); perm == nil {
			return perm
		}
	}
	return nil
}

func (groups groupPermissionChecker) ChildRoles() []rbac.Role {
	var roles []rbac.Role
	for _, group := range groups {
		switch role := group.(type) {
		case nil:
		case rbac.Role:
			roles = append(roles, role)
		}
		roles = append(roles, group.ChildRoles()...)
	}
	return roles
}

func (groups groupPermissionChecker) ChildPermissions() []rbac.Permission {
	var perms []rbac.Permission
	for _, group := range groups {
		switch perm := group.(type) {
		case rbac.Permission:
			perms = append(perms, perm)
		}
		perms = append(perms, group.ChildPermissions()...)
	}
	return perms
}
