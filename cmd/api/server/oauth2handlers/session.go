package oauth2handlers

import (
	"time"

	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

func newSession(user string) *openid.DefaultSession {
	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      "https://domain",
			Subject:     user,
			Audience:    []string{"https://domain"},
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
		},
		Headers: &jwt.Headers{
			Extra: make(map[string]interface{}),
		},
	}
}
