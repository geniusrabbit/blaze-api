package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

func TestAuthClientMutationsDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	clients := NewQueryResolver(nil)
	sessions := NewSessionQueryResolver(nil)
	ctx := session.WithAuthorizerCode(context.Background(), session.AuthorizerCodeDirectToken)
	mutations := map[string]func() error{
		"createAuthClient": func() error {
			_, err := clients.CreateAuthClient(ctx, &gqlmodels.AuthClientCreateInput{})
			return err
		},
		"updateAuthClient": func() error {
			_, err := clients.UpdateAuthClient(ctx, "client", &gqlmodels.AuthClientUpdateInput{})
			return err
		},
		"deleteAuthClient": func() error {
			_, err := clients.DeleteAuthClient(ctx, "client", nil)
			return err
		},
		"deleteAuthSession": func() error {
			_, err := sessions.DeleteAuthSession(ctx, 1)
			return err
		},
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := mutation(); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
				t.Fatalf("err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
			}
		})
	}
}
