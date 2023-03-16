package oauth2handlers

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/geniusrabbit/api-template-base/internal/oauth2"
)

// Revoke endpoint
func Revoke(oauth2provider fosite.OAuth2Provider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// This context will be passed to all methods.
		ctx := oauth2.NewContext()

		// This will accept the token revocation request and validate various parameters.
		err := oauth2provider.NewRevocationRequest(ctx, req)

		// All done, send the response.
		oauth2provider.WriteRevocationResponse(ctx, rw, err)
	}
}
