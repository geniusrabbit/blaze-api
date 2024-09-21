package graphql

import (
	"context"
	"reflect"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/model"
)

type (
	accountOwnerSetter interface {
		SetAccountOwnerID(uint64)
	}
	userOwnerSetter interface {
		SetUserOwnerID(uint64)
	}
)

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
