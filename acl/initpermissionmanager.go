package acl

import (
	"context"
	"reflect"
	"strings"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/permissions"
)

type owner interface {
	OwnerAccountID() uint64
}

type creator interface {
	CreatorUserID() uint64
}

// InitModelPermissions for particular models
func InitModelPermissions(pm *permissions.Manager, models ...any) {
	for _, modelLink := range models {
		pm.RegisterObject(modelLink, commonPermissionCheck)
	}
}

func commonPermissionCheck(ctx context.Context, resource any, perm rbac.Permission) bool {
	var (
		user, account = session.UserAccount(ctx)
		extData       *permissions.ExtData
	)
	// // Exclude all anonimous users
	// if account.IsAnonimous() {
	// 	return false
	// }
	// In case of create we don't need to check the owner because it`s don`t exists
	if perm.Name() == PermCreate {
		return true
	}
	// Check if resource belongs to the account
	// @TODO: check only for account Cover type
	if checkOwnerOrDefault(resource, account.ID, false) {
		return true
	}
	if ext := perm.Ext(); ext != nil {
		// We can be sure in the type because of we define it our selfs in "internal/permissions"
		extData = ext.(*permissions.ExtData)
	}
	// Check that this cover the whole system
	if extData != nil && extData.Cover == `system` {
		return true
	}
	// check if this is mode which no belongs to anyone
	// or in case if the user is the owner of the object and the same account ID
	return isEmptyOwner(resource) ||
		(checkCreatorOrDefault(resource, user.ID, false) &&
			checkOwnerOrDefault(resource, account.ID, true))
}

func checkOwnerOrDefault(resource any, ownerID uint64, def bool) bool {
	own, _ := resource.(owner)
	return (own != nil && own.OwnerAccountID() == ownerID) || def
}

func checkCreatorOrDefault(resource any, creatorID uint64, def bool) bool {
	crt, _ := resource.(creator)
	return (crt != nil && crt.CreatorUserID() == creatorID) || def
}

func isEmptyOwner(resource any) bool {
	if resource == nil {
		return true
	}
	emptyObject := false
	res := reflectTarget(reflect.ValueOf(resource))
	// Check if model has been saved
	if res.Kind() == reflect.Struct {
		typ := res.Type()
		for i := 0; i < typ.NumField(); i++ {
			if isPKField(typ.Field(i)) {
				if gocast.IsEmpty(res.Field(i).Interface()) {
					emptyObject = true
				}
				break
			}
		}
	}
	own, _ := resource.(owner)
	crt, _ := resource.(creator)
	return emptyObject &&
		(own == nil || own.OwnerAccountID() == 0) &&
		(crt == nil || crt.CreatorUserID() == 0)
}

func reflectTarget(r reflect.Value) reflect.Value {
	for reflect.Ptr == r.Kind() || reflect.Interface == r.Kind() {
		r = r.Elem()
	}
	return r
}

func isPKField(field reflect.StructField) bool {
	return strings.EqualFold(field.Name, "id") ||
		strings.EqualFold(field.Tag.Get(`db`), `id`) ||
		strings.Contains(field.Tag.Get(`gorm`), "primaryKey")
}
