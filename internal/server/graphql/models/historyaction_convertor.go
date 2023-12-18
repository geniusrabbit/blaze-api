package models

import (
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/types"
	"github.com/geniusrabbit/api-template-base/model"
)

// FromHistoryAction converts HistoryAction to HistoryAction
func FromHistoryAction(action *model.HistoryAction) *HistoryAction {
	return &HistoryAction{
		ID:      action.ID,
		Name:    action.Name,
		Message: action.Message,

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
	actions := make([]*HistoryAction, 0, len(list))
	for _, u := range list {
		actions = append(actions, FromHistoryAction(u))
	}
	return actions
}
