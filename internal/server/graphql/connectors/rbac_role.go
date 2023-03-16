package connectors

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository/rbac"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// Role model aliase
type Role = model.Role

// RBACRoleConnection implements collection accessor interface with pagination.
type RBACRoleConnection struct {
	ctx           context.Context
	rolesAccessor rbac.Usecase

	totalCount int64
	roles      []*model.Role

	// The edges for each of the roles's lists
	edges []*gqlmodels.RBACRoleEdge

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewRBACRoleConnection based on query object
func NewRBACRoleConnection(ctx context.Context, rolesAccessor rbac.Usecase) *RBACRoleConnection {
	return &RBACRoleConnection{
		ctx:           ctx,
		rolesAccessor: rolesAccessor,
		totalCount:    -1,
	}
}

// TotalCount returns number of campaigns
func (c *RBACRoleConnection) TotalCount() int {
	return int(c.totalCount)
}

// The edges for each of the campaigs's lists
func (c *RBACRoleConnection) Edges() []*gqlmodels.RBACRoleEdge {
	return c.edges
}

// PageInfo returns information about pages
func (c *RBACRoleConnection) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{}
	}
	return c.pageInfo
}

// Roles returns list of the roles, as a convenience when edges are not needed.
func (c *RBACRoleConnection) Roles() []*gqlmodels.RBACRole {
	var err error
	if c.roles == nil {
		c.roles, err = c.rolesAccessor.FetchList(c.ctx, nil)
		panicError(err)
	}
	return gqlmodels.FromRBACRoleModelList(c.roles)
}
