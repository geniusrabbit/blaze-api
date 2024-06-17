package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// RBACRoleConnection implements collection accessor interface with pagination
type RBACRoleConnection = CollectionConnection[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]

// NewRBACRoleConnection based on query object
func NewRBACRoleConnection(ctx context.Context, rolesAccessor rbac.Usecase, filter *gqlmodels.RBACRoleListFilter, order *gqlmodels.RBACRoleListOrder, page *gqlmodels.Page) *RBACRoleConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			roles, err := rolesAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromRBACRoleModelList(ctx, roles), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return rolesAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.RBACRole) *gqlmodels.RBACRoleEdge {
			return &gqlmodels.RBACRoleEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}

// NewRBACRoleConnectionByIDs based on query object
func NewRBACRoleConnectionByIDs(ctx context.Context, rolesPepo rbac.Repository, ids []uint64, order *gqlmodels.RBACRoleListOrder) *RBACRoleConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			var (
				roles []*model.Role
				err   error
			)
			if len(ids) > 0 {
				roles, err = rolesPepo.FetchList(ctx, &rbac.Filter{ID: ids}, order.Order(), nil)
			}
			return gqlmodels.FromRBACRoleModelList(ctx, roles), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return int64(len(ids)), nil
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.RBACRole) *gqlmodels.RBACRoleEdge {
			return &gqlmodels.RBACRoleEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, nil)
}
