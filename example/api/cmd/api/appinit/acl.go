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

// InitModelPermissions models
func InitModelPermissions(pm *permissions.Manager) {
	crudModels := []any{
		&model.User{},
		&model.Role{},
		&model.AuthClient{},
	}

	acl.InitModelPermissions(pm,
		append(crudModels,
			&model.Account{},
			&model.AccountMember{},
			&model.AccountSocialSession{},
			&model.AccountSocial{},
			&model.HistoryAction{},
			&model.Option{},
		)...,
	)

	// Register basic models CRUD permissions
	for _, model := range crudModels {
		_ = pm.RegisterNewOwningPermissions(model, []string{"view", "list", "count", "update", "create"})
	}

	// Register basic models CRUD permissions for Account with member checks
	_ = pm.RegisterNewOwningPermissions(&model.Account{},
		[]string{"view", "list", "count", "update", "create", "delete"},
		rbac.WithCustomCheck(func(ctx context.Context, resource any, perm rbac.Permission) bool {
			account, _ := resource.(*model.Account)
			user := session.User(ctx)
			return account.IsOwnerUser(user.ID) || repository.New().IsMember(ctx, user, account)
		}),
	)

	// Extend user permissions
	_ = pm.RegisterNewPermissions(&model.User{}, []string{"reset_password"})
	_ = pm.RegisterNewPermissions(&model.User{}, []string{"password.reset"})
	_ = pm.RegisterNewOwningPermissions(&model.User{}, []string{"reset_password"})

	// Extend account permissions
	_ = pm.RegisterNewPermission(nil, "account.register", rbac.WithoutCustomCheck)

	// Register basic permissions for the Option model
	_ = pm.RegisterNewOwningPermissions(&model.Option{}, []string{"get", "set", "list"})

	// Register RBAC permissions
	_ = pm.RegisterNewPermission(nil, "permission.list")

	// Register anonymous role and fill permissions for it
	pm.RegisterRole(context.Background(),
		rbac.MustNewRole(session.AnonymousDefaultRole, rbac.WithPermissions(
			`user.view.owner`, `user.reset_password`, `user.reset_password.*`, `user.password.reset`,
			`account.view.owner`, `account.register`,
		)),
	)
}
