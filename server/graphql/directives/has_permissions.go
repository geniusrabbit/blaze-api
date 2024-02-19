package directives

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/permissions"
)

var (
	errAuthorizationRequired = errors.New("authorization required")
	errAccessForbidden       = errors.New("access forbidden")
)

// HasPermissions for this user to the particular permission of the object
// Every module have the list of permissions ["list", "view", "create", "update", "delete", etc]
// This method checks, first of all, that object belongs to the user or have manager access and secondly
// that the user has the requested permissions of the module or several modules
func HasPermissions(ctx context.Context, obj any, next graphql.Resolver, perms []string) (any, error) {
	account := session.Account(ctx)
	pm := permissions.FromContext(ctx)

	if account == nil {
		return nil, errors.Wrap(errAuthorizationRequired, `no correct account`)
	}

	if len(perms) < 1 {
		return nil, errAccessForbidden
	}

	for _, perm := range perms {
		var (
			index   = max(0, strings.Index(perm, "."))
			objName = perm[:index]
			newObj  any
		)
		if objName != `` {
			newObj = pm.ObjectByName(objName)
		}
		if !account.CheckPermissions(ctx, newObj, perm) {
			if account.IsAnonimous() {
				return nil, errAuthorizationRequired
			}
			return nil, errors.Wrap(errAccessForbidden, objName+` [`+strings.Trim(objName[index:], `.`)+`]`)
		}
	}

	return next(ctx)
}
