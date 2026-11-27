package wrapper

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	accountmocks "github.com/geniusrabbit/blaze-api/repository/account/mocks"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testAccount struct {
	accountModels.AccountBase
}

func (a *testAccount) NewWithIDs(id uint64, adminUserIDs ...uint64) account.Model {
	acc := &testAccount{}
	acc.ID = id
	acc.Admins = adminUserIDs
	return acc
}

type fakeTokens struct {
	token *models.DirectAccessToken
	err   error
}

func (f fakeTokens) GetByToken(context.Context, string) (*models.DirectAccessToken, error) {
	return f.token, f.err
}

type wrapperEnv struct {
	loader   *accauth.Loader[*testutil.User, *testAccount]
	users    *usermocks.MockRepository[*testutil.User]
	accounts *accountmocks.MockSessionRepository[*testutil.User, *testAccount]
	members  *accountmocks.MockMemberRepository[*testutil.User, *testAccount]
}

func newWrapperEnv(t *testing.T) *wrapperEnv {
	ctrl := gomock.NewController(t)
	env := &wrapperEnv{
		users:    usermocks.NewMockRepository[*testutil.User](ctrl),
		accounts: accountmocks.NewMockSessionRepository[*testutil.User, *testAccount](ctrl),
		members:  accountmocks.NewMockMemberRepository[*testutil.User, *testAccount](ctrl),
	}
	env.loader = accauth.NewLoader[*testutil.User, *testAccount](env.users, env.accounts, env.members)
	return env
}

func (env *wrapperEnv) serve(tokens tokenGetter, next http.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("D-Access-Token", "key")
	rec := httptest.NewRecorder()
	httpWrapper(next, env.loader, tokens).ServeHTTP(rec, req)
	return rec
}

func validToken() *models.DirectAccessToken {
	return &models.DirectAccessToken{ID: 1, Token: "key", UserID: sql.Null[uint64]{V: 5, Valid: true}, AccountID: 3}
}

func TestHTTPWrapperMarksDirectTokenSession(t *testing.T) {
	t.Parallel()

	env := newWrapperEnv(t)
	acc := &testAccount{}
	acc.ID = 3
	usr := testutil.Stub(5)
	usr.Approve = pkgModels.ApprovedApproveStatus
	env.users.EXPECT().Get(gomock.Any(), uint64(5)).Return(usr, nil)
	env.accounts.EXPECT().Get(gomock.Any(), uint64(3)).Return(acc, nil)
	env.members.EXPECT().CountMembers(gomock.Any(), gomock.Any()).Return(int64(1), nil).AnyTimes()
	env.accounts.EXPECT().LoadPermissions(gomock.Any(), acc, usr).Return(nil).AnyTimes()

	var denyErr error
	called := false
	rec := env.serve(fakeTokens{token: validToken()}, func(_ http.ResponseWriter, r *http.Request) {
		called = true
		denyErr = session.DenyDirectTokenSession(r.Context())
	})
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("status = %d, called = %v", rec.Code, called)
	}
	// Before the fix the API key restrictions didn't apply behind this wrapper
	if !errors.Is(denyErr, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("DenyDirectTokenSession = %v, want %v", denyErr, session.ErrDirectTokenSessionForbidden)
	}
}

func TestHTTPWrapperRejectsBadTokens(t *testing.T) {
	t.Parallel()

	next := func(http.ResponseWriter, *http.Request) { t.Error("next handler was called") }

	if rec := newWrapperEnv(t).serve(fakeTokens{token: &models.DirectAccessToken{}}, next); rec.Code != http.StatusUnauthorized {
		t.Errorf("empty token: status = %d, want 401", rec.Code)
	}
	if rec := newWrapperEnv(t).serve(fakeTokens{err: gorm.ErrRecordNotFound}, next); rec.Code != http.StatusUnauthorized {
		t.Errorf("unknown token: status = %d, want 401", rec.Code)
	}
	if rec := newWrapperEnv(t).serve(fakeTokens{err: errors.New("db down")}, next); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("lookup failure: status = %d, want 503", rec.Code)
	}

	// The token owner is no longer a member of the account
	env := newWrapperEnv(t)
	acc := &testAccount{}
	acc.ID = 3
	env.users.EXPECT().Get(gomock.Any(), uint64(5)).Return(testutil.Stub(5), nil)
	env.accounts.EXPECT().Get(gomock.Any(), uint64(3)).Return(acc, nil)
	env.members.EXPECT().CountMembers(gomock.Any(), gomock.Any()).Return(int64(0), nil)
	if rec := env.serve(fakeTokens{token: validToken()}, next); rec.Code != http.StatusUnauthorized {
		t.Errorf("not a member: status = %d, want 401", rec.Code)
	}
}
