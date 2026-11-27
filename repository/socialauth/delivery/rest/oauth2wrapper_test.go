package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
	socialAccountModels "github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

const (
	testJWTSecret = "test-jwt-secret-0123456789abcdef"
	providerURL   = "https://provider.example/auth"
)

// fakeProvider imitates an OAuth2 provider: LoginURL puts the state into the
// provider URL, UserData returns a fixed social user.
type fakeProvider struct {
	socialID   string
	lastParams []elogin.URLParam
}

func (*fakeProvider) Provider() string { return "fake" }
func (*fakeProvider) Protocol() string { return "oauth2" }
func (p *fakeProvider) LoginURL(params []elogin.URLParam) string {
	p.lastParams = params
	q := url.Values{}
	for _, param := range params {
		q.Set(param.Key, param.Value)
	}
	return providerURL + "?" + q.Encode()
}
func (p *fakeProvider) UserData(context.Context, url.Values, []elogin.URLParam) (*elogin.Token, *elogin.UserData, error) {
	return &elogin.Token{AccessToken: "provider-token"}, &elogin.UserData{ID: p.socialID}, nil
}

// fakeSocialAuth keeps social accounts in memory.
type fakeSocialAuth struct {
	accounts   []*socialAccountModels.AccountSocial
	registered []uint64 // owner user IDs passed to Register
}

func (*fakeSocialAuth) Get(context.Context, uint64) (*socialAccountModels.AccountSocial, error) {
	return nil, nil
}
func (f *fakeSocialAuth) List(_ context.Context, filter *socialauth.Filter) ([]*socialAccountModels.AccountSocial, error) {
	var res []*socialAccountModels.AccountSocial
	for _, acc := range f.accounts {
		if acc.DeletedAt.Valid && !filter.RetrieveDeleted {
			continue
		}
		if len(filter.SocialID) == 1 && acc.SocialID == filter.SocialID[0] {
			res = append(res, acc)
		}
	}
	return res, nil
}
func (f *fakeSocialAuth) Register(_ context.Context, owner user.Model, acc *socialAccountModels.AccountSocial) (uint64, error) {
	ownerID := owner.GetID()
	if ownerID == 0 {
		ownerID = 42 // a newly created user
	}
	f.registered = append(f.registered, ownerID)
	acc.ID = uint64(100 + len(f.accounts))
	acc.UserID = ownerID
	f.accounts = append(f.accounts, acc)
	return acc.ID, nil
}
func (*fakeSocialAuth) Update(context.Context, uint64, *socialAccountModels.AccountSocial) error {
	return nil
}
func (*fakeSocialAuth) Token(context.Context, string, uint64) (*elogin.Token, error) { return nil, nil }
func (*fakeSocialAuth) SetToken(context.Context, string, uint64, *elogin.Token) error {
	return nil
}

type testEnv struct {
	provider   *fakeProvider
	social     *fakeSocialAuth
	wrapper    *Oauth2Wrapper
	handler    http.Handler
	linkedWith []uint64
}

func newTestEnv(t *testing.T, opts ...Option) *testEnv {
	t.Helper()
	env := &testEnv{provider: &fakeProvider{socialID: "fb-1"}, social: &fakeSocialAuth{}}
	opts = append([]Option{
		WithSocialAuthUsecase(env.social),
		WithSessionProvider(jwt.NewDefaultProvider(testJWTSecret, time.Hour, false)),
		WithAllowedRedirectURLs("https://app.example.com/auth"),
	}, opts...)
	env.wrapper = NewWrapper(env.provider, opts...)
	env.handler = env.wrapper.HandleWrapper("/auth/fake")
	return env
}

func anonymousCtx() context.Context {
	ctx := context.Background()
	ctx = permissions.WithManager(ctx, permissions.NewTestManager(ctx))
	return session.WithAnonymousUserAccount(ctx)
}

// loggedInCtx imitates the auth middleware for a request with a valid token.
func loggedInCtx(token string) context.Context {
	ctx := context.Background()
	ctx = permissions.WithManager(ctx, permissions.NewTestManager(ctx))
	return session.WithToken(session.WithUserAccountDevelop(ctx), token)
}

func (env *testEnv) do(ctx context.Context, method, target string, header http.Header, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil).WithContext(ctx)
	for k, v := range header {
		req.Header[k] = v
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	env.handler.ServeHTTP(rec, req)
	return rec
}

// login starts the flow and returns the state passed to the provider and the state cookie.
func (env *testEnv) login(t *testing.T, ctx context.Context, query string, header http.Header) (string, *http.Cookie) {
	t.Helper()
	rec := env.do(ctx, http.MethodGet, "/auth/fake/login?"+query, header)
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	loc, err := url.Parse(rec.Header().Get("Location"))
	if err != nil || !strings.HasPrefix(loc.String(), providerURL) {
		t.Fatalf("login redirect = %q, want provider URL", rec.Header().Get("Location"))
	}
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == stateCookiePrefix+"fake" {
			cookie = c
		}
	}
	if cookie == nil || cookie.Value == "" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("state cookie = %+v, want HttpOnly SameSite=Lax nonce", cookie)
	}
	return loc.Query().Get("state"), cookie
}

