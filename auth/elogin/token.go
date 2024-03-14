package elogin

import (
	"strings"
	"time"
)

// Token represents the token data
type Token struct {
	TokenType    string    `json:"token_type,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scopes       []string  `json:"scopes,omitempty"`
}

// IsExpired checks if the token is expired
func (tok *Token) IsExpired() bool {
	return tok.ExpiresAt.Before(time.Now())
}

// IsBearer checks if the token is a bearer token
func (tok *Token) IsBearer() bool {
	return strings.EqualFold(tok.TokenType, "bearer")
}
