package rest

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/utils"
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/pkg/auth/tokenextractor"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	socialAccountModels "github.com/geniusrabbit/blaze-api/repository/socialaccount/models"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	errUserDateEmpty        = errors.New("user data is empty")
	errInvalidRedirect      = errors.New("redirect URL is not allowed")
	errLinkRequiresHeader   = errors.New("account linking requires the Authorization header")
	errLinkingNotConfigured = errors.New("account linking is not enabled")
	errLinkUserRejected     = errors.New("account linking user is not allowed")
	errLinkedToOtherUser    = errors.New("social account is linked to another user")
	errNoStateSecret        = errors.New("state secret is required: use rest.WithStateSecret or rest.WithSessionProvider")
)

const (
	connectNameKey = "cn"
	redirectKey    = "r"
	scopeKey       = "scope"
	linkUserKey    = "lu"
	linkAccountKey = "la"

	stateCookiePrefix = "elogin_state_"
	stateSecretLabel  = "blaze-api/elogin/state"
)

// Oauth2Wrapper provides a wrapper for oauth2 authentication.
// It does not depend on Email or Password user traits.
type Oauth2Wrapper struct {
	wrapper            *elogin.AuthHTTPWrapper
	sessProvider       *jwt.Provider
	socialAuthUsecase  socialauth.Usecase
	resolveAccountID   func(ctx context.Context, userID uint64) (uint64, error)
	errorRedirectURL   string
	successRedirectURL string
	allowedRedirects   []string

	// OAuth2 state signing and browser binding
	stateSecret []byte
	stateTTL    time.Duration
	stateCodec  *utils.StateCodec

	// linkSession restores the session of the user who started the linking
	linkSession func(ctx context.Context, userID, accountID uint64) (context.Context, error)

	// Optional hooks — set via With* options.
	socialAccountFactory  func(provider string, data *elogin.UserData) *socialAccountModels.AccountSocial
	socialAccountUpdater  func(acc *socialAccountModels.AccountSocial, data *elogin.UserData)
	userProvisionerErased func(ctx context.Context, provider string, data *elogin.UserData) (user.Model, error)

	// sessionUserFn is used internally by WithUserProvisioner to extract the
	// typed session user.  Defaults to session.UserModel.
	sessionUserFn func(ctx context.Context) user.Model
}

// NewWrapper creates a new instance of Oauth2Wrapper
func NewWrapper(auth elogin.AuthAccessor, options ...Option) *Oauth2Wrapper {
	wr := &Oauth2Wrapper{
		sessionUserFn: func(ctx context.Context) user.Model {
			return session.UserModel(ctx)
		},
	}
	for _, opt := range options {
		opt(wr)
	}
	if wr.socialAuthUsecase == nil {
		panic("socialauth usecase is required: use rest.WithSocialAuthUsecase")
	}
	secret := wr.stateSecret
	if len(secret) == 0 && wr.sessProvider != nil && wr.sessProvider.Secret != "" {
		// Domain-separated key, so the state signature can't be used as a JWT one
		mac := hmac.New(sha256.New, []byte(wr.sessProvider.Secret))
		_, _ = mac.Write([]byte(stateSecretLabel))
		secret = mac.Sum(nil)
	}
	if len(secret) == 0 {
		panic(errNoStateSecret)
	}
	codec, err := utils.NewStateCodec(secret, wr.stateTTL)
	if err != nil {
		panic(err)
	}
	wr.stateCodec = codec
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

// RedirectParams returns the redirect parameters for the oauth2 authentication default redirect URL.
// Errors are dropped; the auth wrapper uses ValidateRedirectParams instead.
func (wr *Oauth2Wrapper) RedirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) []elogin.URLParam {
	res, _ := wr.ValidateRedirectParams(w, r, isLogin)
	return res
}

