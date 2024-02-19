package permissions

import (
	"context"
	"database/sql"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
)

const defaultAdminRole = `account:admin`

var (
	// ErrUndefinedRole if not found
	ErrUndefinedRole = errors.New(`undefined role`)
)

// ExtData permission data
type ExtData struct {
	ID     uint64 `json:"id"`
	Object string `json:"object"`
	Cover  string `json:"cover"`
}

type DBRoleLoader struct {
	conn *gorm.DB
}

func (l *DBRoleLoader) ListRoles(ctx context.Context) []rbac.Role {
	var (
		links     []*model.M2MRole
		roles     []*model.Role
		roleCache = make(map[uint64]rbac.Role, 10)
		query     = l.conn.WithContext(ctx)
	)
	err := query.Find(&roles).Error
	if err != nil {
		panic(err)
	}
	err = query.Find(&links).Error
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err)
	}
	for _, role := range roles {
		roleCache[role.ID], err = roleByModel(role, roleCache, links)
		if err != nil {
			panic(err)
		}
	}
	return xtypes.Map[uint64, rbac.Role](roleCache).Values()
}

func roleByModel(role *model.Role, roles map[uint64]rbac.Role, links []*model.M2MRole) (rbac.Role, error) {
	roleList := make([]rbac.Role, 0, len(links))
	for _, link := range links {
		if link.ParentRoleID != role.ID {
			continue
		}
		if rls := roles[link.ChildRoleID]; rls != nil {
			roleList = append(roleList, rls)
		}
	}
	return rbac.NewRole(role.Name, rbac.WithChildRoles(roleList...),
		rbac.WithPermissions(gocast.Slice[any](role.PermissionPatterns)...),
		rbac.WithExtData(&ExtData{ID: role.ID}),
	)
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
			rbac.NewDummyPermission(defaultAdminRole, true),
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

// AsOneRole returns new role object from one or more IDs
func (mng *Manager) AsOneRole(ctx context.Context, isAdmin bool, filter func(context.Context, rbac.Role) bool, id ...uint64) (rbac.Role, error) {
	var roles []rbac.Role
	if isAdmin {
		adminRole := mng.Role(ctx, defaultAdminRole)
		if adminRole == nil {
			return nil, errors.Wrap(ErrUndefinedRole, defaultAdminRole)
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

	if len(roles) == 1 {
		return roles[0], nil
	}
	if len(roles) == 0 {
		return nil, ErrUndefinedRole
	}
	return rbac.NewRole(``, rbac.WithChildRoles(roles...))
}
