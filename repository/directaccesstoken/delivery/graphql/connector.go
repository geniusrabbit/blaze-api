package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// DirectAccessTokenConnection is a GraphQL collection connection for direct access tokens with pagination support.
type DirectAccessTokenConnection = connectors.CollectionConnection[*gqlmodels.DirectAccessToken]

// NewDirectAccessTokenConnection creates a new collection connection for direct access tokens based on the provided query parameters.
//
// Parameters:
//   - ctx: context for the operation
//   - directAccessTokenAccessor: usecase for accessing direct access token data
//   - filter: GraphQL filter criteria
//   - order: GraphQL sort order
//   - page: pagination parameters
//   - fnPrep: optional preparation function to transform tokens before returning
func NewDirectAccessTokenConnection(
	ctx context.Context,
	directAccessTokenAccessor directaccesstoken.Usecase,
	filter *gqlmodels.DirectAccessTokenListFilter,
	order []*gqlmodels.DirectAccessTokenListOrder,
	page *gqlmodels.Page,
	fnPrep func(*models.DirectAccessToken) *models.DirectAccessToken,
) *DirectAccessTokenConnection {
	return connectors.NewCollectionConnection(
		ctx,
		&connectors.DataAccessorFunc[*gqlmodels.DirectAccessToken]{
			// FetchDataListFunc retrieves the list of direct access tokens with applied filters, ordering, and pagination.
			FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.DirectAccessToken, error) {
				opts := []directaccesstoken.QOption{FromFilterGraphQL(filter), page.Pagination()}
				for _, o := range order {
					if ord := FromOrderGraphQL(o); ord != nil {
						opts = append(opts, ord)
					}
				}
				directAccessTokens, err := directAccessTokenAccessor.FetchList(ctx, opts...)
				if fnPrep != nil {
					for i, token := range directAccessTokens {
						directAccessTokens[i] = fnPrep(token)
					}
				}
				return FromDirectAccessTokenModelList(directAccessTokens), err
			},
			// CountDataFunc returns the total count of direct access tokens matching the filter criteria.
			CountDataFunc: func(ctx context.Context) (int64, error) {
				return directAccessTokenAccessor.Count(ctx, FromFilterGraphQL(filter))
			},
		},
		page,
	)
}
