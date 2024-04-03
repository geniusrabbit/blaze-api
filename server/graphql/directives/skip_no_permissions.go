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
		_, obj := objectByPermissionName(pm, perm)
		newObj := ownedObject(ctx, obj, user, account)
		if !account.CheckPermissions(ctx, newObj, perm) {
			return nil, nil
		}
	}

	return next(ctx)
}
