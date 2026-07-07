package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogrepo "github.com/geniusrabbit/blaze-api/repository/historylog/repository"
	historylogusecase "github.com/geniusrabbit/blaze-api/repository/historylog/usecase"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	uc historylog.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver(uc historylog.Usecase) *QueryResolver {
	return &QueryResolver{uc: uc}
}

// NewDefaultQueryResolver returns new API resolver with default usecase
func NewDefaultQueryResolver() *QueryResolver {
	return &QueryResolver{uc: historylogusecase.NewUsecase(historylogrepo.New())}
}

// List changelogs is the resolver for the listChangelogs field.
func (r *QueryResolver) List(ctx context.Context, filter *gqlmodels.HistoryActionListFilter, order []*gqlmodels.HistoryActionListOrder, page *gqlmodels.Page) (*HistoryActionConnection, error) {
	return NewHistoryActionConnection(ctx, r.uc, filter, order, page), nil
}
