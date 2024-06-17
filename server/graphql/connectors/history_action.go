package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// HistoryActionConnection implements collection accessor interface with pagination
type HistoryActionConnection = CollectionConnection[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]

// NewHistoryActionConnection based on query object
func NewHistoryActionConnection(ctx context.Context, historyActionsAccessor historylog.Usecase, filter *gqlmodels.HistoryActionListFilter, order *gqlmodels.HistoryActionListOrder, page *gqlmodels.Page) *HistoryActionConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.HistoryAction, error) {
			historyActions, err := historyActionsAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromHistoryActionModelList(historyActions), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return historyActionsAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.HistoryAction) *gqlmodels.HistoryActionEdge {
			return &gqlmodels.HistoryActionEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
