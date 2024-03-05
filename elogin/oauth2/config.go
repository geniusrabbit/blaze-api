package oauth2

import (
	"context"
	"net/url"
	"strings"

	"github.com/demdxx/xtypes"
	"github.com/geniusrabbit/blaze-api/elogin"
	"golang.org/x/oauth2"
)

// DataExtractor provides a function to extract user data from oauth2 token
type DataExtractor func(ctx context.Context, token *oauth2.Token, oauth2conf *oauth2.Config) (*elogin.UserData, error)

// Config provides a configuration for oauth2 authentication
type Config struct {
	ProviderName string
	OAuth2       *oauth2.Config
	Extractor    DataExtractor
	StateCode    string
}

// Protocol returns the protocol name
func (c *Config) Protocol() string {
	return "oauth2"
}

// Provider returns the provider name
func (c *Config) Provider() string {
	return c.ProviderName
}

// LoginURL returns the login url
func (c *Config) LoginURL(params []elogin.URLParam) string {
	// reauthorize - always has for permissions
	// rerequest - for declined/revoked permissions
	// reauthenticate - always as user to confirm password
	opts := make([]oauth2.AuthCodeOption, 0, 1+len(params))
	opts = append(opts, oauth2.SetAuthURLParam("auth_type", "rerequest"))
	if len(params) > 0 {
		opts = append(opts, oauth2.SetAuthURLParam("redirect_uri", urlSetQueryParams(c.OAuth2.RedirectURL, params)))
	}
	for _, param := range params {
		if param.Key == "scope" {
			opts = append(opts, oauth2.SetAuthURLParam("scope", param.Value))
			break
		}
	}
	return c.OAuth2.AuthCodeURL(c.StateCode, opts...)
}

// OAuth2Config returns the oauth2 configuration
func (c *Config) OAuth2Config() *oauth2.Config {
	return c.OAuth2
}

// UserData returns the user data from the oauth2 token
func (c *Config) UserData(ctx context.Context, values url.Values, params []elogin.URLParam) (*elogin.Token, *elogin.UserData, error) {
	code := values.Get("code")
	scopes := c.OAuth2.Scopes

	// Check state code if it is set
	if c.StateCode != "" && c.StateCode != values.Get("state") {
		return nil, nil, elogin.ErrInvalidState
	}

	var opts []oauth2.AuthCodeOption
	if params != nil {
		opts = append(opts, oauth2.SetAuthURLParam("redirect_uri", urlSetQueryParams(c.OAuth2.RedirectURL, params)))
	}

	// Extract scopes from the params
	for _, param := range params {
		if param.Key == "scope" {
			scopes = xtypes.Slice[string](strings.Split(strings.ReplaceAll(param.Value, " ", ","), ",")).
				Apply(func(s string) string { return strings.TrimSpace(s) }).
				Filter(func(s string) bool { return s != "" }).
				Sort(func(a, b string) bool { return a < b })
			break
		}
	}

	// Exchange code for token
	oa2token, err := c.OAuth2.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, nil, err
	}

	token := &elogin.Token{
		TokenType:    oa2token.TokenType,
		AccessToken:  oa2token.AccessToken,
		RefreshToken: oa2token.RefreshToken,
		ExpiresAt:    oa2token.Expiry,
		Scopes:       scopes,
	}
	if c.Extractor == nil {
		return token, nil, nil
	}

	// Extract user data if extractor is set
	data, err := c.Extractor(ctx, oa2token.WithExtra(map[string]any{"scope": scopes, "newToken": token}), c.OAuth2)
	if err != nil {
		return nil, nil, err
	}

	return token, data, nil
}

func urlSetQueryParams(sUrl string, params []elogin.URLParam) string {
	if len(params) == 0 {
		return sUrl
	}
	query := url.Values{}
	baseURL := strings.SplitN(sUrl, "?", 2)
	if len(baseURL) == 2 {
		query, _ = url.ParseQuery(baseURL[1])
	}
	for _, it := range params {
		query.Set(it.Key, it.Value)
	}
	return baseURL[0] + "?" + query.Encode()
}
