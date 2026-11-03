package accountlogin

import (
	"context"
	"errors"
	"testing"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

var (
	errContinued = errors.New("continued")
	errPassword  = errors.New("invalid password")
)

type loginUser struct {
	userModels.UserBase
}

func (u *loginUser) NewWithID(id uint64) user.Model {
	return &loginUser{UserBase: userModels.UserBase{ID: id}}
}

type loginAccount struct {
	accountModels.AccountBase
}

func (a *loginAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &loginAccount{}
	acc.SetID(id)
	acc.ExtendAdminUsers(adminUserIDs...)
	return acc
}

type stubLogin struct {
	user *loginUser
	err  error
}

func (s stubLogin) Login(context.Context, string, string) (*loginUser, error) {
	return s.user, s.err
}

type stubAccounts struct {
	fetchCalls int
}

func (s *stubAccounts) EmptyObject() *loginAccount { return &loginAccount{} }
func (s *stubAccounts) Get(context.Context, uint64) (*loginAccount, error) {
	return nil, errors.New("unused")
}
func (s *stubAccounts) FetchList(context.Context, ...account.QOption) ([]*loginAccount, error) {
	s.fetchCalls++
	return nil, errContinued
}
func (s *stubAccounts) Count(context.Context, ...account.QOption) (int64, error) { return 0, nil }
func (s *stubAccounts) Create(context.Context, *loginAccount) (uint64, error)    { return 0, nil }
func (s *stubAccounts) Update(context.Context, uint64, *loginAccount) error      { return nil }
func (s *stubAccounts) Delete(context.Context, uint64) error                     { return nil }
func (s *stubAccounts) LoadPermissions(context.Context, *loginAccount, *loginUser) error {
	return nil
}
func (s *stubAccounts) GetByToken(context.Context, string) (*loginUser, *loginAccount, error) {
	return nil, nil, nil
}

func TestLoginRejectsDisapprovedUser(t *testing.T) {
	t.Parallel()

	accounts := &stubAccounts{}
	resolver := New[*loginUser, *loginAccount](nil, stubLogin{
		user: &loginUser{UserBase: userModels.UserBase{
			ID:      4,
			Approve: pkgModels.DisapprovedApproveStatus,
		}},
	}, accounts)

	_, err := resolver.Login(context.Background(), "a@b.c", "secret")
	if !errors.Is(err, errAuthorizationRequired) {
		t.Fatalf("Login() error = %v, want authorization required", err)
	}
	if accounts.fetchCalls != 0 {
		t.Fatalf("FetchList calls = %d, want 0", accounts.fetchCalls)
	}
}

func TestLoginAllowsPendingAndApproved(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		approve pkgModels.ApproveStatus
	}{
		{name: "pending", approve: pkgModels.PendingApproveStatus},
		{name: "approved", approve: pkgModels.ApprovedApproveStatus},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			accounts := &stubAccounts{}
			resolver := New[*loginUser, *loginAccount](nil, stubLogin{
				user: &loginUser{UserBase: userModels.UserBase{ID: 5, Approve: tc.approve}},
			}, accounts)

			_, err := resolver.Login(context.Background(), "a@b.c", "secret")
			if !errors.Is(err, errContinued) {
				t.Fatalf("Login() error = %v, want continued past the reject check", err)
			}
			if errors.Is(err, errAuthorizationRequired) {
				t.Fatal("Login() refused a user who is not rejected")
			}
			if accounts.fetchCalls != 1 {
				t.Fatalf("FetchList calls = %d, want 1", accounts.fetchCalls)
			}
		})
	}
}

func TestLoginWrongPasswordIsUnchanged(t *testing.T) {
	t.Parallel()

	accounts := &stubAccounts{}
	resolver := New[*loginUser, *loginAccount](nil, stubLogin{err: errPassword}, accounts)

	_, err := resolver.Login(context.Background(), "a@b.c", "wrong")
	if !errors.Is(err, errPassword) {
		t.Fatalf("Login() error = %v, want invalid password", err)
	}
	if accounts.fetchCalls != 0 {
		t.Fatalf("FetchList calls = %d, want 0", accounts.fetchCalls)
	}
}
