package model

import (
	"time"
)

// UserPasswordReset direct defenition
type UserPasswordReset struct {
	UserID uint64 `json:"user_id" gorm:"primaryKey"`
	Token  string `json:"token" gorm:"index:,unique" limit:"128"`

	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TableName returns the name in database
func (u *UserPasswordReset) TableName() string {
	return "account_user_password_reset"
}
