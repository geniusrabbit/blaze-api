package context

import "context"

// SessionModel is stored in request context for the active account.
type SessionModel interface {
	GetID() uint64
}

var ctxAccountKey = &struct{ s string }{"account:account"}

// WithSessionAccount puts to the context account model
func WithSessionAccount(ctx context.Context, accountObj SessionModel) context.Context {
	return context.WithValue(ctx, ctxAccountKey, accountObj)
}

// SessionAccount returns current account model
func SessionAccount(ctx context.Context) SessionModel {
	if acc, ok := ctx.Value(ctxAccountKey).(SessionModel); ok {
		return acc
	}
	return nil
}
