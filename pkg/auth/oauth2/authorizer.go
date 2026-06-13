package oauth2

import (
	"errors"
	"net/http"

	"github.com/ory/fosite"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/oauth2/serverprovider"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	accountRepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

var (
	errAccessTokensOnlyAllows = errors.New("only access tokens are allowed in the authorization header")
)

type Authorizer struct {
	provider fosite.OAuth2Provider
}

func NewAuthorizer(provider fosite.OAuth2Provider) *Authorizer {
	return &Authorizer{
		provider: provider,
	}
}

func (au *Authorizer) AuthorizerCode() string {
	return "oauth2"
}

func (au *Authorizer) Authorize(w http.ResponseWriter, r *http.Request) (string, *userModels.User, *accountModels.Account, error) {
	var (
		userObj    *userModels.User
		accountObj *accountModels.Account
		ctx        = r.Context()
		token, err = tokenextractor.DefaultExtractor(r)
		accounts   = accountRepo.NewAccountRepository()
	)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", nil, nil, nil
	}
	if token == "" {
		return "", nil, nil, nil
	}

	oauth2Ctx := serverprovider.NewContext(ctx)
	tokenType, accessReq, errToken := au.provider.IntrospectToken(
		oauth2Ctx, token, fosite.AccessToken, &fosite.DefaultSession{})
	if errToken != nil {
		ctxlogger.Get(r.Context()).Debug("token introspection", zap.Error(errToken))
		return "", nil, nil, nil
	}
	if tokenType != fosite.AccessToken {
		return "", nil, nil, errAccessTokensOnlyAllows
	}
	session := accessReq.GetSession().(*serverprovider.Session)
	userObj, accountObj, err = accounts.GetByToken(ctx, session.AccessToken)

	return token, userObj, accountObj, err
}
