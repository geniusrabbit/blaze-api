package socialauth

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
)

type Repository interface {
	Get(ctx context.Context, id uint64) (*model.AccountSocial, error)
	List(ctx context.Context, filter *Filter) ([]*model.AccountSocial, error)
	Create(ctx context.Context, account *model.AccountSocial) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.AccountSocial) error
	Token(ctx context.Context, name string, accountSocialID uint64) (*elogin.Token, error)
	SetToken(ctx context.Context, name string, accountSocialID uint64, token *elogin.Token) error
}
