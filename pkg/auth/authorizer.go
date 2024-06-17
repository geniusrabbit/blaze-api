package auth

import (
	"net/http"

	"github.com/demdxx/xtypes"

	"github.com/geniusrabbit/blaze-api/model"
)

type Authorizer interface {
	AuthorizerCode() string
	Authorize(w http.ResponseWriter, r *http.Request) (string /* token */, *model.User, *model.Account, error)
}

type AuthorizeWrapper struct {
	authorizers []Authorizer
}

func NewAuthorizeWrapper(authorizers ...Authorizer) *AuthorizeWrapper {
	return &AuthorizeWrapper{
		authorizers: xtypes.Slice[Authorizer](authorizers).
			Filter(func(a Authorizer) bool { return a != nil }),
	}
}

func (a *AuthorizeWrapper) Authorize(w http.ResponseWriter, r *http.Request) (string, *model.User, *model.Account, error) {
	for _, authorizer := range a.authorizers {
		token, user, account, err := authorizer.Authorize(w, r)
		if err != nil {
			return token, nil, nil, err
		}
		if user != nil || account != nil {
			return token, user, account, nil
		}
	}
	return "", nil, nil, nil
}
