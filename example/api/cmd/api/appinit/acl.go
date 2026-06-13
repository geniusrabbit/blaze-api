package appinit

import (
	"context"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/repository"
	daModels "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
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
func InitModelPermissions(pm *permissions.Manager) {
	// Register permission objects
	acl.InitModelPermissions(pm,
		&user.User{},
		&rbacModels.Role{},
		&model.AuthClient{},
		&account.Account{},
		&account.AccountMember{},
		&model.AccountSocialSession{},
		&model.AccountSocial{},
		&model.HistoryAction{},
		&model.Option{},
		&model.DirectAccessToken{},
	)

	// Register user permissions
	_ = pm.RegisterNewOwningPermissions(&user.User{}, append(crudPermissions, PermUserPassReset, PermUserPassSet))

	// Register basic models CRUD permissions for Account with member checks
	_ = pm.RegisterNewOwningPermissions(&account.Account{}, crudPermissionsWithApprove, rbac.WithCustomCheck(accountCustomCheck))
	_ = pm.RegisterNewPermission(nil, PermAccountRegister, rbac.WithoutCustomCheck)

	// Register basic roles permissions
	_ = pm.RegisterNewOwningPermissions(&rbacModels.Role{}, crudPermissions)
	_ = pm.RegisterNewPermission(&rbacModels.Role{}, `check`,
		rbac.WithDescription("Check role permissions is assigned to the user"))
	_ = pm.RegisterNewPermission(nil, PermPermissionList, rbac.WithDescription("List all permissions"))

	// Register basic permissions for the AuthClient model
	_ = pm.RegisterNewOwningPermissions(&model.AuthClient{}, crudPermissions)

	// Register basic permissions for the AccountMember model
	_ = pm.RegisterNewOwningPermissions(&account.AccountMember{}, crudPermissionsWithApprove)
	_ = pm.RegisterNewPermissions(&account.AccountMember{}, []string{`roles.set.account`, `roles.set.all`, `invite`})

	// Register basic permissions for the HistoryAction model
	_ = pm.RegisterNewOwningPermissions(&model.HistoryAction{}, []string{acl.PermView, acl.PermList, acl.PermCount})

	// Register basic permissions for the Option model
	_ = pm.RegisterNewOwningPermissions(&model.Option{}, []string{acl.PermGet, acl.PermSet, acl.PermList, acl.PermCount})

	// Register basic permissions for the DirectAccessToken model
	_ = pm.RegisterNewOwningPermissions(&daModels.DirectAccessToken{}, []string{acl.PermGet, acl.PermList, acl.PermCount, acl.PermCreate, acl.PermDelete})

	// Register anonymous role and fill permissions for it
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

func accountCustomCheck(ctx context.Context, resource any, perm rbac.Permission) bool {
	account, _ := resource.(*model.Account)
	user := session.User(ctx)
	if account.IsOwnerUser(user.ID) {
		return true
	}
	members := repository.NewMemberRepository()
	if perm.MatchPermissionPattern(`*.{view|list|count}.*`) {
		return members.IsMember(ctx, user.ID, account.ID)
	}
	return members.IsAdmin(ctx, user.ID, account.ID)
}
