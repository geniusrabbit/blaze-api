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
	return connectors.NewHistoryActionConnection(
		ctx, r.uc,
		filterFrom(filter),
		orderFrom(order),
		page), nil
}

func filterFrom(filter *models.HistoryActionListFilter) *historylog.Filter {
	if filter == nil {
		return nil
	}
	return &historylog.Filter{
		ID:          filter.ID,
		UserID:      filter.UserID,
		AccountID:   filter.AccountID,
		ObjectID:    filter.ObjectID,
		ObjectIDStr: filter.ObjectIDs,
		ObjectType:  filter.ObjectType,
	}
}

func orderFrom(order *models.HistoryActionListOrder) *historylog.Order {
	if order == nil {
		return nil
	}
	return &historylog.Order{
		ID:          order.ID.Int8(),
		Name:        order.Name.Int8(),
		UserID:      order.UserID.Int8(),
		AccountID:   order.AccountID.Int8(),
		ObjectID:    order.ObjectID.Int8(),
		ObjectIDStr: order.ObjectIDs.Int8(),
		ObjectType:  order.ObjectType.Int8(),
		ActionAt:    order.ActionAt.Int8(),
	}
}
