package directtoken

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	userRepository "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

type Authorizer struct{}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

func (au *Authorizer) AuthorizerCode() string {
	return "directtoken"
}

func (au *Authorizer) Authorize(w http.ResponseWriter, r *http.Request) (string, *model.User, *model.Account, error) {
	ctx := r.Context()
	token, err := tokenextractor.DefaultExtractor(r)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", nil, nil, nil
	}
	if token == "" {
		return "", nil, nil, nil
	}
	userObj, accountObj, err := userRepository.New().GetByToken(ctx, token)
	return token, userObj, accountObj, err
}
