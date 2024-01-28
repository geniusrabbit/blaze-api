package permissions

import (
	"context"
	"testing"

	"github.com/demdxx/rbac"
	"github.com/geniusrabbit/gosql/v2"
	"github.com/stretchr/testify/assert"

	"github.com/geniusrabbit/blaze-api/model"
)

type TestObject struct{}

func TestObjectName(t *testing.T) {
	assert.Equal(t, `permissions:TestObject`, objectName(&TestObject{}))
}

func TestManager(t *testing.T) {
	ctx := context.TODO()
	mng := NewManager(nil, 0)
	mng.RegisterObject(&model.Role{}, func(ctx context.Context, obj *model.Role, perm rbac.Permission) bool { return true })

	dataCtx1, _ := gosql.NewNullableJSON[map[string]any](map[string]any{`object`: `model:Role`})
	permObj1 := &model.Role{ID: 1, Name: `view`, Type: model.RbacPermissionType, Context: *dataCtx1}
	perm1, err := mng.permissionByModel(permObj1, map[uint64]rbac.Role{}, map[uint64]rbac.Permission{}, nil)
	assert.NoError(t, err, `permissionByModel:perm1`)
	assert.True(t, perm1.CheckPermissions(ctx, &model.Role{}, `view`), `CheckPermissions`)

	dataCtx2, _ := gosql.NewNullableJSON[map[string]any](map[string]any{`object`: `model:User`})
	permObj2 := &model.Role{ID: 2, Name: `view`, Type: model.RbacPermissionType, Context: *dataCtx2}
	perm2, err := mng.permissionByModel(permObj2, map[uint64]rbac.Role{}, map[uint64]rbac.Permission{}, nil)
	assert.NoError(t, err, `permissionByModel:perm2`)
	assert.True(t, perm2.CheckPermissions(ctx, &model.Role{}, `view`), `CheckPermissions`)

	roleObj := &model.Role{ID: 10, Name: `role1`, Type: model.RbacRoleType}
	role1, err := mng.roleByModel(roleObj,
		map[uint64]rbac.Role{}, map[uint64]rbac.Permission{1: perm1, 2: perm2},
		[]*model.M2MRole{{ParentRoleID: 10, ChildRoleID: 1}, {ParentRoleID: 10, ChildRoleID: 2}})
	assert.NoError(t, err, `permissionByModel:role1`)
	assert.True(t, role1.CheckPermissions(ctx, &model.Role{}, `view`), `CheckPermissions`)
}
