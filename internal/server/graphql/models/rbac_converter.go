package models

import (
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/types"
	"github.com/geniusrabbit/api-template-base/model"
)

// FromRBACRoleModel to local graphql model
func FromRBACRoleModel(role *model.Role) *RBACRole {
	delTime := role.DeletedAt.Time
	return &RBACRole{
		ID:        int(role.ID),
		Name:      role.Name,
		Title:     role.Title,
		Type:      RoleType(role.Type),
		Context:   types.MustNullableJSONFrom(role.Context.Data),
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
		DeletedAt: &delTime,
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
