package models

import "time"

// M2MRole link parent and child role
type M2MRole struct {
	ParentRoleID uint64    `db:"parent_role_id" gorm:"primaryKey"`
	ChildRoleID  uint64    `db:"child_role_id" gorm:"primaryKey"`
	CreatedAt    time.Time `db:"created_at"`
}

// TableName of the model in the database
func (m2m *M2MRole) TableName() string {
	return `m2m_rbac_role`
}
