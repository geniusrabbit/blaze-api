package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// ListHistory is the resolver for the listHistory field.
func (r *queryResolver) ListHistory(ctx context.Context, filter *models.HistoryActionListFilter, order *models.HistoryActionListOrder, page *models.Page) (*connectors.CollectionConnection[models.HistoryAction, models.HistoryActionEdge], error) {
	return r.historylogs.List(ctx, filter, order, page)
}
