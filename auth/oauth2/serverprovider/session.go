package serverprovider

import (
	"context"
	"time"

	"github.com/ory/fosite"
)

// Session object value
type Session struct {
	Username string
	Subject  string

	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time

	ctx context.Context
}

// NewSession returns basic session object
func NewSession(
	ctx context.Context,
	username string,
	subject string,
	accessToken string,
	accessTokenExpiresAt time.Time,
	refreshToken string,
	refreshTokenExpiresAt time.Time,
) *Session {
	return &Session{
		Username:              username,
		Subject:               subject,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
		ctx:                   ctx,
	}
}

// SetExpiresAt sets the expiration time of a token.
func (sess *Session) SetExpiresAt(key fosite.TokenType, exp time.Time) {
	switch key {
	case fosite.AccessToken:
		sess.AccessTokenExpiresAt = exp
	case fosite.RefreshToken:
		sess.RefreshTokenExpiresAt = exp
	}
}

// GetExpiresAt returns the expiration time of a token if set, or time.IsZero() if not.
//
//	session.GetExpiresAt(fosite.AccessToken)
func (sess *Session) GetExpiresAt(key fosite.TokenType) time.Time {
	switch key {
	case fosite.AccessToken:
		return sess.AccessTokenExpiresAt
	case fosite.RefreshToken:
		return sess.RefreshTokenExpiresAt
	}
	return time.Time{}
}

// GetUsername returns the username, if set. This is optional and only used during token introspection.
func (sess *Session) GetUsername() string {
	return sess.Username
}

// GetSubject returns the subject, if set. This is optional and only used during token introspection.
func (sess *Session) GetSubject() string {
	return sess.Subject
}

// Context of the
func (sess *Session) Context() context.Context {
	return sess.ctx
}

// Clone clones the session.
func (sess *Session) Clone() fosite.Session {
	return &Session{
		Username:              sess.Username,
		Subject:               sess.Subject,
		AccessToken:           sess.AccessToken,
		AccessTokenExpiresAt:  sess.AccessTokenExpiresAt,
		RefreshToken:          sess.RefreshToken,
		RefreshTokenExpiresAt: sess.RefreshTokenExpiresAt,
		ctx:                   sess.ctx,
	}
}
