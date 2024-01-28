package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
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
			roles, err := rolesAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
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

// NewRBACRoleConnectionByIDs based on query object
func NewRBACRoleConnectionByIDs(ctx context.Context, rolesPepo rbac.Repository, ids []uint64, order *gqlmodels.RBACRoleListOrder) *RBACRoleConnection {
	return NewCollectionConnection[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge](ctx, &DataAccessorFunc[gqlmodels.RBACRole, gqlmodels.RBACRoleEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.RBACRole, error) {
			roles, err := rolesPepo.FetchList(ctx, &rbac.Filter{ID: ids}, order.Order(), nil)
			return gqlmodels.FromRBACRoleModelList(roles), err
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

// HistoryActionConnection implements collection accessor interface with pagination.
type HistoryActionConnection = CollectionConnection[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]

// NewHistoryActionConnection based on query object
func NewHistoryActionConnection(ctx context.Context, historyActionsAccessor historylog.Usecase, filter *historylog.Filter, order *historylog.Order, page *gqlmodels.Page) *HistoryActionConnection {
	return NewCollectionConnection[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge](ctx, &DataAccessorFunc[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.HistoryAction, error) {
			historyActions, err := historyActionsAccessor.FetchList(ctx, filter, order, page.Pagination())
			return gqlmodels.FromHistoryActionModelList(historyActions), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return historyActionsAccessor.Count(ctx, filter)
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.HistoryAction) *gqlmodels.HistoryActionEdge {
			return &gqlmodels.HistoryActionEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}

// OptionConnection implements collection accessor interface with pagination.
type OptionConnection = CollectionConnection[gqlmodels.Option, gqlmodels.OptionEdge]

// NewOptionConnection based on query object
func NewOptionConnection(ctx context.Context, optionsAccessor option.Usecase, filter *gqlmodels.OptionListFilter, order *gqlmodels.OptionListOrder, page *gqlmodels.Page) *OptionConnection {
	return NewCollectionConnection[gqlmodels.Option, gqlmodels.OptionEdge](ctx, &DataAccessorFunc[gqlmodels.Option, gqlmodels.OptionEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Option, error) {
			options, err := optionsAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromOptionModelList(options), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return optionsAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.Option) *gqlmodels.OptionEdge {
			return &gqlmodels.OptionEdge{
				Cursor: gocast.Str(obj.Name),
				Node:   obj,
			}
		},
	}, page)
}
