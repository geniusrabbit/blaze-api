package middleware

import (
	"net/http"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
)

// RequestID middleware adds a unique request id to the request context
func RequestID(next http.Handler, opts ...requestid.Option) http.Handler {
	return requestid.RequestID(next, opts...)
}
