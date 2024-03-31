package permissions

import (
	"context"
	"time"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	DefaultAdminRole     = `account:admin`
	DefaultRole          = `general`
	AnonymousDefaultRole = `anonymous`
)

var (
	// ErrUndefinedRole if not found
	ErrUndefinedRole = errors.New(`undefined role`)
)

// ExtData permission data
type ExtData struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	AccessLevel int    `json:"access_level"`
}

// Manager provides methods to control and cache permissions
type Manager struct {
	*rbac.Manager
}

// NewManager object to control roles
func NewManager(conn *gorm.DB, cacheLifetime time.Duration) *Manager {
	if cacheLifetime == 0 {
		cacheLifetime = time.Second * 5
	}
	return &Manager{Manager: rbac.NewManagerWithLoader(
		&DBRoleLoader{conn: conn}, cacheLifetime)}
}

// NewTestManager with all permissions
func NewTestManager(ctx context.Context) *Manager {
	return &Manager{
		Manager: rbac.NewManager(nil).RegisterRole(ctx,
			rbac.NewDummyPermission(`test`, true),
			rbac.NewDummyPermission(DefaultAdminRole, true),
			rbac.NewDummyPermission(AnonymousDefaultRole, true),
		),
	}
}

// RoleByID returns role by ID and reload data if necessary
func (mng *Manager) RoleByID(ctx context.Context, id uint64) (rbac.Role, error) {
	roles := mng.RolesByFilter(ctx, func(ctx context.Context, r rbac.Role) bool {
		return r.Ext().(*ExtData).ID == id
	})
	if len(roles) == 1 {
		return roles[0], nil
	}
	return nil, ErrUndefinedRole
}

// DefaultRole returns default role
func (mng *Manager) DefaultRole(ctx context.Context) rbac.Role {
	return mng.Role(ctx, DefaultRole)
}

// AsOneRole returns new role object from one or more IDs
func (mng *Manager) AsOneRole(ctx context.Context, isAdmin bool, filter func(context.Context, rbac.Role) bool, id ...uint64) (rbac.Role, error) {
	var roles []rbac.Role
	if isAdmin {
		adminRole := mng.Role(ctx, DefaultAdminRole)
		if adminRole == nil {
			return nil, errors.Wrap(ErrUndefinedRole, DefaultAdminRole)
		}
		roles = append(roles, adminRole)
	}

	if len(id) > 0 {
		roles = append(roles, mng.RolesByFilter(ctx, func(ctx context.Context, r rbac.Role) bool {
			switch data := r.Ext().(type) {
			case *ExtData:
				return xtypes.Slice[uint64](id).Has(func(val uint64) bool {
					return data.ID == val && (filter == nil || filter(ctx, r))
				})
			}
			return false
		})...)
	}
	if len(roles) == 0 && len(id) != 0 {
		return nil, ErrUndefinedRole
	}
	return rbac.NewRole(``, rbac.WithChildRoles(append(roles, mng.DefaultRole(ctx))...))
}
