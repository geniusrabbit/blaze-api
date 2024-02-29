package models

import (
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
)

// FromHistoryAction converts HistoryAction to HistoryAction
func FromHistoryAction(action *model.HistoryAction) *HistoryAction {
	if action == nil {
		return nil
	}
	return &HistoryAction{
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

// FromHistoryActionModelList converts list of HistoryAction to list of HistoryAction
func FromHistoryActionModelList(list []*model.HistoryAction) []*HistoryAction {
	return xtypes.SliceApply(list, FromHistoryAction)
}

func (filter *HistoryActionListFilter) Filter() *historylog.Filter {
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

func (order *HistoryActionListOrder) Order() *historylog.Order {
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
