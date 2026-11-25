package authorizer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

type testUser struct {
	userModels.UserBase
}

func (u *testUser) NewWithID(id uint64) user.Model {
	return &testUser{UserBase: userModels.UserBase{ID: id}}
}

type testAccount struct {
	accountModels.AccountBase
}

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	accountObject := &testAccount{}
	accountObject.SetID(id)
	accountObject.ExtendAdminUsers(adminUserIDs...)
	return accountObject
}

type stubTokens struct {
	tokens map[string]*directaccesstoken.DirectAccessToken
	err    error
}

func (stub stubTokens) GetByToken(_ context.Context, token string) (*directaccesstoken.DirectAccessToken, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	if tokenObject, ok := stub.tokens[token]; ok {
		return tokenObject, nil
	}
	return &directaccesstoken.DirectAccessToken{}, nil
}

type stubUsers struct{}

func (stubUsers) EmptyObject() *testUser { return &testUser{} }
func (stubUsers) Get(_ context.Context, id uint64) (*testUser, error) {
	return &testUser{UserBase: userModels.UserBase{ID: id}}, nil
}
func (stubUsers) FetchList(context.Context, ...user.QOption) ([]*testUser, error) { return nil, nil }
func (stubUsers) Count(context.Context, ...user.QOption) (int64, error)           { return 0, nil }
func (stubUsers) Create(context.Context, *testUser) (uint64, error)               { return 0, nil }
func (stubUsers) Update(context.Context, *testUser) error                         { return nil }
func (stubUsers) Delete(context.Context, uint64) error                            { return nil }

type stubAccounts struct {
	sessionTokenLookups *int
}

func (stubAccounts) EmptyObject() *testAccount { return &testAccount{} }
func (stubAccounts) Get(_ context.Context, id uint64) (*testAccount, error) {
	accountObject := &testAccount{}
	accountObject.SetID(id)
	return accountObject, nil
}
func (stubAccounts) FetchList(context.Context, ...account.QOption) ([]*testAccount, error) {
	return nil, nil
}
func (stubAccounts) Count(context.Context, ...account.QOption) (int64, error) { return 0, nil }
func (stubAccounts) Create(context.Context, *testAccount) (uint64, error)     { return 0, nil }
func (stubAccounts) Update(context.Context, uint64, *testAccount) error       { return nil }
func (stubAccounts) Delete(context.Context, uint64) error                     { return nil }
func (stubAccounts) LoadPermissions(context.Context, *testAccount, *testUser) error {
	return nil
}
func (stub stubAccounts) GetByToken(context.Context, string) (*testUser, *testAccount, error) {
	if stub.sessionTokenLookups != nil {
		*stub.sessionTokenLookups++
	}
	return nil, nil, nil
}

type stubMembers struct {
	member     bool
	countError error
}

func (stubMembers) EmptyObject() *account.Member[*testUser, *testAccount] { return nil }
func (stubMembers) IsAdmin(context.Context, uint64, uint64) bool          { return false }
func (stub stubMembers) IsMember(context.Context, uint64, uint64) bool    { return stub.member }
func (stubMembers) FetchListMembers(context.Context, ...account.QOption) ([]*account.Member[*testUser, *testAccount], error) {
	return nil, nil
}
func (stub stubMembers) CountMembers(context.Context, ...account.QOption) (int64, error) {
	if stub.countError != nil {
		return 0, stub.countError
	}
	if stub.member {
		return 1, nil
	}
	return 0, nil
}
func (stubMembers) Member(context.Context, uint64, uint64) (*account.Member[*testUser, *testAccount], error) {
	return nil, nil
}
func (stubMembers) MemberByID(context.Context, uint64) (*account.Member[*testUser, *testAccount], error) {
	return nil, nil
}
func (stubMembers) LinkMember(context.Context, *testAccount, bool, ...*testUser) error { return nil }
func (stubMembers) UnlinkMember(context.Context, *testAccount, ...*testUser) error     { return nil }
func (stubMembers) SetMemberRoles(context.Context, *testAccount, *testUser, ...string) error {
	return nil
}

func newBearerRequest(token string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func TestDirectTokenAuthorizerResolvesTokenFromDirectAccessTokens(t *testing.T) {
	t.Parallel()

	sessionTokenLookups := 0
	loader := accauth.NewLoader[*testUser, *testAccount](
		stubUsers{}, stubAccounts{sessionTokenLookups: &sessionTokenLookups}, stubMembers{member: true})
	tokens := stubTokens{tokens: map[string]*directaccesstoken.DirectAccessToken{
		"api-key": {ID: 1, Token: "api-key", UserID: sql.Null[uint64]{V: 7, Valid: true}, AccountID: 9},
	}}
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, tokens)

	token, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("api-key"))
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if token != "api-key" {
		t.Fatalf("token = %q, want api-key", token)
	}
	if userObject == nil || userObject.GetID() != 7 {
		t.Fatalf("user = %v, want id 7", userObject)
	}
	if accountObject == nil || accountObject.GetID() != 9 {
		t.Fatalf("account = %v, want id 9", accountObject)
	}
	if sessionTokenLookups != 0 {
		t.Fatalf("auth_session lookup called %d times", sessionTokenLookups)
	}
}

func TestDirectTokenAuthorizerAccountOnlyToken(t *testing.T) {
	t.Parallel()

	loader := accauth.NewLoader[*testUser, *testAccount](nil, stubAccounts{}, stubMembers{member: true})
	tokens := stubTokens{tokens: map[string]*directaccesstoken.DirectAccessToken{
		"account-key": {ID: 2, Token: "account-key", AccountID: 11},
	}}
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, tokens)

	_, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("account-key"))
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if userObject != nil {
		t.Fatalf("user = %v, want nil", userObject)
	}
	if accountObject == nil || accountObject.GetID() != 11 {
		t.Fatalf("account = %v, want id 11", accountObject)
	}
}

