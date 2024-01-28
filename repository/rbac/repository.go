// Package account present full API functionality of the specific object
package rbac

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	ID    []uint64
	Names []string
	Types []model.RoleType
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl != nil {
		if len(fl.Types) > 0 {
			query = query.Where(`type IN (?)`, fl.Types)
		}
		if len(fl.Names) > 0 {
			query = query.Where(`name IN (?)`, fl.Names)
		}
		if len(fl.ID) > 0 {
			query = query.Where(`id IN (?)`, fl.ID)
		}
	}
	return query
}

// Order of the objects list
type Order struct {
	ID       model.Order
	ParentID model.Order
	Name     model.Order
	Title    model.Order
	Type     model.Order
}

func (o *Order) PrepareQuery(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = o.ID.PrepareQuery(query, `id`)
	query = o.ParentID.PrepareQuery(query, `parent_id`)
	query = o.Name.PrepareQuery(query, `name`)
	query = o.Title.PrepareQuery(query, `title`)
	query = o.Type.PrepareQuery(query, `type`)
	return query
}

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, pagination *repository.Pagination) ([]*model.Role, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, role *model.Role) (uint64, error)
	Update(ctx context.Context, id uint64, role *model.Role) error
	Delete(ctx context.Context, id uint64) error
}
