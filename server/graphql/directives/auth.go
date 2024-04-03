package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
)

// Auth directive checks that user is authenticated
func Auth(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	if session.User(ctx).IsAnonymous() {
		return nil, errAuthorizationRequired
	}
	return next(ctx)
}
