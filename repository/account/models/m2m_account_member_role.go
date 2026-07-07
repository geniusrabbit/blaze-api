package models

import "time"

// M2MAccountMemberRole m2m link between members and roles|permissions.
type M2MAccountMemberRole struct {
	MemberID  uint64    `db:"member_id" gorm:"primaryKey"`
	RoleID    uint64    `db:"role_id" gorm:"primaryKey"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName of the model in the database.
func (member *M2MAccountMemberRole) TableName() string {
	return `m2m_account_member_role`
}
