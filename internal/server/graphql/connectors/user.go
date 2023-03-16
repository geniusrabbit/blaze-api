package connectors

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository/user"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// User model aliase
type User = model.User

// UserConnection implements collection accessor interface with pagination.
type UserConnection struct {
	ctx           context.Context
	usersAccessor user.Usecase

	totalCount int64
	users      []*model.User

	// The edges for each of the campaigs's lists
	edges []*gqlmodels.UserEdge

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewUserConnection based on query object
func NewUserConnection(ctx context.Context, usersAccessor user.Usecase) *UserConnection {
	return &UserConnection{
		ctx:           ctx,
		usersAccessor: usersAccessor,
		totalCount:    -1,
	}
}

// TotalCount returns number of campaigns
func (c *UserConnection) TotalCount() int {
	return int(c.totalCount)
}

// The edges for each of the campaigs's lists
func (c *UserConnection) Edges() []*gqlmodels.UserEdge {
	return c.edges
}

// PageInfo returns information about pages
func (c *UserConnection) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{}
	}
	return c.pageInfo
}

// Users returns list of the users, as a convenience when edges are not needed.
func (c *UserConnection) Users() []*gqlmodels.User {
	var err error
	if c.users == nil {
		c.users, err = c.usersAccessor.FetchList(c.ctx, 0, 0, 0)
		panicError(err)
	}
	return gqlmodels.FromUserModelList(c.users)
}
