package models

import (
	"time"

	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// UserBase is the minimal embeddable user identity (Member anchor).
type UserBase struct {
	ID        uint64                  `json:"id" gorm:"primaryKey"`
	Approve   pkgModels.ApproveStatus `gorm:"column:approve_status" db:"approve_status" json:"approve_status"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	DeletedAt gorm.DeletedAt          `json:"deleted_at"`
}

// GetID returns user ID.
func (u *UserBase) GetID() uint64 {
	if u == nil {
		return 0
	}
	return u.ID
}

// SetID sets user primary key.
func (u *UserBase) SetID(id uint64) {
	if u != nil {
		u.ID = id
	}
}

// IsNil checks if the user is nil.
func (u *UserBase) IsNil() bool {
	return u == nil
}

// IsAnonymous reports whether this is an anonymous user.
func (u *UserBase) IsAnonymous() bool {
	return u == nil || u.ID == 0
}

// CreatorUserID returns the user id for ACL ownership checks.
func (u *UserBase) CreatorUserID() uint64 {
	if u == nil {
		return 0
	}
	return u.ID
}

// RBACResourceName returns the default RBAC resource name (override on consumer type).
func (u *UserBase) RBACResourceName() string {
	return "user"
}

// TableName returns the default database table (override on consumer type).
func (u *UserBase) TableName() string {
	return "account_user"
}

// SetApprove sets approval status.
func (u *UserBase) SetApprove(status pkgModels.ApproveStatus) {
	if u != nil {
		u.Approve = status
	}
}

// GetApprove returns approval status.
func (u *UserBase) GetApprove() pkgModels.ApproveStatus {
	if u == nil {
		return pkgModels.UndefinedApproveStatus
	}
	return u.Approve
}

// SetCreatedAt sets created timestamp.
func (u *UserBase) SetCreatedAt(t time.Time) { u.CreatedAt = t }

// GetCreatedAt returns created timestamp.
func (u *UserBase) GetCreatedAt() time.Time { return u.CreatedAt }

// SetUpdatedAt sets updated timestamp.
func (u *UserBase) SetUpdatedAt(t time.Time) { u.UpdatedAt = t }

// GetUpdatedAt returns updated timestamp.
func (u *UserBase) GetUpdatedAt() time.Time { return u.UpdatedAt }
