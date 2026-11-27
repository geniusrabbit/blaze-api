package oauth2

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin"
)

func newTestConfig() *Config {
	return &Config{
		ProviderName: "test",
		StateCode:    "fixed",
		OAuth2: &oauth2.Config{
			ClientID: "id",
			Endpoint: oauth2.Endpoint{AuthURL: "https://provider.example/auth", TokenURL: "http://127.0.0.1:0/token"},
		},
	}
}

func TestLoginURLUsesManagedState(t *testing.T) {
	t.Parallel()

	loginURL, err := url.Parse(newTestConfig().LoginURL([]elogin.URLParam{
		{Key: elogin.ParamState, Value: "signed.state"},
		{Key: "scope", Value: "email"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got := loginURL.Query().Get("state"); got != "signed.state" {
		t.Fatalf("state = %q, want the managed state", got)
	}
	if got := loginURL.Query().Get("scope"); got != "email" {
		t.Fatalf("scope = %q", got)
	}
}

func TestUserDataStateCode(t *testing.T) {
	t.Parallel()

	// Without a managed state the fixed state code is still checked
	_, _, err := newTestConfig().UserData(context.Background(), url.Values{"state": {"bogus"}}, nil)
	if !errors.Is(err, elogin.ErrInvalidState) {
		t.Fatalf("error = %v, want ErrInvalidState", err)
	}

	// A managed (already verified) state skips it and goes on to the code exchange
	_, _, err = newTestConfig().UserData(context.Background(), url.Values{"state": {"signed.state"}},
		[]elogin.URLParam{{Key: elogin.ParamState, Value: "signed.state"}})
	if err == nil || errors.Is(err, elogin.ErrInvalidState) {
		t.Fatalf("error = %v, want an exchange error", err)
	}
}
