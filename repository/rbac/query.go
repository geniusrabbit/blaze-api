package rbac

import (
	"github.com/geniusrabbit/blaze-api/model"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	ID             []uint64
	Names          []string
	MinAccessLevel int
	MaxAccessLevel int
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
	// Here we can have logic error but it's not critical. Example `access_level BETWEEN 2 AND 1`
	if fl.MinAccessLevel > 0 && fl.MaxAccessLevel > 0 {
		query = query.Where(`access_level BETWEEN ? AND ?`, fl.MinAccessLevel, fl.MaxAccessLevel)
	} else if fl.MinAccessLevel > 0 {
		query = query.Where(`access_level >= ?`, fl.MinAccessLevel)
	} else if fl.MaxAccessLevel > 0 {
		query = query.Where(`access_level <= ?`, fl.MaxAccessLevel)
	}
	return query
}

// Order of the objects list
type Order struct {
	ID          model.Order
	Name        model.Order
	Title       model.Order
	AccessLevel model.Order
	CreatedAt   model.Order
	UpdatedAt   model.Order
}

func (o *Order) PrepareQuery(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = o.ID.PrepareQuery(query, `id`)
	query = o.Name.PrepareQuery(query, `name`)
	query = o.Title.PrepareQuery(query, `title`)
	query = o.AccessLevel.PrepareQuery(query, `access_level`)
	query = o.CreatedAt.PrepareQuery(query, `created_at`)
	query = o.UpdatedAt.PrepareQuery(query, `updated_at`)
	return query
}
