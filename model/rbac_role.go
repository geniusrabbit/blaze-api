package model

import (
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/gosql/v2"
	"gorm.io/gorm"
)

const (
	AccessLevelBasic       = 0
	AccessLevelNoAnonymous = 1
	AccessLevelAccount     = 2
	AccessLevelSystem      = 3
)

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

// Role base model
type Role struct {
	ID    uint64 `db:"id"`
	Name  string `db:"name"`
	Title string `db:"title"`

	Description string `db:"description"`

	// Contains additional data for the role
	Context gosql.NullableJSON[map[string]any] `db:"context"`

	ChildRoles         []*Role                   `db:"-" gorm:"many2many:m2m_rbac_role;ForeignKey:ID;joinForeignKey:parent_role_id;joinReferences:child_role_id;References:ID"`
	PermissionPatterns gosql.NullableStringArray `db:"permissions" gorm:"column:permissions;type:text[]"`

	AccessLevel int `db:"access_level"` // 0 - any, 1 - no anonymous, 2 - account, >=3 - system

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// GetTitle from role object
// nolint:unused // exported
func (role *Role) GetTitle() string {
	return gocast.Or(role.Title, role.Name)
}

// TableName of the model in the database
func (role *Role) TableName() string {
	return `rbac_role`
}

// RBACResourceName returns the name of the resource for the RBAC
func (role *Role) RBACResourceName() string {
	return "role"
}

// ContextMap returns the map from the context
func (role *Role) ContextMap() map[string]any {
	if role == nil || role.Context.Data == nil {
		return nil
	}
	return *role.Context.Data
}

// ContextItem returns one value by name from context
func (role *Role) ContextItem(name string) any {
	return role.ContextMap()[name]
}

// ContextItemString returns one string value by name from context
func (role *Role) ContextItemString(name string) string {
	return gocast.Str(role.ContextItem(name))
}
