package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// UserConnection implements collection accessor interface with pagination
type UserConnection = CollectionConnection[gqlmodels.User, gqlmodels.UserEdge]

