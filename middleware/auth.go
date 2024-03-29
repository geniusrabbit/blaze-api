// Package middleware provides auth middle procedures
package middleware

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	// grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	// "github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/demdxx/gocast/v2"
	"github.com/ory/fosite"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/geniusrabbit/blaze-api/auth/elogin/utils"
	"github.com/geniusrabbit/blaze-api/auth/jwt"
	"github.com/geniusrabbit/blaze-api/auth/oauth2/serverprovider"
	"github.com/geniusrabbit/blaze-api/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userRepository "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

var testMode = false

// SetTestMode state
// nolint:unused // used in tests
func SetTestMode(test bool) { testMode = test }

var (
	errAccessTokensOnlyAllows       = errors.New("only access tokens are allowed in the authorization header")
	errAuthUserIsNotMemberOfAccount = errors.New("user is not a member of the account")
	errNoCrossAuthPermission        = errors.New("user don't have cross auth permissions")
)

// AuthOption to access to default user
type AuthOption struct {
	DevToken     string
	DevUserID    uint64
	DevAccountID uint64
}

type authWrapper struct {
	authSuccess prometheus.Counter
	authError   prometheus.Counter
}

func newAuthWrapper(prefix string) *authWrapper {
	return &authWrapper{
		authSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: prefix + "auth_success_count",
			Help: "Auth success count",
		}),
		authError: promauto.NewCounter(prometheus.CounterOpts{
			Name: prefix + "auth_error_count",
			Help: "Auth error count",
		}),
	}
}

// AuthHTTP middleware
func AuthHTTP(metricsPrefix string, next http.Handler, oauth2provider fosite.OAuth2Provider, jwtProvider *jwt.Provider, authOpt *AuthOption) http.Handler {
	jwtmiddleware := jwtProvider.Middleware()
	authWr := newAuthWrapper(metricsPrefix)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err          error
			ctx          = r.Context()
			isJWTSession = false
			token        = fosite.AccessTokenFromRequest(r)
			authorized   = false
		)
		// If authroization by social network then all parameters will be passed in the state
		if token == "" && r.URL.Query().Get("state") != "" {
			state := utils.DecodeState(r.URL.Query().Get("state"))
			token = state.Get("access_token")
		}
		// If token is empty then it's anonymous user
		if token == "" {
			ctx = session.WithAnonymousUserAccount(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if authOpt != nil && authOpt.DevToken == token {
			if ctx, err = authWr.authContextAcc(ctx, authOpt.DevUserID, authOpt.DevAccountID); err != nil {
				ctxlogger.Get(r.Context()).Error("JWT authorization", zap.Error(err))
			} else {
				authorized = true
			}
		} else if !testMode && jwtmiddleware.CheckJWT(w, r) == nil {
			jwtToken := r.Context().Value(jwtmiddleware.Options.UserProperty)
			switch t := jwtToken.(type) {
			case nil:
			case *jwt.Token:
				isJWTSession = true
				if ctx, err = authWr.authContextJWT(ctx, jwtProvider, t); err != nil {
					ctxlogger.Get(r.Context()).Error("JWT authorization", zap.Error(err))
				} else {
					authorized = true
				}
			}
		}
		if !isJWTSession && !authorized {
			if ctx, err = authWr.authContext(ctx, oauth2provider, token, r.Header.Get(session.CrossAuthHeader)); err != nil {
				ctxlogger.Get(r.Context()).Error("authorization", zap.Error(err))
			} else {
				authorized = true
			}
		}
		if !authorized {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"errors":[{"message":"Unauthorized","code":401}]}`))
			return
		}
		next.ServeHTTP(w, r.WithContext(session.WithToken(
			ctx, gocast.IfThen(authorized, token, ""),
		)))
	})
}

func (w *authWrapper) authContext(ctx context.Context, oauth2provider fosite.OAuth2Provider, token, crossAccountID string) (_ context.Context, err error) {
	var (
		userObj    *model.User
		accountObj *model.Account
	)

	defer func() {
		if err != nil {
			w.authError.Inc()
		} else {
			w.authSuccess.Inc()
		}
	}()

	if !testMode {
		oauth2Ctx := serverprovider.NewContext(ctx)
		tokenType, accessReq, errToken := oauth2provider.IntrospectToken(
			oauth2Ctx, token, fosite.AccessToken, &fosite.DefaultSession{})
		if errToken != nil {
			return nil, errToken
		}
		if tokenType != fosite.AccessToken {
			return nil, errAccessTokensOnlyAllows
		}
		session := accessReq.GetSession().(*serverprovider.Session)
		users := userRepository.New()
		userObj, accountObj, err = users.GetByToken(ctx, session.AccessToken)
	} else {
		users := userRepository.New()
		userObj, accountObj, err = users.GetByToken(ctx, token)
	}
	if userObj == nil {
		userObj = &model.User{}
		accountObj = &model.Account{}
	}

	if err != nil {
		return nil, err
	}

	userObj, accountObj, err = crossAccountConnect(ctx, crossAccountID, userObj, accountObj)
	if err != nil {
		return nil, err
	}

	return session.WithUserAccount(ctx, userObj, accountObj), nil
}

func (w *authWrapper) authContextJWT(ctx context.Context, jwtProvider *jwt.Provider, token *jwt.Token) (_ context.Context, err error) {
	jwtData, err := jwtProvider.ExtractTokenData(token)
	if err != nil {
		return nil, err
	}
	return w.authContextAcc(ctx, jwtData.UserID, jwtData.AccountID)
}

func (w *authWrapper) authContextAcc(ctx context.Context, userID, accountID uint64) (_ context.Context, err error) {
	userObj, accountObj, err := userAccountByID(ctx, userID, accountID, nil, nil)
	if err != nil {
		return nil, err
	}
	return session.WithUserAccount(ctx, userObj, accountObj), nil
}

func userAccountByID(ctx context.Context, uid, accid uint64, preUser *model.User, prevAccount *model.Account) (*model.User, *model.Account, error) {
	var (
		err      error
		users    = userRepository.New()
		accounts = accountRepository.New()
		account  = prevAccount
		userObj  = preUser
	)
	if uid > 0 && (preUser == nil || preUser.ID != uid) {
		if userObj, err = users.Get(ctx, uid); err != nil {
			return nil, nil, err
		}
	}
	if accid > 0 && (prevAccount == nil || prevAccount.ID != accid) {
		if account, err = accounts.Get(ctx, accid); err != nil {
			return nil, nil, err
		}
	}
	if account != nil {
		if userObj != nil && !accounts.IsMember(ctx, userObj.ID, account.ID) {
			return nil, nil, errAuthUserIsNotMemberOfAccount
		}
		if prevAccount != nil && prevAccount.ID != account.ID &&
			!prevAccount.CheckPermissions(ctx, account, session.PermAuthCross) {
			return nil, nil, errNoCrossAuthPermission
		}
		err = accounts.LoadPermissions(ctx, account, userObj)
		if err != nil {
			return nil, nil, err
		}
		if prevAccount != nil {
			account.ExtendPermissions(prevAccount.Permissions)
		}
	}
	return userObj, account, nil
}

func crossAccountConnect(ctx context.Context, crossAccountID string, userObj *model.User, accountObj *model.Account) (*model.User, *model.Account, error) {
	if crossAccountID != "" {
		userID, accountID := session.ParseCrossAuthHeader(crossAccountID)
		if userID > 0 || accountID > 0 {
			return userAccountByID(ctx, userID, accountID, userObj, accountObj)
		}
	}
	return userObj, accountObj, nil
}
