// Package account present full API functionality of the specific object
package historylog

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Filter of the objects list
type Filter struct {
	ID          []uuid.UUID
	Name        []string
	UserID      []uint64
	AccountID   []uint64
	ObjectID    []uint64
	ObjectIDStr []string
	ObjectType  []string
}

func (filter *Filter) Query(query *gorm.DB) *gorm.DB {
	if filter == nil {
		return query
	}
	if len(filter.ID) > 0 {
		query = query.Where(`id IN (?)`, filter.ID)
	}
	if len(filter.Name) > 0 {
		query = query.Where(`name IN (?)`, filter.Name)
	}
	if len(filter.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, filter.UserID)
	}
	if len(filter.AccountID) > 0 {
		query = query.Where(`account_id IN (?)`, filter.AccountID)
	}
	if len(filter.ObjectID) > 0 {
		query = query.Where(`object_id IN (?)`, filter.ObjectID)
	}
	if len(filter.ObjectIDStr) > 0 {
		query = query.Where(`object_ids IN (?)`, filter.ObjectIDStr)
	}
	if len(filter.ObjectType) > 0 {
		query = query.Where(`object_type IN (?)`, filter.ObjectType)
	}
	return query
}

type Order struct {
	ID          int8
	Name        int8
	UserID      int8
	AccountID   int8
	ObjectID    int8
	ObjectIDStr int8
	ObjectType  int8
	ActionAt    int8
}

func (o *Order) Query(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = orderS(query, `id`, o.ID)
	query = orderS(query, `name`, o.Name)
	query = orderS(query, `user_id`, o.UserID)
	query = orderS(query, `account_id`, o.AccountID)
	query = orderS(query, `object_id`, o.ObjectID)
	query = orderS(query, `object_ids`, o.ObjectIDStr)
	query = orderS(query, `object_type`, o.ObjectType)
	query = orderS(query, `action_at`, o.ActionAt)
	return query
}

func orderS(query *gorm.DB, n string, o int8) *gorm.DB {
	if o > 0 {
		return query.Order(n + ` ASC`)
	} else if o < 0 {
		return query.Order(n + ` DESC`)
	}
	return query
}

// Repository of access to the changelog
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Count(ctx context.Context, filter *Filter) (int64, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, pagination *repository.Pagination) ([]*model.HistoryAction, error)
}
