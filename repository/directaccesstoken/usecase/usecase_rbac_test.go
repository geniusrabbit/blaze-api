package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

func realRoleContext(t *testing.T, permissionNames ...string) context.Context {
	t.Helper()
	ctx := context.Background()
	manager := permissions.NewTestManager(ctx)
	acl.InitModelPermissions(manager, &models.DirectAccessToken{})
	err := manager.RegisterNewOwningPermissions(&models.DirectAccessToken{},
		[]string{acl.PermView, acl.PermList, acl.PermCount, acl.PermCreate, acl.PermDelete})
	if err != nil {
		t.Fatalf("register permissions: %v", err)
	}
	rolePermissions := make([]any, 0, len(permissionNames))
	for _, permissionName := range permissionNames {
		rolePermissions = append(rolePermissions, permissionName)
	}
	manager.RegisterRole(ctx, rbac.MustNewRole("token-test-role", rbac.WithPermissions(rolePermissions...)))

	accountObject := &testAccount{}
	accountObject.SetID(sessionAccountID)
	accountObject.SetPermissions(manager.Role(ctx, "token-test-role"))
	ctx = permissions.WithManager(ctx, manager)
	ctx = session.WithUserAccount(ctx, testutil.Stub(sessionUserID), accountObject)
	return session.WithAuthorizerCode(ctx, "jwt")
}

func TestListScopeWithRealRoles(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		permission    string
		scoped        bool
		scopedAccount uint64
		scopedUser    uint64
	}{
		"owner":   {permission: "directaccesstoken.list.owner", scoped: true, scopedAccount: sessionAccountID, scopedUser: sessionUserID},
		"account": {permission: "directaccesstoken.list.account", scoped: true, scopedAccount: sessionAccountID},
		"all":     {permission: "directaccesstoken.list.all"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repository := &recordingRepository{}
			if _, err := New(repository).FetchList(realRoleContext(t, testCase.permission)); err != nil {
				t.Fatalf("list: %v", err)
			}
			if !testCase.scoped {
				if len(scopeFilters(repository.listOptions)) != 0 {
					t.Fatal("list.all must not be scoped")
				}
				return
			}
			if !hasScope(repository.listOptions, testCase.scopedAccount, testCase.scopedUser) {
				t.Fatalf("list is not scoped to account %d user %d", testCase.scopedAccount, testCase.scopedUser)
			}
		})
	}
}

func TestRevokeScopeWithRealOwnerRole(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	err := New(repository).Revoke(realRoleContext(t, "directaccesstoken.list.owner", "directaccesstoken.delete.owner"))
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !hasScope(repository.revokeOptions, sessionAccountID, sessionUserID) {
		t.Fatal("revoke with delete.owner is not scoped to the session user")
	}
}

func TestGenerateForAnotherUserWithRealRoles(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	_, err := New(repository).Generate(realRoleContext(t, "directaccesstoken.create.owner"), otherUserID, sessionAccountID, "", expiry())
	if !errors.Is(err, ErrTokenForAnotherUser) {
		t.Fatalf("create.owner: err = %v, want %v", err, ErrTokenForAnotherUser)
	}

	_, err = New(repository).Generate(realRoleContext(t, "directaccesstoken.create.all"), otherUserID, sessionAccountID, "", expiry())
	if err != nil {
		t.Fatalf("create.all: %v", err)
	}
	if repository.generatedUser == nil || *repository.generatedUser != otherUserID {
		t.Fatalf("generated user = %v, want %d", repository.generatedUser, otherUserID)
	}
}
