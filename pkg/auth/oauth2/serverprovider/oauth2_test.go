package serverprovider

import (
	"testing"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/pkce"
	"github.com/stretchr/testify/require"
)

func TestNewProviderPKCEAfterExplicit(t *testing.T) {
	config := &fosite.Config{}
	_ = NewProvider(config, &DatabaseStorage{}, &compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(config),
	}, nil)

	explicitIdx, pkceIdx := -1, -1
	for i, h := range config.AuthorizeEndpointHandlers {
		switch h.(type) {
		case *oauth2.AuthorizeExplicitGrantHandler:
			explicitIdx = i
		case *pkce.Handler:
			pkceIdx = i
		}
	}

	require.GreaterOrEqual(t, explicitIdx, 0, "authorize explicit handler must be registered")
	require.Greater(t, pkceIdx, explicitIdx, "PKCE handler must be registered after authorize explicit")
}
