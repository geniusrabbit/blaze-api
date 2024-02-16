package elogin

import (
	"net/http"
)

type URLParam struct {
	Key   string
	Value string
}

type ErrorHandler interface {
	Error(w http.ResponseWriter, r *http.Request, err error)
}

type SuccessHandler interface {
	Success(w http.ResponseWriter, r *http.Request, token *Token, data *UserData)
}

type RedirectParamsExtractor interface {
	RedirectParams(w http.ResponseWriter, r *http.Request, login bool) []URLParam
}

// AuthHTTPWrapper provides a wrapper for auth authentication
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

// Protocol returns the protocol name
func (wr *AuthHTTPWrapper) Protocol() string {
	return wr.Auth.Protocol()
}

// Provider returns the provider name
func (wr *AuthHTTPWrapper) Provider() string {
	return wr.Auth.Provider()
}

// HandleWrapper returns the http handler which handles the auth authentication
func (wr *AuthHTTPWrapper) HandleWrapper(prefix string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", wr.Login)
	mux.HandleFunc("/callback", wr.Callback)
	if prefix != "" {
		return http.StripPrefix(prefix, mux)
	}
	return mux
}

// Login handles the login request
func (wr *AuthHTTPWrapper) Login(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, wr.Auth.LoginURL(wr.rediParams(w, r, true)), http.StatusTemporaryRedirect)
}

// Callback handles the callback request
func (wr *AuthHTTPWrapper) Callback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		wr.Error.Error(w, r, err)
		return
	}

	token, data, err := wr.Auth.UserData(r.Context(), r.Form, wr.rediParams(w, r, false))
	if err != nil {
		wr.Error.Error(w, r, err)
		return
	}
	wr.Success.Success(w, r, token, data)
}

func (wr *AuthHTTPWrapper) rediParams(w http.ResponseWriter, r *http.Request, isLogin bool) []URLParam {
	if wr.RedirectParams != nil {
		return wr.RedirectParams.RedirectParams(w, r, isLogin)
	}
	return nil
}
