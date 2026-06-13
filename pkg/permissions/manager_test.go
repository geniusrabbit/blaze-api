package permissions

import (
	"context"
	"testing"

	"github.com/demdxx/rbac"
	"github.com/stretchr/testify/assert"

	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

type TestObject struct{}

func TestManager(t *testing.T) {
	ctx := context.TODO()
	mng := NewManager(nil, 0)

	mng.RegisterObject(&rbacModels.Role{}, func(ctx context.Context, obj *rbacModels.Role, perm rbac.Permission) bool { return true })

	_ = mng.RegisterNewPermission(&rbacModels.Role{}, `view`)
	_ = mng.RegisterNewPermission(&userModels.User{}, `view`)

	perm1 := mng.Permission(`role.view`)
	assert.True(t, perm1.CheckPermissions(ctx, &rbacModels.Role{}, `view`), `CheckPermissions`)

	perm2 := mng.Permission(`user.view`)
	assert.True(t, perm2.CheckPermissions(ctx, &userModels.User{}, `view`), `CheckPermissions`)

	roleObj := &rbacModels.Role{ID: 10, Name: `role1`, PermissionPatterns: []string{`role.*`}}
	role1, err := roleByModel(roleObj, map[uint64]rbac.Role{}, nil)
	assert.NoError(t, err, `permissionByModel:role1`)

	mng.RegisterRole(ctx, role1)
	assert.True(t, role1.CheckPermissions(ctx, &rbacModels.Role{}, `view`), `CheckPermissions`)
}
