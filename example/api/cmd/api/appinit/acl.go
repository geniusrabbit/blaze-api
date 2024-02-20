package appinit

import (
	"context"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/acl"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/permissions"
)

// InitModelPermissions models
func InitModelPermissions(pm *permissions.Manager) {
	crudModels := []any{
		&model.User{},
		&model.Account{},
		&model.Role{},
		&model.AuthClient{},
	}

	acl.InitModelPermissions(pm,
		append(crudModels,
			&model.HistoryAction{},
			&model.Option{},
			&model.AccountMember{},
			&model.AccountSocialSession{},
			&model.AccountSocial{},
		)...,
	)

	// Register basic models CRUD permissions
	for _, model := range crudModels {
		_ = pm.RegisterNewOwningPermissions(model, []string{"view", "list", "update", "create"})
	}

	// Extend user permissions
	_ = pm.RegisterNewPermissions(&model.User{}, []string{"reset_password"})
	_ = pm.RegisterNewOwningPermissions(&model.User{}, []string{"reset_password"})

	// Extend account permissions
	_ = pm.RegisterNewPermission(&model.Account{}, "register", rbac.WithoutCustomCheck)

	// Register basic permissions for the Option model
	_ = pm.RegisterNewOwningPermissions(&model.Option{}, []string{"get", "set", "list"})

	// Register RBAC permissions
	_ = pm.RegisterNewPermission(nil, "permission.list")

	// Register anonymous role and fill permissions for it
	pm.RegisterRole(context.Background(),
		rbac.MustNewRole(session.AnonymousDefaultRole, rbac.WithPermissions(
			`user.view.owner`, `user.reset_password.*`,
			`account.view.owner`, `account.register`,
		)),
	)
}
