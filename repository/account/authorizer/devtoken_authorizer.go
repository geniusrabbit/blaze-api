package authorizer

import (
	"net/http"

	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// DevTokenAuthorizer handles development token authentication.
type DevTokenAuthorizer[TUser user.Model, TAccount account.Model] struct {
	options   AuthOption
	extractor TokenExtractor
	loader    *accauth.Loader[TUser, TAccount]
}

// NewDevTokenAuthorizer creates a new DevTokenAuthorizer with the given options.
func NewDevTokenAuthorizer[TUser user.Model, TAccount account.Model](
	opts *AuthOption,
	loader *accauth.Loader[TUser, TAccount],
) *DevTokenAuthorizer[TUser, TAccount] {
	return &DevTokenAuthorizer[TUser, TAccount]{
		options: gocast.IfThenExec(opts != nil,
			func() AuthOption { return *opts },
			func() AuthOption { return AuthOption{} }),
		extractor: tokenextractor.DefaultExtractor,
		loader:    loader,
	}
}

// AuthorizerCode returns the authorizer identifier.
func (au *DevTokenAuthorizer[TUser, TAccount]) AuthorizerCode() string {
	return "devtoken"
}

// Authorize validates a development token from the request and returns the associated user and account.
func (au *DevTokenAuthorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (string, TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	if au.options.DevToken == "" {
		return "", zeroUser, zeroAcc, nil
	}

	ctx := r.Context()
	token, err := au.extractor(r)
	if err != nil {
		ctxlogger.Get(ctx).Error("token extraction failed", zap.Error(err))
		return "", zeroUser, zeroAcc, nil
	}
	if token == "" {
		return "", zeroUser, zeroAcc, nil
	}

	if au.options.DevToken == token {
		usr, acc, err := au.loader.UserAccountByID(ctx, au.options.DevUserID, au.options.DevAccountID, zeroUser, zeroAcc)
		if err != nil {
			ctxlogger.Get(ctx).Error("failed to fetch user and account", zap.Error(err))
			return "", zeroUser, zeroAcc, nil
		}
		return token, usr, acc, nil
	}

	return "", zeroUser, zeroAcc, nil
}
