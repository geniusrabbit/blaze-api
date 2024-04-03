package socialauth

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
)

// Usecase of the socialauth account which provides bussiness logic for socialauth access
// and connection to the user account by social network
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.AccountSocial, error)
	List(ctx context.Context, filter *Filter) ([]*model.AccountSocial, error)
	Register(ctx context.Context, user *model.User, account *model.AccountSocial) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.AccountSocial) error
	Token(ctx context.Context, name string, accountSocialID uint64) (*elogin.Token, error)
	SetToken(ctx context.Context, name string, accountSocialID uint64, token *elogin.Token) error
}