// ValidateRedirectParams builds the signed state on login and verifies it on callback.
//
// Login query parameters:
//   - redirect: where to send the user afterwards; must pass the redirect allowlist
//   - connect_name: name of the token connection (default "default")
//   - scope: extra OAuth2 scopes
//
// When the login request is authenticated, the social account is linked to that
// user. Linking requires WithLinkSession and the Authorization header: a token
// in the URL is rejected, as anyone can make a victim's browser open such a link.
func (wr *Oauth2Wrapper) ValidateRedirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) ([]elogin.URLParam, error) {
	if !isLogin {
		raw := r.URL.Query().Get("state")
		state, err := wr.verifyState(w, r, raw)
		if err != nil {
			return nil, err
		}
		res := []elogin.URLParam{{Key: elogin.ParamState, Value: raw}}
		for _, p := range state {
			if p.Key == utils.StateNonceKey || p.Key == utils.StateExpiresKey {
				continue
			}
			res = append(res, elogin.URLParam{Key: p.Key, Value: p.Value})
		}
		return res, nil
	}

	var (
		ctx            = r.Context()
		query          = r.URL.Query()
		connectionName = query.Get("connect_name")
		redirectURL    = query.Get("redirect")
		scopes         = strings.Join(
			xtypes.Slice[string](
				strings.Split(strings.ReplaceAll(query.Get("scope"), " ", ","), ",")).
				Apply(func(s string) string { return strings.TrimSpace(s) }).
				Filter(func(s string) bool { return s != "" }).
				Sort(func(a, b string) bool { return a < b }),
			" ")
		state utils.State
	)

	if redirectURL != "" {
		if !wr.isAllowedRedirect(redirectURL) {
			return nil, errInvalidRedirect
		}
		state = state.Extend(redirectKey, redirectURL)
	}
	if connectionName != "" {
		state = state.Extend(connectNameKey, connectionName)
	}
	if scopes != "" {
		state = state.Extend(scopeKey, scopes)
	}

	// Linking to the current user
	if token := session.Token(ctx); token != "" {
		userObj, accountObj := session.UserAccount(ctx)
		if !gocast.IsNil(userObj) && !userObj.IsAnonymous() && userObj.GetID() > 0 {
			if headerToken, _ := tokenextractor.AuthorizationHeaderExtractor(r); headerToken != token {
				return nil, errLinkRequiresHeader
			}
			if wr.linkSession == nil {
				return nil, errLinkingNotConfigured
			}
			state = state.Extend(linkUserKey, strconv.FormatUint(userObj.GetID(), 10))
			if !gocast.IsNil(accountObj) && accountObj.GetID() > 0 {
				state = state.Extend(linkAccountKey, strconv.FormatUint(accountObj.GetID(), 10))
			}
		}
	}

	encoded, nonce, err := wr.stateCodec.Encode(state)
	if err != nil {
		return nil, err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wr.stateCookieName(),
		Value:    nonce,
		Path:     "/",
		MaxAge:   int(wr.stateCodec.TTL().Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})

	res := []elogin.URLParam{{Key: elogin.ParamState, Value: encoded}}
	if scopes != "" {
		res = append(res, elogin.URLParam{Key: scopeKey, Value: scopes})
	}
	return res, nil
}

// verifyState checks the state signature, lifetime and the browser binding
// cookie, and clears the cookie so the state can't be used twice.
func (wr *Oauth2Wrapper) verifyState(w http.ResponseWriter, r *http.Request, raw string) (utils.State, error) {
	state, err := wr.stateCodec.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", elogin.ErrInvalidState, err)
	}
	cookie, err := r.Cookie(wr.stateCookieName())
	if err != nil || cookie.Value == "" ||
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state.Get(utils.StateNonceKey))) != 1 {
		return nil, fmt.Errorf("%w: state is not bound to this browser", elogin.ErrInvalidState)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     wr.stateCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
	return state, nil
}

// callbackState returns the verified state of the callback request, if valid.
// The browser binding is checked (and the cookie cleared) in verifyState before.
func (wr *Oauth2Wrapper) callbackState(r *http.Request) utils.State {
	state, err := wr.stateCodec.Decode(r.URL.Query().Get("state"))
	if err != nil {
		return nil
	}
	return state
}

func (wr *Oauth2Wrapper) stateCookieName() string {
	return stateCookiePrefix + wr.wrapper.Provider()
}

