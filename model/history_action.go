package model

import (
	"encoding/json"
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/google/uuid"
)

// HistoryAction model used for store history of actions.
type HistoryAction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;"`
	RequestID string    `gorm:"type:varchar(255);not null;index:idx_history_actions_request_id;"`

	UserID    uint64 `json:"user_id" gorm:"index:idx_history_actions_user_id;not null;"`
	AccountID uint64 `json:"account_id" gorm:"index:idx_history_actions_account_id;not null;"`

	Name    string `gorm:"type:varchar(255);not null;index:idx_history_actions_name;"`
	Message string `gorm:"type:text;not null;"`

	ObjectType string                             `gorm:"type:varchar(255);not null;index:idx_history_actions_object_type;"`
	ObjectID   uint64                             `gorm:"type:bigint;not null;index:idx_history_actions_object_id;"`
	ObjectIDs  string                             `gorm:"type:varchar(255);not null;index:idx_history_actions_object_ids;"`
	Data       gosql.NullableJSON[map[string]any] `gorm:"type:jsonb;not null;"`

	ActionAt time.Time `gorm:"type:timestamp;not null;index:idx_history_actions_at;"`
}

// TableName returns name of table.
func (*HistoryAction) TableName() string {
	return "history_actions"
}

// DataMap returns data as map.
func (act *HistoryAction) DataMap() map[string]any {
	if dt := act.Data.Data; dt != nil {
		return *dt
	}
	return nil
}

// DataTo unmarshal data to dest.
func (act *HistoryAction) DataTo(dest any) error {
	vl, err := act.Data.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(vl, dest)
}

func (act *HistoryAction) CreatorUserID() uint64 {
	return act.UserID
}

func (act *HistoryAction) OwnerAccountID() uint64 {
	return act.AccountID
}

// RBACResourceName returns the name of the resource for the RBAC
func (*HistoryAction) RBACResourceName() string {
	return `history_log`
}
