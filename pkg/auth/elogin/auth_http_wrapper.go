package elogin

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ParamState is the URL parameter key for a state managed by the caller.
// When a RedirectParamsExtractor returns it, the AuthAccessor must use the value
// as the OAuth2 state verbatim on login, and must not do its own state check on
// callback, because the extractor has already verified it.
const ParamState = "state"

// URLParam represents a URL parameter with a key-value pair
type URLParam struct {
	Key   string
	Value string
}

// ErrorHandler defines the interface for handling authentication errors
type ErrorHandler interface {
	Error(w http.ResponseWriter, r *http.Request, err error)
}

// SuccessHandler defines the interface for handling successful authentication
type SuccessHandler interface {
	Success(w http.ResponseWriter, r *http.Request, token *Token, data *UserData)
}

// RedirectParamsExtractor defines the interface for extracting redirect parameters
type RedirectParamsExtractor interface {
	RedirectParams(w http.ResponseWriter, r *http.Request, login bool) []URLParam
}

// RedirectParamsValidator is an optional extension of RedirectParamsExtractor.
// When implemented, it is used instead of RedirectParams and can reject the
// request (an invalid redirect target or state, for example).
type RedirectParamsValidator interface {
	ValidateRedirectParams(w http.ResponseWriter, r *http.Request, login bool) ([]URLParam, error)
}

// AuthHTTPWrapper provides HTTP handler wrapping for authentication flows
type AuthHTTPWrapper struct {
	Auth           AuthAccessor
	Error          ErrorHandler
	Success        SuccessHandler
	RedirectParams RedirectParamsExtractor
}

// NewWrapper creates a new instance of AuthHTTPWrapper
func NewWrapper(auth AuthAccessor, err ErrorHandler, success SuccessHandler, redirectParams RedirectParamsExtractor) *AuthHTTPWrapper {
	return &AuthHTTPWrapper{
		Auth:           auth,
		Error:          err,
		Success:        success,
		RedirectParams: redirectParams,
	}
}

// Protocol returns the authentication protocol name
func (wr *AuthHTTPWrapper) Protocol() string {
	return wr.Auth.Protocol()
}

// Provider returns the authentication provider name
func (wr *AuthHTTPWrapper) Provider() string {
	return wr.Auth.Provider()
}

// HandleWrapper returns an HTTP handler for the authentication routes
func (wr *AuthHTTPWrapper) HandleWrapper(prefix string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", wr.Login)
	mux.HandleFunc("/callback", wr.Callback)
	if prefix != "" {
		return http.StripPrefix(prefix, mux)
	}
	return mux
}

// Login handles the login request and redirects to the provider.
//
// With `?response=json` or `Accept: application/json` it answers with
// `{"url": "<provider login URL>"}` instead of a redirect. That lets a frontend
// start the flow with a fetch request authenticated by the Authorization header
// and then navigate to the URL.
func (wr *AuthHTTPWrapper) Login(w http.ResponseWriter, r *http.Request) {
	params, err := wr.redirectParams(w, r, true)
	if err != nil {
		wr.Error.Error(w, r, err)
		return
	}
	loginURL := wr.Auth.LoginURL(params)
	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"url": loginURL})
		return
	}
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

// Callback handles the provider callback and authenticates the user
func (wr *AuthHTTPWrapper) Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		wr.Error.Error(w, r, err)
		return
	}

	params, err := wr.redirectParams(w, r, false)
	if err != nil {
		wr.Error.Error(w, r, err)
		return
	}

	token, data, err := wr.Auth.UserData(r.Context(), r.Form, params)
	if err != nil {
		wr.Error.Error(w, r, err)
		return
	}
	wr.Success.Success(w, r, token, data)
}

// redirectParams retrieves redirect parameters from the extractor
func (wr *AuthHTTPWrapper) redirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) ([]URLParam, error) {
	if validator, ok := wr.RedirectParams.(RedirectParamsValidator); ok {
		return validator.ValidateRedirectParams(w, r, isLogin)
	}
	if wr.RedirectParams != nil {
		return wr.RedirectParams.RedirectParams(w, r, isLogin), nil
	}
	return nil, nil
}

func wantsJSON(r *http.Request) bool {
	return r.URL.Query().Get("response") == "json" ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}
