package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// SocialAccountConnection is a GraphQL connection type for paginated social account collections.
type SocialAccountConnection = connectors.CollectionConnection[gqlmodels.SocialAccount, gqlmodels.SocialAccountEdge]

// NewSocialAccountConnection creates a new paginated connection for social accounts.
// It configures data fetching, counting, and edge conversion based on the provided filter, order, and page parameters.
func NewSocialAccountConnection(
	ctx context.Context,
	accountsAccessor socialaccount.Usecase,
	filter *gqlmodels.SocialAccountListFilter,
	order []*gqlmodels.SocialAccountListOrder,
	page *gqlmodels.Page,
) *SocialAccountConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.SocialAccount, gqlmodels.SocialAccountEdge]{
		// FetchDataListFunc retrieves the paginated list of social accounts.
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.SocialAccount, error) {
			opts := []socialaccount.QOption{FromGQLFilter(filter), page.Pagination()}
			for _, o := range order {
				if ord := FromGQLOrder(o); ord != nil {
					opts = append(opts, ord)
				}
			}
			accounts, err := accountsAccessor.FetchList(ctx, opts...)
			return FromSocialAccountModelList(accounts), err
		},
		// CountDataFunc returns the total count of social accounts matching the filter.
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return accountsAccessor.Count(ctx, FromGQLFilter(filter))
		},
		// ConvertToEdgeFunc transforms a social account into a GraphQL edge with cursor.
		ConvertToEdgeFunc: func(obj *gqlmodels.SocialAccount) *gqlmodels.SocialAccountEdge {
			return &gqlmodels.SocialAccountEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
