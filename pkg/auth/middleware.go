package auth

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/authutils"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
)

func Middelware(next http.Handler, authorizers ...Authorizer) http.Handler {
	wr := NewAuthorizeWrapper(authorizers...)
	accounts := accountRepository.New()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token, user, acc, err := wr.Authorize(w, r)
		if err != nil {
			ctxlogger.Get(ctx).Error("authorize", zap.Error(err))
			unauthorized(w)
			return
		}
		if user == nil && acc == nil {
			ctx = session.WithAnonymousUserAccount(ctx)
		} else {
			user, acc, err = authutils.CrossAccountConnect(ctx, r.Header.Get(session.CrossAuthHeader), user, acc)
			if err != nil {
				ctxlogger.Get(ctx).Error("cross account connect", zap.Error(err))
				unauthorized(w)
				return
			}

			// Check if user is a member of the account and load permissions
			if acc != nil {
				if user != nil && !accounts.IsMember(ctx, user.ID, acc.ID) {
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

		next.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}
