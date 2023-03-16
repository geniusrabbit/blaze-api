package model

import (
	"context"
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
