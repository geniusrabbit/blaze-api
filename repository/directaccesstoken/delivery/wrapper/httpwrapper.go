package wrapper

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/models"
	directAccRepository "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// TokenSource defines where and how to extract a token from an HTTP request.
type TokenSource struct {
	Type string `json:"type"` // Type of token source: "query" or "header"
	Name string `json:"name"` // Name of query parameter or header field
}

// Extract retrieves the token value from the request based on the TokenSource configuration.
func (ts TokenSource) Extract(r *http.Request) string {
	switch ts.Type {
	case "query":
		return r.URL.Query().Get(ts.Name)
	case "header":
		return r.Header.Get(ts.Name)
	}
	return ""
}

type tokenGetter interface {
	GetByToken(ctx context.Context, token string) (*models.DirectAccessToken, error)
}

// HTTPWrapper is middleware that validates direct access tokens and injects user/account context.
//
// The session is marked with the direct token authorizer code, so the
// API key restrictions (session.DenyDirectTokenSession) apply as with the
// auth middleware. Prefer header sources: tokens in a query string end up in
// access and proxy logs.
func HTTPWrapper[TUser user.Model, TAccount account.Model](
	h http.Handler,
	loader *accauth.Loader[TUser, TAccount],
	sources ...TokenSource,
) http.Handler {
	return httpWrapper(h, loader, directAccRepository.NewDirectAccessTokenRepository(), sources...)
}

func httpWrapper[TUser user.Model, TAccount account.Model](
	h http.Handler,
	loader *accauth.Loader[TUser, TAccount],
	actokens tokenGetter,
	sources ...TokenSource,
) http.Handler {
	if len(sources) == 0 {
		sources = append(sources, TokenSource{Type: "header", Name: "D-Access-Token"})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		ctx := r.Context()

		for _, source := range sources {
			if token = source.Extract(r); token != "" {
				break
			}
		}

		if token == "" {
			badRequest(w)
			return
		}

		tokenObj, err := actokens.GetByToken(ctx, token)
		if err != nil {
			ctxlogger.Get(ctx).Error("invalid token load", zap.Error(err))
			unavailable(w)
			return
		}
		if tokenObj == nil || tokenObj.ID == 0 {
			unauthorized(w)
			return
		}

		var userID uint64
		if tokenObj.UserID.Valid {
			userID = tokenObj.UserID.V
		}

		var zeroUser TUser
		var zeroAcc TAccount
		userObj, acc, err := loader.UserAccountByID(ctx, userID, tokenObj.AccountID, zeroUser, zeroAcc)
		if err != nil {
			ctxlogger.Get(ctx).Error("invalid user load", zap.Error(err))
			if errors.Is(err, auth.ErrAuthorizationUnavailable) {
				unavailable(w)
			} else {
				unauthorized(w)
			}
			return
		}

		if any(acc) == any(zeroAcc) {
			ctxlogger.Get(ctx).Info("user and account not found")
			unauthorized(w)
			return
		}

		if any(userObj) != any(zeroUser) {
			if user.SessionRejected(userObj) {
				unauthorized(w)
				return
			}
			isMember, err := loader.CheckMembership(ctx, userObj.GetID(), acc.GetID())
			if err != nil {
				ctxlogger.Get(ctx).Error("membership check", zap.Error(err))
				unavailable(w)
				return
			}
			if !isMember {
				ctxlogger.Get(ctx).Error("user is not a member of the account")
				unauthorized(w)
				return
			}
		}

		if err = loader.Accounts.LoadPermissions(ctx, acc, userObj); err != nil {
			ctxlogger.Get(ctx).Error("load permissions", zap.Error(err))
			unavailable(w)
			return
		}

		ctx = session.WithUserAccount(ctx, userObj, acc)
		ctx = session.WithAuthorizerCode(ctx, session.AuthorizerCodeDirectToken)
		h.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

// unavailable responds with a 503 Service Unavailable error.
func unavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Service Unavailable","code":503}]}`))
}

// badRequest responds with a 400 Bad Request error.
func badRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Bad Request","code":400}]}`))
}

// unauthorized responds with a 401 Unauthorized error.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}
