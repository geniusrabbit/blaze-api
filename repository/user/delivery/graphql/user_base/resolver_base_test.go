package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

func TestUserMutationsDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := &QueryResolverBase[*testutil.User, any, any, any, any, any, any]{}
	ctx := session.WithAuthorizerCode(context.Background(), session.AuthorizerCodeDirectToken)
	mutations := map[string]func() error{
		"updateUser": func() error {
			_, err := resolver.UpdateUser(ctx, 7, nil)
			return err
		},
		"approveUser": func() error {
			_, err := resolver.ApproveUser(ctx, 7, nil)
			return err
		},
		"rejectUser": func() error {
			_, err := resolver.RejectUser(ctx, 7, nil)
			return err
		},
	}
	for name, mutation := range mutations {
		if err := mutation(); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
			t.Fatalf("%s: err = %v, want %v", name, err, session.ErrDirectTokenSessionForbidden)
		}
	}
}
