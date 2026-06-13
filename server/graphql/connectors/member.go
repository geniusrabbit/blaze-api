package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// MemberConnection implements collection accessor interface with pagination
type MemberConnection = CollectionConnection[gqlmodels.Member, gqlmodels.MemberEdge]
