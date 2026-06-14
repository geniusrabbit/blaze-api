package elogin

import (
	"net/http"
)

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

// Login handles the login request and redirects to the provider
func (wr *AuthHTTPWrapper) Login(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, wr.Auth.LoginURL(wr.redirectParams(w, r, true)), http.StatusTemporaryRedirect)
}

// Callback handles the provider callback and authenticates the user
func (wr *AuthHTTPWrapper) Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		wr.Error.Error(w, r, err)
		return
	}

	token, data, err := wr.Auth.UserData(r.Context(), r.Form, wr.redirectParams(w, r, false))
	if err != nil {
		wr.Error.Error(w, r, err)
		return
	}
	wr.Success.Success(w, r, token, data)
}

// redirectParams retrieves redirect parameters from the extractor
func (wr *AuthHTTPWrapper) redirectParams(w http.ResponseWriter, r *http.Request, isLogin bool) []URLParam {
	if wr.RedirectParams != nil {
		return wr.RedirectParams.RedirectParams(w, r, isLogin)
	}
	return nil
}
