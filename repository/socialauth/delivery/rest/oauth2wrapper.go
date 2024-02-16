package rest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/demdxx/gocast/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/acl"
	"github.com/geniusrabbit/blaze-api/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/elogin"
	"github.com/geniusrabbit/blaze-api/jwt"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/socialauth/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialauth/usecase"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

var errUserDateEmpty = errors.New("user data is empty")

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
func (wr *Oauth2Wrapper) RedirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) []elogin.URLParam {
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL != "" {
		return []elogin.URLParam{{Key: "redirect", Value: redirectURL}}
	}
	return nil
}

// Error handles the error occurred during the oauth2 authentication
func (wr *Oauth2Wrapper) Error(w http.ResponseWriter, r *http.Request, err error) {
	ctxlogger.Get(r.Context()).Error("Auth error",
		zap.String(`protocol`, wr.wrapper.Protocol()),
		zap.String(`provider`, wr.wrapper.Provider()),
		zap.Error(err))
	if wr.errorRedirectURL != "" {
		http.Redirect(w, r, wr.errorRedirectURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "error",
		"protocol": wr.wrapper.Protocol(),
		"provider": wr.wrapper.Provider(),
		"error":    err.Error(),
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
		ctx       = acl.WithNoPermCheck(r.Context())
	)

	// Check if user already exists (awoid permission check)
	list, err := wr.socialAuthUsecase.List(ctx, &socialauth.Filter{
		SocialID: []string{userData.ID},
		Provider: []string{wr.Provider()},
	})
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows)) {
		wr.Error(w, r, err)
		return
	}

	// If user already exists, update the token
	if len(list) > 0 {
		accSocial = list[0]
		if err := wr.updateSocialAccount(ctx, list[0], token, userData); err != nil {
			wr.Error(w, r, err)
			return
		}
	} else if accSocial, err = wr.createSocialAccountAndUser(r.Context(), token, userData); err != nil {
		wr.Error(w, r, err)
		return
	}

	// Update token if provided
	if accSocial != nil && token != nil {
		if err := wr.socialAuthUsecase.SetToken(ctx, accSocial.ID, token); err != nil {
			wr.Error(w, r, err)
			return
		}
	}

	// Session token initialization
	var sessToken string

	// Create session if provided
	if wr.sessProvider != nil && session.User(ctx).IsAnonymous() {
		sessToken, err = wr.sessProvider.CreateToken(accSocial.UserID, 0)
		if err != nil {
			wr.Error(w, r, err)
			return
		}
	}

	// Redirect to the success URL if provided
	if red := gocast.Or(r.URL.Query().Get("redirect"), wr.successRedirectURL); red != "" {
		redirectURL := urlSetQueryParams(red, map[string]string{"token": sessToken})
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"protocol": wr.wrapper.Protocol(),
		"provider": wr.wrapper.Provider(),
		"token":    sessToken,
	})
}

func (wr *Oauth2Wrapper) createSocialAccountAndUser(ctx context.Context, token *elogin.Token, userData *elogin.UserData) (*model.AccountSocial, error) {
	user := session.User(ctx)
	// Create new user or connect to the existing one
	if user.IsAnonymous() {
		user = &model.User{
			Email:   userData.Email,
			Approve: model.ApprovedApproveStatus,
		}
	}

	var scopes []string
	if userData.OAuth2conf != nil {
		scopes = userData.OAuth2conf.Scopes
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
		Scope:     scopes,
	}

	// Execute all operations in transaction
	_, err := wr.socialAuthUsecase.Register(ctx, user, socAcc)
	return socAcc, err
}

func (wr *Oauth2Wrapper) updateSocialAccount(ctx context.Context, socAcc *model.AccountSocial, token *elogin.Token, userData *elogin.UserData) error {
	socAcc.Email = gocast.Or(userData.Email, socAcc.Email)
	socAcc.FirstName = gocast.Or(userData.FirstName, socAcc.FirstName)
	socAcc.LastName = gocast.Or(userData.LastName, socAcc.LastName)
	socAcc.Avatar = gocast.Or(userData.AvatarURL, socAcc.Avatar)
	socAcc.Link = gocast.Or(userData.Link, socAcc.Link)
	socAcc.Scope = gocast.IfThen(userData.OAuth2conf != nil, userData.OAuth2conf.Scopes, socAcc.Scope)
	return wr.socialAuthUsecase.Update(ctx, socAcc.ID, socAcc)
}

func urlSetQueryParams(sUrl string, params map[string]string) string {
	if len(params) == 0 {
		return sUrl
	}
	query := url.Values{}
	baseURL := strings.SplitN(sUrl, "?", 2)
	if len(baseURL) == 2 {
		query, _ = url.ParseQuery(baseURL[1])
	}
	for k, v := range params {
		query.Set(k, v)
	}
	return baseURL[0] + "?" + query.Encode()
}
