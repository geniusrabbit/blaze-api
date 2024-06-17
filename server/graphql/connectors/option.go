package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/repository/option"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// OptionConnection implements collection accessor interface with pagination
type OptionConnection = CollectionConnection[gqlmodels.Option, gqlmodels.OptionEdge]

// NewOptionConnection based on query object
func NewOptionConnection(ctx context.Context, optionsAccessor option.Usecase, filter *gqlmodels.OptionListFilter, order *gqlmodels.OptionListOrder, page *gqlmodels.Page) *OptionConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.Option, gqlmodels.OptionEdge]{
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
