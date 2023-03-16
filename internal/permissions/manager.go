package permissions

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/demdxx/rbac"
	"github.com/guregu/null"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/api-template-base/model"
	"github.com/geniusrabbit/gosql/v2"
)

const defaultAdminRole = `account:admin`

// Cover permission type
const (
	CoverTypeSystem  = `system`  // The total system access
	CoverTypeAccount = `account` // The total account access
)

// ErrUndefinedRole if not found
var ErrUndefinedRole = errors.New(`undefined role`)

type roleID interface {
	RoleID() uint64
}

// ExtData permission data
type ExtData struct {
	Object string `json:"object"`
	Cover  string `json:"cover"`
}

type objectItem struct {
	objType      any
	checkCallbac any
}

// Manager provides methods to control and cache permissions
type Manager struct {
	mx sync.RWMutex

	preinitedRoles []rbac.Role

	// Object context data
	objects map[string]*objectItem

	// Connection to main database
	conn *gorm.DB

	// Role cache of roles
	roleCache       map[uint64]rbac.Role
	lastCacheUpdate time.Time
	lifetimeCache   time.Duration
}

// NewManager object to control roles
func NewManager(conn *gorm.DB, cacheLifetime time.Duration, preinitedRoles ...rbac.Role) *Manager {
	if cacheLifetime == 0 {
		cacheLifetime = time.Second * 5
	}
	return &Manager{
		conn:           conn,
		preinitedRoles: preinitedRoles,
		objects:        map[string]*objectItem{},
		roleCache:      map[uint64]rbac.Role{},
		lifetimeCache:  cacheLifetime,
	}
}

// NewManagerWithRoles object to control roles
func NewManagerWithRoles(roles map[uint64]rbac.Role) *Manager {
	return &Manager{
		objects:         map[string]*objectItem{},
		roleCache:       roles,
		lastCacheUpdate: time.Now(),
		lifetimeCache:   time.Hour * 24 * 365 * 10,
	}
}

// NewTestManager with all permissions
func NewTestManager() *Manager {
	return NewManagerWithRoles(
		map[uint64]rbac.Role{
			1: rbac.NewDummyPermission(`test`, true),
			2: rbac.NewDummyPermission(defaultAdminRole, true),
		},
	)
}

// RegisterObject for processing
func (mng *Manager) RegisterObject(objType, checkCallbac any) {
	mng.objects[objectName(objType)] = &objectItem{
		objType:      objType,
		checkCallbac: checkCallbac,
	}
}

// RegisterRole preinted in advanced
func (mng *Manager) RegisterRole(role rbac.Role) {
	mng.preinitedRoles = append(mng.preinitedRoles, role)
}

// Role returns role by ID and reload data if necessary
func (mng *Manager) Role(ctx context.Context, id uint64) (rbac.Role, error) {
	if err := mng.refresh(ctx); err != nil {
		return nil, err
	}
	mng.mx.RLock()
	defer mng.mx.RUnlock()
	if role := mng.roleCache[id]; role != nil {
		return role, nil
	}
	return nil, ErrUndefinedRole
}

// RoleByName returns role by Name and reload data if necessary
func (mng *Manager) RoleByName(ctx context.Context, name string) (rbac.Role, error) {
	if err := mng.refresh(ctx); err != nil {
		return nil, err
	}
	mng.mx.RLock()
	defer mng.mx.RUnlock()
	for _, role := range mng.roleCache {
		if role.Name() == name {
			return role, nil
		}
	}
	return nil, errors.Wrap(ErrUndefinedRole, name)
}

// AsOneRole returns new role object from one or more IDs
func (mng *Manager) AsOneRole(ctx context.Context, isAdmin bool, filter func(rbac.Role) bool, id ...uint64) (rbac.Role, error) {
	var roles []rbac.Role
	if err := mng.refresh(ctx); err != nil {
		return nil, err
	}
	if isAdmin {
		admin, err := mng.RoleByName(ctx, defaultAdminRole)
		if err != nil {
			return nil, err
		}
		roles = append(roles, admin)
	}
	mng.mx.RLock()
	defer mng.mx.RUnlock()
	for _, idVal := range id {
		if role := mng.roleCache[idVal]; role != nil && (filter == nil || filter(role)) {
			roles = append(roles, role)
		}
	}
	if len(roles) == 1 {
		return roles[0], nil
	}
	if len(roles) == 0 {
		return nil, ErrUndefinedRole
	}
	return rbac.NewRole(``, rbac.WithChildRoles(roles...))
}

// ObjectByName returns registered object with :name
func (mng *Manager) ObjectByName(name string) any {
	item := mng.objects[name]
	if item == nil {
		return nil
	}
	return item.objType
}

