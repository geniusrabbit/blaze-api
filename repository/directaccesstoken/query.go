package directaccesstoken

import (
	"time"

	"github.com/geniusrabbit/blaze-api/model"
	"gorm.io/gorm"
)

type Filter struct {
	ID           []uint64
	Token        []string
	UserID       []uint64
	AccountID    []uint64
	MinExpiresAt time.Time
	MaxExpiresAt time.Time
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.ID) > 0 {
		query = query.Where(`id IN (?)`, fl.ID)
	}
	if len(fl.Token) > 0 {
		query = query.Where(`token IN (?)`, fl.Token)
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, fl.UserID)
	}
	if len(fl.AccountID) > 0 {
		query = query.Where(`account_id IN (?)`, fl.AccountID)
	}
	if !fl.MinExpiresAt.IsZero() {
		query = query.Where(`expires_at >= ?`, fl.MinExpiresAt)
	}
	if !fl.MaxExpiresAt.IsZero() {
		query = query.Where(`expires_at <= ?`, fl.MaxExpiresAt)
	}
	return query
}

type Order struct {
	ID        model.Order
	Token     model.Order
	UserID    model.Order
	AccountID model.Order
	CreatedAt model.Order
	ExpiresAt model.Order
}

func (ord *Order) PrepareQuery(query *gorm.DB) *gorm.DB {
	if ord == nil {
		return query
	}
	query = ord.ID.PrepareQuery(query, "id")
	query = ord.Token.PrepareQuery(query, "token")
	query = ord.UserID.PrepareQuery(query, "user_id")
	query = ord.AccountID.PrepareQuery(query, "account_id")
	query = ord.CreatedAt.PrepareQuery(query, "created_at")
	query = ord.ExpiresAt.PrepareQuery(query, "expires_at")
	return query
}
