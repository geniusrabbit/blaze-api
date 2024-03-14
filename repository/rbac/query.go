package rbac

import (
	"github.com/geniusrabbit/blaze-api/model"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	ID    []uint64
	Names []string
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.Names) > 0 {
		query = query.Where(`name IN (?)`, fl.Names)
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	return query
}

// Order of the objects list
type Order struct {
	ID    model.Order
	Name  model.Order
	Title model.Order
}

func (o *Order) PrepareQuery(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = o.ID.PrepareQuery(query, `id`)
	query = o.Name.PrepareQuery(query, `name`)
	query = o.Title.PrepareQuery(query, `title`)
	return query
}
