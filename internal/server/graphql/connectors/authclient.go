package connectors

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository/authclient"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// AuthClient model aliase
type AuthClient = model.AuthClient

// AuthClientConnection implements collection accessor interface with pagination.
type AuthClientConnection struct {
	ctx                 context.Context
	authClientsAccessor authclient.Usecase

	totalCount  int64
	authClients []*model.AuthClient

	// The edges for each of the authClients's lists
	edges []*gqlmodels.AuthClientEdge

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewAuthClientConnection based on query object
func NewAuthClientConnection(ctx context.Context, authClientsAccessor authclient.Usecase) *AuthClientConnection {
	return &AuthClientConnection{
		ctx:                 ctx,
		authClientsAccessor: authClientsAccessor,
		totalCount:          -1,
	}
}

// TotalCount returns number of campaigns
func (c *AuthClientConnection) TotalCount() int {
	return int(c.totalCount)
}

// The edges for each of the campaigs's lists
func (c *AuthClientConnection) Edges() []*gqlmodels.AuthClientEdge {
	return c.edges
}

// PageInfo returns information about pages
func (c *AuthClientConnection) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{}
	}
	return c.pageInfo
}

// AuthClients returns list of the authClients, as a convenience when edges are not needed.
func (c *AuthClientConnection) AuthClients() []*gqlmodels.AuthClient {
	var err error
	if c.authClients == nil {
		c.authClients, err = c.authClientsAccessor.FetchList(c.ctx, nil)
		panicError(err)
	}
	return gqlmodels.FromAuthClientModelList(c.authClients)
}
