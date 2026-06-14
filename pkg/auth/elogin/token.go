package elogin

import (
	"strings"
	"time"
)

// Token represents OAuth token data with access and refresh tokens.
type Token struct {
	TokenType    string    `json:"token_type,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scopes       []string  `json:"scopes,omitempty"`
}

// IsExpired returns true if the token has expired based on the ExpiresAt time.
func (tok *Token) IsExpired() bool {
	return tok.ExpiresAt.Before(time.Now())
}

// IsBearer returns true if the token type is "bearer" (case-insensitive).
func (tok *Token) IsBearer() bool {
	return strings.EqualFold(tok.TokenType, "bearer")
}
