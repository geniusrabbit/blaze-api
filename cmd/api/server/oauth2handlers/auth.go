package oauth2handlers

import (
	"net/http"

	"github.com/ory/fosite"
)

// Auth endpoint
func Auth(oauth2 fosite.OAuth2Provider) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		http.NotFound(rw, req)
	}
}
