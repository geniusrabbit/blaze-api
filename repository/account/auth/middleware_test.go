package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	blazeauth "github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

const unauthorizedBody = `{"errors":[{"message":"Unauthorized","code":401}]}`

type mwUser struct {
	userModels.UserBase
}

func (u *mwUser) NewWithID(id uint64) user.Model {
	return &mwUser{UserBase: userModels.UserBase{ID: id}}
}

type mwAccount struct {
	accountModels.AccountBase
}

func (a *mwAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &mwAccount{}
	acc.SetID(id)
	acc.ExtendAdminUsers(adminUserIDs...)
	return acc
}

type fakeAuthorizer struct {
	user *mwUser
	acc  *mwAccount
}

func (fakeAuthorizer) AuthorizerCode() string { return "fake" }

func (f fakeAuthorizer) Authorize(http.ResponseWriter, *http.Request) (string, *mwUser, *mwAccount, error) {
	return "tok", f.user, f.acc, nil
}

type stubMembers struct {
	countError error
}

func (stubMembers) EmptyObject() *account.Member[*mwUser, *mwAccount] { return nil }
func (stubMembers) IsAdmin(context.Context, uint64, uint64) bool      { return false }
func (stubMembers) IsMember(context.Context, uint64, uint64) bool     { return true }
func (stubMembers) FetchListMembers(context.Context, ...account.QOption) ([]*account.Member[*mwUser, *mwAccount], error) {
	return nil, nil
}
func (stub stubMembers) CountMembers(context.Context, ...account.QOption) (int64, error) {
	if stub.countError != nil {
		return 0, stub.countError
	}
	return 1, nil
}
func (stubMembers) Member(context.Context, uint64, uint64) (*account.Member[*mwUser, *mwAccount], error) {
	return nil, nil
}
func (stubMembers) MemberByID(context.Context, uint64) (*account.Member[*mwUser, *mwAccount], error) {
	return nil, nil
}
func (stubMembers) LinkMember(context.Context, *mwAccount, bool, ...*mwUser) error { return nil }
func (stubMembers) UnlinkMember(context.Context, *mwAccount, ...*mwUser) error     { return nil }
func (stubMembers) SetMemberRoles(context.Context, *mwAccount, *mwUser, ...string) error {
	return nil
}

type stubSession struct{}

func (stubSession) EmptyObject() *mwAccount                         { return &mwAccount{} }
func (stubSession) Get(context.Context, uint64) (*mwAccount, error) { return nil, nil }
func (stubSession) FetchList(context.Context, ...account.QOption) ([]*mwAccount, error) {
	return nil, nil
}
func (stubSession) Count(context.Context, ...account.QOption) (int64, error) { return 0, nil }
func (stubSession) Create(context.Context, *mwAccount) (uint64, error)       { return 0, nil }
func (stubSession) Update(context.Context, uint64, *mwAccount) error         { return nil }
func (stubSession) Delete(context.Context, uint64) error                     { return nil }
func (stubSession) LoadPermissions(context.Context, *mwAccount, *mwUser) error {
	return nil
}
func (stubSession) GetByToken(context.Context, string) (*mwUser, *mwAccount, error) {
	return nil, nil, nil
}

func TestMiddlewareRejectedUser(t *testing.T) {
	t.Parallel()

	nextCalled := false
	handler := Middleware[*mwUser, *mwAccount](
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }),
		nil,
		fakeAuthorizer{user: &mwUser{UserBase: userModels.UserBase{
			ID:      4,
			Approve: pkgModels.DisapprovedApproveStatus,
		}}},
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if rec.Body.String() != unauthorizedBody {
		t.Fatalf("body = %s", rec.Body.String())
	}
	if nextCalled {
		t.Fatal("next handler was called for a rejected user")
	}
}

func TestMiddlewareAllowsPendingApprovedAndAnonymous(t *testing.T) {
	t.Parallel()

	accountOK := &mwAccount{}
	accountOK.SetID(9)

	cases := []struct {
		name string
		user *mwUser
		acc  *mwAccount
	}{
		{name: "pending", user: &mwUser{UserBase: userModels.UserBase{ID: 5, Approve: pkgModels.PendingApproveStatus}}, acc: accountOK},
		{name: "approved", user: &mwUser{UserBase: userModels.UserBase{ID: 6, Approve: pkgModels.ApprovedApproveStatus}}, acc: accountOK},
		{name: "anonymous"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			loader := NewLoader[*mwUser, *mwAccount](nil, stubSession{}, stubMembers{})
			handler := Middleware[*mwUser, *mwAccount](
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					nextCalled = true
					w.WriteHeader(http.StatusNoContent)
				}),
				loader,
				fakeAuthorizer{user: tc.user, acc: tc.acc},
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(permissions.WithManager(req.Context(), permissions.NewTestManager(req.Context())))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if !nextCalled {
				t.Fatal("next handler was not called")
			}
			if rec.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want 204", rec.Code)
			}
		})
	}
}

