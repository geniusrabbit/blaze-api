package model

import (
	"time"

	"gorm.io/gorm"
)

// M2MAccountMemberRole m2m link between members and roles|permissions
type M2MAccountMemberRole struct {
	MemberID  uint64    `db:"member_id" gorm:"primaryKey"`
	RoleID    uint64    `db:"role_id" gorm:"primaryKey"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName of the model in the database
func (member *M2MAccountMemberRole) TableName() string {
	return `m2m_account_member_role`
}

// AccountMember contains reference from user to account as memeber
type AccountMember struct {
	ID      uint64        `db:"id" gorm:"primaryKey"`
	Approve ApproveStatus `db:"approve_status" gorm:"column:approve_status"`

	AccountID uint64   `db:"account_id"`
	Account   *Account `db:"-" gorm:"foreignKey:AccountID;references:ID"`
	UserID    uint64   `db:"user_id"`
	User      *User    `db:"-" gorm:"foreignKey:UserID;references:ID"`

	// Superuser permissions for the current account
	// Despite of that optinion that better to use roles as the only way of permission issue
	//   the Owner flag in most of cases is very useful approach which prevent many problems related to
	//   permission updates.
	// Admin permission restricted by some limits which available only to superusers and managers.
	IsAdmin bool `db:"is_admin"`

	// Roles of the member
	Roles []*Role `gorm:"many2many:m2m_account_member_role;foreignKey:ID;joinForeignKey:MemberID;references:ID;joinReferences:RoleID"`

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// TableName of the model in the database
func (member *AccountMember) TableName() string {
	return `account_member`
}

func (member *AccountMember) OwnerAccountID() uint64 {
	return member.AccountID
}

// RBACResourceName returns the name of the resource for the RBAC
func (member *AccountMember) RBACResourceName() string {
	return "account.member"
}
