package graphql

import (
	"context"

	"github.com/demdxx/gocast/v2"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// HistoryActionConnection implements collection accessor interface with pagination
type HistoryActionConnection = connectors.CollectionConnection[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]

// NewHistoryActionConnection based on query object
func NewHistoryActionConnection(ctx context.Context, historyActionsAccessor historylog.Usecase, filter *gqlmodels.HistoryActionListFilter, order *gqlmodels.HistoryActionListOrder, page *gqlmodels.Page) *HistoryActionConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.HistoryAction, error) {
			historyActions, err := historyActionsAccessor.FetchList(ctx, HistoryActionFilter(filter), HistoryActionOrder(order), page.Pagination())
			return FromHistoryActionModelList(historyActions), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return historyActionsAccessor.Count(ctx, HistoryActionFilter(filter))
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.HistoryAction) *gqlmodels.HistoryActionEdge {
			return &gqlmodels.HistoryActionEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
