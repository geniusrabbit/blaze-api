package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	"github.com/geniusrabbit/api-template-base/internal/repository/authclient"
	"github.com/geniusrabbit/api-template-base/internal/repository/rbac"
	"github.com/geniusrabbit/api-template-base/internal/repository/user"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
)

// AccountConnection implements collection accessor interface with pagination.
type AccountConnection = CollectionConnection[gqlmodels.Account, gqlmodels.AccountEdge]

// NewAccountConnection based on query object
func NewAccountConnection(ctx context.Context, accountsAccessor account.Usecase, filter *gqlmodels.AccountListFilter, order *gqlmodels.AccountListOrder, page *gqlmodels.Page) *AccountConnection {
	return NewCollectionConnection[gqlmodels.Account, gqlmodels.AccountEdge](ctx, &DataAccessorFunc[gqlmodels.Account, gqlmodels.AccountEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Account, error) {
			accounts, err := accountsAccessor.FetchList(ctx, filter.Filter(), page.Pagination())
			return gqlmodels.FromAccountModelList(accounts), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return accountsAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.Account) *gqlmodels.AccountEdge {
			return &gqlmodels.AccountEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}

// RBACRoleConnection implements collection accessor interface with pagination.
type RBACRoleConnection = CollectionConnection[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]

// NewRBACRoleConnection based on query object
func NewRBACRoleConnection(ctx context.Context, rolesAccessor rbac.Usecase, filter *gqlmodels.RBACRoleListFilter, order *gqlmodels.RBACRoleListOrder, page *gqlmodels.Page) *RBACRoleConnection {
	return NewCollectionConnection[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge](ctx, &DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			roles, err := rolesAccessor.FetchList(ctx, filter.Filter(), page.Pagination())
			return gqlmodels.FromRBACRoleModelList(roles), err
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

// AuthClientConnection implements collection accessor interface with pagination.
type AuthClientConnection = CollectionConnection[gqlmodels.AuthClient, gqlmodels.AuthClientEdge]

// NewAuthClientConnection based on query object
func NewAuthClientConnection(ctx context.Context, authClientsAccessor authclient.Usecase, page *gqlmodels.Page) *AuthClientConnection {
	return NewCollectionConnection[gqlmodels.AuthClient, gqlmodels.AuthClientEdge](ctx, &DataAccessorFunc[gqlmodels.AuthClient, gqlmodels.AuthClientEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.AuthClient, error) {
			clients, err := authClientsAccessor.FetchList(ctx, nil)
			return gqlmodels.FromAuthClientModelList(clients), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return authClientsAccessor.Count(ctx, nil)
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.AuthClient) *gqlmodels.AuthClientEdge {
			return &gqlmodels.AuthClientEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}

// UserConnection implements collection accessor interface with pagination.
type UserConnection = CollectionConnection[gqlmodels.User, gqlmodels.UserEdge]

// NewUserConnection based on query object
func NewUserConnection(ctx context.Context, usersAccessor user.Usecase, filter *gqlmodels.UserListFilter, order *gqlmodels.UserListOrder, page *gqlmodels.Page) *UserConnection {
	return NewCollectionConnection[gqlmodels.User, gqlmodels.UserEdge](ctx, &DataAccessorFunc[gqlmodels.User, gqlmodels.UserEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.User, error) {
			users, err := usersAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromUserModelList(users), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return usersAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.User) *gqlmodels.UserEdge {
			return &gqlmodels.UserEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
