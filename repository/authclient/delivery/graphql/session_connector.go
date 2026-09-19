package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AuthSessionConnection implements collection accessor interface with pagination
type AuthSessionConnection = connectors.CollectionConnection[*gqlmodels.AuthSession]

// NewAuthSessionConnection based on query object
func NewAuthSessionConnection(
	ctx context.Context,
	sessions authclient.SessionUsecase,
	filter *gqlmodels.AuthSessionListFilter,
	order []*gqlmodels.AuthSessionListOrder,
	page *gqlmodels.Page,
) *AuthSessionConnection {
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[*gqlmodels.AuthSession]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.AuthSession, error) {
			opts := []authclient.QOption{FromSessionFilterGraphQL(filter), page.Pagination()}
			for _, o := range order {
				if ord := FromSessionOrderGraphQL(o); ord != nil {
					opts = append(opts, ord)
				}
			}
			list, err := sessions.FetchList(ctx, opts...)
			return FromAuthSessionModelList(list), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return sessions.Count(ctx, FromSessionFilterGraphQL(filter))
		},
	}, page)
}
