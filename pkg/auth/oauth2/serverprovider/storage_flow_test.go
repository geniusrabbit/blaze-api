package serverprovider

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/geniusrabbit/blaze-api/pkg/cache/dummy"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/user"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

const (
	flowClientID    = "spa"
	flowRedirectURI = "https://app.example.com/cb"
	flowVerifier    = "verifier-0123456789-0123456789-0123456789-0123456789"
)

// flowEnv runs a real fosite provider on top of DatabaseStorage and SQLite.
type flowEnv struct {
	t        *testing.T
	db       *gorm.DB
	provider fosite.OAuth2Provider
}

func newFlowEnv(t *testing.T) *flowEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&authclient.AuthClient{}, &authclient.AuthSession{}); err != nil {
		t.Fatal(err)
	}
	err = db.Create(&authclient.AuthClient{
		ID:            flowClientID,
		UserID:        1,
		Title:         "SPA",
		RedirectURIs:  []string{flowRedirectURI},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code", "token"},
		Scope:         "offline",
		Public:        true,
		ExpiresAt:     time.Now().Add(time.Hour),
	}).Error
	if err != nil {
		t.Fatal(err)
	}

	config := &fosite.Config{
		GlobalSecret:          []byte("flow-test-global-secret-32-bytes"),
		AccessTokenLifespan:   time.Hour,
		RefreshTokenLifespan:  24 * time.Hour,
		AuthorizeCodeLifespan: 10 * time.Minute,
		EnforcePKCE:           true,
	}
	storage := NewDatabaseStorage(db, nil, dummy.New(), time.Hour)
	provider := NewProvider(config, storage, &compose.CommonStrategy{CoreStrategy: compose.NewOAuth2HMACStrategy(config)}, nil)
	return &flowEnv{t: t, db: db, provider: provider}
}

func newFlowSession() *Session {
	return NewSession(context.Background(), "user", "1", "", time.Time{}, "", time.Time{})
}

// authorize returns a fresh authorization code for the SPA client with PKCE.
func (env *flowEnv) authorize() string {
	env.t.Helper()
	challenge := sha256.Sum256([]byte(flowVerifier))
	q := url.Values{
		"client_id":             {flowClientID},
		"response_type":         {"code"},
		"redirect_uri":          {flowRedirectURI},
		"scope":                 {"offline"},
		"state":                 {"state-0123456789"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
	}
	ctx := NewContext(context.Background())
	ar, err := env.provider.NewAuthorizeRequest(ctx, httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+q.Encode(), nil))
	if err != nil {
		env.t.Fatalf("authorize request: %v", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
	}
	ar.GrantScope("offline")
	resp, err := env.provider.NewAuthorizeResponse(ctx, ar, newFlowSession())
	if err != nil {
		env.t.Fatalf("authorize response: %v", fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
	}
	return resp.GetCode()
}

// token calls the token endpoint and returns the access and refresh tokens.
func (env *flowEnv) token(form url.Values) (access, refresh string, err error) {
	form.Set("client_id", flowClientID)
	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := NewContext(context.Background())
	accessReq, err := env.provider.NewAccessRequest(ctx, req, newFlowSession())
	if err != nil {
		return "", "", err
	}
	resp, err := env.provider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		return "", "", err
	}
	refresh, _ = resp.GetExtra("refresh_token").(string)
	return resp.GetAccessToken(), refresh, nil
}

func (env *flowEnv) exchangeCode(code string) (string, string, error) {
	return env.token(url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {flowRedirectURI},
		"code_verifier": {flowVerifier},
	})
}

func (env *flowEnv) refresh(token string) (string, string, error) {
	return env.token(url.Values{"grant_type": {"refresh_token"}, "refresh_token": {token}})
}

func (env *flowEnv) accessTokenValid(token string) bool {
	_, _, err := env.provider.IntrospectToken(NewContext(context.Background()), token, fosite.AccessToken, newFlowSession())
	return err == nil
}

func mustNoErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
	}
}

func TestFlowAuthorizationCodeWithPKCE(t *testing.T) {
	env := newFlowEnv(t)

	access, refresh, err := env.exchangeCode(env.authorize())
	mustNoErr(t, err, "exchange code")
	if access == "" || refresh == "" {
		t.Fatalf("access = %q, refresh = %q", access, refresh)
	}
	if !env.accessTokenValid(access) {
		t.Fatal("issued access token is not valid")
	}
}

func TestFlowAuthorizationCodeCannotBeReplayed(t *testing.T) {
	env := newFlowEnv(t)

	code := env.authorize()
	_, _, err := env.exchangeCode(code)
	mustNoErr(t, err, "first exchange")
	if _, _, err = env.exchangeCode(code); err == nil {
		t.Fatal("the authorization code was accepted twice")
	}
}

