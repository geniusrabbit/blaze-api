package connectors

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// Account model aliase
type Account = model.Account

// AccountConnection implements collection accessor interface with pagination.
type AccountConnection struct {
	ctx              context.Context
	accountsAccessor account.Usecase

	totalCount int64
	accounts   []*model.Account

	// The edges for each of the accounts's lists
	edges []*gqlmodels.AccountEdge

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewAccountConnection based on query object
func NewAccountConnection(ctx context.Context, accountsAccessor account.Usecase) *AccountConnection {
	return &AccountConnection{
		ctx:              ctx,
		accountsAccessor: accountsAccessor,
		totalCount:       -1,
	}
}

// TotalCount returns number of campaigns
func (c *AccountConnection) TotalCount() int {
	return int(c.totalCount)
}

// The edges for each of the campaigs's lists
func (c *AccountConnection) Edges() []*gqlmodels.AccountEdge {
	return c.edges
}

// PageInfo returns information about pages
func (c *AccountConnection) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{}
	}
	return c.pageInfo
}

// Accounts returns list of the accounts, as a convenience when edges are not needed.
func (c *AccountConnection) Accounts() []*gqlmodels.Account {
	var err error
	if c.accounts == nil {
		c.accounts, err = c.accountsAccessor.FetchList(c.ctx, nil)
		panicError(err)
	}
	return gqlmodels.FromAccountModelList(c.accounts)
}
