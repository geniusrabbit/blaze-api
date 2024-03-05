package model

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

type AccountSocialSession struct {
	// Unique name of the session to destinguish between different sessions with different scopes
	Name            string `db:"name" gorm:"primaryKey"`
	AccountSocialID uint64 `db:"account_social_id" gorm:"primaryKey;autoIncrement:false"`

	TokenType    string                    `db:"token_type" json:"token_type,omitempty"`
	AccessToken  string                    `db:"access_token" json:"access_token"`
	RefreshToken string                    `db:"refresh_token" json:"refresh_token"`
	Scopes       gosql.NullableStringArray `db:"scopes" json:"scopes,omitempty" gorm:"type:text[]"`

	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	ExpiresAt null.Time      `db:"expires_at" json:"expires_at,omitempty"`
	DeletedAt gorm.DeletedAt `db:"deleted_at" json:"deleted_at,omitempty"`
}

// TableName in database
func (m *AccountSocialSession) TableName() string {
	return `account_social_session`
}
