package models

import (
	"context"
	"strings"

	mrbac "github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

func FromRBACPermissionModel(perm mrbac.Permission) *RBACPermission {
	type rname interface {
		ResourceName() string
	}
	var (
		name    = perm.Name()
		objName string
		access  string
	)
	if r, ok := perm.(rname); ok {
		objName = r.ResourceName()
		name = name[len(objName)+1:]
		if strings.HasSuffix(name, `.owner`) || strings.HasSuffix(name, `.account`) ||
			strings.HasSuffix(name, `.all`) || strings.HasSuffix(name, `.system`) {
			access = name[strings.LastIndex(name, `.`)+1:]
			name = name[:len(name)-len(access)-1]
		}
	}
	return &RBACPermission{
		Fullname:    perm.Name(),
		Name:        name,
		Object:      objName,
		Access:      access,
		Description: s2ptr(perm.Description()),
	}
}

// FromRBACPermissionModelList converts model list to local model list
func FromRBACPermissionModelList(perms []mrbac.Permission) []*RBACPermission {
	return xtypes.SliceApply(perms, FromRBACPermissionModel).
		Sort(func(a, b *RBACPermission) bool {
			if a.Object < b.Object {
				return true
			}
			if a.Object != b.Object {
				return false
			}
			if a.Name < b.Name {
				return true
			}
			return a.Name == b.Name && a.Access < b.Access
		})
}

// FromRBACRoleModel to local graphql model
func FromRBACRoleModel(ctx context.Context, role *model.Role) *RBACRole {
	perms := permissions.FromContext(ctx).Permissions(role.PermissionPatterns...)
	return &RBACRole{
		ID:    role.ID,
		Name:  role.Name,
		Title: role.Title,

		Description: s2ptr(role.Description),

		Context: types.MustNullableJSONFrom(role.Context.Data),

		Permissions:        FromRBACPermissionModelList(perms),
		PermissionPatterns: role.PermissionPatterns,

		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: DeletedAt(role.DeletedAt),
	}
}

// FromRBACRoleModelList converts model list to local model list
func FromRBACRoleModelList(ctx context.Context, list []*model.Role) []*RBACRole {
	return xtypes.SliceApply(list, func(val *model.Role) *RBACRole {
		return FromRBACRoleModel(ctx, val)
	})
}

func (fl *RBACRoleListFilter) Filter() *rbac.Filter {
	if fl == nil {
		return nil
	}
	return &rbac.Filter{
		ID:    fl.ID,
		Names: fl.Name,
	}
}

func (ol *RBACRoleListOrder) Order() *rbac.Order {
	if ol == nil {
		return nil
	}
	return &rbac.Order{
		ID:    ol.ID.AsOrder(),
		Name:  ol.Name.AsOrder(),
		Title: ol.Title.AsOrder(),
	}
}
