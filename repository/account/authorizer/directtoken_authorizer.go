package authorizer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// DirectAccessTokenGetter looks up a non-expired direct access token by its value.
type DirectAccessTokenGetter interface {
	GetByToken(ctx context.Context, token string) (*directaccesstoken.DirectAccessToken, error)
}

// DirectTokenAuthorizer implements authorization using direct access tokens (API keys).
type DirectTokenAuthorizer[TUser user.Model, TAccount account.Model] struct {
	extractor TokenExtractor
	loader    *accauth.Loader[TUser, TAccount]
	tokens    DirectAccessTokenGetter
}

// NewDirectTokenAuthorizer creates a new instance of DirectTokenAuthorizer.
func NewDirectTokenAuthorizer[TUser user.Model, TAccount account.Model](
	loader *accauth.Loader[TUser, TAccount],
	tokens DirectAccessTokenGetter,
) *DirectTokenAuthorizer[TUser, TAccount] {
	return &DirectTokenAuthorizer[TUser, TAccount]{
		extractor: tokenextractor.AuthorizationHeaderExtractor,
		loader:    loader,
		tokens:    tokens,
	}
}

// AuthorizerCode returns the identifier code for this authorizer.
func (au *DirectTokenAuthorizer[TUser, TAccount]) AuthorizerCode() string {
	return session.AuthorizerCodeDirectToken
}

// Authorize looks the `Authorization: Bearer` token up in the direct access tokens
// and loads the user and account the token belongs to.
// A token that is not a direct access token yields an empty result without an error.
func (au *DirectTokenAuthorizer[TUser, TAccount]) Authorize(w http.ResponseWriter, r *http.Request) (string, TUser, TAccount, error) {
	var zeroUser TUser
	var zeroAccount TAccount
	ctx := r.Context()

	token, err := au.extractor(r)
	if err != nil {
		ctxlogger.Get(ctx).Error("token extraction", zap.Error(err))
		return "", zeroUser, zeroAccount, nil
	}
	if token == "" {
		return "", zeroUser, zeroAccount, nil
	}

	tokenObject, err := au.tokens.GetByToken(ctx, token)
	if err != nil {
		return "", zeroUser, zeroAccount, fmt.Errorf("%w: direct access token lookup: %w", auth.ErrAuthorizationUnavailable, err)
	}
	if tokenObject == nil || tokenObject.ID == 0 {
		return "", zeroUser, zeroAccount, nil
	}

	var userID uint64
	if tokenObject.UserID.Valid {
		userID = tokenObject.UserID.V
	}

	userObject, accountObject, err := au.loader.UserAccountByID(ctx, userID, tokenObject.AccountID, zeroUser, zeroAccount)
	if err != nil {
		if isRejectedCredential(err) {
			return "", zeroUser, zeroAccount, err
		}
		return "", zeroUser, zeroAccount, fmt.Errorf("%w: direct access token owner: %w", auth.ErrAuthorizationUnavailable, err)
	}
	return token, userObject, accountObject, nil
}

func isRejectedCredential(err error) bool {
	return errors.Is(err, accauth.ErrUserIsNotMemberOfAccount) ||
		errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, sql.ErrNoRows)
}
