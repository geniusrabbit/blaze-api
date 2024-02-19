package acl

import (
	"context"
	"strings"

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
		// or user have access to the whole `system` | `all` (alias for `system`)
		// or user have the permission to create the object, in that case doesn't matter who is the owner
		// becase the object is not exists yet
		if cover == `all` || cover == `system` || perm.MatchPermissionPattern("*.create.*") {
			return true
		}

		// Check if resource belongs to the account.
		// If the user have the permission to the account we can allow access
		// even if the resource is not belongs to the user
		if cover == `account` && checkOwnerAccount(resource, account.ID) == 1 {
			return true
		}

		// Check if resource belongs to the specific user and account.
		ccu := checkCreatorUser(resource, user.ID)
		coa := checkOwnerAccount(resource, account.ID)
		if (ccu == 1 && coa >= 0) || (ccu >= 0 && coa == 1) {
			return true
		}

		// Check if resource have custom check function
		if customCheck != nil {
			return customCheck(ctx, resource, perm)
		}

		// check if this is mode which no belongs to anyone.
		// Here we are expecting that user have the required permission
		// and as the object is not belongs to anyone we can allow access
		return isEmptyOwner(resource)
	}
}

func permExtractCover(perm rbac.Permission) string {
	namea := strings.Split(perm.Name(), ".")
	return namea[len(namea)-1]
}

func checkOwnerAccount(resource any, ownerID uint64) int {
	own, _ := resource.(owner)
	if own == nil {
		return 0
	}
	if own.OwnerAccountID() == ownerID {
		return 1
	}
	return -1
}

func checkCreatorUser(resource any, creatorID uint64) int {
	crt, _ := resource.(creator)
	if crt == nil {
		return 0
	}
	if crt.CreatorUserID() == creatorID {
		return 1
	}
	return -1
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
