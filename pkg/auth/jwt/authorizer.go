package jwt

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Authorizer handles JWT-based authorization for API requests.
type Authorizer[TUser user.Model, TAccount account.Model] struct {
	provider *Provider
	jmid     *jwtmiddleware.JWTMiddleware
	loader   *accauth.Loader[TUser, TAccount]
}

// NewAuthorizer creates a new JWT authorizer instance.
func NewAuthorizer[TUser user.Model, TAccount account.Model](jwtProvider *Provider, loader *accauth.Loader[TUser, TAccount]) *Authorizer[TUser, TAccount] {
	return &Authorizer[TUser, TAccount]{
		provider: jwtProvider,
		jmid:     jwtProvider.Middleware(),
		loader:   loader,
	}
}

// AuthorizerCode returns the identifier for this authorizer.
func (au *Authorizer[TUser, TAccount]) AuthorizerCode() string {
	return "jwt"
}

// Authorize validates the JWT token from the request and retrieves associated user and account data.
func (au *Authorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (token string, usr TUser, acc TAccount, err error) {
	var zeroUser TUser
	var zeroAcc TAccount
	if err = au.jmid.CheckJWT(w, r); err != nil {
		ctxlogger.Get(r.Context()).Debug("JWT authorization", zap.Error(err))
		return "", zeroUser, zeroAcc, nil
	}

	ctx := r.Context()
	jwtToken := ctx.Value(au.jmid.Options.UserProperty)

	switch t := jwtToken.(type) {
	case nil:
	case *Token:
		token = t.Raw
		usr, acc, err = au.authContextJWT(ctx, t)
	}

	return token, usr, acc, err
}

func (au *Authorizer[TUser, TAccount]) authContextJWT(ctx context.Context, token *Token) (TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	jwtData, err := au.provider.ExtractTokenData(token)
	if err != nil {
		return zeroUser, zeroAcc, err
	}
	return au.loader.UserAccountByID(ctx, jwtData.UserID, jwtData.AccountID, zeroUser, zeroAcc)
}
