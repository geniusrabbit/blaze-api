package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/rbac/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// RBACRoleConnection implements collection accessor interface with pagination
type RBACRoleConnection = connectors.CollectionConnection[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]

// NewRBACRoleConnection based on query object
func NewRBACRoleConnection(ctx context.Context, rolesAccessor rbac.Usecase, filter *gqlmodels.RBACRoleListFilter, order []*gqlmodels.RBACRoleListOrder, page *gqlmodels.Page) *RBACRoleConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			opts := []rbac.QOption{filter.Filter(), page.Pagination()}
			for _, o := range order {
				if ord := o.Order(); ord != nil {
					opts = append(opts, ord)
				}
			}
			roles, err := rolesAccessor.FetchList(ctx, opts...)
			return FromRBACRoleModelList(ctx, roles), err
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
func NewRBACRoleConnectionByIDs(ctx context.Context, rolesPepo rbac.Repository, ids []uint64, order []*gqlmodels.RBACRoleListOrder) *RBACRoleConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			var (
				roles []*models.Role
				err   error
			)
			if len(ids) > 0 {
				opts := []rbac.QOption{&rbac.Filter{ID: ids}}
				for _, o := range order {
					if ord := o.Order(); ord != nil {
						opts = append(opts, ord)
					}
				}
				roles, err = rolesPepo.FetchList(ctx, opts...)
			}
			return FromRBACRoleModelList(ctx, roles), err
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
