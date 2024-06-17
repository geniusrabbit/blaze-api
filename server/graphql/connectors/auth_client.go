package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AuthClientConnection implements collection accessor interface with pagination
type AuthClientConnection = CollectionConnection[gqlmodels.AuthClient, gqlmodels.AuthClientEdge]

// NewAuthClientConnection based on query object
func NewAuthClientConnection(ctx context.Context, authClientsAccessor authclient.Usecase, page *gqlmodels.Page) *AuthClientConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.AuthClient, gqlmodels.AuthClientEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.AuthClient, error) {
			clients, err := authClientsAccessor.FetchList(ctx, nil)
			return gqlmodels.FromAuthClientModelList(clients), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return authClientsAccessor.Count(ctx, nil)
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.AuthClient) *gqlmodels.AuthClientEdge {
			return &gqlmodels.AuthClientEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
