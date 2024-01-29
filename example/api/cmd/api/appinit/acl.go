package appinit

import (
	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/acl"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/permissions"
)

// InitModelPermissions models
func InitModelPermissions(pm *permissions.Manager) {
	acl.InitModelPermissions(pm,
		&model.User{},
		&model.Account{},
		&model.AccountMember{},
		&model.Role{},
		&model.AuthClient{},
		&model.HistoryAction{},
		&model.Option{},
	)

	// Register anonymous role and fill permissions for it
	pm.RegisterRole(rbac.MustNewRole(session.AnonymousDefaultRole, rbac.WithSubPermissins(
		pm.MustNewResourcePermission("view", (*model.User)(nil)),
		pm.MustNewResourcePermission("reset_password", (*model.User)(nil)),
		pm.MustNewResourcePermission("view", (*model.Account)(nil)),
		pm.MustNewResourcePermission("register", (*model.Account)(nil)),
		rbac.MustNewSimplePermission("account.register"),
	)))
}
