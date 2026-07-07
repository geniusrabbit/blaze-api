package oauth2

import (
	"errors"
	"net/http"

	"github.com/ory/fosite"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/oauth2/serverprovider"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	errAccessTokensOnlyAllows = errors.New("only access tokens are allowed in the authorization header")
)

type Authorizer[TUser user.Model, TAccount account.Model] struct {
	provider fosite.OAuth2Provider
	accounts account.SessionRepository[TUser, TAccount]
}

func NewAuthorizer[TUser user.Model, TAccount account.Model](
	provider fosite.OAuth2Provider,
	accounts account.SessionRepository[TUser, TAccount],
) *Authorizer[TUser, TAccount] {
	return &Authorizer[TUser, TAccount]{
		provider: provider,
		accounts: accounts,
	}
}

func (au *Authorizer[TUser, TAccount]) AuthorizerCode() string {
	return "oauth2"
}

func (au *Authorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (string, TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	ctx := r.Context()

	token, err := tokenextractor.DefaultExtractor(r)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", zeroUser, zeroAcc, nil
	}

	if token == "" {
		return "", zeroUser, zeroAcc, nil
	}

	oauth2Ctx := serverprovider.NewContext(ctx)
	tokenType, accessReq, errToken := au.provider.IntrospectToken(
		oauth2Ctx, token, fosite.AccessToken, &fosite.DefaultSession{})
	if errToken != nil {
		ctxlogger.Get(r.Context()).Debug("token introspection", zap.Error(errToken))
		return "", zeroUser, zeroAcc, nil
	}

	if tokenType != fosite.AccessToken {
		return "", zeroUser, zeroAcc, errAccessTokensOnlyAllows
	}

	session := accessReq.GetSession().(*serverprovider.Session)
	userObj, accountObj, err := au.accounts.GetByToken(ctx, session.AccessToken)

	return token, userObj, accountObj, err
}
