package directives

import (
	"context"
	"errors"
	"testing"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// permUser is a consumer user type registered as the "user" resource.
type permUser struct{ userModels.UserBase }

func (*permUser) RBACResourceName() string { return "user" }
func (*permUser) NewWithID(id uint64) user.Model {
	return &permUser{UserBase: userModels.UserBase{ID: id}}
}

func TestHasPermissionsAnonymousUserResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pm := &permissions.Manager{Manager: rbac.NewManager(nil)}
	acl.InitModelPermissions(pm, &permUser{})
	_ = pm.RegisterNewOwningPermissions(&permUser{}, []string{"password.reset"})
	pm.RegisterRole(ctx,
		rbac.MustNewRole(session.AnonymousDefaultRole, rbac.WithPermissions("user.password.reset.all")))
	ctx = session.WithAnonymousUserAccount(permissions.WithManager(ctx, pm))

	called := false
	next := func(context.Context) (any, error) { called = true; return "ok", nil }

	// Before the fix the object was built from the anonymous session user
	// type, which never matches the registered "user" resource
	if _, err := HasPermissions[*permUser, *cacheTestAccount](ctx, nil, next, []string{"user.password.reset.*"}); err != nil || !called {
		t.Fatalf("anonymous with user.password.reset.all: err = %v, called = %v", err, called)
	}

	called = false
	if _, err := HasPermissions[*permUser, *cacheTestAccount](ctx, nil, next, []string{"user.update.*"}); !errors.Is(err, errAuthorizationRequired) || called {
		t.Fatalf("permission the role doesn't have: err = %v, called = %v", err, called)
	}
}
