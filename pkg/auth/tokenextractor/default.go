package tokenextractor

import (
	"net/http"

	"github.com/ory/fosite"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/utils"
)

func DefaultExtractor(r *http.Request) (string, error) {
	token := fosite.AccessTokenFromRequest(r)
	if token == "" {
		state := utils.DecodeState(r.URL.Query().Get("state"))
		token = state.Get(`access_token`)
	}
	return token, nil
}
