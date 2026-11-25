package graphql

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	datModels "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

func TestTokenMutationsDeniedForDirectTokenSession(t *testing.T) {
	t.Parallel()

	resolver := NewQueryResolver(nil)
	ctx := session.WithAuthorizerCode(context.Background(), session.AuthorizerCodeDirectToken)

	if _, err := resolver.Generate(ctx, nil, "", nil); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("generate err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
	}
	if _, err := resolver.Revoke(ctx, models.DirectAccessTokenListFilter{}); !errors.Is(err, session.ErrDirectTokenSessionForbidden) {
		t.Fatalf("revoke err = %v, want %v", err, session.ErrDirectTokenSessionForbidden)
	}
}

type stubTokenUsecase struct {
	token *datModels.DirectAccessToken
}

func (stub stubTokenUsecase) Get(context.Context, uint64) (*datModels.DirectAccessToken, error) {
	copied := *stub.token
	return &copied, nil
}
func (stub stubTokenUsecase) FetchList(context.Context, ...directaccesstoken.QOption) ([]*datModels.DirectAccessToken, error) {
	copied := *stub.token
	return []*datModels.DirectAccessToken{&copied}, nil
}
func (stubTokenUsecase) Count(context.Context, ...directaccesstoken.QOption) (int64, error) {
	return 1, nil
}
func (stub stubTokenUsecase) Generate(context.Context, uint64, uint64, string, time.Time) (*datModels.DirectAccessToken, error) {
	copied := *stub.token
	return &copied, nil
}
func (stubTokenUsecase) Revoke(context.Context, ...directaccesstoken.QOption) error { return nil }

const secretToken = "secret-token-value"

func freshTokenUsecase() stubTokenUsecase {
	return stubTokenUsecase{token: &datModels.DirectAccessToken{ID: 1, Token: secretToken, AccountID: 9, CreatedAt: time.Now()}}
}

func TestTokenValueIsReturnedOnlyByGenerate(t *testing.T) {
	t.Parallel()

	resolver := NewQueryResolver(freshTokenUsecase())
	ctx := context.Background()

	generated, err := resolver.Generate(ctx, nil, "", nil)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if generated.Token.Token != secretToken {
		t.Fatalf("generate token = %q, want the value", generated.Token.Token)
	}

	fetched, err := resolver.Get(ctx, 1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if strings.Contains(fetched.Token.Token, secretToken) || fetched.Token.Token == "" {
		t.Fatalf("get token = %q, want a masked value", fetched.Token.Token)
	}

	connection, err := resolver.List(ctx, nil, nil, nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, token := range connection.List() {
		if strings.Contains(token.Token, secretToken) {
			t.Fatalf("list token = %q, want a masked value", token.Token)
		}
	}
}
