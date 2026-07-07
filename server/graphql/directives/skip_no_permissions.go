package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
)

// SkipNoPermissions directive to skip resolver if no permissions
func SkipNoPermissions(ctx context.Context, obj any, next graphql.Resolver, perms []string) (any, error) {
	user, account := session.UserAccount(ctx)

	if account == nil {
		return nil, nil
	}

	pm := permissions.FromContext(ctx)
	for _, perm := range perms {
		objName, permObj := objectByPermissionName(pm, perm)
		newObj := ownedObject(ctx, permObj, user, account, objName)
		if !account.CheckPermissions(ctx, newObj, perm) {
			return nil, nil
		}
	}

	return next(ctx)
}
