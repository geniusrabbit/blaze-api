package model

import (
	"time"

	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

// AccountSocial object represents a social network account
type AccountSocial struct {
	ID     uint64 `db:"id"`
	UserID uint64 `db:"user_id"`

	SocialID  string                    `db:"social_id"` // social network user id
	Provider  string                    `db:"provider"`  // facebook, google, twitter, github, etc
	Email     string                    `db:"email"`
	FirstName string                    `db:"first_name"`
	LastName  string                    `db:"last_name"`
	Username  string                    `db:"username"`
	Avatar    string                    `db:"avatar"`
	Link      string                    `db:"link"`
	Scope     gosql.NullableStringArray `db:"scope" gorm:"type:text[]"` // requested scope

	// Data is a JSON object with additional data
	Data gosql.NullableJSON[map[string]any] `db:"data" gorm:"type:jsonb"`

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// TableName in database
func (m *AccountSocial) TableName() string {
	return `account_social`
}

// RBACResourceName returns the name of the resource for the RBAC
func (m *AccountSocial) RBACResourceName() string {
	return `account_social`
}
