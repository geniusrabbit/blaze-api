package session

import (
	"context"

	scs "github.com/alexedwards/scs/v2"
)

var ctxSessionStorage = struct{ s string }{"session"}

// WithSession context wrapper
func WithSession(ctx context.Context, manager *scs.SessionManager) context.Context {
	return context.WithValue(ctx, ctxSessionStorage, manager)
}

// Get session storage
func Get(ctx context.Context) *scs.SessionManager {
	return ctx.Value(ctxSessionStorage).(*scs.SessionManager)
}
