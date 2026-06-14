package models

import (
	"database/sql"
	"time"
)

// DirectAccessToken represents a direct access token entity.
type DirectAccessToken struct {
	ID          uint64           `json:"id"`
	Token       string           `json:"token"`
	Description string           `json:"description"`
	UserID      sql.Null[uint64] `json:"user_id"`
	AccountID   uint64           `json:"account_id"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TableName specifies the database table name for DirectAccessToken.
func (m *DirectAccessToken) TableName() string {
	return "direct_access_tokens"
}

// RBACResourceName returns the RBAC resource name for access control.
func (m *DirectAccessToken) RBACResourceName() string {
	return "directaccesstoken"
}