func TestFlowRefreshTokenRotation(t *testing.T) {
	env := newFlowEnv(t)

	oldAccess, oldRefresh, err := env.exchangeCode(env.authorize())
	mustNoErr(t, err, "exchange code")

	// Before the fix a valid refresh token was rejected as expired, and the
	// rotation panicked.
	newAccess, newRefresh, err := env.refresh(oldRefresh)
	mustNoErr(t, err, "refresh")
	if newAccess == "" || newRefresh == "" || newRefresh == oldRefresh {
		t.Fatalf("access = %q, refresh = %q", newAccess, newRefresh)
	}
	if !env.accessTokenValid(newAccess) {
		t.Fatal("new access token is not valid")
	}
	if env.accessTokenValid(oldAccess) {
		t.Fatal("old access token is still valid after refresh")
	}
	if _, _, err = env.refresh(oldRefresh); err == nil {
		t.Fatal("the old refresh token was accepted after rotation")
	}

	// The new refresh token keeps working
	_, _, err = env.refresh(newRefresh)
	mustNoErr(t, err, "second refresh")
}

func TestFlowRevokedRefreshTokenIsRejected(t *testing.T) {
	env := newFlowEnv(t)

	access, refresh, err := env.exchangeCode(env.authorize())
	mustNoErr(t, err, "exchange code")

	form := url.Values{"token": {refresh}, "token_type_hint": {"refresh_token"}, "client_id": {flowClientID}}
	req := httptest.NewRequest(http.MethodPost, "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mustNoErr(t, env.provider.NewRevocationRequest(NewContext(context.Background()), req), "revoke")

	// Before the fix the inverted expiry check accepted revoked refresh tokens
	if _, _, err = env.refresh(refresh); err == nil {
		t.Fatal("a revoked refresh token was accepted")
	}
	if env.accessTokenValid(access) {
		t.Fatal("the access token of a revoked grant is still valid")
	}
}

func TestFlowImplicitGrantIsNotSupported(t *testing.T) {
	env := newFlowEnv(t)

	q := url.Values{
		"client_id":     {flowClientID},
		"response_type": {"token"},
		"redirect_uri":  {flowRedirectURI},
		"state":         {"state-0123456789"},
	}
	ctx := NewContext(context.Background())
	ar, err := env.provider.NewAuthorizeRequest(ctx, httptest.NewRequest(http.MethodGet, "/oauth2/auth?"+q.Encode(), nil))
	if err == nil {
		// Before the fix this reached CreateImplicitAccessTokenSession, which panicked
		_, err = env.provider.NewAuthorizeResponse(ctx, ar, newFlowSession())
	}
	if err == nil {
		t.Fatal("the implicit grant issued a token")
	}

	// The storage method returns an error instead of panicking
	storage := NewDatabaseStorage(env.db, nil, dummy.New(), time.Hour)
	if err := storage.CreateImplicitAccessTokenSession(ctx, "sig", nil); !errors.Is(err, fosite.ErrUnsupportedResponseType) {
		t.Fatalf("err = %v, want ErrUnsupportedResponseType", err)
	}
}

// stubUserAccessor imitates the user repositories for Authenticate.
type stubUserAccessor struct {
	users     map[string]uint64
	passwords map[uint64]string
	failWith  error
}

func (s stubUserAccessor) GetByEmail(_ context.Context, email string) (user.Model, error) {
	if s.failWith != nil {
		return nil, s.failWith
	}
	id, ok := s.users[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return testutil.Stub(id), nil
}

func (s stubUserAccessor) GetByPassword(_ context.Context, userID uint64, password string) (user.Model, error) {
	if s.passwords[userID] != password {
		return nil, userrepo.ErrInvalidPassword
	}
	return testutil.Stub(userID), nil
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()

	accessor := stubUserAccessor{
		users:     map[string]uint64{"user@example.com": 7},
		passwords: map[uint64]string{7: "correct-password"},
	}
	storage := NewDatabaseStorage(nil, accessor, dummy.New(), time.Hour)

	sub, err := storage.Authenticate(NewContext(context.Background()), "user@example.com", "correct-password")
	if err != nil || sub != "7" {
		t.Fatalf("valid credentials: subject = %q, err = %v", sub, err)
	}
	// Before the fix a wrong password was a success and a right one an error
	for name, creds := range map[string][2]string{
		"wrong password": {"user@example.com", "wrong"},
		"unknown user":   {"nobody@example.com", "correct-password"},
	} {
		if _, err := storage.Authenticate(NewContext(context.Background()), creds[0], creds[1]); !errors.Is(err, fosite.ErrNotFound) {
			t.Errorf("%s: err = %v, want fosite.ErrNotFound", name, err)
		}
	}

	failing := NewDatabaseStorage(nil, stubUserAccessor{failWith: errors.New("db down")}, dummy.New(), time.Hour)
	if _, err := failing.Authenticate(NewContext(context.Background()), "user@example.com", "x"); err == nil || errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("storage failure: err = %v, want a non-NotFound error", err)
	}
}
