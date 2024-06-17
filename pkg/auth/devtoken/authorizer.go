package devtoken

import (
	"net/http"

	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/auth/authutils"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
)

type Authorizer struct {
	options AuthOption
}

func NewAuthorizer(opts *AuthOption) *Authorizer {
	return &Authorizer{
		options: gocast.IfThenExec(opts != nil,
			func() AuthOption { return *opts },
			func() AuthOption { return AuthOption{} }),
	}
}

func (au *Authorizer) AuthorizerCode() string {
	return "devtoken"
}

func (au *Authorizer) Authorize(w http.ResponseWriter, r *http.Request) (string, *model.User, *model.Account, error) {
	if au.options.DevToken == "" {
		return "", nil, nil, nil
	}
	ctx := r.Context()
	token, err := tokenextractor.DefaultExtractor(r)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", nil, nil, nil
	}
	if token == "" {
		return "", nil, nil, nil
	}
	if au.options.DevToken == token {
		usr, acc, err := authutils.UserAccountByID(ctx, au.options.DevUserID, au.options.DevAccountID, nil, nil)
		if err != nil {
			ctxlogger.Get(r.Context()).Error("get user acc", zap.Error(err))
			return "", nil, nil, nil
		}
		return token, usr, acc, nil
	}
	return "", nil, nil, nil
}
