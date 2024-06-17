package jwt

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/auth/authutils"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

type Authorizer struct {
	provider *Provider
	jmid     *jwtmiddleware.JWTMiddleware
}

func NewAuthorizer(jwtProvider *Provider) *Authorizer {
	return &Authorizer{
		provider: jwtProvider,
		jmid:     jwtProvider.Middleware(),
	}
}

func (au *Authorizer) AuthorizerCode() string {
	return "jwt"
}

func (au *Authorizer) Authorize(w http.ResponseWriter, r *http.Request) (token string, usr *model.User, acc *model.Account, err error) {
	if err = au.jmid.CheckJWT(w, r); err != nil {
		ctxlogger.Get(r.Context()).Debug("JWT authorization", zap.Error(err))
		return "", nil, nil, nil
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

func (au *Authorizer) authContextJWT(ctx context.Context, token *Token) (*model.User, *model.Account, error) {
	jwtData, err := au.provider.ExtractTokenData(token)
	if err != nil {
		return nil, nil, err
	}
	return authutils.UserAccountByID(ctx, jwtData.UserID, jwtData.AccountID, nil, nil)
}
