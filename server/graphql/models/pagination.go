package models

import (
	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Information for paginating
type Page struct {
	// Start after the cursor ID
	After *string `json:"after,omitempty"`
	// Start after some records
	Offset *int `json:"offset,omitempty"`
	// Page number to start at (0-based), defaults to 0 (0, 1, 2, etc.)
	StartPage *int `json:"startPage,omitempty"`
	// Maximum number of items to return
	Size *int `json:"size,omitempty"`
}

func (p *Page) Pagination() *repository.Pagination {
	if p == nil {
		return nil
	}
	return &repository.Pagination{
		After:  gocast.PtrAsValue(p.After, ""),
		Offset: gocast.PtrAsValue(p.Offset, 0),
		Page:   gocast.PtrAsValue(p.StartPage, 1),
		Size:   gocast.PtrAsValue(p.Size, 0),
	}
}

// Information for paginating
type PageInfo struct {
	// When paginating backwards, the cursor to continue.
	StartCursor string `json:"startCursor"`
	// When paginating forwards, the cursor to continue.
	EndCursor string `json:"endCursor"`
	// When paginating backwards, are there more items?
	HasPreviousPage bool `json:"hasPreviousPage"`
	// When paginating forwards, are there more items?
	HasNextPage bool `json:"hasNextPage"`
	// Total number of pages available
	Total int `json:"total"`
	// Current page number
	Page int `json:"page"`
	// Number of pages
	Count int `json:"count"`
}
