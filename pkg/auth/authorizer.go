package auth

import (
	"errors"
	"net/http"

	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
)

// ErrAuthorizationUnavailable marks an authorizer error caused by the infrastructure,
// not by the credential. The middleware answers it with 503 instead of 401.
var ErrAuthorizationUnavailable = errors.New("authorization is temporarily unavailable")

// IsNillable describes a type that can report whether it is nil/empty.
type IsNillable interface {
	IsNil() bool
}

// Authorizer defines a single authorization strategy.
type Authorizer[User IsNillable, Account IsNillable] interface {
	// AuthorizerCode returns a unique identifier for the authorizer.
	AuthorizerCode() string
	// Authorize validates request credentials and returns token, user, account, and error.
	Authorize(w http.ResponseWriter, r *http.Request) (string /* token */, User, Account, error)
}

// AuthorizeWrapper combines multiple authorizers and executes them in order.
type AuthorizeWrapper[User IsNillable, Account IsNillable] struct {
	authorizers []Authorizer[User, Account]
}

// NewAuthorizeWrapper creates a wrapper from the provided non-nil authorizers.
func NewAuthorizeWrapper[User IsNillable, Account IsNillable](authorizers ...Authorizer[User, Account]) *AuthorizeWrapper[User, Account] {
	return &AuthorizeWrapper[User, Account]{
		authorizers: xtypes.Slice[Authorizer[User, Account]](authorizers).
			Filter(func(a Authorizer[User, Account]) bool { return a != nil }),
	}
}

// Authorize runs each authorizer until one returns a non-nil user or account.
// If any authorizer returns an error, authorization stops and the error is returned.
func (a *AuthorizeWrapper[User, Account]) Authorize(w http.ResponseWriter, r *http.Request) (string, User, Account, error) {
	_, token, user, account, err := a.AuthorizeWithCode(w, r)
	return token, user, account, err
}

// AuthorizeWithCode works like Authorize and also returns the code of the authorizer that matched.
// The code is empty when no authorizer matched.
func (a *AuthorizeWrapper[User, Account]) AuthorizeWithCode(w http.ResponseWriter, r *http.Request) (string, string, User, Account, error) {
	var (
		zeroUser    User
		zeroAccount Account
	)

	for _, authorizer := range a.authorizers {
		token, user, account, err := authorizer.Authorize(w, r)
		if err != nil {
			return "", token, zeroUser, zeroAccount, err
		}
		if (!gocast.IsNil(user) && !user.IsNil()) || (!gocast.IsNil(account) && !account.IsNil()) {
			return authorizer.AuthorizerCode(), token, user, account, nil
		}
	}

	return "", "", zeroUser, zeroAccount, nil
}
