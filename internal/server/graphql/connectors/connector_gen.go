package connectors

import (
	"context"

	gqlmodels "github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
)

// DataAccessor is a generic interface for data accessors
type DataAccessor[M any, EdgeT any] interface {
	FetchDataList(ctx context.Context) ([]*M, error)
	CountData(ctx context.Context) (int64, error)
	ConvertToEdge(obj *M) *EdgeT
}

// DataAccessorFunc provides a generic implementation of DataAccessor as a function
type DataAccessorFunc[M any, EdgeT any] struct {
	FetchDataListFunc func(ctx context.Context) ([]*M, error)
	CountDataFunc     func(ctx context.Context) (int64, error)
	ConvertToEdgeFunc func(obj *M) *EdgeT
}

func (d *DataAccessorFunc[M, EdgeT]) FetchDataList(ctx context.Context) ([]*M, error) {
	return d.FetchDataListFunc(ctx)
}

func (d *DataAccessorFunc[M, EdgeT]) CountData(ctx context.Context) (int64, error) {
	return d.CountDataFunc(ctx)
}

func (d *DataAccessorFunc[M, EdgeT]) ConvertToEdge(obj *M) *EdgeT {
	return d.ConvertToEdgeFunc(obj)
}

// CollectionConnection implements collection accessor interface with pagination
type CollectionConnection[GQLM any, EdgeT any] struct {
	ctx          context.Context
	dataAccessor DataAccessor[GQLM, EdgeT]

	totalCount int64
	list       []*GQLM

	// The edges for each of the accounts's lists
	edges []*EdgeT

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewCollectionConnection based on query object
func NewCollectionConnection[GQLM any, EdgeT any](ctx context.Context, dataAccessor DataAccessor[GQLM, EdgeT]) *CollectionConnection[GQLM, EdgeT] {
	return &CollectionConnection[GQLM, EdgeT]{
		ctx:          ctx,
		dataAccessor: dataAccessor,
		totalCount:   -1,
		list:         nil,
		edges:        nil,
		pageInfo:     nil,
	}
}

// TotalCount returns number of campaigns
func (c *CollectionConnection[GQLM, EdgeT]) TotalCount() int {
	if c.totalCount < 0 {
		var err error
		c.totalCount, err = c.dataAccessor.CountData(c.ctx)
		panicError(err)
	}
	return int(c.totalCount)
}

// The edges for each of the campaigs's lists
func (c *CollectionConnection[GQLM, EdgeT]) Edges() []*EdgeT {
	if c.edges == nil {
		for _, obj := range c.List() {
			c.edges = append(c.edges, c.dataAccessor.ConvertToEdge(obj))
		}
	}
	return c.edges
}

// PageInfo returns information about pages
func (c *CollectionConnection[GQLM, EdgeT]) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{
			StartCursor:     "",
			EndCursor:       "",
			HasNextPage:     false,
			HasPreviousPage: false,
			Count:           c.TotalCount(),
		}
	}
	return c.pageInfo
}

// List returns list of the accounts, as a convenience when edges are not needed.
func (c *CollectionConnection[GQLM, EdgeT]) List() []*GQLM {
	if c.list == nil {
		var err error
		c.list, err = c.dataAccessor.FetchDataList(c.ctx)
		panicError(err)
	}
	return c.list
}
