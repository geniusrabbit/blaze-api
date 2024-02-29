package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/historylog/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	uc historylog.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		uc: usecase.NewUsecase(repository.New()),
	}
}

// List changelogs is the resolver for the listChangelogs field.
func (r *QueryResolver) List(ctx context.Context, filter *models.HistoryActionListFilter, order *models.HistoryActionListOrder, page *models.Page) (*connectors.HistoryActionConnection, error) {
	return connectors.NewHistoryActionConnection(ctx, r.uc, filter, order, page), nil
}
