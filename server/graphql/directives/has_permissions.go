package directives

import (
	"context"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
	"github.com/pkg/errors"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
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
func HasPermissions[T user.Model, A account.Model](ctx context.Context, obj any, next graphql.Resolver, perms []string) (any, error) {
	user, accountObj := session.UserAccount(ctx)
	pm := permissions.FromContext(ctx)

	if accountObj == nil {
		return nil, errors.Wrap(errAuthorizationRequired, `no correct account`)
	}

	if len(perms) < 1 {
		return nil, errAccessForbidden
	}

	for _, perm := range perms {
		objName, obj := objectByPermissionName(pm, perm)
		newObj := ownedObject(ctx, obj, user, accountObj, objName)
		if !accountObj.CheckPermissions(ctx, newObj, perm) {
			if user.IsAnonymous() {
				return nil, errAuthorizationRequired
			}
			return nil, errors.Wrap(errAccessForbidden, objName+` [`+strings.Trim(perm[len(objName):], `.`)+`]`)
		}
	}

	return next(ctx)
}

func ownedObject(ctx context.Context, obj any, usr user.Model, acc account.Model, objName string) any {
	switch {
	case obj == nil:
		return nil
	case objName == "account":
		return acc.NewWithIDs(acc.GetID(), usr.GetID())
	case objName == "user":
		return usr.NewWithID(usr.GetID())
	}

	tp := reflect.TypeOf(obj).Elem()
	for tp.Kind() == reflect.Pointer || tp.Kind() == reflect.Interface {
		tp = tp.Elem()
	}
	if tp.Kind() != reflect.Struct {
		return obj
	}

	newObj := reflect.New(tp).Interface()

	if setter, ok := newObj.(accountOwnerSetter); ok {
		setter.SetAccountOwnerID(acc.GetID())
	} else {
		_ = gocast.SetStructFieldValue(ctx, newObj, `AccountID`, acc.GetID())
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerAccountID`, acc.GetID())
	}

	if setter, ok := newObj.(userOwnerSetter); ok {
		setter.SetUserOwnerID(usr.GetID())
	} else {
		_ = gocast.SetStructFieldValue(ctx, newObj, `UserID`, usr.GetID())
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerID`, usr.GetID())
		_ = gocast.SetStructFieldValue(ctx, newObj, `OwnerUserID`, usr.GetID())
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
