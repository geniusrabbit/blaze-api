package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testIdentity struct {
	empty bool
}

func (identity *testIdentity) IsNil() bool { return identity == nil || identity.empty }

type testAuthorizer struct {
	code     string
	user     *testIdentity
	identity *testIdentity
}

func (authorizer testAuthorizer) AuthorizerCode() string { return authorizer.code }

func (authorizer testAuthorizer) Authorize(http.ResponseWriter, *http.Request) (string, *testIdentity, *testIdentity, error) {
	return authorizer.code + "-token", authorizer.user, authorizer.identity, nil
}

func TestAuthorizeWithCodeReturnsMatchingAuthorizer(t *testing.T) {
	t.Parallel()

	wrapper := NewAuthorizeWrapper[*testIdentity, *testIdentity](
		testAuthorizer{code: "jwt"},
		testAuthorizer{code: "directtoken", user: &testIdentity{}},
		testAuthorizer{code: "last", user: &testIdentity{}},
	)

	code, token, user, _, err := wrapper.AuthorizeWithCode(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if code != "directtoken" {
		t.Fatalf("code = %q, want directtoken", code)
	}
	if token != "directtoken-token" {
		t.Fatalf("token = %q, want directtoken-token", token)
	}
	if user == nil {
		t.Fatal("user is nil")
	}
}

func TestAuthorizeWithCodeAnonymous(t *testing.T) {
	t.Parallel()

	wrapper := NewAuthorizeWrapper[*testIdentity, *testIdentity](testAuthorizer{code: "jwt"})

	code, token, _, _, err := wrapper.AuthorizeWithCode(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if code != "" || token != "" {
		t.Fatalf("code = %q token = %q, want empty", code, token)
	}
}
