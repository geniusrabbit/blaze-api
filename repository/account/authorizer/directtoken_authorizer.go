package authorizer

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// DirectTokenAuthorizer implements authorization using direct token authentication.
type DirectTokenAuthorizer[TUser user.Model, TAccount account.Model] struct {
	extractor TokenExtractor
	loader    *accauth.Loader[TUser, TAccount]
}

// NewDirectTokenAuthorizer creates a new instance of DirectTokenAuthorizer.
func NewDirectTokenAuthorizer[TUser user.Model, TAccount account.Model](loader *accauth.Loader[TUser, TAccount]) *DirectTokenAuthorizer[TUser, TAccount] {
	return &DirectTokenAuthorizer[TUser, TAccount]{
		extractor: tokenextractor.DefaultExtractor,
		loader:    loader,
	}
}

// AuthorizerCode returns the identifier code for this authorizer.
func (au *DirectTokenAuthorizer[TUser, TAccount]) AuthorizerCode() string {
	return "directtoken"
}

// Authorize validates the request by extracting and verifying the token,
// then retrieves the associated user and account information.
func (au *DirectTokenAuthorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (string, TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAcc TAccount
	ctx := r.Context()

	token, err := au.extractor(r)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", zeroUser, zeroAcc, nil
	}
	if token == "" {
		return "", zeroUser, zeroAcc, nil
	}

	userObj, accountObj, err := au.loader.Accounts.GetByToken(ctx, token)
	return token, userObj, accountObj, err
}
