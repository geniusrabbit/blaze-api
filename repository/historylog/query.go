package historylog

import (
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	ID          []uuid.UUID
	RequestID   []string
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
	if len(filter.RequestID) > 0 {
		query = query.Where(`request_id IN (?)`, filter.RequestID)
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
	ID          model.Order
	RequestID   model.Order
	Name        model.Order
	UserID      model.Order
	AccountID   model.Order
	ObjectID    model.Order
	ObjectIDStr model.Order
	ObjectType  model.Order
	ActionAt    model.Order
}

func (o *Order) Query(query *gorm.DB) *gorm.DB {
	if o == nil {
		return query
	}
	query = o.ID.PrepareQuery(query, `id`)
	query = o.RequestID.PrepareQuery(query, `request_id`)
	query = o.Name.PrepareQuery(query, `name`)
	query = o.UserID.PrepareQuery(query, `user_id`)
	query = o.AccountID.PrepareQuery(query, `account_id`)
	query = o.ObjectID.PrepareQuery(query, `object_id`)
	query = o.ObjectIDStr.PrepareQuery(query, `object_ids`)
	query = o.ObjectType.PrepareQuery(query, `object_type`)
	query = o.ActionAt.PrepareQuery(query, `action_at`)
	return query
}
