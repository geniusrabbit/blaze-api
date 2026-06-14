package context

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/account/models"
)

var ctxAccountKey = &struct{ s string }{"account:account"}

// WithSessionAccount puts to the context account model
func WithSessionAccount(ctx context.Context, accountObj *models.Account) context.Context {
	return context.WithValue(ctx, ctxAccountKey, accountObj)
}

// SessionAccount returns current account model
// nolint:unused // temporary
func SessionAccount(ctx context.Context) *models.Account {
	return ctx.Value(ctxAccountKey).(*models.Account)
}
