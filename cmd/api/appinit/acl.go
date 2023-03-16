package appinit

import (
	"github.com/demdxx/rbac"
	"github.com/geniusrabbit/api-template-base/internal/acl"
	"github.com/geniusrabbit/api-template-base/internal/permissions"
	"github.com/geniusrabbit/api-template-base/model"
)

// InitModelPermissions models
func InitModelPermissions(pm *permissions.Manager) {
	acl.InitModelPermissions(pm,
		&model.User{},
		&model.Account{},
		&model.AccountMember{},
		&model.Role{},
		&model.AuthClient{},
	)
	pm.RegisterRole(rbac.MustNewRole("anonymous", rbac.WithSubPermissins(
		pm.MustNewRosourcePermission("view", (*model.User)(nil)),
		pm.MustNewRosourcePermission("view", (*model.Account)(nil)),
	)))
}
