package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/demdxx/rbac"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

var errNotFound = errors.New("not found")

// crossPerms grants exactly the listed permission patterns for any resource.
type crossPerms struct{ allowed map[string]bool }

func (p crossPerms) CheckPermissions(_ context.Context, _ any, patterns ...string) bool {
	return p.HasPermission(patterns...)
}
func (p crossPerms) CheckedPermissions(context.Context, any, ...string) rbac.Permission { return nil }
func (crossPerms) ChildRoles() []rbac.Role                                              { return nil }
func (crossPerms) ChildPermissions() []rbac.Permission                                  { return nil }
func (crossPerms) Permissions(...string) []rbac.Permission                              { return nil }
func (p crossPerms) HasPermission(patterns ...string) bool {
	for _, pattern := range patterns {
		if p.allowed[pattern] {
			return true
		}
	}
	return false
}

type mapUsers struct{ users map[uint64]*mwUser }

func (mapUsers) EmptyObject() *mwUser { return &mwUser{} }
func (r mapUsers) Get(_ context.Context, id uint64) (*mwUser, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, errNotFound
}
func (mapUsers) FetchList(context.Context, ...user.QOption) ([]*mwUser, error) { return nil, nil }
func (mapUsers) Count(context.Context, ...user.QOption) (int64, error)         { return 0, nil }
func (mapUsers) Create(context.Context, *mwUser) (uint64, error)               { return 0, nil }
func (mapUsers) Update(context.Context, *mwUser) error                         { return nil }
func (mapUsers) Delete(context.Context, uint64) error                          { return nil }

type mapAccounts struct {
	stubSession
	accounts map[uint64]*mwAccount
}

func (r mapAccounts) Get(_ context.Context, id uint64) (*mwAccount, error) {
	if a, ok := r.accounts[id]; ok {
		return a, nil
	}
	return nil, errNotFound
}

// mapMembers reports membership from a set of "userID:accountID" pairs.
type mapMembers struct {
	stubMembers
	members map[[2]uint64]bool
}

func (r mapMembers) CountMembers(_ context.Context, opts ...account.QOption) (int64, error) {
	for _, opt := range opts {
		if filter, ok := opt.(*account.MemberFilter); ok && len(filter.UserID) == 1 && len(filter.AccountID) == 1 {
			if r.members[[2]uint64{filter.UserID[0], filter.AccountID[0]}] {
				return 1, nil
			}
			return 0, nil
		}
	}
	return 0, nil
}

func newAccount(id uint64, perms ...string) *mwAccount {
	acc := &mwAccount{}
	acc.SetID(id)
	allowed := map[string]bool{}
	for _, p := range perms {
		allowed[p] = true
	}
	acc.SetPermissions(crossPerms{allowed: allowed})
	return acc
}

func newUser(id uint64) *mwUser {
	return &mwUser{UserBase: userModels.UserBase{ID: id, Approve: pkgModels.ApprovedApproveStatus}}
}

// Fixture: user 6 is a member of accounts 9 and 10; user 7 (the victim admin) is a member of account 9.
func newCrossLoader() *Loader[*mwUser, *mwAccount] {
	return NewLoader[*mwUser, *mwAccount](
		mapUsers{users: map[uint64]*mwUser{6: newUser(6), 7: newUser(7), 8: newUser(8)}},
		mapAccounts{accounts: map[uint64]*mwAccount{9: newAccount(9), 10: newAccount(10)}},
		mapMembers{members: map[[2]uint64]bool{{6, 9}: true, {6, 10}: true, {7, 9}: true}},
	)
}

