package jwt

import (
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/repository/account/auth"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// Authorizer handles JWT-based authorization for API requests.
type Authorizer struct {
	provider *Provider
	jmid     *jwtmiddleware.JWTMiddleware
}

// NewAuthorizer creates a new JWT authorizer instance.
func NewAuthorizer(jwtProvider *Provider) *Authorizer {
	return &Authorizer{
		provider: jwtProvider,
		jmid:     jwtProvider.Middleware(),
	}
}

// AuthorizerCode returns the identifier for this authorizer.
func (au *Authorizer) AuthorizerCode() string {
	return "jwt"
}

// Authorize validates the JWT token from the request and retrieves associated user and account data.
func (au *Authorizer) Authorize(w http.ResponseWriter, r *http.Request) (token string, usr *userModels.User, acc *accountModels.Account, err error) {
	// Validate JWT token.
	// CheckJWT updates *r in-place (via *r = *r.WithContext(newCtx)) so the parsed
	// token is stored in r.Context() AFTER this call — not in a context captured before.
	if err = au.jmid.CheckJWT(w, r); err != nil {
		ctxlogger.Get(r.Context()).Debug("JWT authorization", zap.Error(err))
		return "", nil, nil, nil
	}

	// Read the updated context only AFTER CheckJWT has modified *r.
	ctx := r.Context()

	// Extract token from context
	jwtToken := ctx.Value(au.jmid.Options.UserProperty)

	// Parse token and fetch user/account data
	switch t := jwtToken.(type) {
	case nil:
	case *Token:
		token = t.Raw
		usr, acc, err = au.authContextJWT(ctx, t)
	}

	return token, usr, acc, err
}

// authContextJWT extracts user and account information from the JWT token.
func (au *Authorizer) authContextJWT(ctx context.Context, token *Token) (*userModels.User, *accountModels.Account, error) {
	jwtData, err := au.provider.ExtractTokenData(token)
	if err != nil {
		return nil, nil, err
	}
	return auth.UserAccountByID(ctx, jwtData.UserID, jwtData.AccountID, nil, nil)
}
