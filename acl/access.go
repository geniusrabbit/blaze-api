package acl

import (
	"context"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/permissions"
)

var ctxNoPermCheck = struct{ s string }{`no-perm-check`}

// WithNoPermCheck returns new context with disabled permission check
func WithNoPermCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxNoPermCheck, true)
}

// IsNoPermCheck returns `true` if the permission check is disabled
func IsNoPermCheck(ctx context.Context) bool {
	return ctx.Value(ctxNoPermCheck) != nil
}

// The permission list
const (
	PermView      = `view`
	PermCreate    = `create`
	PermUpdate    = `update`
	PermDelete    = `delete`
	PermList      = `list`
	PermAuthCross = session.PermAuthCross
	PermCount     = `count`
)

// HavePermissions returns `true` if the `user` have all permissions from the list
func HavePermissions(ctx context.Context, permissions ...string) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, nil, permissions...)
}

// HaveObjectPermissions returns `true` if the `user` have all permissions from the list for the object
func HaveObjectPermissions(ctx context.Context, obj any, permissions ...string) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, permissions...)
}

// HaveAccessView to the object returns `true` if user can read of the object
func HaveAccessView(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermView+`.*`)
}

// HaveAccessList to the object returns `true` if user can read list of the object
func HaveAccessList(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermList+`.*`)
}

// HaveAccessCount of the object returns `true` if user can count the object
func HaveAccessCount(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermCount+`.*`)
}

// HaveAccessCreate of the object returns `true` if user can create this type of object
func HaveAccessCreate(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermCreate+`.*`)
}

// HaveAccessUpdate of the object returns `true` if user can update the object
func HaveAccessUpdate(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermUpdate+`.*`)
}

// HaveAccessDelete of the object returns `true` if user can delite the object
func HaveAccessDelete(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || checkPermissions(ctx, obj, PermDelete+`.*`)
}

// HaveAccountLink of the object to the current account
func HaveAccountLink(ctx context.Context, obj any) bool {
	if IsNoPermCheck(ctx) {
		return true
	}
	// Check if I am is owner or have some `account` or `system` access to the object
	return checkPermissions(ctx, obj, PermView+`.*`, PermList+`.*`, PermUpdate+`.*`)
}

func checkPermissions(ctx context.Context, obj any, names ...string) bool {
	account := session.Account(ctx)
	if account == nil {
		return false
	}
	if account.Permissions == nil {
		return permissions.FromContext(ctx).
			DefaultRole(ctx).CheckPermissions(ctx, obj, names...)
	}
	return account.CheckPermissions(ctx, obj, names...)
}