// isAllowedRedirect reports whether the user may be sent to the target.
func (wr *Oauth2Wrapper) isAllowedRedirect(target string) bool {
	if strings.ContainsAny(target, "\\\r\n\t") {
		return false
	}
	u, err := url.Parse(target)
	if err != nil || u.User != nil {
		return false
	}
	// Relative path on the same origin, but not a scheme-relative "//host"
	if u.Scheme == "" && u.Host == "" {
		return strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "//")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return false
	}
	allowed := append([]string{}, wr.allowedRedirects...)
	for _, base := range []string{wr.successRedirectURL, wr.errorRedirectURL} {
		if b, err := url.Parse(base); err == nil && b.Host != "" {
			allowed = append(allowed, b.Scheme+"://"+b.Host)
		}
	}
	for _, entry := range allowed {
		a, err := url.Parse(entry)
		if err != nil || a.Host == "" {
			continue
		}
		if !strings.EqualFold(a.Scheme, u.Scheme) || !strings.EqualFold(a.Host, u.Host) {
			continue
		}
		prefix := strings.TrimSuffix(a.Path, "/")
		if prefix == "" || u.Path == prefix || strings.HasPrefix(u.Path, prefix+"/") {
			return true
		}
	}
	return false
}

// redirectTarget returns the validated redirect target from the state or the default.
func (wr *Oauth2Wrapper) redirectTarget(state utils.State, def string) string {
	if red := state.Get(redirectKey); red != "" && wr.isAllowedRedirect(red) {
		return red
	}
	return def
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// Error handles the error occurred during the oauth2 authentication
func (wr *Oauth2Wrapper) Error(w http.ResponseWriter, r *http.Request, err error) {
	state := wr.callbackState(r)
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
	if red := wr.redirectTarget(state, wr.successRedirectURL); red != "" {
		redirectURL := urlSetQueryParams(red, map[string]string{"error": err.Error()})
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorStatus(err))
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "error",
		"connect_name": connectName,
		"protocol":     wr.wrapper.Protocol(),
		"provider":     wr.wrapper.Provider(),
		"error":        err.Error(),
	})
}

