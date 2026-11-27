package tokenextractor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geniusrabbit/blaze-api/pkg/auth/elogin/utils"
)

func TestDefaultExtractorIgnoresState(t *testing.T) {
	t.Parallel()

	// A token in the OAuth2 state must not authenticate the request
	state := utils.NewState(utils.Param{Key: "access_token", Value: "attacker-token"}).Encode() //nolint:staticcheck // the old unsigned format is the attack input
	req := httptest.NewRequest(http.MethodGet, "/auth/facebook/callback?state="+state, nil)
	if token, _ := DefaultExtractor(req); token != "" {
		t.Fatalf("token = %q, want none", token)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	if token, _ := DefaultExtractor(req); token != "header-token" {
		t.Fatalf("token = %q, want header-token", token)
	}
}
