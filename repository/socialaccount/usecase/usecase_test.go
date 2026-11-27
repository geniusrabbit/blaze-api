package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type fakeRepo struct {
	socialaccount.Repository
	accounts       []*models.AccountSocial
	sessionsCalled bool
}

func (f *fakeRepo) FetchList(_ context.Context, opts ...socialaccount.QOption) ([]*models.AccountSocial, error) {
	var res []*models.AccountSocial
	for _, opt := range opts {
		if fl, ok := opt.(*socialaccount.Filter); ok {
			for _, acc := range f.accounts {
				for _, id := range fl.ID {
					if acc.ID == id {
						res = append(res, acc)
					}
				}
			}
		}
	}
	return res, nil
}

func (f *fakeRepo) FetchSessionList(context.Context, []uint64) ([]*models.AccountSocialSession, error) {
	f.sessionsCalled = true
	return []*models.AccountSocialSession{{Name: "default"}}, nil
}

// ownPerms allows viewing only the social accounts of the user 5.
type ownPerms struct{}

func (ownPerms) CheckPermissions(_ context.Context, resource any, _ ...string) bool {
	acc, ok := resource.(*models.AccountSocial)
	return ok && acc.UserID == 5
}
func (ownPerms) CheckedPermissions(context.Context, any, ...string) rbac.Permission { return nil }
func (ownPerms) ChildRoles() []rbac.Role                                            { return nil }
func (ownPerms) ChildPermissions() []rbac.Permission                                { return nil }
func (ownPerms) Permissions(...string) []rbac.Permission                            { return nil }
func (ownPerms) HasPermission(...string) bool                                       { return false }

type testAccount struct{ accountModels.AccountBase }

func (a *testAccount) NewWithIDs(id uint64, _ ...uint64) account.Model {
	acc := &testAccount{}
	acc.ID = id
	return acc
}

func TestFetchSessionListChecksAccess(t *testing.T) {
	t.Parallel()

	acc := &testAccount{}
	acc.ID = 1
	acc.SetPermissions(ownPerms{})
	ctx := session.WithUserAccount(context.Background(), testutil.Stub(5), acc)
	repo := &fakeRepo{accounts: []*models.AccountSocial{{ID: 1, UserID: 5}, {ID: 2, UserID: 6}}}
	uc := NewSocaccUsecase(repo)

	if _, err := uc.FetchSessionList(ctx, []uint64{1}); err != nil || !repo.sessionsCalled {
		t.Fatalf("own account: err = %v, called = %v", err, repo.sessionsCalled)
	}

	// Before the fix the sessions (with provider tokens) of any account were returned
	for _, ids := range [][]uint64{{2}, {1, 2}, {404}} {
		repo.sessionsCalled = false
		if _, err := uc.FetchSessionList(ctx, ids); !errors.Is(err, acl.ErrNoPermissions) || repo.sessionsCalled {
			t.Errorf("ids %v: err = %v, called = %v", ids, err, repo.sessionsCalled)
		}
	}
}
