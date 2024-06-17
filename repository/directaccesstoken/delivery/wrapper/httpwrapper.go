package wrapper

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/pkg/auth/authutils"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	directAccRepository "github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
)

type TokenSource struct {
	Type string `json:"type"` // Type of the token source: `query`, `header`
	Name string `json:"name"` // Name of the token source
}

func (ts TokenSource) Extract(r *http.Request) string {
	switch ts.Type {
	case "query":
		return r.URL.Query().Get(ts.Name)
	case "header":
		return r.Header.Get(ts.Name)
	}
	return ""
}

func HTTPWrapper(h http.Handler, sources ...TokenSource) http.Handler {
	if len(sources) == 0 {
		sources = append(sources, TokenSource{Type: "header", Name: "D-Access-Token"})
	}
	actokens := directAccRepository.New()
	accounts := accountRepository.New()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		ctx := r.Context()

		// Extract token
		for _, source := range sources {
			if token = source.Extract(r); token != "" {
				break
			}
		}

		// Check token
		if token == "" {
			badRequest(w)
			return
		}

		// Load direct token object
		tokenObj, err := actokens.GetByToken(ctx, token)
		if err != nil {
			ctxlogger.Get(ctx).Error(`Invalid token load`, zap.Error(err))
			unauthorized(w)
			return
		}

		// Load user and account
		user, acc, err := authutils.UserAccountByID(ctx, tokenObj.UserID.V, tokenObj.AccountID, nil, nil)
		if err != nil {
			ctxlogger.Get(ctx).Error(`Invalid user load`, zap.Error(err))
			unauthorized(w)
			return
		}

		// Check if user and account not found
		if acc == nil {
			ctxlogger.Get(ctx).Info(`User and account not found`)
			unauthorized(w)
			return
		}

		// Check if user is a member of the account and load permissions
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

		ctx = session.WithUserAccount(ctx, user, acc)
		h.ServeHTTP(w, r.WithContext(session.WithToken(ctx, token)))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
}

func badRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"errors":[{"message":"Bad request","code":400}]}`))
}
