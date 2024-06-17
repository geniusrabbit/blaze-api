package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/account"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AccountConnection implements collection accessor interface with pagination
type AccountConnection = CollectionConnection[gqlmodels.Account, gqlmodels.AccountEdge]

// NewAccountConnection based on query object
func NewAccountConnection(ctx context.Context, accountsAccessor account.Usecase, filter *gqlmodels.AccountListFilter, order *gqlmodels.AccountListOrder, page *gqlmodels.Page) *AccountConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.Account, gqlmodels.AccountEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Account, error) {
			accounts, err := accountsAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
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
