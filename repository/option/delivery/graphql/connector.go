package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// OptionConnection implements collection accessor interface with pagination
type OptionConnection = connectors.CollectionConnection[*gqlmodels.Option]

// NewOptionConnection based on query object
func NewOptionConnection(ctx context.Context, optionsAccessor option.Usecase, filter *gqlmodels.OptionListFilter, order []*gqlmodels.OptionListOrder, page *gqlmodels.Page) *OptionConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.Option]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.Option, error) {
			opts := []option.QOption{filter.Filter(), page.Pagination()}
			for _, o := range order {
				if ord := o.Order(); ord != nil {
					opts = append(opts, ord)
				}
			}
			options, err := optionsAccessor.FetchList(ctx, opts...)
			return FromOptionModelList(options), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return optionsAccessor.Count(ctx, filter.Filter())
		},
	}, page)
}
