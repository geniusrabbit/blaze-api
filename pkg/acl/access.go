package acl

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
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
	PermApprove   = `approve`
	PermReject    = `reject`
	PermGet       = `get`
	PermSet       = `set`
)

// HavePermissions returns `true` if the `user` have all permissions from the list
func HavePermissions(ctx context.Context, permissions ...string) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, nil, permissions...)
}

// HaveObjectPermissions returns `true` if the `user` have all permissions from the list for the object
func HaveObjectPermissions(ctx context.Context, obj any, permissions ...string) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, permissions...)
}

// HaveAccessView to the object returns `true` if user can read of the object
func HaveAccessView(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermView+`.*`)
}

// HaveAccessList to the object returns `true` if user can read list of the object
func HaveAccessList(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermList+`.*`)
}

// HaveAccessCount of the object returns `true` if user can count the object
func HaveAccessCount(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermCount+`.*`)
}

// HaveAccessCreate of the object returns `true` if user can create this type of object
func HaveAccessCreate(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermCreate+`.*`)
}

// HaveAccessUpdate of the object returns `true` if user can update the object
func HaveAccessUpdate(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermUpdate+`.*`)
}

// HaveAccessDelete of the object returns `true` if user can delite the object
func HaveAccessDelete(ctx context.Context, obj any) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).CheckPermissions(ctx, obj, PermDelete+`.*`)
}

// HaveAccountLink of the object to the current account
func HaveAccountLink(ctx context.Context, obj any) bool {
	if IsNoPermCheck(ctx) {
		return true
	}
	// Check if I am is owner or have some `account` or `system` access to the object
	account := session.Account(ctx)
	return false ||
		account.CheckPermissions(ctx, obj, PermView+`.*`, PermList+`.*`, PermUpdate+`.*`)
}

// HasPermission returns `true` if the `user` have all permissions from the list (without custom check)
func HasPermission(ctx context.Context, permissions ...string) bool {
	return IsNoPermCheck(ctx) || session.Account(ctx).HasPermission(permissions...)
}
