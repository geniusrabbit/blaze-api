package permissions

import (
	"context"
	"testing"

	"github.com/demdxx/rbac"
	"github.com/stretchr/testify/assert"

	"github.com/geniusrabbit/blaze-api/model"
)

type TestObject struct{}

func TestManager(t *testing.T) {
	ctx := context.TODO()
	mng := NewManager(nil, 0)

	mng.RegisterObject(&model.Role{}, func(ctx context.Context, obj *model.Role, perm rbac.Permission) bool { return true })

	_ = mng.RegisterNewPermission(&model.Role{}, `view`)
	_ = mng.RegisterNewPermission(&model.User{}, `view`)

	perm1 := mng.Permission(`role.view`)
	assert.True(t, perm1.CheckPermissions(ctx, &model.Role{}, `view`), `CheckPermissions`)

	perm2 := mng.Permission(`user.view`)
	assert.True(t, perm2.CheckPermissions(ctx, &model.User{}, `view`), `CheckPermissions`)

	roleObj := &model.Role{ID: 10, Name: `role1`, PermissionPatterns: []string{`role.*`}}
	role1, err := roleByModel(roleObj, map[uint64]rbac.Role{}, nil)
	assert.NoError(t, err, `permissionByModel:role1`)

	mng.RegisterRole(ctx, role1)
	assert.True(t, role1.CheckPermissions(ctx, &model.Role{}, `view`), `CheckPermissions`)
}