func (env *testEnv) callback(ctx context.Context, state string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	return env.do(ctx, http.MethodGet, "/auth/fake/callback?code=c&state="+url.QueryEscape(state), nil, cookies...)
}

func TestLoginRejectsForeignRedirect(t *testing.T) {
	t.Parallel()

	for _, target := range []string{
		"https://evil.example/steal",
		"//evil.example/steal",
		"/\\evil.example",
		"https://app.example.com.evil.example/auth",
		"https://app.example.com/other",
		"https://user@app.example.com/auth",
		"javascript:alert(1)",
	} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			env := newTestEnv(t)
			rec := env.do(anonymousCtx(), http.MethodGet, "/auth/fake/login?redirect="+url.QueryEscape(target), nil)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; location = %q", rec.Code, rec.Header().Get("Location"))
			}
			if loc := rec.Header().Get("Location"); loc != "" {
				t.Fatalf("redirected to %q", loc)
			}
		})
	}
}

func TestIsAllowedRedirect(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t, WithSuccessRedirectURL("https://front.example.com/done"))
	allowed := []string{
		"/dashboard",
		"https://app.example.com/auth",
		"https://app.example.com/auth/callback?x=1",
		"https://front.example.com/anything",
	}
	for _, target := range allowed {
		if !env.wrapper.isAllowedRedirect(target) {
			t.Errorf("%q is rejected, want allowed", target)
		}
	}
	for _, target := range []string{"https://app.example.com/authx", "http://app.example.com/auth", "ftp://app.example.com/auth", "/x\r\nLocation: y"} {
		if env.wrapper.isAllowedRedirect(target) {
			t.Errorf("%q is allowed, want rejected", target)
		}
	}
}

func TestLoginStateCarriesNoToken(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t, WithLinkSession(func(ctx context.Context, _, _ uint64) (context.Context, error) { return ctx, nil }))
	header := http.Header{"Authorization": {"Bearer session-jwt-token"}}
	state, _ := env.login(t, loggedInCtx("session-jwt-token"), "redirect=%2Fhome", header)

	decoded, err := env.wrapper.stateCodec.Decode(state)
	if err != nil {
		t.Fatalf("decode state: %v", err)
	}
	for _, p := range decoded {
		if strings.Contains(p.Value, "session-jwt-token") || p.Key == "access_token" {
			t.Fatalf("state leaks the session token: %+v", decoded)
		}
	}
	if decoded.Get(linkUserKey) != "1" || decoded.Get(linkAccountKey) != "1" {
		t.Fatalf("state = %+v, want link user 1 and account 1", decoded)
	}
}

func TestLoginLinkingRequiresAuthorizationHeader(t *testing.T) {
	t.Parallel()

	// The auth middleware accepted the token from ?access_token=, so anyone
	// could make the victim's browser open such a link.
	env := newTestEnv(t, WithLinkSession(func(ctx context.Context, _, _ uint64) (context.Context, error) { return ctx, nil }))
	rec := env.do(loggedInCtx("attacker-token"), http.MethodGet, "/auth/fake/login?access_token=attacker-token", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	// Linking is off unless WithLinkSession is set
	env = newTestEnv(t)
	header := http.Header{"Authorization": {"Bearer user-token"}}
	if rec = env.do(loggedInCtx("user-token"), http.MethodGet, "/auth/fake/login", header); rec.Code != http.StatusBadRequest {
		t.Fatalf("status without linking = %d, want 400", rec.Code)
	}
}

func TestLoginJSONResponse(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	rec := env.do(anonymousCtx(), http.MethodGet, "/auth/fake/login?response=json", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || !strings.HasPrefix(body["url"], providerURL) {
		t.Fatalf("body = %v, err = %v", body, err)
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("state cookie is not set")
	}
}

func TestCallbackRejectsBadState(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	state, cookie := env.login(t, anonymousCtx(), "", nil)
	otherState, _ := newTestEnv(t, WithStateSecret("another-secret-0123456789")).login(t, anonymousCtx(), "", nil)

	cases := map[string]struct {
		state   string
		cookies []*http.Cookie
	}{
		"no cookie (login CSRF)": {state: state},
		"wrong cookie":           {state: state, cookies: []*http.Cookie{{Name: cookie.Name, Value: "0011"}}},
		"tampered state":         {state: "x" + state, cookies: []*http.Cookie{cookie}},
		"foreign signature":      {state: otherState, cookies: []*http.Cookie{cookie}},
		"legacy unsigned state":  {state: "JwAAAAJrZXkxAAcAAAB2YWx1ZTEAAmtleTIABwAAAHZhbHVlMgAA", cookies: []*http.Cookie{cookie}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := env.callback(anonymousCtx(), tc.state, tc.cookies...)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; location = %q", rec.Code, rec.Header().Get("Location"))
			}
		})
	}
	if len(env.social.registered) != 0 {
		t.Fatalf("social accounts were registered: %v", env.social.registered)
	}

	// With a validated redirect in the state, errors go there and nowhere else
	state, _ = env.login(t, anonymousCtx(), "redirect=%2Fhome", nil)
	rec := env.callback(anonymousCtx(), state)
	if loc := rec.Header().Get("Location"); rec.Code != http.StatusTemporaryRedirect || !strings.HasPrefix(loc, "/home?error=") {
		t.Fatalf("status = %d, location = %q; want error redirect to /home", rec.Code, loc)
	}
}

