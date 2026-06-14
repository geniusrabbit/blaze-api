package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/models"
)

// Anonymous user object
var Anonymous = User{ID: 0}

// User direct defenition
type User struct {
	ID    uint64 `json:"id" gorm:"primaryKey"`
	Email string `json:"email"`

	Password              string `json:"password"`
	RequiredPasswordReset bool   `json:"required_password_reset"`

	Approve models.ApproveStatus `gorm:"column:approve_status" db:"approve_status" json:"approve_status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// GetID returns user ID
func (u *User) GetID() uint64 {
	if u == nil {
		return 0
	}
	return u.ID
}

// IsNil checks if the user is nil
func (u *User) IsNil() bool {
	return u == nil
}

// TableName returns the name in database
func (u *User) TableName() string {
	return "account_user"
}

// IsAnonymous user object
// nolint:unused // temporary
func (u *User) IsAnonymous() bool {
	return u == nil || u.ID == 0
}

// CreatorUserID returns the user id
func (u *User) CreatorUserID() uint64 {
	if u == nil {
		return 0
	}
	return u.ID
}

// RBACResourceName returns the name of the resource for the RBAC
func (u *User) RBACResourceName() string {
	return "user"
}
