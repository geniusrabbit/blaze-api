package permissions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/rbac"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
)

// DBRoleLoader provides roles from database.
//
// When the database fails, the error is logged and the roles of the last
// successful load are returned, so a database outage doesn't panic every
// request (or drop all roles); before the first successful load no database
// roles are available.
type DBRoleLoader struct {
	conn *gorm.DB

	mx   sync.Mutex
	last []rbac.Role
}

// ListRoles returns all roles from database
func (l *DBRoleLoader) ListRoles(ctx context.Context) []rbac.Role {
	roles, err := l.load(ctx)
	l.mx.Lock()
	defer l.mx.Unlock()
	if err != nil {
		ctxlogger.Get(ctx).Error("Failed to load roles, using the previous ones",
			zap.Int("previous_roles", len(l.last)), zap.Error(err))
		return l.last
	}
	l.last = roles
	return roles
}

func (l *DBRoleLoader) load(ctx context.Context) ([]rbac.Role, error) {
	if l.conn == nil {
		return nil, errors.New("permissions: no database connection")
	}
	var (
		links []*rbacModels.M2MRole
		roles []*rbacModels.Role
		query = l.conn.WithContext(ctx)
	)
	if err := query.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("load roles: %w", err)
	}
	err := query.Find(&links).Error
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("load role links: %w", err)
	}
	return buildRoles(roles, links)
}

// buildRoles converts role models to rbac roles. Child roles are built
// first, whatever the order of the rows, so inherited permissions are
// never lost; a link cycle is an error.
func buildRoles(models []*rbacModels.Role, links []*rbacModels.M2MRole) ([]rbac.Role, error) {
	byID := make(map[uint64]*rbacModels.Role, len(models))
	for _, m := range models {
		byID[m.ID] = m
	}
	children := make(map[uint64][]uint64, len(links))
	for _, link := range links {
		children[link.ParentRoleID] = append(children[link.ParentRoleID], link.ChildRoleID)
	}

	built := make(map[uint64]rbac.Role, len(models))
	visiting := make(map[uint64]bool)
	var build func(id uint64) (rbac.Role, error)
	build = func(id uint64) (rbac.Role, error) {
		if role, ok := built[id]; ok {
			return role, nil
		}
		model := byID[id]
		if model == nil {
			return nil, nil // link to a missing (e.g. deleted) role
		}
		if visiting[id] {
			return nil, fmt.Errorf("role %q: cycle in child roles", model.Name)
		}
		visiting[id] = true
		defer delete(visiting, id)

		childRoles := make([]rbac.Role, 0, len(children[id]))
		for _, childID := range children[id] {
			child, err := build(childID)
			if err != nil {
				return nil, err
			}
			if child != nil {
				childRoles = append(childRoles, child)
			}
		}
		role, err := rbac.NewRole(model.Name, rbac.WithChildRoles(childRoles...),
			rbac.WithPermissions(gocast.Slice[any](model.PermissionPatterns)...),
			rbac.WithExtData(&ExtData{ID: model.ID, Title: model.Title, AccessLevel: model.AccessLevel}),
			rbac.WithDescription(model.Description),
		)
		if err != nil {
			return nil, fmt.Errorf("role %q: %w", model.Name, err)
		}
		built[id] = role
		return role, nil
	}

	result := make([]rbac.Role, 0, len(models))
	for _, m := range models {
		role, err := build(m.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, nil
}
