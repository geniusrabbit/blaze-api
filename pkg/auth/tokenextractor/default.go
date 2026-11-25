package tokenextractor

import (
	"net/http"
	"strings"

	"github.com/ory/fosite"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/utils"
)

// DefaultExtractor is the default token extractor that looks for the token in the Authorization header and then in the state query parameter.
// Parse token from state query parameter is for the case when the token is passed in the state parameter during the e-login flow, which is used by the frontend to pass the token to the backend after successful login.
func DefaultExtractor(r *http.Request) (string, error) {
	token := fosite.AccessTokenFromRequest(r)
	if token == "" {
		state := utils.DecodeState(r.URL.Query().Get("state"))
		token = state.Get(`access_token`)
	}
	return token, nil
}

// AuthorizationHeaderExtractor reads the token from the `Authorization: Bearer` header only.
func AuthorizationHeaderExtractor(r *http.Request) (string, error) {
	scheme, token, found := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", nil
	}
	return strings.TrimSpace(token), nil
}
