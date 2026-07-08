package appinit

import (
	"context"
	"strings"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	daModels "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/option"
	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
)

var (
	crudPermissions = []string{
		acl.PermView, acl.PermList, acl.PermCount, acl.PermUpdate, acl.PermCreate, acl.PermDelete,
	}
	crudPermissionsWithApprove = append(crudPermissions, acl.PermApprove, acl.PermReject)
)

const (
	PermAccountRegister = `account.register`
	PermPermissionList  = `permission.list`
	PermUserPassReset   = `password.reset`
	PermUserPassSet     = `password.set`
)

// InitModelPermissions models
func InitModelPermissions(pm *permissions.Manager, deps *Deps) {
	acl.InitModelPermissions(pm,
		&domain.User{},
		&rbacModels.Role{},
		&authclient.AuthClient{},
		&domain.Account{},
		&domain.AccountMember{},
		&socialaccount.AccountSocialSession{},
		&socialaccount.AccountSocial{},
		&historylog.HistoryAction{},
		&option.Option{},
		&daModels.DirectAccessToken{},
	)

	_ = pm.RegisterNewOwningPermissions(&domain.User{}, append(crudPermissions, PermUserPassReset, PermUserPassSet))

	_ = pm.RegisterNewOwningPermissions(&domain.Account{}, crudPermissionsWithApprove, rbac.WithCustomCheck(func(ctx context.Context, resource any, perm rbac.Permission) bool {
		return accountCustomCheck(ctx, resource, perm, deps)
	}))
	_ = pm.RegisterNewPermission(nil, PermAccountRegister, rbac.WithoutCustomCheck)

	_ = pm.RegisterNewOwningPermissions(&rbacModels.Role{}, crudPermissions)
	_ = pm.RegisterNewPermission(&rbacModels.Role{}, `check`,
		rbac.WithDescription("Check role permissions is assigned to the user"))
	_ = pm.RegisterNewPermission(nil, PermPermissionList, rbac.WithDescription("List all permissions"))

	_ = pm.RegisterNewOwningPermissions(&authclient.AuthClient{}, crudPermissions)

	_ = pm.RegisterNewOwningPermissions(&domain.AccountMember{}, crudPermissionsWithApprove)
	_ = pm.RegisterNewPermissions(&domain.AccountMember{}, []string{`roles.set.account`, `roles.set.all`, `invite`})

	_ = pm.RegisterNewOwningPermissions(&historylog.HistoryAction{}, []string{acl.PermView, acl.PermList, acl.PermCount})
	_ = pm.RegisterNewOwningPermissions(&option.Option{}, []string{acl.PermGet, acl.PermSet, acl.PermList, acl.PermCount})
	_ = pm.RegisterNewOwningPermissions(&daModels.DirectAccessToken{}, []string{acl.PermGet, acl.PermList, acl.PermCount, acl.PermCreate, acl.PermDelete})

	pm.RegisterRole(context.Background(),
		rbac.MustNewRole(session.AnonymousDefaultRole,
			rbac.WithDescription("Anonymous user role"),
			rbac.WithPermissions(
				`user.view.owner`, `user.list.owner`, `user.count.owner`,
				`user.password.reset.owner`, `user.password.set.owner`, PermAccountRegister,
				`account.view.owner`, `account.list.owner`, `account.count.owner`,
				`directaccesstoken.view.owner`, `directaccesstoken.list.owner`, `directaccesstoken.count.owner`,
				`role.check`,
			),
		),
		rbac.MustNewRole(permissions.DefaultRole,
			rbac.WithDescription("Default user role"),
			rbac.WithPermissions(
				`user.view.owner`, `user.list.owner`, `user.count.owner`,
				`user.password.reset.owner`, `user.password.set.owner`, PermAccountRegister,
				`account.view.owner`, `account.list.owner`, `account.count.owner`,
				`directaccesstoken.view.owner`, `directaccesstoken.list.owner`, `directaccesstoken.count.owner`, `directaccesstoken.create.owner`, `directaccesstoken.update.owner`, `directaccesstoken.delete.owner`,
				`role.check`,
			),
		),
	)
}

func accountCustomCheck(ctx context.Context, resource any, perm rbac.Permission, deps *Deps) bool {
	if strings.HasSuffix(perm.Name(), `.system`) || strings.HasSuffix(perm.Name(), `.all`) {
		return true
	}
	acc, _ := resource.(*domain.Account)
	userObj := session.User(ctx)
	if acc.IsOwnerUser(userObj.GetID()) {
		return true
	}
	if acc.ID > 0 {
		if perm.MatchPermissionPattern(`*.{view|list|count}.*`) {
			return deps.MemberRepo.IsMember(ctx, userObj.GetID(), acc.ID)
		}
		return deps.MemberRepo.IsAdmin(ctx, userObj.GetID(), acc.ID)
	}
	return false
}
