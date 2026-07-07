package graphql

import (
	"context"
	"reflect"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
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

type userACLSubject struct {
	userModels.UserBase
}

func (userACLSubject) TableName() string        { return "account_user" }
func (userACLSubject) RBACResourceName() string { return "user" }

func ownedObject(ctx context.Context, obj any, usr user.Model, acc account.Model) any {
	switch rbacObjectName(obj) {
	case "account":
		return account.ACLAccountStub(acc.GetID(), usr.GetID())
	case "user":
		return userACLSubject{UserBase: userModels.UserBase{ID: usr.GetID()}}
	}

	tp := reflect.TypeOf(obj).Elem()
	for tp.Kind() == reflect.Ptr || tp.Kind() == reflect.Interface {
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

func rbacObjectName(obj any) string {
	if obj == nil {
		return ""
	}
	if checker, ok := obj.(interface{ RBACResourceName() string }); ok {
		return checker.RBACResourceName()
	}
	return ""
}
