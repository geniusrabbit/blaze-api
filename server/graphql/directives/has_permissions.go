package directives

import (
	"context"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
)

var (
	errAuthorizationRequired = errors.New("authorization required")
	errAccessForbidden       = errors.New("access forbidden")
)

type (
	accountOwnerSetter interface {
		SetAccountOwnerID(uint64)
	}
	userOwnerSetter interface {
		SetUserOwnerID(uint64)
	}
)

// HasPermissions for this user to the particular permission of the object
// Every module have the list of permissions ["list", "view", "create", "update", "delete", etc]
// This method checks, first of all, that object belongs to the user or have manager access and secondly
// that the user has the requested permissions of the module or several modules
func HasPermissions(ctx context.Context, obj any, next graphql.Resolver, perms []string) (any, error) {
	user, account := session.UserAccount(ctx)
	pm := permissions.FromContext(ctx)

	if account == nil {
		return nil, errors.Wrap(errAuthorizationRequired, `no correct account`)
	}

	if len(perms) < 1 {
		return nil, errAccessForbidden
	}

	for _, perm := range perms {
		objName, obj := objectByPermissionName(pm, perm)
		newObj := ownedObject(ctx, obj, user, account)
		if !account.CheckPermissions(ctx, newObj, perm) {
			if user.IsAnonymous() {
				return nil, errAuthorizationRequired
			}
			return nil, errors.Wrap(errAccessForbidden, objName+` [`+strings.Trim(perm[len(objName):], `.`)+`]`)
		}
	}

	return next(ctx)
}

// here we need to check that object belongs to the user or have manager access
// as it's the basic level of access to any object
func ownedObject(ctx context.Context, obj any, user *model.User, acc *model.Account) any {
	switch obj.(type) {
	case nil:
		return nil
	case *model.Account, model.Account:
		return &model.Account{ID: acc.ID, Admins: []uint64{user.ID}}
	case *model.User, model.User:
		return &model.User{ID: user.ID}
	}

	// Get object struct type value
	tp := reflect.TypeOf(obj).Elem()
	for tp.Kind() == reflect.Ptr || tp.Kind() == reflect.Interface {
		tp = tp.Elem()
	}
	if tp.Kind() != reflect.Struct {
		return obj
	}

	// Create new object with the same type
	newObj := reflect.New(tp).Interface()

	// Set account and user owner IDs
	if setter, ok := newObj.(accountOwnerSetter); ok {
		setter.SetAccountOwnerID(acc.ID)
	} else {
		_ = gocast.SetStructFieldValue(ctx, newObj, `AccountID`, acc.ID)
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerAccountID`, acc.ID)
	}

	// Set user owner ID
	if setter, ok := newObj.(userOwnerSetter); ok {
		setter.SetUserOwnerID(user.ID)
	} else {
		_ = gocast.SetStructFieldValue(ctx, newObj, `UserID`, user.ID)
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerID`, user.ID)
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerUserID`, user.ID)
	}
	return newObj
}

// ObjectByPermissionName returns object by permission name
func objectByPermissionName(mng *permissions.Manager, name string) (string, any) {
	for i := len(name) - 1; i > 0; i-- {
		if name[i] == '.' {
			if obj := mng.ObjectByName(name[:i]); obj != nil {
				return name[:i], obj
			}
		}
	}
	return "", nil
}
