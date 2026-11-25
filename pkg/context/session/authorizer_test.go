package session

import (
	"context"
	"errors"
	"testing"
)

func TestAuthorizerCodeRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithAuthorizerCode(context.Background(), AuthorizerCodeDirectToken)
	if code := AuthorizerCode(ctx); code != AuthorizerCodeDirectToken {
		t.Fatalf("code = %q, want %q", code, AuthorizerCodeDirectToken)
	}
	if AuthorizerCode(context.Background()) != "" {
		t.Fatal("empty context must have no authorizer code")
	}
}

func TestIsDirectTokenSession(t *testing.T) {
	t.Parallel()

	if IsDirectTokenSession(context.Background()) {
		t.Fatal("context without authorizer code reported as direct token session")
	}
	if IsDirectTokenSession(WithAuthorizerCode(context.Background(), "jwt")) {
		t.Fatal("jwt session reported as direct token session")
	}
	if !IsDirectTokenSession(WithAuthorizerCode(context.Background(), AuthorizerCodeDirectToken)) {
		t.Fatal("direct token session not detected")
	}
}

func TestDenyDirectTokenSession(t *testing.T) {
	t.Parallel()

	if err := DenyDirectTokenSession(WithAuthorizerCode(context.Background(), "jwt")); err != nil {
		t.Fatalf("jwt session denied: %v", err)
	}
	err := DenyDirectTokenSession(WithAuthorizerCode(context.Background(), AuthorizerCodeDirectToken))
	if !errors.Is(err, ErrDirectTokenSessionForbidden) {
		t.Fatalf("err = %v, want %v", err, ErrDirectTokenSessionForbidden)
	}
}
