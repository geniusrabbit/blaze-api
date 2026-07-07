package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/rbac/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// RBACRoleConnection implements collection accessor interface with pagination
type RBACRoleConnection = connectors.CollectionConnection[*gqlmodels.RBACRole]

// NewRBACRoleConnection based on query object
func NewRBACRoleConnection(ctx context.Context, rolesAccessor rbac.Usecase, filter *gqlmodels.RBACRoleListFilter, order []*gqlmodels.RBACRoleListOrder, page *gqlmodels.Page) *RBACRoleConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.RBACRole]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			opts := []rbac.QOption{FromGQLFilter(filter), page.Pagination()}
			for _, o := range order {
				if ord := FromGQLOrder(o); ord != nil {
					opts = append(opts, ord)
				}
			}
			roles, err := rolesAccessor.FetchList(ctx, opts...)
			return FromRBACRoleModelList(ctx, roles), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return rolesAccessor.Count(ctx, FromGQLFilter(filter))
		},
	}, page)
}

// NewRBACRoleConnectionByIDs based on query object
func NewRBACRoleConnectionByIDs(ctx context.Context, rolesPepo rbac.Repository, ids []uint64, order []*gqlmodels.RBACRoleListOrder) *RBACRoleConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.RBACRole]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			var (
				roles []*models.Role
				err   error
			)
			if len(ids) > 0 {
				opts := []rbac.QOption{&rbac.Filter{ID: ids}}
				for _, o := range order {
					if ord := FromGQLOrder(o); ord != nil {
						opts = append(opts, ord)
					}
				}
				if roles, err = rolesPepo.FetchList(ctx, opts...); err != nil {
					return nil, err
				}
			}
			return FromRBACRoleModelList(ctx, roles), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return int64(len(ids)), nil
		},
	}, nil)
}
