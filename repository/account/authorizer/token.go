package authorizer

import "net/http"

// AuthOption to access to default user
type AuthOption struct {
	DevToken     string
	DevUserID    uint64
	DevAccountID uint64
}

// TokenExtractor defines a function type for extracting tokens from HTTP requests.
type TokenExtractor func(r *http.Request) (string, error)
