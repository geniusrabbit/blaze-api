package authorizer

import (
	"net/http"

	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// DevTokenAuthorizer handles development token authentication.
type DevTokenAuthorizer struct {
	options   AuthOption
	extractor TokenExtractor
}

// NewDevTokenAuthorizer creates a new DevTokenAuthorizer with the given options.
func NewDevTokenAuthorizer(opts *AuthOption) *DevTokenAuthorizer {
	return &DevTokenAuthorizer{
		options: gocast.IfThenExec(opts != nil,
			func() AuthOption { return *opts },
			func() AuthOption { return AuthOption{} }),
		extractor: tokenextractor.DefaultExtractor,
	}
}

// AuthorizerCode returns the authorizer identifier.
func (au *DevTokenAuthorizer) AuthorizerCode() string {
	return "devtoken"
}

// Authorize validates a development token from the request and returns the associated user and account.
func (au *DevTokenAuthorizer) Authorize(w http.ResponseWriter, r *http.Request) (string, *userModels.User, *models.Account, error) {
	// Return early if dev token is not configured
	if au.options.DevToken == "" {
		return "", nil, nil, nil
	}

	ctx := r.Context()

	// Extract token from request
	token, err := au.extractor(r)
	if err != nil {
		ctxlogger.Get(ctx).Error("token extraction failed", zap.Error(err))
		return "", nil, nil, nil
	}

	// Return early if no token provided
	if token == "" {
		return "", nil, nil, nil
	}

	// Validate token matches configured dev token
	if au.options.DevToken == token {
		usr, acc, err := auth.UserAccountByID(ctx, au.options.DevUserID, au.options.DevAccountID, nil, nil)
		if err != nil {
			ctxlogger.Get(ctx).Error("failed to fetch user and account", zap.Error(err))
			return "", nil, nil, nil
		}
		return token, usr, acc, nil
	}

	return "", nil, nil, nil
}
