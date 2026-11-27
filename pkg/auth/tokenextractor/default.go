package tokenextractor

import (
	"net/http"
	"strings"

	"github.com/ory/fosite"
)

// DefaultExtractor is the default token extractor. It reads the token from the
// `Authorization: Bearer` header or the `access_token` request parameter.
//
// It no longer reads the token from the OAuth2 `state` parameter: the state is
// sent to the provider and the token would leak to it and let anyone who sends
// a crafted callback link authenticate the victim's browser as the attacker.
func DefaultExtractor(r *http.Request) (string, error) {
	return fosite.AccessTokenFromRequest(r), nil
}

// AuthorizationHeaderExtractor reads the token from the `Authorization: Bearer` header only.
func AuthorizationHeaderExtractor(r *http.Request) (string, error) {
	scheme, token, found := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", nil
	}
	return strings.TrimSpace(token), nil
}
