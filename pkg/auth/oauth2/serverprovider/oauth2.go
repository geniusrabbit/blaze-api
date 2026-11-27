package serverprovider

import (
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
)

// NewProvider returns oauth2 provider
func NewProvider(config *fosite.Config, store *DatabaseStorage, strat *compose.CommonStrategy, hasher fosite.Hasher) fosite.OAuth2Provider {
	return compose.Compose(
		config,
		store,
		strat,
		// hasher,

		// enabled handlers
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2PKCEFactory, // after explicit: PKCE reads the authorize code
		// The implicit grant is not supported (removed in OAuth 2.1): use the
		// authorization code flow with PKCE.
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		// compose.OAuth2ResourceOwnerPasswordCredentialsFactory,

		compose.OAuth2TokenRevocationFactory,
		compose.OAuth2TokenIntrospectionFactory,
	)
}
