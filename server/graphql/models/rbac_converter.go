package models

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// FromRBACRoleModel to local graphql model
func FromRBACRoleModel(role *model.Role) *RBACRole {
	return &RBACRole{
		ID:        role.ID,
		Name:      role.Name,
		Title:     role.Title,
		Type:      RoleType(role.Type),
		Context:   types.MustNullableJSONFrom(role.Context.Data),
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: deletedAt(role.DeletedAt),
	}
}

// FromRBACRoleModelList converts model list to local model list
func FromRBACRoleModelList(list []*model.Role) []*RBACRole {
	roles := make([]*RBACRole, 0, len(list))
	for _, role := range list {
		roles = append(roles, FromRBACRoleModel(role))
	}
	return roles
}

func (fl *RBACRoleListFilter) Filter() *rbac.Filter {
	if fl == nil {
		return nil
	}
	return &rbac.Filter{
		ID:    fl.ID,
		Names: fl.Name,
		Types: xtypes.SliceApply[RoleType](fl.Type, func(v RoleType) model.RoleType {
			switch v {
			case RoleTypeRole:
				return model.RbacRoleType
			case RoleTypePermission:
				return model.RbacPermissionType
			}
			return model.RoleType(v)
		}),
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
		Type:  ol.Type.AsOrder(),
	}
}
