package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AccountConnection implements collection accessor interface with pagination
type AccountConnection = CollectionConnection[gqlmodels.Account, gqlmodels.AccountEdge]

