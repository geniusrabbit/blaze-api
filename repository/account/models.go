package account

import "github.com/geniusrabbit/blaze-api/repository/account/models"

type (
	M2MAccountMemberRole = models.M2MAccountMemberRole
	PermissionChecker    = models.PermissionChecker
)

// CtxPermissionCheckAccount is the context key for account permission checks.
var CtxPermissionCheckAccount = models.CtxPermissionCheckAccount
