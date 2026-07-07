package wrapper

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accauth "github.com/geniusrabbit/blaze-api/repository/account/auth"
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

// HTTPWrapper is middleware that validates direct access tokens and injects user/account context.
func HTTPWrapper[TUser user.Model, TAccount account.Model](
	h http.Handler,
	loader *accauth.Loader[TUser, TAccount],
	sources ...TokenSource,
) http.Handler {
	if len(sources) == 0 {
		sources = append(sources, TokenSource{Type: "header", Name: "D-Access-Token"})
	}

	actokens := directAccRepository.NewDirectAccessTokenRepository()

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
			unauthorized(w)
			return
		}

		var zeroUser TUser
		var zeroAcc TAccount
		userObj, acc, err := loader.UserAccountByID(ctx, tokenObj.UserID.V, tokenObj.AccountID, zeroUser, zeroAcc)
		if err != nil {
			ctxlogger.Get(ctx).Error("invalid user load", zap.Error(err))
			unauthorized(w)
			return
		}

		if any(acc) == any(zeroAcc) {
			ctxlogger.Get(ctx).Info("user and account not found")
			unauthorized(w)
			return
		}

		if any(userObj) != any(zeroUser) && !loader.Members.IsMember(ctx, userObj.GetID(), acc.GetID()) {
			ctxlogger.Get(ctx).Error("user is not a member of the account")
			unauthorized(w)
			return
		}

		if err = loader.Accounts.LoadPermissions(ctx, acc, userObj); err != nil {
			ctxlogger.Get(ctx).Error("load permissions", zap.Error(err))
			unauthorized(w)
			return
		}

		ctx = session.WithUserAccount(ctx, userObj, acc)
		h.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
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
