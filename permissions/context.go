package permissions

import (
	"context"
)

var (
	// CtxPermissionManagerObject reference to the permission manager
	CtxPermissionManagerObject = struct{ s string }{"permissionmanager"}
)

// FromContext permission manager object
func FromContext(ctx context.Context) *Manager {
	return ctx.Value(CtxPermissionManagerObject).(*Manager)
}

// WithManager puts permission manager to context
func WithManager(ctx context.Context, manager *Manager) context.Context {
	return context.WithValue(ctx, CtxPermissionManagerObject, manager)
}
