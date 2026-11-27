package jwt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/cache/memory"
)

const testSecret = "test-secret-0123456789-0123456789-abcdef"

func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	return &Provider{TokenLifetime: time.Hour, Secret: testSecret, Issuer: "blaze", Audience: "api"}
}

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func validClaims() jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{"uid": 5, "acc": 3, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix(), "iss": "blaze", "aud": "api", "jti": "id-1"}
}

func TestCreateAndParseToken(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	raw, expireAt, err := p.CreateToken(5, 3, 7)
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.ParseToken(context.Background(), raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if data.UserID != 5 || data.AccountID != 3 || data.SocialAccountID != 7 || data.ID == "" || data.ExpireAt != expireAt.Unix() {
		t.Fatalf("data = %+v", data)
	}

	other, _, _ := p.CreateToken(5, 3, 7)
	if other == raw {
		t.Fatal("two tokens are identical: jti is not random")
	}
}

func TestParseTokenRejects(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	expired := validClaims()
	expired["exp"] = time.Now().Add(-time.Hour).Unix()
	noExp := validClaims()
	delete(noExp, "exp")
	noUser := validClaims()
	delete(noUser, "uid")
	foreignIss := validClaims()
	foreignIss["iss"] = "other"
	foreignAud := validClaims()
	foreignAud["aud"] = "other"
	future := validClaims()
	future["iat"] = time.Now().Add(time.Hour).Unix()

	cases := map[string]string{
		"alg none":         signed(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims()),
		"other HMAC alg":   signed(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims()),
		"wrong secret":     signed(t, jwt.SigningMethodHS256, []byte("another-secret-0123456789-0123456789"), validClaims()),
		"expired":          signed(t, jwt.SigningMethodHS256, []byte(testSecret), expired),
		"no expiration":    signed(t, jwt.SigningMethodHS256, []byte(testSecret), noExp),
		"no user":          signed(t, jwt.SigningMethodHS256, []byte(testSecret), noUser),
		"foreign issuer":   signed(t, jwt.SigningMethodHS256, []byte(testSecret), foreignIss),
		"foreign audience": signed(t, jwt.SigningMethodHS256, []byte(testSecret), foreignAud),
		"issued in future": signed(t, jwt.SigningMethodHS256, []byte(testSecret), future),
		"garbage":          "not.a.token",
	}
	for name, raw := range cases {
		if _, err := p.ParseToken(context.Background(), raw); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: err = %v, want ErrInvalidToken", name, err)
		}
	}
}

func TestShortSecretIsRejected(t *testing.T) {
	t.Parallel()

	// Before the fix an empty secret signed forgeable tokens
	for _, secret := range []string{"", "short"} {
		p := &Provider{Secret: secret}
		if _, _, err := p.CreateToken(1, 0, 0); !errors.Is(err, ErrSecretTooShort) {
			t.Errorf("secret %q: create err = %v", secret, err)
		}
		if _, err := p.ParseToken(context.Background(), signed(t, jwt.SigningMethodHS256, []byte(secret), validClaims())); err == nil {
			t.Errorf("secret %q: token accepted", secret)
		}
	}
}

func TestRevocation(t *testing.T) {
	t.Parallel()

	c, err := memory.NewTimeout(context.Background(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	p := newTestProvider(t)
	p.Revocation = NewCacheRevocationStore(c)
	ctx := context.Background()

	first, _, _ := p.CreateToken(5, 3, 0)
	second, _, _ := p.CreateToken(5, 3, 0)
	if err := p.Revoke(ctx, first); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := p.ParseToken(ctx, first); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("revoked token: err = %v", err)
	}
	if _, err := p.ParseToken(ctx, second); err != nil {
		t.Fatalf("another token of the user was revoked: %v", err)
	}

	// All tokens issued before the user-wide revocation are rejected
	old := signed(t, jwt.SigningMethodHS256, []byte(testSecret), func() jwt.MapClaims {
		cl := validClaims()
		cl["iat"] = time.Now().Add(-time.Minute).Unix()
		cl["jti"] = "old"
		return cl
	}())
	if err := p.RevokeUserTokens(ctx, 5); err != nil {
		t.Fatal(err)
	}
	if _, err := p.ParseToken(ctx, old); !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("token issued before the user revocation: err = %v", err)
	}

	if err := (&Provider{Secret: testSecret}).Revoke(ctx, second); !errors.Is(err, ErrRevocationDisabled) {
		t.Fatalf("revoke without store: err = %v", err)
	}
}

type failingStore struct{ RevocationStore }

func (failingStore) IsTokenRevoked(context.Context, string) (bool, error) {
	return false, errors.New("redis down")
}

func TestRevocationStoreFailureIsUnavailable(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	p.Revocation = failingStore{}
	raw, _, _ := p.CreateToken(5, 3, 0)
	if _, err := p.ParseToken(context.Background(), raw); !errors.Is(err, auth.ErrAuthorizationUnavailable) {
		t.Fatalf("err = %v, want ErrAuthorizationUnavailable", err)
	}
}

func TestExtractToken(t *testing.T) {
	t.Parallel()

	p := newTestProvider(t)
	req := httptest.NewRequest(http.MethodGet, "/graphql?access_token=from-query", nil)
	if token, _ := p.ExtractToken(req); token != "" {
		t.Fatalf("query token accepted by default: %q", token)
	}
	req.Header.Set("Authorization", "Bearer from-header")
	if token, _ := p.ExtractToken(req); token != "from-header" {
		t.Fatalf("header token = %q", token)
	}

	p.Extractor = tokenextractor.HeaderOrWebsocketQuery("access_token")
	plain := httptest.NewRequest(http.MethodPost, "/graphql?access_token=from-query", strings.NewReader("{}"))
	if token, _ := p.ExtractToken(plain); token != "" {
		t.Fatalf("query token accepted on a plain request: %q", token)
	}
	ws := httptest.NewRequest(http.MethodGet, "/graphql?access_token=from-query", nil)
	ws.Header.Set("Upgrade", "websocket")
	if token, _ := p.ExtractToken(ws); token != "from-query" {
		t.Fatalf("websocket query token = %q", token)
	}
}
