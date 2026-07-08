package account

import (
	"context"
	"time"

	"github.com/demdxx/rbac"
	"github.com/geniusrabbit/blaze-api/pkg/auth"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// Model is the compile-time constraint for core account repository/usecase operations.
// RBACResourceName/TableName stay on concrete structs (used by ACL/GORM), not on this interface.
type Model interface {
	auth.IsNillable
	GetID() uint64

	IsAnonymous() bool
	OwnerAccountID() uint64

	GetApprove() pkgModels.ApproveStatus
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time

	NewWithIDs(id uint64, adminUserIDs ...uint64) Model

	CheckPermissions(ctx context.Context, resource any, patterns ...string) bool
	CheckedPermissions(ctx context.Context, resource any, patterns ...string) rbac.Permission
	IsAdminUser(userID uint64) bool
	IsOwnerUser(userID uint64) bool
	ExtendAdminUsers(ids ...uint64)
	SetPermissions(perm PermissionChecker)
	ExtendPermissions(perm PermissionChecker)
	PermissionsChecker() PermissionChecker
	ListPermissions(patterns ...string) []rbac.Permission
	HasPermission(patterns ...string) bool
}

// MemberModel is the compile-time constraint for account member join records.
type MemberModel interface {
	GetID() uint64
	OwnerAccountID() uint64
}
