package socialaccount

import (
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
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
	ID        models.Order
	UserID    models.Order
	Provider  models.Order
	Email     models.Order
	Username  models.Order
	FirstName models.Order
	LastName  models.Order
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

// Type aliases for common repository types.
type (
	Pagination  = repository.Pagination
	QOption     = repository.QOption
	ListOptions = repository.ListOptions
)
