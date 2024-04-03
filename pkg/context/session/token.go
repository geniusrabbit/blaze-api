package session

import "context"

var ctxTokenKey = &struct{ s string }{"token"}

// WithToken set token to context
func WithToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxTokenKey, token)
}

// Token get token from context
func Token(ctx context.Context) string {
	switch tok := ctx.Value(ctxTokenKey).(type) {
	case nil:
	case string:
		return tok
	}
	return ""
}
