package context

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user/models"
)

var ctxUserKey = &struct{ s string }{"account:user"}

// WithSessionUser puts to the context user model
func WithSessionUser(ctx context.Context, userObj *models.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, userObj)
}

// SessionUser returns current user model
// nolint:unused // temporary
func SessionUser(ctx context.Context) *models.User {
	return ctx.Value(ctxUserKey).(*models.User)
}
