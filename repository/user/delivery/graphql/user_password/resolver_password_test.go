package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

func TestChangeUserPasswordDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := &QueryResolverPassword[*testutil.User, any, any, any, any, any]{}
	ctx := session.WithAuthorizerCode(context.Background(), session.AuthorizerCodeDirectToken)

	_, err := resolver.ChangeUserPassword(ctx, "current", "new-password")
	if !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
	}
}
