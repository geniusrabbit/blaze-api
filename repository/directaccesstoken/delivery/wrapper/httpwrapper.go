package wrapper

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account/auth"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	directAccRepository "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
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
func HTTPWrapper(h http.Handler, sources ...TokenSource) http.Handler {
	// Default to header-based token if no sources specified
	if len(sources) == 0 {
		sources = append(sources, TokenSource{Type: "header", Name: "D-Access-Token"})
	}

	actokens := directAccRepository.NewDirectAccessTokenRepository()
	accounts := accountRepository.NewAccountRepository()
	members := accountRepository.NewMemberRepository()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		ctx := r.Context()

		// Extract token from request
		for _, source := range sources {
			if token = source.Extract(r); token != "" {
				break
			}
		}

		// Validate token exists
		if token == "" {
			badRequest(w)
			return
		}

		// Load token object from repository
		tokenObj, err := actokens.GetByToken(ctx, token)
		if err != nil {
			ctxlogger.Get(ctx).Error("invalid token load", zap.Error(err))
			unauthorized(w)
			return
		}

		// Load associated user and account
		user, acc, err := auth.UserAccountByID(ctx, tokenObj.UserID.V, tokenObj.AccountID, nil, nil)
		if err != nil {
			ctxlogger.Get(ctx).Error("invalid user load", zap.Error(err))
			unauthorized(w)
			return
		}

		// Validate account found
		if acc == nil {
			ctxlogger.Get(ctx).Info("user and account not found")
			unauthorized(w)
			return
		}

		// Verify user is member of account
		if user != nil && !members.IsMember(ctx, user.ID, acc.ID) {
			ctxlogger.Get(ctx).Error("user is not a member of the account")
			unauthorized(w)
			return
		}

		// Load user permissions for the account
		err = accounts.LoadPermissions(ctx, acc, user)
		if err != nil {
			ctxlogger.Get(ctx).Error("load permissions", zap.Error(err))
			unauthorized(w)
			return
		}

		// Inject user and account into context
		ctx = session.WithUserAccount(ctx, user, acc)
		h.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

// unauthorized responds with a 401 Unauthorized error.
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}

// badRequest responds with a 400 Bad Request error.
func badRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Bad request","code":400}]}`))
}
