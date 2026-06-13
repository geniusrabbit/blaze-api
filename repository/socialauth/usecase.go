package socialauth

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	socialAccountModels "github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// Usecase of the socialauth account which provides bussiness logic for socialauth access
// and connection to the user account by social network
type Usecase interface {
	Get(ctx context.Context, id uint64) (*socialAccountModels.AccountSocial, error)
	List(ctx context.Context, filter *Filter) ([]*socialAccountModels.AccountSocial, error)
	Register(ctx context.Context, user *userModels.User, account *socialAccountModels.AccountSocial) (uint64, error)
	Update(ctx context.Context, id uint64, account *socialAccountModels.AccountSocial) error
	Token(ctx context.Context, name string, accountSocialID uint64) (*elogin.Token, error)
	SetToken(ctx context.Context, name string, accountSocialID uint64, token *elogin.Token) error
}