func TestParseCrossAuthHeaderStrict(t *testing.T) {
	t.Parallel()

	ok := map[string][2]uint64{
		"9":      {9, 0},
		"9:7":    {9, 7},
		" 9 : 7": {9, 7},
	}
	for val, want := range ok {
		accID, userID, err := session.ParseCrossAuthHeaderStrict(val)
		if err != nil || accID != want[0] || userID != want[1] {
			t.Errorf("parse(%q) = %d, %d, %v; want %d, %d", val, accID, userID, err, want[0], want[1])
		}
	}
	for _, val := range []string{"", "0", "abc", "9:", "9:0", "9:x", "-1", "9:7:1", ":7"} {
		if _, _, err := session.ParseCrossAuthHeaderStrict(val); !errors.Is(err, session.ErrInvalidCrossAuthHeader) {
			t.Errorf("parse(%q) error = %v; want ErrInvalidCrossAuthHeader", val, err)
		}
	}
	if accID, userID := session.ParseCrossAuthHeader(session.NewCrossAuthHeader(10, 6)); accID != 10 || userID != 6 {
		t.Errorf("round trip = %d, %d; want 10, 6", accID, userID)
	}
}

func TestCrossAccountConnect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		header    string
		sessUser  *mwUser
		sessAcc   *mwAccount
		wantErr   error
		wantUser  uint64
		wantAccID uint64
	}{
		{name: "no header keeps session", header: "", sessUser: newUser(6), sessAcc: newAccount(9), wantUser: 6, wantAccID: 9},
		{name: "same account and user needs no permission", header: "9:6", sessUser: newUser(6), sessAcc: newAccount(9), wantUser: 6, wantAccID: 9},
		{name: "same account only needs no permission", header: "9", sessUser: newUser(6), sessAcc: newAccount(9), wantUser: 6, wantAccID: 9},
		{name: "malformed header", header: "abc", sessUser: newUser(6), sessAcc: newAccount(9), wantErr: session.ErrInvalidCrossAuthHeader},

		// Regression: a plain member could become another member of the same
		// account by sending the victim's user ID.
		{name: "impersonate user in same account without permission", header: "9:7", sessUser: newUser(6), sessAcc: newAccount(9), wantErr: errNoCrossAuthPermission},
		{name: "bare victim id is parsed as account, not user", header: "7", sessUser: newUser(6), sessAcc: newAccount(9), wantErr: errNotFound},
		{name: "switch account without permission", header: "10", sessUser: newUser(6), sessAcc: newAccount(9), wantErr: errNoCrossAuthPermission},
		{name: "session without account cannot switch", header: "9:7", sessUser: newUser(8), sessAcc: nil, wantErr: errNoCrossAuthPermission},

		{name: "switch user with permission", header: "9:7", sessUser: newUser(6), sessAcc: newAccount(9, session.PermAuthCross), wantUser: 7, wantAccID: 9},
		{name: "switch account with permission", header: "10", sessUser: newUser(6), sessAcc: newAccount(9, session.PermAuthCross), wantUser: 6, wantAccID: 10},
		{name: "target user must be member", header: "10:7", sessUser: newUser(6), sessAcc: newAccount(9, session.PermAuthCross), wantErr: ErrUserIsNotMemberOfAccount},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, acc, err := newCrossLoader().CrossAccountConnect(context.Background(), tc.header, tc.sessUser, tc.sessAcc)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v; want %v", err, tc.wantErr)
				}
				if u != nil || acc != nil {
					t.Fatalf("got session %v/%v on error; want none", u, acc)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.GetID() != tc.wantUser || acc.GetID() != tc.wantAccID {
				t.Fatalf("session = user %d account %d; want user %d account %d", u.GetID(), acc.GetID(), tc.wantUser, tc.wantAccID)
			}
		})
	}
}

func TestMiddlewareRejectsCrossAuthImpersonation(t *testing.T) {
	t.Parallel()

	// "7" is the exploit used against the old parser (it read the value as a
	// user ID); "9:7" is the documented AccountID:UserID form.
	for _, header := range []string{"7", session.NewCrossAuthHeader(9, 7)} {
		t.Run(header, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			handler := Middleware[*mwUser, *mwAccount](
				http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }),
				newCrossLoader(),
				fakeAuthorizer{user: newUser(6), acc: newAccount(9)},
			)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(permissions.WithManager(request.Context(), permissions.NewTestManager(request.Context())))
			request.Header.Set(session.CrossAuthHeader, header)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", recorder.Code)
			}
			if nextCalled {
				t.Fatal("next handler was called for an impersonation attempt")
			}
		})
	}
}
