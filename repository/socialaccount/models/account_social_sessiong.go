package models

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

// AccountSocialSession represents a social account authentication session
// with OAuth tokens and scope information.
type AccountSocialSession struct {
	// Name uniquely identifies the session to distinguish between different
	// sessions with different scopes for the same social account.
	Name string `db:"name" gorm:"primaryKey"`

	// AccountSocialID is the foreign key to the social account.
	AccountSocialID uint64 `db:"account_social_id" gorm:"primaryKey;autoIncrement:false"`

	// TokenType specifies the OAuth token type (e.g., "Bearer").
	TokenType string `db:"token_type" json:"token_type,omitempty"`

	// AccessToken is the OAuth access token for API requests.
	AccessToken string `db:"access_token" json:"access_token"`

	// RefreshToken is used to obtain a new access token when it expires.
	RefreshToken string `db:"refresh_token" json:"refresh_token"`

	// Scopes are the OAuth permission scopes granted for this session.
	Scopes gosql.NullableStringArray `db:"scopes" json:"scopes,omitempty" gorm:"type:text[]"`

	// CreatedAt is the session creation timestamp.
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// ExpiresAt is when the access token expires (nullable).
	ExpiresAt null.Time `db:"expires_at" json:"expires_at,omitempty"`

	// DeletedAt is the soft delete timestamp.
	DeletedAt gorm.DeletedAt `db:"deleted_at" json:"deleted_at,omitempty"`
}

// TableName returns the database table name for AccountSocialSession.
func (m *AccountSocialSession) TableName() string {
	return `account_social_session`
}
