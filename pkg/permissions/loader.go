package permissions

import (
	"context"
	"database/sql"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

// DBRoleLoader provides roles from database
type DBRoleLoader struct {
	conn *gorm.DB
}

// ListRoles returns all roles from database
func (l *DBRoleLoader) ListRoles(ctx context.Context) []rbac.Role {
	var (
		links     []*model.M2MRole
		roles     []*model.Role
		roleCache = make(map[uint64]rbac.Role, 10)
		query     = l.conn.WithContext(ctx)
	)
	// Load roles from database
	err := query.Find(&roles).Error
	if err != nil {
		panic(err)
	}
	// Load links between roles from database
	err = query.Find(&links).Error
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctxlogger.Get(ctx).Error("Failed to load roles links", zap.Error(err))
		panic(err)
	}
	// Convert roles to rbac roles
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
		rbac.WithExtData(&ExtData{ID: role.ID, Title: role.Title, AccessLevel: role.AccessLevel}),
		rbac.WithDescription(role.Description),
	)
}
