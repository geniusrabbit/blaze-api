package models

import (
	"time"

	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
)

// MemberBase is the embeddable account-member join (Account ↔ User + RBAC).
type MemberBase struct {
	ID        uint64                `db:"id" gorm:"primaryKey"`
	Approve   pkgModels.ApproveStatus `db:"approve_status" gorm:"column:approve_status"`
	AccountID uint64                `db:"account_id"`
	UserID    uint64                `db:"user_id"`
	IsAdmin   bool                  `db:"is_admin"`
	Roles     []*rbacModels.Role    `gorm:"many2many:m2m_account_member_role;foreignKey:ID;joinForeignKey:MemberID;references:ID;joinReferences:RoleID"`
	CreatedAt time.Time             `db:"created_at"`
	UpdatedAt time.Time             `db:"updated_at"`
	DeletedAt gorm.DeletedAt        `db:"deleted_at"`
}

// GetID returns member ID.
func (m *MemberBase) GetID() uint64 {
	if m == nil {
		return 0
	}
	return m.ID
}

// OwnerAccountID returns the account this member belongs to.
func (m *MemberBase) OwnerAccountID() uint64 {
	if m == nil {
		return 0
	}
	return m.AccountID
}

// RBACResourceName returns the default RBAC resource name (override on consumer type).
func (m *MemberBase) RBACResourceName() string {
	return "account.member"
}

// TableName returns the default database table (override on consumer type).
func (m *MemberBase) TableName() string {
	return "account_member"
}

// MemberTableName returns the default member join table name (nil-safe for query builders).
func MemberTableName() string {
	return "account_member"
}
