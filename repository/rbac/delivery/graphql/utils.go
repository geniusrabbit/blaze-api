package graphql

import (
	"context"
	"reflect"

	"github.com/demdxx/gocast/v2"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

type (
	accountOwnerSetter interface {
		SetAccountOwnerID(uint64)
	}
	userOwnerSetter interface {
		SetUserOwnerID(uint64)
	}
)

func ownedObject(ctx context.Context, obj any, user *userModels.User, acc *accountModels.Account) any {
	switch obj.(type) {
	case nil:
		return nil
	case *accountModels.Account, accountModels.Account:
		return &accountModels.Account{ID: acc.ID, Admins: []uint64{user.ID}}
	case *userModels.User, userModels.User:
		return &userModels.User{ID: user.ID}
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
