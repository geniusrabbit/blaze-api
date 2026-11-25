package session

import (
	"context"
	"errors"
)

// AuthorizerCodeDirectToken is the code of the authorizer that accepts direct access tokens (API keys).
const AuthorizerCodeDirectToken = "directtoken"

// ErrDirectTokenSessionForbidden is returned when an operation is not allowed for a session authenticated by an API key.
var ErrDirectTokenSessionForbidden = errors.New("operation is not allowed for API key sessions")

var ctxAuthorizerCodeKey = &struct{ s string }{"authorizer_code"}

// WithAuthorizerCode stores the code of the authorizer that authenticated the request.
func WithAuthorizerCode(ctx context.Context, code string) context.Context {
	if code == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxAuthorizerCodeKey, code)
}

// AuthorizerCode returns the code of the authorizer that authenticated the request.
func AuthorizerCode(ctx context.Context) string {
	code, _ := ctx.Value(ctxAuthorizerCodeKey).(string)
	return code
}

// IsDirectTokenSession reports whether the request is authenticated by an API key.
func IsDirectTokenSession(ctx context.Context) bool {
	return AuthorizerCode(ctx) == AuthorizerCodeDirectToken
}

// DenyDirectTokenSession returns ErrDirectTokenSessionForbidden for a session authenticated by an API key.
func DenyDirectTokenSession(ctx context.Context) error {
	if IsDirectTokenSession(ctx) {
		return ErrDirectTokenSessionForbidden
	}
	return nil
}
