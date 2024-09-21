package models

import (
	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/repository"
)

func (p *Page) Pagination() *repository.Pagination {
	if p == nil {
		return nil
	}
	return &repository.Pagination{
		After: gocast.PtrAsValue(p.After, ""),
		Page:  gocast.PtrAsValue(p.StartPage, 1),
		Size:  gocast.PtrAsValue(p.Size, 0),
	}
}
