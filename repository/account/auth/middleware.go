package auth

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// Authorizer is a type alias for the generic Authorizer with specific user and account types.
type Authorizer = auth.Authorizer[*userModels.User, *models.Account]

// Middleware returns an HTTP handler that performs authorization checks on incoming requests.
// It validates tokens, handles cross-account connections, and loads user permissions.
func Middleware(next http.Handler, authorizers ...Authorizer) http.Handler {
	authWrap := auth.NewAuthorizeWrapper(authorizers...)
	accounts := accountRepository.NewAccountRepository()
	members := accountRepository.NewMemberRepository()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Authorize the request and extract token, user, and account information
		token, user, acc, err := authWrap.Authorize(w, r)
		if err != nil {
			ctxlogger.Get(ctx).Error("authorize", zap.Error(err))
			unauthorized(w)
			return
		}

		// Handle anonymous user if neither user nor account is present
		if user == nil && acc == nil {
			ctx = session.WithAnonymousUserAccount(ctx)
		} else {
			// Handle cross-account connection
			user, acc, err = CrossAccountConnect(ctx, r.Header.Get(session.CrossAuthHeader), user, acc)
			if err != nil {
				ctxlogger.Get(ctx).Error("cross account connect", zap.Error(err))
				unauthorized(w)
				return
			}

			// Validate account membership and load permissions
			if acc != nil {
				if user != nil && !members.IsMember(ctx, user.ID, acc.ID) {
					ctxlogger.Get(ctx).Error("user is not a member of the account")
					unauthorized(w)
					return
				}

				err = accounts.LoadPermissions(ctx, acc, user)
				if err != nil {
					ctxlogger.Get(ctx).Error("load permissions", zap.Error(err))
					unauthorized(w)
					return
				}
			}

			ctx = session.WithUserAccount(ctx, user, acc)
		}

		// Pass the request to the next handler with updated context
		next.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

// unauthorized writes an HTTP 401 Unauthorized response with a JSON error message.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}
