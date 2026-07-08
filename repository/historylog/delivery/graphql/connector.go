package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// HistoryActionConnection implements collection accessor interface with pagination
type HistoryActionConnection = connectors.CollectionConnection[*gqlmodels.HistoryAction]

// NewHistoryActionConnection based on query object
func NewHistoryActionConnection(ctx context.Context, historyActionsAccessor historylog.Usecase, filter *gqlmodels.HistoryActionListFilter, order []*gqlmodels.HistoryActionListOrder, page *gqlmodels.Page) *HistoryActionConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.HistoryAction]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.HistoryAction, error) {
			opts := []historylog.QOption{HistoryActionFilter(filter), page.Pagination()}
			for _, o := range order {
				if ord := HistoryActionOrder(o); ord != nil {
					opts = append(opts, ord)
				}
			}
			historyActions, err := historyActionsAccessor.FetchList(ctx, opts...)
			return FromHistoryActionModelList(historyActions), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return historyActionsAccessor.Count(ctx, HistoryActionFilter(filter))
		},
	}, page)
}
