package model

import (
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/gosql/v2"
	"github.com/guregu/null"
	"gorm.io/gorm"
)

// RoleType type casting
type RoleType string

// RBAC role constant list...
const (
	RbacUndefinedType  RoleType = ``
	RbacRoleType       RoleType = `role`
	RbacPermissionType RoleType = `permission`
)

func (rt RoleType) String() string {
	return string(rt)
}

// IsPermission type
func (rt RoleType) IsPermission() bool {
	return rt == RbacPermissionType
}

// M2MRole link parent and child role
type M2MRole struct {
	ParentRoleID uint64    `db:"parent_role_id"`
	ChildRoleID  uint64    `db:"child_role_id"`
	CreatedAt    time.Time `db:"created_at"`
}

// TableName of the model in the database
func (m2m *M2MRole) TableName() string {
	return `m2m_rbac_role`
}

// Role base model
type Role struct {
	ID       uint64   `db:"id"`
	ParentID null.Int `db:"parent_id"`
	Name     string   `db:"name"`
	Title    string   `db:"title"`
	Type     RoleType `db:"type"`

	// {"cover": "account", "object": "model:User"}
	// "cover" - is a name of the cover area of the object type
	// "object" - is a name of the object type <module>:<object-name>
	Context gosql.NullableJSON[map[string]any] `db:"context"`

	ChildRolesAndPermissions []*Role `db:"-" gorm:"-"`

	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt gorm.DeletedAt `db:"deleted_at"`
}

// GetTitle from role object
// nolint:unused // exported
func (role *Role) GetTitle() string {
	if role.Title != `` {
		return role.Title
	}
	if objName := role.ContextItemString(`object`); objName != `` {
		role.Title = objName + `:` + role.Name
		if coverName := role.ContextItemString(`cover`); coverName != `` {
			role.Title += ` ` + coverName
		}
		return role.Title
	}
	return role.Name
}

// TableName of the model in the database
func (role *Role) TableName() string {
	return `rbac_role`
}

// ContextMap returns the map from the context
func (role *Role) ContextMap() map[string]any {
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
