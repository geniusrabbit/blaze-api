package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/account"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// MemberConnection implements collection accessor interface with pagination
type MemberConnection = CollectionConnection[gqlmodels.Member, gqlmodels.MemberEdge]

// NewMemberConnection based on query object
func NewMemberConnection(ctx context.Context, accountsAccessor account.Usecase, filter *gqlmodels.MemberListFilter, order *gqlmodels.MemberListOrder, page *gqlmodels.Page) *MemberConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.Member, gqlmodels.MemberEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Member, error) {
			members, err := accountsAccessor.FetchListMembers(ctx,
				filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromMemberModelList(ctx, members), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return accountsAccessor.CountMembers(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.Member) *gqlmodels.MemberEdge {
			return &gqlmodels.MemberEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
