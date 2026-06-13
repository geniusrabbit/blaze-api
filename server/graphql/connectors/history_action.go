package connectors

import (
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// HistoryActionConnection implements collection accessor interface with pagination
type HistoryActionConnection = CollectionConnection[gqlmodels.HistoryAction, gqlmodels.HistoryActionEdge]

