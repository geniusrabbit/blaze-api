package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AuthClientConnection implements collection accessor interface with pagination
type AuthClientConnection = connectors.CollectionConnection[*gqlmodels.AuthClient]

// NewAuthClientConnection based on query object
func NewAuthClientConnection(ctx context.Context, authClientsAccessor authclient.Usecase, page *gqlmodels.Page) *AuthClientConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.AuthClient]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.AuthClient, error) {
			clients, err := authClientsAccessor.FetchList(ctx, page.Pagination())
			return FromAuthClientModelList(clients), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return authClientsAccessor.Count(ctx, nil)
		},
	}, page)
}
