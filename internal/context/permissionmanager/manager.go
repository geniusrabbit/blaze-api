package permissionmanager

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/permissions"
)

var (
	// CtxPermissionManagerObject reference to the permission manager
	CtxPermissionManagerObject = struct{ s string }{"permissionmanager"}
)

// Get permission manager object
func Get(ctx context.Context) *permissions.Manager {
	return ctx.Value(CtxPermissionManagerObject).(*permissions.Manager)
}

// WithManager puts permission manager to context
func WithManager(ctx context.Context, manager *permissions.Manager) context.Context {
	return context.WithValue(ctx, CtxPermissionManagerObject, manager)
}
