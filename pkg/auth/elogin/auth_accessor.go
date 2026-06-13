package elogin

import (
	"context"
	"net/url"
)

// AuthAccessor defines the interface for authentication providers.
// It handles OAuth/login flow operations like generating login URLs
// and exchanging codes for user data.
type AuthAccessor interface {
	// Provider returns the name of the authentication provider (e.g., "google", "github").
	Provider() string

	// Protocol returns the protocol used by the provider (e.g., "oauth2", "saml").
	Protocol() string

	// LoginURL generates the login URL with the provided URL parameters.
	LoginURL(urlParams []URLParam) string

	// UserData exchanges authentication values for user information and token.
	// It takes a context, URL values from the callback, and URL parameters,
	// returning the token, user data, or an error.
	UserData(ctx context.Context, values url.Values, urlParams []URLParam) (*Token, *UserData, error)
}
