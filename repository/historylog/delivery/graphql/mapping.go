package graphql

import (
	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// FromHistoryAction converts a historylogModels.HistoryAction to a gqlmodels.HistoryAction.
func FromHistoryAction(action *historylogModels.HistoryAction) *gqlmodels.HistoryAction {
	if action == nil {
		return nil
	}
	return &gqlmodels.HistoryAction{
		ID:        action.ID,
		RequestID: action.RequestID,
		Name:      action.Name,
		Message:   action.Message,

		UserID:    action.UserID,
		AccountID: action.AccountID,

		ObjectID:   action.ObjectID,
		ObjectIDs:  action.ObjectIDs,
		ObjectType: action.ObjectType,
		Data:       *types.MustNullableJSONFrom(&action.Data),

		ActionAt: action.ActionAt,
	}
}

// FromHistoryActionModelList converts a slice of historylogModels.HistoryAction to a slice of gqlmodels.HistoryAction.
func FromHistoryActionModelList(list []*historylogModels.HistoryAction) []*gqlmodels.HistoryAction {
	return xtypes.SliceApply(list, FromHistoryAction)
}

// HistoryActionFilter converts a GraphQL filter to a domain filter.
func HistoryActionFilter(filter *gqlmodels.HistoryActionListFilter) *historylog.Filter {
	if filter == nil {
		return nil
	}
	return &historylog.Filter{
		ID:          filter.ID,
		RequestID:   filter.RequestID,
		UserID:      filter.UserID,
		AccountID:   filter.AccountID,
		ObjectID:    filter.ObjectID,
		ObjectIDStr: filter.ObjectIDs,
		ObjectType:  filter.ObjectType,
	}
}

// HistoryActionOrder converts a GraphQL order to a domain order.
func HistoryActionOrder(order *gqlmodels.HistoryActionListOrder) *historylog.Order {
	if order == nil {
		return nil
	}
	return &historylog.Order{
		ID:          order.ID.AsOrder(),
		RequestID:   order.RequestID.AsOrder(),
		Name:        order.Name.AsOrder(),
		UserID:      order.UserID.AsOrder(),
		AccountID:   order.AccountID.AsOrder(),
		ObjectID:    order.ObjectID.AsOrder(),
		ObjectIDStr: order.ObjectIDs.AsOrder(),
		ObjectType:  order.ObjectType.AsOrder(),
		ActionAt:    order.ActionAt.AsOrder(),
	}
}
