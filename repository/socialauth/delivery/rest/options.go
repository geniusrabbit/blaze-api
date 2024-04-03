package rest

import (
	"github.com/geniusrabbit/blaze-api/pkg/auth/jwt"
	"github.com/geniusrabbit/blaze-api/repository/socialauth"
)

type Option func(*Oauth2Wrapper)

// WithErrorRedirectURL sets the error redirect URL
func WithErrorRedirectURL(url string) Option {
	return func(w *Oauth2Wrapper) {
		w.errorRedirectURL = url
	}
}

// WithSuccessRedirectURL sets the success redirect URL
func WithSuccessRedirectURL(url string) Option {
	return func(w *Oauth2Wrapper) {
		w.successRedirectURL = url
	}
}

// WithSocialAuthUsecase sets the social auth usecase
func WithSocialAuthUsecase(usecase socialauth.Usecase) Option {
	return func(w *Oauth2Wrapper) {
		w.socialAuthUsecase = usecase
	}
}

// WithSessionProvider sets the session provider
func WithSessionProvider(provider *jwt.Provider) Option {
	return func(w *Oauth2Wrapper) {
		w.sessProvider = provider
	}
}
