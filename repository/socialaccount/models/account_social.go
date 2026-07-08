package models

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

// AccountSocial represents a user's social network account connection.
type AccountSocial struct {
	ID     uint64 `db:"id" gorm:"primaryKey"`
	UserID uint64 `db:"user_id"`

	// Social network credentials and profile information
	SocialID  string `db:"social_id"` // unique identifier from the social provider
	Provider  string `db:"provider"`  // provider name (facebook, google, twitter, github, etc.)
	Email     string `db:"email"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Username  string `db:"username"`
	Avatar    string `db:"avatar"`
	Link      string `db:"link"` // profile URL

	// Data stores additional provider-specific information as JSON
	Data gosql.NullableJSON[map[string]any] `db:"data" gorm:"type:jsonb"`

	// Sessions contains active sessions linked to this social account
	Sessions []*AccountSocialSession `db:"-" gorm:"foreignKey:AccountSocialID;references:ID"`

	// Timestamps for audit tracking
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"` // soft delete
}

// TableName returns the database table name for AccountSocial.
func (m *AccountSocial) TableName() string {
	return `account_social`
}

// RBACResourceName returns the RBAC resource identifier.
func (m *AccountSocial) RBACResourceName() string {
	return `account_social`
}

// CreatorUserID returns the ID of the account owner.
func (m *AccountSocial) CreatorUserID() uint64 {
	if m == nil {
		return 0
	}
	return m.UserID
}
