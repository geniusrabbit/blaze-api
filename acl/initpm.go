package acl

import (
	"context"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/permissions"
)

type checkFnk func(ctx context.Context, resource any, perm rbac.Permission) bool

type owner interface {
	OwnerAccountID() uint64
}

type creator interface {
	CreatorUserID() uint64
}

// InitModelPermissions for particular models
func InitModelPermissions(pm *permissions.Manager, models ...any) {
	checkerFnk := commonPermissionCheck()
	for _, modelLink := range models {
		pm.RegisterObject(modelLink, checkerFnk)
	}
}

// InitModelPermissionsWithCustomCheck for particular models and extra custom check function
func InitModelPermissionsWithCustomCheck(pm *permissions.Manager, customCheck checkFnk, models ...any) {
	for _, modelLink := range models {
		pm.RegisterObject(modelLink, commonPermissionCheck(customCheck))
	}
}

func commonPermissionCheck(custCheck ...checkFnk) checkFnk {
	var customCheck checkFnk
	if len(custCheck) > 0 {
		customCheck = custCheck[0]
	}
	return func(ctx context.Context, resource any, perm rbac.Permission) bool {
		var (
			user, account = session.UserAccount(ctx)
			cover         = permExtractCover(perm)
		)

		// In case of create we don't need to check the owner because it`s don`t exists
		// or user have access to the whole `system`
		if cover == `system` || perm.Name() == PermCreate {
			return true
		}

		// Check if resource belongs to the account
		if cover == `account` && checkOwner(resource, account.ID) {
			return true
		}

		// Check if resource belongs to the specific user and account
		if checkCreator(resource, user.ID) && checkOwner(resource, account.ID) {
			return true
		}

		// Check if resource have custom check function
		if customCheck != nil {
			return customCheck(ctx, resource, perm)
		}

		// check if this is mode which no belongs to anyone
		return isEmptyOwner(resource)
	}
}

func permExtractCover(perm rbac.Permission) string {
	if ext := perm.Ext(); ext != nil {
		// We can be sure in the type because of we define it our selfs in "internal/permissions"
		if extData, _ := ext.(*permissions.ExtData); extData != nil {
			return extData.Cover
		}
	}
	return ``
}

func checkOwner(resource any, ownerID uint64) bool {
	own, _ := resource.(owner)
	return own != nil && own.OwnerAccountID() == ownerID
}

func checkCreator(resource any, creatorID uint64) bool {
	crt, _ := resource.(creator)
	return crt != nil && crt.CreatorUserID() == creatorID
}

func isEmptyOwner(resource any) bool {
	if resource == nil {
		return true
	}
	// emptyObject := false
	// res := reflectTarget(reflect.ValueOf(resource))
	// // Check if model has been saved
	// if res.Kind() == reflect.Struct {
	// 	typ := res.Type()
	// 	for i := 0; i < typ.NumField(); i++ {
	// 		if isPKField(typ.Field(i)) {
	// 			if gocast.IsEmpty(res.Field(i).Interface()) {
	// 				emptyObject = true
	// 			}
	// 			break
	// 		}
	// 	}
	// }
	own, _ := resource.(owner)
	crt, _ := resource.(creator)
	return own == nil && crt == nil
}

// func reflectTarget(r reflect.Value) reflect.Value {
// 	for reflect.Ptr == r.Kind() || reflect.Interface == r.Kind() {
// 		r = r.Elem()
// 	}
// 	return r
// }

// func isPKField(field reflect.StructField) bool {
// 	return strings.EqualFold(field.Name, "id") ||
// 		strings.EqualFold(field.Tag.Get(`db`), `id`) ||
// 		strings.Contains(field.Tag.Get(`gorm`), "primaryKey")
// }
