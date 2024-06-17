package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// SocialAccountConnection implements collection accessor interface with pagination
type SocialAccountConnection = CollectionConnection[gqlmodels.SocialAccount, gqlmodels.SocialAccountEdge]

// NewSocialAccountConnection based on query object
func NewSocialAccountConnection(ctx context.Context, accountsAccessor socialaccount.Usecase, filter *gqlmodels.SocialAccountListFilter, order *gqlmodels.SocialAccountListOrder, page *gqlmodels.Page) *SocialAccountConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.SocialAccount, gqlmodels.SocialAccountEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.SocialAccount, error) {
			accounts, err := accountsAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromSocialAccountModelList(accounts), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return accountsAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.SocialAccount) *gqlmodels.SocialAccountEdge {
			return &gqlmodels.SocialAccountEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
