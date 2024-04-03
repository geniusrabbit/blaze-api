package rest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/utils"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/socialauth/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialauth/usecase"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

var errUserDateEmpty = errors.New("user data is empty")

const (
	connectNameKey = "cn"
	redirectKey    = "r"
)

// Oauth2Wrapper provides a wrapper for oauth2 authentication
type Oauth2Wrapper struct {
	wrapper            *elogin.AuthHTTPWrapper
	sessProvider       *jwt.Provider
	socialAuthUsecase  socialauth.Usecase
	errorRedirectURL   string
	successRedirectURL string
}

// NewWrapper creates a new instance of Oauth2Wrapper
func NewWrapper(auth elogin.AuthAccessor, options ...Option) *Oauth2Wrapper {
	wr := &Oauth2Wrapper{}
	for _, opt := range options {
		opt(wr)
	}
	if wr.socialAuthUsecase == nil {
		wr.socialAuthUsecase = usecase.New(userrepo.New(), repository.New())
	}
	wr.wrapper = elogin.NewWrapper(auth, wr, wr, wr)
	return wr
}

// Provider returns the provider name
func (wr *Oauth2Wrapper) Provider() string {
	return wr.wrapper.Provider()
}

// HandleWrapper returns the http handler which handles the oauth2 authentication
// endpoints like /login and /callback with the given prefix
func (wr *Oauth2Wrapper) HandleWrapper(prefix string) http.Handler {
	return wr.wrapper.HandleWrapper(prefix)
}

// RedirectParams returns the redirect parameters for the oauth2 authentication default redirect URL
func (wr *Oauth2Wrapper) RedirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) (res []elogin.URLParam) {
	query := r.URL.Query()

	if !isLogin {
		state := utils.DecodeState(query.Get("state"))
		for _, p := range state {
			res = append(res, elogin.URLParam{Key: p.Key, Value: p.Value})
		}
		return res
	}

	var (
		connectionName = query.Get("connect_name")
		redirectURL    = query.Get("redirect")
		scopes         = strings.Join(
			xtypes.Slice[string](
				strings.Split(strings.ReplaceAll(query.Get("scope"), " ", ","), ",")).
				Apply(func(s string) string { return strings.TrimSpace(s) }).
				Filter(func(s string) bool { return s != "" }).
				Sort(func(a, b string) bool { return a < b }),
			" ")
	)
	if token := session.Token(r.Context()); token != "" {
		res = append(res, elogin.URLParam{Key: "access_token", Value: token})
	}
	if redirectURL != "" {
		res = append(res, elogin.URLParam{Key: redirectKey, Value: redirectURL})
	}
	if connectionName != "" {
		res = append(res, elogin.URLParam{Key: connectNameKey, Value: connectionName})
	}
	if scopes != "" {
		res = append(res, elogin.URLParam{Key: "scope", Value: scopes})
	}
	return res
}

