package jwt

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Authorizer handles JWT-based authorization for API requests.
type Authorizer[TUser user.Model, TAccount account.Model] struct {
	provider *Provider
	loader   *accauth.Loader[TUser, TAccount]
}

// NewAuthorizer creates a new JWT authorizer instance.
func NewAuthorizer[TUser user.Model, TAccount account.Model](jwtProvider *Provider, loader *accauth.Loader[TUser, TAccount]) *Authorizer[TUser, TAccount] {
	return &Authorizer[TUser, TAccount]{
		provider: jwtProvider,
		loader:   loader,
	}
}

// AuthorizerCode is the code of the JWT authorizer in the session context.
const AuthorizerCode = "jwt"

// AuthorizerCode returns the identifier for this authorizer.
func (au *Authorizer[TUser, TAccount]) AuthorizerCode() string {
	return AuthorizerCode
}

// Authorize validates the JWT token from the request and retrieves associated user and account data.
// A missing or invalid token yields no session and no error, so the next
// authorizer can try the same credential (e.g. an OAuth2 access token).
func (au *Authorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (string, TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	ctx := r.Context()

	raw, err := au.provider.ExtractToken(r)
	if err != nil || raw == "" {
		return "", zeroUser, zeroAcc, nil
	}

	data, err := au.provider.ParseToken(ctx, raw)
	if err != nil {
		if errors.Is(err, auth.ErrAuthorizationUnavailable) {
			return "", zeroUser, zeroAcc, err
		}
		ctxlogger.Get(ctx).Debug("JWT authorization", zap.Error(err))
		return "", zeroUser, zeroAcc, nil
	}

	usr, acc, err := au.authContextJWT(ctx, data)
	return raw, usr, acc, err
}

func (au *Authorizer[TUser, TAccount]) authContextJWT(ctx context.Context, data *TokenData) (TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	return au.loader.UserAccountByID(ctx, data.UserID, data.AccountID, zeroUser, zeroAcc)
}
