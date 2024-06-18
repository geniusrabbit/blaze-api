package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// DirectAccessTokenConnection implements collection accessor interface with pagination
type DirectAccessTokenConnection = CollectionConnection[gqlmodels.DirectAccessToken, gqlmodels.DirectAccessTokenEdge]

// NewDirectAccessTokenConnection based on query object
func NewDirectAccessTokenConnection(ctx context.Context, directAccessTokenAccessor directaccesstoken.Usecase, filter *gqlmodels.DirectAccessTokenListFilter, order *gqlmodels.DirectAccessTokenListOrder, page *gqlmodels.Page, fnPrep func(*model.DirectAccessToken) *model.DirectAccessToken) *DirectAccessTokenConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.DirectAccessToken, gqlmodels.DirectAccessTokenEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.DirectAccessToken, error) {
			directAccessTokens, err := directAccessTokenAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			if fnPrep != nil {
				for i, token := range directAccessTokens {
					directAccessTokens[i] = fnPrep(token)
				}
			}
			return gqlmodels.FromDirectAccessTokenModelList(directAccessTokens), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return directAccessTokenAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.DirectAccessToken) *gqlmodels.DirectAccessTokenEdge {
			return &gqlmodels.DirectAccessTokenEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
