package model

import (
	"time"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

type AccountSocialSession struct {
	AccountSocialID uint64 `db:"account_social_id"`

	TokenType    string `json:"token_type,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	ExpiresAt null.Time      `json:"expires_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}

// TableName in database
func (m *AccountSocialSession) TableName() string {
	return `account_social_session`
}
