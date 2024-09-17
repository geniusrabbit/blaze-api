package connectors

import (
	"context"
	"errors"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
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
	page       *gqlmodels.Page
	list       []*GQLM

	// The edges for each of the accounts's lists
	edges []*EdgeT

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewCollectionConnection based on query object
func NewCollectionConnection[GQLM any, EdgeT any](ctx context.Context, dataAccessor DataAccessor[GQLM, EdgeT], page *gqlmodels.Page) *CollectionConnection[GQLM, EdgeT] {
	return &CollectionConnection[GQLM, EdgeT]{
		ctx:          ctx,
		dataAccessor: dataAccessor,
		totalCount:   -1,
		page:         page,
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
		if errors.Is(err, acl.ErrNoPermissions) {
			c.totalCount = -1
		} else {
			panicError(err)
		}
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
			Total:           c.TotalCount(),
			Page:            0,
			Count:           0,
		}
		if edges := c.Edges(); len(edges) > 0 {
			cur1, _ := gocast.StructFieldValue(edges[0], "Cursor")
			cur2, _ := gocast.StructFieldValue(edges[len(edges)-1], "Cursor")
			c.pageInfo.StartCursor = gocast.Str(cur1)
			c.pageInfo.EndCursor = gocast.Str(cur2)
		}
		if c.page != nil && c.page.Size != nil {
			c.pageInfo.Page = gocast.PtrAsValue(c.page.StartPage, 0)
			c.pageInfo.Count = c.pageInfo.Total/(*c.page.Size) + gocast.Int(c.pageInfo.Total%(*c.page.Size) > 0)
			c.pageInfo.HasNextPage = c.pageInfo.Total > c.pageInfo.Page
			c.pageInfo.HasPreviousPage = c.pageInfo.Page > 1
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
