package permissions

import (
	"context"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
)

func loaderDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&rbacModels.Role{}, &rbacModels.M2MRole{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestListRolesKeepsPreviousOnError(t *testing.T) {
	db := loaderDB(t)
	if err := db.Create(&rbacModels.Role{Name: "viewer", PermissionPatterns: []string{"user.view.*"}}).Error; err != nil {
		t.Fatal(err)
	}
	loader := &DBRoleLoader{conn: db}
	if roles := loader.ListRoles(context.Background()); len(roles) != 1 {
		t.Fatalf("roles = %d, want 1", len(roles))
	}

	// Before the fix a database error panicked inside every permission check
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	roles := loader.ListRoles(context.Background())
	if len(roles) != 1 || roles[0].Name() != "viewer" {
		t.Fatalf("roles after a DB failure = %v, want the previous ones", roles)
	}

	if roles := (&DBRoleLoader{}).ListRoles(context.Background()); roles != nil {
		t.Fatalf("no connection: roles = %v", roles)
	}
}

func TestBuildRolesIndependentOfOrder(t *testing.T) {
	t.Parallel()

	// The parent row comes before its child: the old loader built roles in
	// row order and silently dropped the child's permissions
	models := []*rbacModels.Role{
		{ID: 1, Name: "parent", PermissionPatterns: []string{"a.view.*"}},
		{ID: 2, Name: "child", PermissionPatterns: []string{"b.view.*"}},
	}
	links := []*rbacModels.M2MRole{{ParentRoleID: 1, ChildRoleID: 2}}
	roles, err := buildRoles(models, links)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range roles {
		if role.Name() == "parent" {
			if len(role.ChildRoles()) != 1 || role.ChildRoles()[0].Name() != "child" {
				t.Fatalf("parent children = %v, want [child]", role.ChildRoles())
			}
		}
	}

	if _, err := buildRoles(models, append(links, &rbacModels.M2MRole{ParentRoleID: 2, ChildRoleID: 1})); err == nil {
		t.Fatal("a role cycle was accepted")
	}
}
