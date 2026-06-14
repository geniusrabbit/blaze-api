package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
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

// List changelogs is the resolver for the listChangelogs field.
func (r *QueryResolver) List(ctx context.Context, filter *gqlmodels.HistoryActionListFilter, order *gqlmodels.HistoryActionListOrder, page *gqlmodels.Page) (*HistoryActionConnection, error) {
	return NewHistoryActionConnection(ctx, r.uc, filter, order, page), nil
}
