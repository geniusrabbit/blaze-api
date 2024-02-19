package model

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

// AuthSession describes session object of the external applications
// which are authenticated by the oauth2 protocol with the current service
type AuthSession struct {
	ID     uint64 `db:"id"`
	Active bool   `db:"active"`

	ClientID string `db:"client_id"` // Internal AuthClient id
	Username string `db:"username"`
	Subject  string `db:"subject"`

	RequestID string `db:"request_id"`

	// AccessToken is the main access token for the session
	AccessToken           string      `db:"access_token"`
	AccessTokenExpiresAt  time.Time   `db:"access_token_expires_at"`
	RefreshToken          null.String `db:"refresh_token" gorm:"type:text"`
	RefreshTokenExpiresAt time.Time   `db:"refresh_token_expires_at"`

	Form              string                    `db:"form"`
	RequestedScope    gosql.NullableStringArray `db:"requested_scope" gorm:"type:text[]"`
	GrantedScope      gosql.NullableStringArray `db:"granted_scope" gorm:"type:text[]"`
	RequestedAudience gosql.NullableStringArray `db:"requested_audience" gorm:"type:text[]"`
	GrantedAudience   gosql.NullableStringArray `db:"granted_audience" gorm:"type:text[]"`

	CreatedAt time.Time      `db:"created_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// TableName in database
func (m *AuthSession) TableName() string {
	return `auth_session`
}

// RBACResourceName returns the name of the resource for the RBAC
func (m *AuthSession) RBACResourceName() string {
	return `auth_session`
}
