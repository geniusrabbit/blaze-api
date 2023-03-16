package graphql

import (
	"context"

	"github.com/guregu/null"

	"github.com/geniusrabbit/api-template-base/internal/repository/rbac"
	"github.com/geniusrabbit/api-template-base/internal/repository/rbac/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/rbac/usecase"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	roles rbac.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		roles: usecase.NewRoleUsecase(repository.New()),
	}
}

// Role is the resolver for the Role field.
func (r *QueryResolver) Role(ctx context.Context, id uint64) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: int(role.ID),
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

// ListRoles is the resolver for the listRoles field.
func (r *QueryResolver) ListRoles(ctx context.Context,
	filter *gqlmodels.RBACRoleListFilter,
	order []*gqlmodels.RBACRoleListOrder,
	page *gqlmodels.Page) (*connectors.RBACRoleConnection, error) {
	// roles, err := r.roles.FetchList(ctx, &rbac.Filter{
	// 	ID:    gocast.Slice[uint64](filter.ID),
	// 	Names: filter.Name,
	// 	Types: gocast.Slice[model.RoleType](filter.Type),
	// })
	// if err != nil {
	// 	return nil, err
	// }
	return connectors.NewRBACRoleConnection(ctx, r.roles), nil
}

// CreateRole is the resolver for the createRole field.
func (r *QueryResolver) CreateRole(ctx context.Context, input *gqlmodels.RBACRoleInput) (*gqlmodels.RBACRolePayload, error) {
	roleObj := &model.Role{
		ParentID: null.IntFromPtr(int64Ptr(input.ParentID)),
		Name:     input.Name,
		Title:    input.Title,
		Type:     model.RoleType(input.Type),
	}
	if input.Context != nil {
		if err := roleObj.Context.SetValue(input.Context.Data); err != nil {
			return nil, err
		}
	}
	id, err := r.roles.Create(ctx, roleObj)
	if err != nil {
		return nil, err
	}
	// role, err := r.roles.Get(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }
	return &gqlmodels.RBACRolePayload{
		RoleID: int(id),
		Role:   gqlmodels.FromRBACRoleModel(roleObj),
	}, nil
}

// UpdateRole is the resolver for the updateRole field.
func (r *QueryResolver) UpdateRole(ctx context.Context, id uint64, input *gqlmodels.RBACRoleInput) (*gqlmodels.RBACRolePayload, error) {
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update object fields
	role.ParentID = null.IntFromPtr(int64Ptr(input.ParentID))
	role.Name = input.Name
	role.Title = input.Title
	role.Type = model.RoleType(input.Type)
	if input.Context != nil {
		if err := role.Context.SetValue(input.Context.Data); err != nil {
			return nil, err
		}
	}
	if err := r.roles.Update(ctx, id, role); err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: int(id),
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

// DeleteRole is the resolver for the deleteRole field.
func (r *QueryResolver) DeleteRole(ctx context.Context, id uint64, msg *string) (*gqlmodels.RBACRolePayload, error) {
	err := r.roles.Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	role, err := r.roles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.RBACRolePayload{
		RoleID: int(id),
		Role:   gqlmodels.FromRBACRoleModel(role),
	}, nil
}

func int64Ptr(v *int) *int64 {
	if v == nil {
		return nil
	}
	u := int64(*v)
	return &u
}
