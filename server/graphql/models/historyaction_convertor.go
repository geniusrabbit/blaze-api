package models

import (
	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/server/graphql/types"
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
	return xtypes.SliceApply[*model.HistoryAction, *HistoryAction](list, FromHistoryAction)
}
