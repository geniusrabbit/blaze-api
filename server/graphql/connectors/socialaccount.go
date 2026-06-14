package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// SocialAccountConnection implements collection accessor interface with pagination
type SocialAccountConnection = CollectionConnection[gqlmodels.SocialAccount, gqlmodels.SocialAccountEdge]

