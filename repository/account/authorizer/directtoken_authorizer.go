package authorizer

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	accountRepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// DirectTokenAuthorizer implements authorization using direct token authentication.
type DirectTokenAuthorizer struct {
	extractor TokenExtractor
}

// NewDirectTokenAuthorizer creates a new instance of DirectTokenAuthorizer.
func NewDirectTokenAuthorizer() *DirectTokenAuthorizer {
	return &DirectTokenAuthorizer{
		extractor: tokenextractor.DefaultExtractor,
	}
}

// AuthorizerCode returns the identifier code for this authorizer.
func (au *DirectTokenAuthorizer) AuthorizerCode() string {
	return "directtoken"
}

// Authorize validates the request by extracting and verifying the token,
// then retrieves the associated user and account information.
func (au *DirectTokenAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) (string, *userModels.User, *accountModels.Account, error) {
	ctx := r.Context()

	// Extract token from the request
	token, err := au.extractor(r)
	if err != nil {
		ctxlogger.Get(r.Context()).Error("token extraction", zap.Error(err))
		return "", nil, nil, nil
	}

	// Return early if no token is provided
	if token == "" {
		return "", nil, nil, nil
	}

	// Retrieve user and account information by token
	userObj, accountObj, err := accountRepo.NewAccountRepository().GetByToken(ctx, token)
	return token, userObj, accountObj, err
}
