package elogin

import (
	"context"
	"net/url"
)

type AuthAccessor interface {
	Provider() string
	Protocol() string
	LoginURL(urlParams []URLParam) string
	UserData(ctx context.Context, values url.Values, urlParams []URLParam) (*Token, *UserData, error)
}