func TestCallbackLoginFlow(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	state, cookie := env.login(t, anonymousCtx(), "redirect="+url.QueryEscape("https://app.example.com/auth/done"), nil)

	rec := env.callback(anonymousCtx(), state, cookie)
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	if loc.Host != "app.example.com" || loc.Path != "/auth/done" || loc.Query().Get("access_token") == "" {
		t.Fatalf("location = %q, want allowed redirect with a new session token", loc)
	}
	if len(env.social.registered) != 1 || env.social.registered[0] != 42 {
		t.Fatalf("registered owners = %v, want a new user", env.social.registered)
	}

	// The state cookie is cleared, so the state can't be replayed in this browser
	cleared := false
	for _, c := range rec.Result().Cookies() {
		cleared = cleared || (c.Name == cookie.Name && c.MaxAge < 0)
	}
	if !cleared {
		t.Fatal("state cookie is not cleared")
	}
}

func TestCallbackLinksToUserFromState(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	env.wrapper.linkSession = env.withLinkingFn()
	header := http.Header{"Authorization": {"Bearer user-token"}}
	state, cookie := env.login(t, loggedInCtx("user-token"), "", header)

	// The callback request comes from the provider redirect: no Authorization header
	rec := env.callback(anonymousCtx(), state, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(env.linkedWith) != 2 || env.linkedWith[0] != 1 || env.linkedWith[1] != 1 {
		t.Fatalf("link session = %v, want user 1 account 1", env.linkedWith)
	}
	if len(env.social.registered) != 1 || env.social.registered[0] != 1 {
		t.Fatalf("registered owners = %v, want the linking user 1", env.social.registered)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if _, ok := body["access_token"]; ok {
		t.Fatalf("linking response has a new token: %v", body)
	}
}

func TestCallbackRefusesSocialAccountOfAnotherUser(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	env.wrapper.linkSession = env.withLinkingFn()
	env.social.accounts = []*socialAccountModels.AccountSocial{{ID: 7, SocialID: "fb-1", UserID: 99}}

	header := http.Header{"Authorization": {"Bearer user-token"}}
	state, cookie := env.login(t, loggedInCtx("user-token"), "", header)
	if rec := env.callback(anonymousCtx(), state, cookie); rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCallbackLinkSessionError(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t, WithLinkSession(func(ctx context.Context, _, _ uint64) (context.Context, error) {
		return ctx, errors.New("user is not a member of the account")
	}))
	header := http.Header{"Authorization": {"Bearer user-token"}}
	state, cookie := env.login(t, loggedInCtx("user-token"), "", header)
	if rec := env.callback(anonymousCtx(), state, cookie); rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if len(env.social.registered) != 0 {
		t.Fatalf("social account registered: %v", env.social.registered)
	}
}

func TestNewWrapperRequiresStateSecret(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("NewWrapper without a state secret didn't panic")
		}
	}()
	NewWrapper(&fakeProvider{}, WithSocialAuthUsecase(&fakeSocialAuth{}))
}

func (env *testEnv) withLinkingFn() func(ctx context.Context, userID, accountID uint64) (context.Context, error) {
	return func(ctx context.Context, userID, accountID uint64) (context.Context, error) {
		env.linkedWith = append(env.linkedWith, userID, accountID)
		return session.WithUserAccountDevelop(ctx), nil
	}
}

func TestCallbackIgnoresDisconnectedSocialAccount(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)
	// The user 99 disconnected this Facebook account earlier
	env.social.accounts = []*socialAccountModels.AccountSocial{{
		ID: 7, SocialID: "fb-1", UserID: 99,
		DeletedAt: gorm.DeletedAt{Valid: true, Time: time.Now()},
	}}
	state, cookie := env.login(t, anonymousCtx(), "", nil)
	rec := env.callback(anonymousCtx(), state, cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	// Before the fix the deleted record was revived and logged in as user 99
	if len(env.social.registered) != 1 || env.social.registered[0] == 99 {
		t.Fatalf("registered owners = %v, want a new owner, not the former user", env.social.registered)
	}
}
