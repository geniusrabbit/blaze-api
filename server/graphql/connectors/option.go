package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// OptionConnection implements collection accessor interface with pagination
type OptionConnection = CollectionConnection[gqlmodels.Option, gqlmodels.OptionEdge]

