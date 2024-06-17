package connectors

import (
	"context"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository/user"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// UserConnection implements collection accessor interface with pagination
type UserConnection = CollectionConnection[gqlmodels.User, gqlmodels.UserEdge]

// NewUserConnection based on query object
func NewUserConnection(ctx context.Context, usersAccessor user.Usecase, filter *gqlmodels.UserListFilter, order *gqlmodels.UserListOrder, page *gqlmodels.Page) *UserConnection {
	return NewCollectionConnection(ctx, &DataAccessorFunc[gqlmodels.User, gqlmodels.UserEdge]{
		FetchDataListFunc: func(ctx context.Context) ([]*gqlmodels.User, error) {
			users, err := usersAccessor.FetchList(ctx, filter.Filter(), order.Order(), page.Pagination())
			return gqlmodels.FromUserModelList(users), err
		},
		CountDataFunc: func(ctx context.Context) (int64, error) {
			return usersAccessor.Count(ctx, filter.Filter())
		},
		ConvertToEdgeFunc: func(obj *gqlmodels.User) *gqlmodels.UserEdge {
			return &gqlmodels.UserEdge{
				Cursor: gocast.Str(obj.ID),
				Node:   obj,
			}
		},
	}, page)
}
