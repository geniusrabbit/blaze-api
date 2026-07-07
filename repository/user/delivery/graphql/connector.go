package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// UserConnection implements collection accessor interface with pagination (base schema types).
type UserConnection[TGQLUser any] = connectors.CollectionConnection[TGQLUser]

// NewUserConnection based on query object.
func NewUserConnection[
	TDomain user.Model,
	TGQLUser any,
](
	ctx context.Context,
	usersAccessor user.Usecase[TDomain],
	filter user.QOption,
	order []user.QOption,
	page *gqlmodels.Page,
	toGraphQL UserGraphQLConverter[TDomain, TGQLUser],
) *UserConnection[TGQLUser] {
	toList := UserGraphQLListConverter(toGraphQL)
	return connectors.NewCollectionConnection(ctx, &connectors.DataAccessorFunc[TGQLUser]{
		FetchDataListFunc: func(ctx context.Context) ([]TGQLUser, error) {
			opts := append(order, filter, page.Pagination())
			users, err := usersAccessor.FetchList(ctx, opts...)
			if err != nil {
				return nil, err
			}
			return toList(users), nil
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return usersAccessor.Count(ctx, filter)
		},
	}, page)
}
