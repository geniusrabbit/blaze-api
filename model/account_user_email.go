package model

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// Email linked to user
type Email struct {
	Email      string         `json:"email" gorm:"primaryKey"`
	UserID     uint64         `json:"user_id"`
	Primary    bool           `json:"primary"`
	VerifiedAt sql.NullTime   `json:"verified_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

// TableName returns the name in database
func (e *Email) TableName() string {
	return "account_email"
}
