package graphql

import (
	"context"
	"errors"
	"time"

	"github.com/geniusrabbit/api-template-base/internal/jwt"
	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	accountrepo "github.com/geniusrabbit/api-template-base/internal/repository/account/repository"
	userrepo "github.com/geniusrabbit/api-template-base/internal/repository/user/repository"
	"github.com/geniusrabbit/api-template-base/model"
)

var (
	errInvalidAccountTarget = errors.New(`invalid account target`)
)

// AuthResolver is the resolver for the Auth type.
type AuthResolver struct {
	provider *jwt.Provider
}

// NewAuthResolver creates new resolver for the Auth type.
func NewAuthResolver(provider *jwt.Provider) *AuthResolver {
	return &AuthResolver{provider: provider}
}

// Login is the resolver for the login field.
func (r *AuthResolver) Login(ctx context.Context, login string, password string) (string, time.Duration, error) {
	accountID := uint64(0)
	usersRepo := userrepo.New()
	accountRepo := accountrepo.New()

	user, err := usersRepo.GetByPassword(ctx, login, password)
	if err != nil {
		return "", 0, err
	}

	account, err := accountForUser(ctx, accountRepo, user.ID, accountID)
	if err != nil {
		return "", 0, err
	}
	if account != nil {
		accountID = account.ID
	}

	token, err := r.provider.CreateToken(user.ID, accountID)
	if err != nil {
		return "", 0, err
	}

	return token, r.provider.TokenLifetime, nil
}

// Logout is the resolver for the logout field.
func (r *AuthResolver) Logout(ctx context.Context) (bool, error) {
	return true, nil
}

func accountForUser(ctx context.Context, accountRepo account.Repository, userID, accountID uint64) (*model.Account, error) {
	accounts, err := accountRepo.FetchList(ctx, &account.Filter{UserID: []uint64{userID}})
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		if accountID != 0 {
			return nil, errInvalidAccountTarget
		}
		return nil, nil
	}
	return accounts[0], nil
}
