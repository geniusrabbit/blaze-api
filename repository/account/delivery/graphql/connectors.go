package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AccountConnection implements collection accessor interface with pagination
type AccountConnection = connectors.CollectionConnection[gqlmodels.Account, gqlmodels.AccountEdge]

// NewAccountConnection based on query object
func NewAccountConnection(ctx context.Context, accountsAccessor account.Usecase, filter *gqlmodels.AccountListFilter, order *gqlmodels.AccountListOrder, page *gqlmodels.Page) *AccountConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.Account, gqlmodels.AccountEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Account, error) {
			accounts, err := accountsAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return FromAccountModelList(accounts), err
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

// MemberConnection implements collection accessor interface with pagination
type MemberConnection = connectors.CollectionConnection[gqlmodels.Member, gqlmodels.MemberEdge]

// NewMemberConnection based on query object
func NewMemberConnection(ctx context.Context, membersAccessor account.MemberUsecase, filter *gqlmodels.MemberListFilter, order *gqlmodels.MemberListOrder, page *gqlmodels.Page) *MemberConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.Member, gqlmodels.MemberEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Member, error) {
			members, err := membersAccessor.FetchListMembers(ctx,
				filter.Filter(), order.Order(), page.Pagination())
			return FromMemberModelList(ctx, members), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return membersAccessor.CountMembers(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.Member) *gqlmodels.MemberEdge {
			return &gqlmodels.MemberEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