func TestDirectTokenAuthorizerUnknownTokenIsNotMine(t *testing.T) {
	t.Parallel()

	loader := accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: true})
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, stubTokens{})

	token, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("unknown"))
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if token != "" || userObject != nil || accountObject != nil {
		t.Fatalf("got token=%q user=%v account=%v, want empty", token, userObject, accountObject)
	}
}

func TestDirectTokenAuthorizerNoToken(t *testing.T) {
	t.Parallel()

	loader := accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: true})
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, stubTokens{err: errors.New("must not be called")})

	request := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	_, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), request)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if userObject != nil || accountObject != nil {
		t.Fatalf("got user=%v account=%v, want empty", userObject, accountObject)
	}
}

func TestDirectTokenAuthorizerLookupError(t *testing.T) {
	t.Parallel()

	lookupError := errors.New("database is down")
	loader := accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: true})
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, stubTokens{err: lookupError})

	_, _, _, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("api-key"))
	if !errors.Is(err, lookupError) {
		t.Fatalf("err = %v, want %v", err, lookupError)
	}
	if !errors.Is(err, auth.ErrAuthorizationUnavailable) {
		t.Fatalf("err = %v, want %v", err, auth.ErrAuthorizationUnavailable)
	}
}

type failingUsers struct {
	stubUsers
	err error
}

func (users failingUsers) Get(context.Context, uint64) (*testUser, error) { return nil, users.err }

func TestDirectTokenAuthorizerClassifiesLoaderErrors(t *testing.T) {
	t.Parallel()

	tokens := stubTokens{tokens: map[string]*directaccesstoken.DirectAccessToken{
		"api-key": {ID: 1, Token: "api-key", UserID: sql.Null[uint64]{V: 7, Valid: true}, AccountID: 9},
	}}
	cases := map[string]struct {
		loader          *accauth.Loader[*testUser, *testAccount]
		wantUnavailable bool
	}{
		"user is not a member": {
			loader: accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: false}),
		},
		"user was deleted": {
			loader: accauth.NewLoader[*testUser, *testAccount](failingUsers{err: fmt.Errorf("load user: %w", gorm.ErrRecordNotFound)}, stubAccounts{}, stubMembers{member: true}),
		},
		"database is down": {
			loader:          accauth.NewLoader[*testUser, *testAccount](failingUsers{err: errors.New("connection refused")}, stubAccounts{}, stubMembers{member: true}),
			wantUnavailable: true,
		},
		"membership check fails": {
			loader:          accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: true, countError: errors.New("replica timeout")}),
			wantUnavailable: true,
		},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](testCase.loader, tokens)
			_, _, _, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("api-key"))
			if err == nil {
				t.Fatal("authorize succeeded")
			}
			if errors.Is(err, auth.ErrAuthorizationUnavailable) != testCase.wantUnavailable {
				t.Fatalf("err = %v, unavailable = %v, want %v", err, !testCase.wantUnavailable, testCase.wantUnavailable)
			}
		})
	}
}

func TestDirectTokenAuthorizerReadsAuthorizationHeaderOnly(t *testing.T) {
	t.Parallel()

	tokens := stubTokens{tokens: map[string]*directaccesstoken.DirectAccessToken{
		"api-key": {ID: 1, Token: "api-key", AccountID: 9},
	}}
	loader := accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: true})
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, tokens)

	requests := map[string]*http.Request{
		"query parameter": httptest.NewRequest(http.MethodGet, "/graphql?access_token=api-key", nil),
		"oauth state":     httptest.NewRequest(http.MethodGet, "/auth/facebook/callback?state=access_token%3Dapi-key", nil),
	}
	for name, request := range requests {
		_, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), request)
		if err != nil {
			t.Fatalf("%s: authorize: %v", name, err)
		}
		if userObject != nil || accountObject != nil {
			t.Fatalf("%s: key accepted outside the Authorization header", name)
		}
	}

	lowerCaseScheme := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	lowerCaseScheme.Header.Set("Authorization", "bearer api-key")
	_, _, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), lowerCaseScheme)
	if err != nil || accountObject == nil {
		t.Fatalf("lower-case bearer scheme rejected: account=%v err=%v", accountObject, err)
	}
}

func TestDirectTokenAuthorizerUserNotMember(t *testing.T) {
	t.Parallel()

	loader := accauth.NewLoader[*testUser, *testAccount](stubUsers{}, stubAccounts{}, stubMembers{member: false})
	tokens := stubTokens{tokens: map[string]*directaccesstoken.DirectAccessToken{
		"api-key": {ID: 1, Token: "api-key", UserID: sql.Null[uint64]{V: 7, Valid: true}, AccountID: 9},
	}}
	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](loader, tokens)

	_, userObject, accountObject, err := authorizer.Authorize(httptest.NewRecorder(), newBearerRequest("api-key"))
	if err == nil {
		t.Fatal("authorize succeeded for a user that is not a member of the account")
	}
	if userObject != nil || accountObject != nil {
		t.Fatalf("got user=%v account=%v, want empty", userObject, accountObject)
	}
}

func TestDirectTokenAuthorizerCode(t *testing.T) {
	t.Parallel()

	authorizer := NewDirectTokenAuthorizer[*testUser, *testAccount](nil, stubTokens{})
	if authorizer.AuthorizerCode() != "directtoken" {
		t.Fatalf("code = %q, want directtoken", authorizer.AuthorizerCode())
	}
}
