package models

import (
	"context"
	"time"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// AccountBase is the minimal embeddable account (tenant anchor for Members).
type AccountBase struct {
	ID        uint64                `json:"id" gorm:"primaryKey"`
	Approve   pkgModels.ApproveStatus `json:"approved" db:"approve_status" gorm:"column:approve_status"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt gorm.DeletedAt        `json:"deleted_at"`

	Permissions PermissionChecker `json:"-" gorm:"-"`
	Admins      []uint64          `json:"-" gorm:"-"`
}

// GetID returns account ID.
func (acc *AccountBase) GetID() uint64 {
	if acc == nil {
		return 0
	}
	return acc.ID
}

// IsNil checks if the account is nil.
func (acc *AccountBase) IsNil() bool {
	return acc == nil
}

// IsAnonymous account.
func (acc *AccountBase) IsAnonymous() bool {
	return acc == nil || acc.ID == 0
}

// RBACResourceName returns the default RBAC resource name (override on consumer type).
func (acc *AccountBase) RBACResourceName() string {
	return "account"
}

// TableName returns the default database table (override on consumer type).
func (acc *AccountBase) TableName() string {
	return "account_base"
}

// IsAdminUser reports whether userID is an admin of this account.
func (acc *AccountBase) IsAdminUser(userID uint64) bool {
	if acc == nil || len(acc.Admins) == 0 {
		return false
	}
	return xtypes.Slice[uint64](acc.Admins).Has(func(id uint64) bool { return id == userID })
}

// ExtendAdminUsers to the account.
func (acc *AccountBase) ExtendAdminUsers(ids ...uint64) {
	if acc == nil {
		return
	}
	acc.Admins = xtypes.SliceUnique(append(acc.Admins, ids...))
}

// CheckPermissions for some specific resource.
func (acc *AccountBase) CheckPermissions(ctx context.Context, resource any, patterns ...string) bool {
	if acc == nil || acc.Permissions == nil {
		return false
	}
	ctx = context.WithValue(ctx, CtxPermissionCheckAccount, acc)
	return acc.Permissions.CheckPermissions(ctx, resource, patterns...)
}

// CheckedPermissions for some specific resource.
func (acc *AccountBase) CheckedPermissions(ctx context.Context, resource any, patterns ...string) rbac.Permission {
	if acc == nil || acc.Permissions == nil {
		return nil
	}
	ctx = context.WithValue(ctx, CtxPermissionCheckAccount, acc)
	return acc.Permissions.CheckedPermissions(ctx, resource, patterns...)
}

// ListPermissions for the account.
func (acc *AccountBase) ListPermissions(patterns ...string) []rbac.Permission {
	if acc == nil || acc.Permissions == nil {
		return nil
	}
	return acc.Permissions.Permissions(patterns...)
}

// HasPermission for the account.
func (acc *AccountBase) HasPermission(patterns ...string) bool {
	return acc.Permissions.HasPermission(patterns...)
}

// OwnerAccountID returns the account ID which belongs the object.
func (acc *AccountBase) OwnerAccountID() uint64 {
	if acc == nil {
		return 0
	}
	return acc.ID
}

// IsOwnerUser of the account.
func (acc *AccountBase) IsOwnerUser(userID uint64) bool {
	return acc.IsAdminUser(userID)
}

// PermissionsChecker returns the account permission checker.
func (acc *AccountBase) PermissionsChecker() PermissionChecker {
	if acc == nil {
		return nil
	}
	return acc.Permissions
}

// SetPermissions of the account for the user.
func (acc *AccountBase) SetPermissions(perm PermissionChecker) {
	if acc == nil {
		return
	}
	if perm == nil {
		acc.Permissions = nil
		return
	}
	acc.Permissions = perm
}

// ExtendPermissions of the account for the user.
func (acc *AccountBase) ExtendPermissions(perm PermissionChecker) {
	if acc == nil || perm == nil {
		return
	}
	switch prev := acc.Permissions.(type) {
	case groupPermissionChecker:
		prev = append(prev, perm)
		acc.Permissions = prev
	case nil:
		acc.Permissions = perm
	default:
		acc.Permissions = groupPermissionChecker{prev, perm}
	}
}

// SetApprove sets approval status.
func (acc *AccountBase) SetApprove(status pkgModels.ApproveStatus) {
	if acc != nil {
		acc.Approve = status
	}
}

// GetApprove returns approval status.
func (acc *AccountBase) GetApprove() pkgModels.ApproveStatus {
	if acc == nil {
		return pkgModels.UndefinedApproveStatus
	}
	return acc.Approve
}

// GetCreatedAt returns created timestamp.
func (acc *AccountBase) GetCreatedAt() time.Time {
	if acc == nil {
		return time.Time{}
	}
	return acc.CreatedAt
}

// GetUpdatedAt returns updated timestamp.
func (acc *AccountBase) GetUpdatedAt() time.Time {
	if acc == nil {
		return time.Time{}
	}
	return acc.UpdatedAt
}

// SetID sets account primary key.
func (acc *AccountBase) SetID(id uint64) {
	if acc != nil {
		acc.ID = id
	}
}
