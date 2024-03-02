package socialaccount

import (
	"context"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

type Filter struct {
	ID       []uint64
	Provider []string
	Username []string
	Email    []string
	UserID   []uint64
}

func (f *Filter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if f == nil {
		return q
	}
	if len(f.ID) > 0 {
		q = q.Where(`id IN (?)`, f.ID)
	}
	if len(f.Provider) > 0 {
		q = q.Where(`provider IN (?)`, f.Provider)
	}
	if len(f.Username) > 0 {
		q = q.Where(`username IN (?)`, f.Username)
	}
	if len(f.Email) > 0 {
		q = q.Where(`email IN (?)`, f.Email)
	}
	if len(f.UserID) > 0 {
		q = q.Where(`user_id IN (?)`, f.UserID)
	}
	return q
}

type Order struct {
	ID        model.Order
	UserID    model.Order
	Provider  model.Order
	Email     model.Order
	Username  model.Order
	FirstName model.Order
	LastName  model.Order
}

func (o *Order) PrepareQuery(q *gorm.DB) *gorm.DB {
	if o == nil {
		return q
	}
	q = o.ID.PrepareQuery(q, `id`)
	q = o.UserID.PrepareQuery(q, `user_id`)
	q = o.Provider.PrepareQuery(q, `provider`)
	q = o.Email.PrepareQuery(q, `email`)
	q = o.Username.PrepareQuery(q, `username`)
	q = o.FirstName.PrepareQuery(q, `first_name`)
	q = o.LastName.PrepareQuery(q, `last_name`)
	return q

}

type Repository interface {
	Get(ctx context.Context, id uint64) (*model.AccountSocial, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, pagination *repository.Pagination) ([]*model.AccountSocial, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Disconnect(ctx context.Context, id uint64) error
	FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*model.AccountSocialSession, error)
}