type codedAuthorizer struct {
	fakeAuthorizer
	code string
}

func (authorizer codedAuthorizer) AuthorizerCode() string { return authorizer.code }

func TestMiddlewareStoresAuthorizerCode(t *testing.T) {
	t.Parallel()

	accountObject := &mwAccount{}
	accountObject.SetID(9)
	userObject := &mwUser{UserBase: userModels.UserBase{ID: 6, Approve: pkgModels.ApprovedApproveStatus}}

	for _, code := range []string{"jwt", session.AuthorizerCodeDirectToken} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()

			storedCode := ""
			loader := NewLoader[*mwUser, *mwAccount](nil, stubSession{}, stubMembers{})
			handler := Middleware[*mwUser, *mwAccount](
				http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
					storedCode = session.AuthorizerCode(request.Context())
				}),
				loader,
				codedAuthorizer{fakeAuthorizer: fakeAuthorizer{user: userObject, acc: accountObject}, code: code},
			)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(permissions.WithManager(request.Context(), permissions.NewTestManager(request.Context())))
			handler.ServeHTTP(httptest.NewRecorder(), request)

			if storedCode != code {
				t.Fatalf("stored code = %q, want %q", storedCode, code)
			}
		})
	}
}

func TestMiddlewareRejectsCrossAccountForDirectToken(t *testing.T) {
	t.Parallel()

	accountObject := &mwAccount{}
	accountObject.SetID(9)
	userObject := &mwUser{UserBase: userModels.UserBase{ID: 6, Approve: pkgModels.ApprovedApproveStatus}}

	nextCalled := false
	loader := NewLoader[*mwUser, *mwAccount](nil, stubSession{}, stubMembers{})
	handler := Middleware[*mwUser, *mwAccount](
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }),
		loader,
		codedAuthorizer{
			fakeAuthorizer: fakeAuthorizer{user: userObject, acc: accountObject},
			code:           session.AuthorizerCodeDirectToken,
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(session.CrossAuthHeader, session.NewCrossAuthHeader(10, 6))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if nextCalled {
		t.Fatal("next handler was called for a cross-account API key request")
	}
}

type failingAuthorizer struct {
	err error
}

func (failingAuthorizer) AuthorizerCode() string { return "failing" }

func (authorizer failingAuthorizer) Authorize(http.ResponseWriter, *http.Request) (string, *mwUser, *mwAccount, error) {
	return "", nil, nil, authorizer.err
}

func TestMiddlewareStatusForAuthorizerErrors(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err        error
		wantStatus int
	}{
		"rejected credential": {err: errors.New("user is not a member of the account"), wantStatus: http.StatusUnauthorized},
		"service unavailable": {err: fmt.Errorf("%w: connection refused", blazeauth.ErrAuthorizationUnavailable), wantStatus: http.StatusServiceUnavailable},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			handler := Middleware[*mwUser, *mwAccount](
				http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }),
				nil,
				failingAuthorizer{err: testCase.err},
			)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
			if nextCalled {
				t.Fatal("next handler was called")
			}
		})
	}
}

func TestMiddlewareAnswersUnavailableWhenMembershipCheckFails(t *testing.T) {
	t.Parallel()

	accountObject := &mwAccount{}
	accountObject.SetID(9)
	userObject := &mwUser{UserBase: userModels.UserBase{ID: 6, Approve: pkgModels.ApprovedApproveStatus}}

	nextCalled := false
	loader := NewLoader[*mwUser, *mwAccount](nil, stubSession{}, stubMembers{countError: errors.New("replica timeout")})
	handler := Middleware[*mwUser, *mwAccount](
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }),
		loader,
		fakeAuthorizer{user: userObject, acc: accountObject},
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
	if nextCalled {
		t.Fatal("next handler was called")
	}
}
