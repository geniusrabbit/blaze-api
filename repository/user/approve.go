package user

import (
	"reflect"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

type approveStatusGetter interface {
	IsAnonymous() bool
	GetApprove() pkgModels.ApproveStatus
}

// SessionRejected reports whether obj is a non-anonymous user with approve status rejected.
// A value that does not expose GetApprove is not rejected.
func SessionRejected(obj any) bool {
	if obj == nil {
		return false
	}
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return false
	}
	getter, ok := obj.(approveStatusGetter)
	if !ok || getter.IsAnonymous() {
		return false
	}
	return getter.GetApprove().IsRejected()
}
