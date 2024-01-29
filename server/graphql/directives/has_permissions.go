package directives

import (
	"context"
	"reflect"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/context/permissionmanager"
	"github.com/geniusrabbit/blaze-api/context/session"
	gmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

var (
	errAuthorizationRequired = errors.New("authorization required")
	errAccessForbidden       = errors.New("access forbidden")
)

// HasPermissions for this user to the particular permission of the object
// Every module have the list of permissions ["list", "view", "create", "update", "delete", etc]
// This method checks, first of all, that object belongs to the user or have manager access and secondly
// that the user has the requested permissions of the module or several modules
func HasPermissions(ctx context.Context, obj any, next graphql.Resolver, permissions []*gmodels.Permission) (any, error) {
	account := session.Account(ctx)
	pm := permissionmanager.Get(ctx)

	if account == nil {
		return nil, errors.Wrap(errAuthorizationRequired, `no correct account`)
	}

	if len(permissions) < 1 {
		return nil, errAccessForbidden
	}

	for _, perm := range permissions {
		newObj := pm.ObjectByName(perm.Key)

		// Check if user have permission to the whole module
		// for example: users.*, campaigns.*, banners.*, etc.
		if account.CheckPermissions(ctx, newObj, "*") {
			continue
		} else if len(perm.Access) == 0 {
			if account.IsAnonimous() {
				return nil, errAuthorizationRequired
			}
			return nil, errors.Wrap(errAccessForbidden, objectName(newObj)+` [*]`)
		}

		// Check permission one by one if user doesn't have at least one of them then return the error
		for _, access := range perm.Access {
			if !account.CheckPermissions(ctx, newObj, access) {
				if account.IsAnonimous() {
					return nil, errAuthorizationRequired
				}
				return nil, errors.Wrap(errAccessForbidden, objectName(newObj)+` [`+access+`]`)
			}
		}
	}

	return next(ctx)
}

func objectName(obj any) string {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Interface || t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