func errorStatus(err error) int {
	switch {
	case errors.Is(err, errInvalidRedirect), errors.Is(err, errLinkRequiresHeader),
		errors.Is(err, errLinkingNotConfigured), errors.Is(err, elogin.ErrInvalidState):
		return http.StatusBadRequest
	case errors.Is(err, errLinkUserRejected), errors.Is(err, errLinkedToOtherUser):
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

// Success handles the success of the oauth2 authentication
func (wr *Oauth2Wrapper) Success(w http.ResponseWriter, r *http.Request, token *elogin.Token, userData *elogin.UserData) {
	if userData == nil || userData.ID == "" {
		wr.Error(w, r, errUserDateEmpty)
		return
	}

	var (
		accSocial *socialAccountModels.AccountSocial
		expiresAt time.Time
		baseCtx   = r.Context()
		state     = wr.callbackState(r)
		err       error
	)

	// Restore the session of the user who started the linking
	if linkUserID := gocast.Uint64(state.Get(linkUserKey)); linkUserID > 0 {
		if wr.linkSession == nil {
			wr.Error(w, r, errLinkingNotConfigured)
			return
		}
		if baseCtx, err = wr.linkSession(baseCtx, linkUserID, gocast.Uint64(state.Get(linkAccountKey))); err != nil {
			wr.Error(w, r, fmt.Errorf("%w: %w", errLinkUserRejected, err))
			return
		}
	}
	ctx := acl.WithNoPermCheck(baseCtx)

	// Get session connection name
	connectName := gocast.Or(state.Get(connectNameKey), "default")

	// Check if user already exists (awoid permission check)
	list, err := wr.socialAuthUsecase.List(ctx, &socialauth.Filter{
		SocialID:        []string{userData.ID},
		Provider:        []string{wr.Provider()},
		// A disconnected (soft-deleted) social account must not log in as
		// its former user; it is registered again like a new one.
		RetrieveDeleted: false,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, sql.ErrNoRows) {
		wr.Error(w, r, err)
		return
	}

	// If user already exists, update the token
	if len(list) > 0 {
		accSocial = list[0]
		if sessUser := session.UserModel(ctx); sessUser != nil && !sessUser.IsAnonymous() &&
			sessUser.GetID() > 0 && sessUser.GetID() != accSocial.UserID {
			wr.Error(w, r, errLinkedToOtherUser)
			return
		}
		if err := wr.updateSocialAccount(ctx, list[0], userData); err != nil {
			wr.Error(w, r, err)
			return
		}
	} else if accSocial, err = wr.createSocialAccountAndUser(baseCtx, userData); err != nil {
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
	var sessToken string

	// Create internal session if provided
	if wr.sessProvider != nil && session.User(ctx).IsAnonymous() {
		// Get preoritized user account
		accountID := uint64(0)
		if wr.resolveAccountID != nil {
			accountID, _ = wr.resolveAccountID(ctx, accSocial.UserID)
		}

		// Create new session token for the user and social account connection
		sessToken, expiresAt, err = wr.sessProvider.CreateToken(accSocial.UserID, accountID, accSocial.ID)
		if err != nil {
			wr.Error(w, r, err)
			return
		}
	}

	// Redirect to the success URL if provided
	if red := wr.redirectTarget(state, wr.successRedirectURL); red != "" {
		params := map[string]string{}
		if sessToken != "" {
			params["access_token"] = sessToken
		}
		http.Redirect(w, r, urlSetQueryParams(red, params), http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	resp := map[string]any{
		"status":       "ok",
		"protocol":     wr.wrapper.Protocol(),
		"provider":     wr.wrapper.Provider(),
		"connect_name": connectName,
	}
	if sessToken != "" {
		resp["access_token"] = sessToken
		resp["expires_at"] = expiresAt
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// createSocialAccountAndUser builds an AccountSocial record and registers the
// owner user via the usecase.  The oauthUser internal type is intentionally
// absent — owner resolution is delegated to the usecase's provisioner or
// default flow.  The caller may supply a custom factory via WithSocialAccountFactory.
func (wr *Oauth2Wrapper) createSocialAccountAndUser(ctx context.Context, userData *elogin.UserData) (*socialAccountModels.AccountSocial, error) {
	// Build the AccountSocial record — use the factory hook when provided.
	var socAcc *socialAccountModels.AccountSocial
	if wr.socialAccountFactory != nil {
		socAcc = wr.socialAccountFactory(wr.Provider(), userData)
	} else {
		socAcc = &socialAccountModels.AccountSocial{
			Provider:  wr.Provider(),
			SocialID:  userData.ID,
			Email:     userData.Email,
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Avatar:    userData.AvatarURL,
			Link:      userData.Link,
		}
	}

	// Determine the owner user.  When a type-erased provisioner is present we
	// call it here; otherwise the usecase's built-in flow applies when the
	// owner has no ID (see usecase.Register / ensureOwner).
	var owner user.Model
	if wr.userProvisionerErased != nil {
		var err error
		owner, err = wr.userProvisionerErased(ctx, wr.Provider(), userData)
		if err != nil {
			return nil, err
		}
	} else {
		owner = wr.sessionUserFn(ctx)
		if owner == nil {
			return nil, errUserDateEmpty
		}
	}

	_, err := wr.socialAuthUsecase.Register(ctx, owner, socAcc)
	return socAcc, err
}

// updateSocialAccount refreshes an existing AccountSocial with the latest
// provider data.  Uses the updater hook when set, otherwise applies defaults.
func (wr *Oauth2Wrapper) updateSocialAccount(ctx context.Context, socAcc *socialAccountModels.AccountSocial, userData *elogin.UserData) error {
	if wr.socialAccountUpdater != nil {
		wr.socialAccountUpdater(socAcc, userData)
	} else {
		socAcc.Email = gocast.Or(userData.Email, socAcc.Email)
		socAcc.FirstName = gocast.Or(userData.FirstName, socAcc.FirstName)
		socAcc.LastName = gocast.Or(userData.LastName, socAcc.LastName)
		socAcc.Avatar = gocast.Or(userData.AvatarURL, socAcc.Avatar)
		socAcc.Link = gocast.Or(userData.Link, socAcc.Link)
	}
	socAcc.DeletedAt = gorm.DeletedAt{Valid: false, Time: time.Now()}

	// Update social account
	return wr.socialAuthUsecase.Update(ctx, socAcc.ID, socAcc)
}
