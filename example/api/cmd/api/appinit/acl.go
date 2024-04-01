package appinit

import (
	"context"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/acl"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account/repository"
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
		&model.User{},
		&model.Role{},
		&model.AuthClient{},
		&model.Account{},
		&model.AccountMember{},
		&model.AccountSocialSession{},
		&model.AccountSocial{},
		&model.HistoryAction{},
		&model.Option{},
	)

	// Register user permissions
	_ = pm.RegisterNewOwningPermissions(&model.User{}, append(crudPermissions, PermUserPassReset, PermUserPassSet))

	// Register basic models CRUD permissions for Account with member checks
	_ = pm.RegisterNewOwningPermissions(&model.Account{}, crudPermissionsWithApprove, rbac.WithCustomCheck(accountCustomCheck))
	_ = pm.RegisterNewPermission(nil, PermAccountRegister, rbac.WithoutCustomCheck)

	// Register basic roles permissions
	_ = pm.RegisterNewOwningPermissions(&model.Role{}, crudPermissions)
	_ = pm.RegisterNewPermission(nil, PermPermissionList)

	// Register basic permissions for the AuthClient model
	_ = pm.RegisterNewOwningPermissions(&model.AuthClient{}, crudPermissions)

	// Register basic permissions for the AccountMember model
	_ = pm.RegisterNewOwningPermissions(&model.AccountMember{}, crudPermissionsWithApprove)
	_ = pm.RegisterNewPermissions(&model.AccountMember{}, []string{`roles.set.account`, `roles.set.all`, `invite`})

	// Register basic permissions for the HistoryAction model
	_ = pm.RegisterNewOwningPermissions(&model.HistoryAction{}, []string{acl.PermView, acl.PermList, acl.PermCount})

	// Register basic permissions for the Option model
	_ = pm.RegisterNewOwningPermissions(&model.Option{}, []string{acl.PermGet, acl.PermSet, acl.PermList, acl.PermCount})

	// Register anonymous role and fill permissions for it
	pm.RegisterRole(context.Background(),
		rbac.MustNewRole(session.AnonymousDefaultRole, rbac.WithPermissions(
			`user.view.owner`, `user.list.owner`, `user.count.owner`,
			`user.password.reset.owner`, `user.password.set.owner`, PermAccountRegister,
			`account.view.owner`, `account.list.owner`, `account.count.owner`,
		)),
	)
}

func accountCustomCheck(ctx context.Context, resource any, perm rbac.Permission) bool {
	account, _ := resource.(*model.Account)
	user := session.User(ctx)
	if account.IsOwnerUser(user.ID) {
		return true
	}
	repo := repository.New()
	if perm.MatchPermissionPattern(`*.{view|list|count}.*`) {
		return repo.IsMember(ctx, user.ID, account.ID)
	}
	return repo.IsAdmin(ctx, user.ID, account.ID)
}