func (mng *Manager) refresh(ctx context.Context) (err error) {
	mng.mx.Lock()
	defer mng.mx.Unlock()
	if mng.lastCacheUpdate.Add(mng.lifetimeCache).Before(time.Now()) {
		err = mng.refreshRoles(ctx)
	}
	return err
}

func (mng *Manager) refreshRoles(ctx context.Context) error {
	var (
		links     []*model.M2MRole
		roles     []*model.Role
		permCache = map[uint64]rbac.Permission{}
		roleCache = map[uint64]rbac.Role{}
		query     = mng.conn.WithContext(ctx)
	)
	for _, role := range mng.preinitedRoles {
		if rid, _ := role.(roleID); rid != nil {
			roleCache[rid.RoleID()] = role
		} else {
			roleCache[0] = role
		}
	}
	mng.roleCache = roleCache
	mng.lastCacheUpdate = time.Now()
	err := query.Order(`parent_id ASC`).Find(&roles).Error
	if err != nil {
		return err
	}
	err = query.Find(&links).Error
	if err != nil && err != sql.ErrNoRows && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, role := range roles {
		if role.Type.IsPermission() {
			permCache[role.ID], err = mng.permissionByModel(role, roleCache, permCache, links)
		}
		if err != nil {
			return err
		}
	}
	for _, role := range roles {
		if !role.Type.IsPermission() {
			roleCache[role.ID], err = mng.roleByModel(role, roleCache, permCache, links)
		}
		if err != nil {
			return err
		}
	}
	mng.roleCache = roleCache
	mng.lastCacheUpdate = time.Now()
	return nil
}

func (mng *Manager) roleByModel(role *model.Role, roles map[uint64]rbac.Role, perms map[uint64]rbac.Permission, links []*model.M2MRole) (rbac.Role, error) {
	var (
		permList = make([]rbac.Permission, 0, len(links))
		roleList = make([]rbac.Role, 0, len(links))
	)
	for _, link := range links {
		if link.ParentRoleID != role.ID {
			continue
		}
		if perm := perms[link.ChildRoleID]; perm != nil {
			permList = append(permList, perm)
		}
		if rls := roles[link.ChildRoleID]; rls != nil {
			roleList = append(roleList, rls)
		}
	}
	return rbac.NewRole(role.Name, rbac.WithChildRoles(roleList...), rbac.WithSubPermissins(permList...))
}

// NewResourcePermission returns new permission for resource
func (mng *Manager) NewRosourcePermission(name string, resType any, options ...rbac.Option) (rbac.Permission, error) {
	jCtx, _ := gosql.NewNullableJSON[map[string]any](map[string]any{`object`: objectName(resType)})
	if jCtx == nil {
		jCtx = &gosql.NullableJSON[map[string]any]{}
	}
	return mng.permissionByModel(&model.Role{
		ID:       0,
		Name:     name,
		Title:    ``,
		ParentID: null.Int{},
		Type:     model.RbacPermissionType,
		Context:  *jCtx,
	}, nil, nil, nil)
}

// MustNewResourcePermission returns new permission for resource or panic
func (mng *Manager) MustNewRosourcePermission(name string, resType any, options ...rbac.Option) rbac.Permission {
	perm, err := mng.NewRosourcePermission(name, resType, options...)
	if err != nil {
		panic(err)
	}
	return perm
}

func (mng *Manager) permissionByModel(role *model.Role, roles map[uint64]rbac.Role, perms map[uint64]rbac.Permission, links []*model.M2MRole) (rbac.Permission, error) {
	permList := make([]rbac.Permission, 0, len(links))
	for _, link := range links {
		if link.ParentRoleID != role.ID {
			continue
		}
		if perm := perms[link.ChildRoleID]; perm != nil {
			permList = append(permList, perm)
		}
		if rls := roles[link.ChildRoleID]; rls != nil {
			permList = append(permList, rls)
		}
	}

	var (
		options = []rbac.Option{}
		data    = &ExtData{
			Object: role.ContextItemString("object"),
			Cover:  role.ContextItemString("cover"),
		}
	)

	// Init from object reference
	if obj := mng.objects[data.Object]; obj != nil {
		options = append(options, rbac.WithSubPermissins(permList...))
		if obj.checkCallbac != nil {
			options = append(options, rbac.WithCustomCheck(obj.checkCallbac, data))
		}
		return rbac.NewRosourcePermission(role.Name, obj.objType, options...)
	}

	return rbac.NewSimplePermission(role.Name, append(options, rbac.WithSubPermissins(permList...))...)
}

func objectName(obj any) string {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Interface || t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return filepath.Base(t.PkgPath()) + `:` + t.Name()
}
