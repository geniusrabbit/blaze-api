package context

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ctxUserKey = &struct{ s string }{"account:user"}

// WithSessionUser puts the user model into context.
func WithSessionUser(ctx context.Context, userObj user.Model) context.Context {
	return context.WithValue(ctx, ctxUserKey, userObj)
}

// SessionUser returns current user model from context.
func SessionUser(ctx context.Context) user.Model {
	if v := ctx.Value(ctxUserKey); v != nil {
		if u, ok := v.(user.Model); ok {
			return u
		}
	}
	return nil
}
