package connectors

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// DataAccessor is a generic interface for data accessors
type DataAccessor[M any] interface {
	FetchDataList(ctx context.Context) ([]M, error)
	CountData(ctx context.Context) (int64, error)
}

// DataAccessorFunc provides a generic implementation of DataAccessor as a function
type DataAccessorFunc[M any] struct {
	FetchDataListFunc func(ctx context.Context) ([]M, error)
	CountDataFunc     func(ctx context.Context) (int64, error)
}

func (d *DataAccessorFunc[M]) FetchDataList(ctx context.Context) ([]M, error) {
	return d.FetchDataListFunc(ctx)
}

func (d *DataAccessorFunc[M]) CountData(ctx context.Context) (int64, error) {
	return d.CountDataFunc(ctx)
}

// CollectionConnection implements collection accessor interface with pagination
type CollectionConnection[GQLM any] struct {
	ctx          context.Context
	dataAccessor DataAccessor[GQLM]

	totalCount int64
	page       *gqlmodels.Page
	list       []GQLM

	// Information for paginating this connection
	pageInfo *gqlmodels.PageInfo
}

// NewCollectionConnection based on query object
func NewCollectionConnection[GQLM any](ctx context.Context, dataAccessor DataAccessor[GQLM], page *gqlmodels.Page) *CollectionConnection[GQLM] {
	return &CollectionConnection[GQLM]{
		ctx:          ctx,
		dataAccessor: dataAccessor,
		totalCount:   -1,
		page:         page,
		list:         nil,
		pageInfo:     nil,
	}
}

// TotalCount returns number of campaigns
func (c *CollectionConnection[GQLM]) TotalCount() int {
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

// PageInfo returns information about pages
func (c *CollectionConnection[GQLM]) PageInfo() *gqlmodels.PageInfo {
	if c.pageInfo == nil {
		c.pageInfo = &gqlmodels.PageInfo{
			StartCursor:     "",
			EndCursor:       "",
			HasNextPage:     false,
			HasPreviousPage: false,
			Total:           c.TotalCount(),
			Page:            1,
			Count:           0,
		}
		if c.page != nil && c.page.Size != nil {
			c.pageInfo.Page = max(1, gocast.PtrAsValue(c.page.StartPage, 1))
			c.pageInfo.Count = c.pageInfo.Total/(*c.page.Size) + gocast.Int(c.pageInfo.Total%(*c.page.Size) > 0)
			c.pageInfo.HasNextPage = c.pageInfo.Count > c.pageInfo.Page
			c.pageInfo.HasPreviousPage = c.pageInfo.Page > 1
		}
	}
	return c.pageInfo
}

// List returns list of the accounts, as a convenience when edges are not needed.
func (c *CollectionConnection[GQLM]) List() []GQLM {
	if c.list == nil {
		var err error
		c.list, err = c.dataAccessor.FetchDataList(c.ctx)
		panicError(err)
	}
	return c.list
}

type _cacheValue[GQLM any] struct {
	TotalCount int64               `json:"totalCount"`
	PageInfo   *gqlmodels.PageInfo `json:"pageInfo"`
	List       []GQLM              `json:"list"`
}

// EncodableCacheValue encodes the collection connection to a byte slice for caching
func (c *CollectionConnection[GQLM]) EncodableCacheValue() ([]byte, error) {
	return json.Marshal(&_cacheValue[GQLM]{
		TotalCount: int64(c.TotalCount()),
		PageInfo:   c.PageInfo(),
		List:       c.List(),
	})
}

// DecodableCacheValue decodes the byte slice from cache back into the collection connection
func (c *CollectionConnection[GQLM]) DecodableCacheValue(data []byte) error {
	var cacheValue _cacheValue[GQLM]
	if err := json.Unmarshal(data, &cacheValue); err != nil {
		return err
	}
	c.totalCount = cacheValue.TotalCount
	c.pageInfo = cacheValue.PageInfo
	c.list = cacheValue.List
	return nil
}