// Error handles the error occurred during the oauth2 authentication
func (wr *Oauth2Wrapper) Error(w http.ResponseWriter, r *http.Request, err error) {
	state := utils.DecodeState(r.URL.Query().Get("state"))
	connectName := gocast.Or(state.Get(connectNameKey), "default")
	ctxlogger.Get(r.Context()).Error("Auth error",
		zap.String(`protocol`, wr.wrapper.Protocol()),
		zap.String(`provider`, wr.wrapper.Provider()),
		zap.String(`connect_name`, connectName),
		zap.Error(err))
	if wr.errorRedirectURL != "" {
		http.Redirect(w, r, wr.errorRedirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Redirect to the error URL if provided
	if red := gocast.Or(state.Get(redirectKey), wr.successRedirectURL); red != "" {
		redirectURL := urlSetQueryParams(red, map[string]string{"error": err.Error()})
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "error",
		"connect_name": connectName,
		"protocol":     wr.wrapper.Protocol(),
		"provider":     wr.wrapper.Provider(),
		"error":        err.Error(),
	})
}

// Success handles the success of the oauth2 authentication
func (wr *Oauth2Wrapper) Success(w http.ResponseWriter, r *http.Request, token *elogin.Token, userData *elogin.UserData) {
	if userData == nil || userData.ID == "" {
		wr.Error(w, r, errUserDateEmpty)
		return
	}

	var (
		accSocial *model.AccountSocial
		expiresAt time.Time
		ctx       = acl.WithNoPermCheck(r.Context())
		state     = utils.DecodeState(r.URL.Query().Get("state"))
	)

	// Get session connection name
	connectName := gocast.Or(state.Get(connectNameKey), "default")

	// Check if user already exists (awoid permission check)
	list, err := wr.socialAuthUsecase.List(ctx, &socialauth.Filter{
		SocialID:        []string{userData.ID},
		Provider:        []string{wr.Provider()},
		RetrieveDeleted: true,
	})
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows)) {
		wr.Error(w, r, err)
		return
	}

	// If user already exists, update the token
	if len(list) > 0 {
		accSocial = list[0]
		if err := wr.updateSocialAccount(ctx, list[0], userData); err != nil {
			wr.Error(w, r, err)
			return
		}
	} else if accSocial, err = wr.createSocialAccountAndUser(r.Context(), userData); err != nil {
		wr.Error(w, r, err)
		return
	}

	// Update token if provided
	if accSocial != nil && token != nil {
		if err := wr.socialAuthUsecase.SetToken(ctx, connectName, accSocial.ID, token); err != nil {
			wr.Error(w, r, err)
			return
		}
	}

	// Internal session token initialization
	sessToken := session.Token(ctx)

	// Create internal session if provided
	if sessToken == "" && wr.sessProvider != nil && session.User(ctx).IsAnonymous() {
		// Get preoritized user account
		accountID := uint64(0)
		acclist, err := accrepo.New().FetchList(ctx, &account.Filter{
			UserID: []uint64{accSocial.UserID},
			Status: []model.ApproveStatus{model.ApprovedApproveStatus, model.PendingApproveStatus},
		}, nil, nil)
		if err == nil && len(acclist) > 0 {
			for _, acc := range acclist {
				if acc.Approve.IsApproved() {
					accountID = acc.ID
					break
				}
				if accountID == 0 {
					accountID = acc.ID
				}
			}
		}

		// Create new session token for the user and social account connection
		sessToken, expiresAt, err = wr.sessProvider.CreateToken(accSocial.UserID, accountID, accSocial.ID)
		if err != nil {
			wr.Error(w, r, err)
			return
		}
	}

	// Redirect to the success URL if provided
	if red := gocast.Or(state.Get(redirectKey), wr.successRedirectURL); red != "" {
		redirectURL := urlSetQueryParams(red, map[string]string{"access_token": sessToken})
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "ok",
		"protocol":     wr.wrapper.Protocol(),
		"provider":     wr.wrapper.Provider(),
		"connect_name": connectName,
		"expires_at":   expiresAt,
		"access_token": sessToken,
	})
}

func (wr *Oauth2Wrapper) createSocialAccountAndUser(ctx context.Context, userData *elogin.UserData) (*model.AccountSocial, error) {
	user := session.User(ctx)

	// Create new user or connect to the existing one
	if user.IsAnonymous() {
		user = &model.User{
			Email:   userData.Email,
			Approve: model.ApprovedApproveStatus,
		}
	}

	// Connect user to the social account
	socAcc := &model.AccountSocial{
		Provider:  wr.Provider(),
		SocialID:  userData.ID,
		Email:     userData.Email,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Avatar:    userData.AvatarURL,
		Link:      userData.Link,
	}

	// Execute all operations in transaction
	_, err := wr.socialAuthUsecase.Register(ctx, user, socAcc)
	return socAcc, err
}

func (wr *Oauth2Wrapper) updateSocialAccount(ctx context.Context, socAcc *model.AccountSocial, userData *elogin.UserData) error {
	socAcc.Email = gocast.Or(userData.Email, socAcc.Email)
	socAcc.FirstName = gocast.Or(userData.FirstName, socAcc.FirstName)
	socAcc.LastName = gocast.Or(userData.LastName, socAcc.LastName)
	socAcc.Avatar = gocast.Or(userData.AvatarURL, socAcc.Avatar)
	socAcc.Link = gocast.Or(userData.Link, socAcc.Link)
	socAcc.DeletedAt = gorm.DeletedAt{Valid: false, Time: time.Now()}

	// Update social account
	return wr.socialAuthUsecase.Update(ctx, socAcc.ID, socAcc)
}
