package auth

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Middleware returns an HTTP handler that performs authorization checks on incoming requests.
func Middleware[TUser user.Model, TAccount account.Model](
	next http.Handler,
	loader *Loader[TUser, TAccount],
	authorizers ...auth.Authorizer[TUser, TAccount],
) http.Handler {
	authWrap := auth.NewAuthorizeWrapper(authorizers...)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, user, acc, err := authWrap.Authorize(w, r)
		if err != nil {
			ctxlogger.Get(ctx).Error("authorize", zap.Error(err))
			unauthorized(w)
			return
		}

		var zeroUser TUser
		var zeroAcc TAccount
		if any(user) == any(zeroUser) && any(acc) == any(zeroAcc) {
			ctx = session.WithAnonymousUserAccount(ctx)
		} else {
			user, acc, err = loader.CrossAccountConnect(ctx, r.Header.Get(session.CrossAuthHeader), user, acc)
			if err != nil {
				ctxlogger.Get(ctx).Error("cross account connect", zap.Error(err))
				unauthorized(w)
				return
			}

			if any(acc) != any(zeroAcc) {
				if any(user) != any(zeroUser) && !loader.Members.IsMember(ctx, user.GetID(), acc.GetID()) {
					ctxlogger.Get(ctx).Error("user is not a member of the account")
					unauthorized(w)
					return
				}

				if err = loader.Accounts.LoadPermissions(ctx, acc, user); err != nil {
					ctxlogger.Get(ctx).Error("load permissions", zap.Error(err))
					unauthorized(w)
					return
				}
			}

			ctx = session.WithUserAccount(ctx, user, acc)
		}

		next.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

// unauthorized writes an HTTP 401 Unauthorized response with a JSON error message.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}
