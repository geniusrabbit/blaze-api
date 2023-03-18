package acl

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/context/session"
)

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
	return session.Account(ctx).CheckPermissions(ctx, nil, permissions...)
}

// HaveObjectPermissions returns `true` if the `user` have all permissions from the list for the object
func HaveObjectPermissions(ctx context.Context, obj any, permissions ...string) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, permissions...)
}

// HaveAccessView to the object returns `true` if user can read of the object
func HaveAccessView(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermView)
}

// HaveAccessList to the object returns `true` if user can read list of the object
func HaveAccessList(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermList)
}

// HaveAccessCount of the object returns `true` if user can count the object
func HaveAccessCount(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermCount)
}

// HaveAccessCreate of the object returns `true` if user can create this type of object
func HaveAccessCreate(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermCreate)
}

// HaveAccessUpdate of the object returns `true` if user can update the object
func HaveAccessUpdate(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermUpdate)
}

// HaveAccessDelete of the object returns `true` if user can delite the object
func HaveAccessDelete(ctx context.Context, obj any) bool {
	return session.Account(ctx).CheckPermissions(ctx, obj, PermDelete)
}

// HaveAccountLink of the object to the current account
func HaveAccountLink(ctx context.Context, obj any) bool {
	// Check if I am is owner or have some `account` or `system` access to the object
	account := session.Account(ctx)
	return account.CheckPermissions(ctx, obj, PermView) ||
		account.CheckPermissions(ctx, obj, PermList) ||
		account.CheckPermissions(ctx, obj, PermUpdate)
}
