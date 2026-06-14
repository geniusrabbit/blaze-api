package user

import (
	"strings"

	"github.com/demdxx/xtypes"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Order Ascending or Descending for query fields
type Order = models.Order

// ListFilter object with filtered values which is not NULL
type ListFilter struct {
	UserID []uint64
	Emails []string
}

// PrepareQuery returns the query with applied filters
func (fl *ListFilter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.UserID) > 0 {
		q = q.Where(`id IN (?)`, fl.UserID)
	}
	if len(fl.Emails) > 0 {
		q = q.Where(`lower(email) IN (?)`, xtypes.SliceApply(fl.Emails, strings.ToLower))
	}
	return q
}

// ListOrder object with order values which is not NULL
type ListOrder struct {
	ID        Order
	Email     Order
	Status    Order
	CreatedAt Order
	UpdatedAt Order
}

// PrepareQuery returns the query with applied order
func (ord *ListOrder) PrepareQuery(q *gorm.DB) *gorm.DB {
	if ord == nil {
		return q
	}
	q = ord.ID.PrepareQuery(q, "id")
	q = ord.Email.PrepareQuery(q, "email")
	q = ord.Status.PrepareQuery(q, "approve_status")
	q = ord.CreatedAt.PrepareQuery(q, "created_at")
	q = ord.UpdatedAt.PrepareQuery(q, "updated_at")
	return q
}

type (
	// Pagination is the pagination object
	Pagination = repository.Pagination

	// QOption is the query option interface
	QOption = repository.QOption

	// ListOptions is the list options struct
	ListOptions = repository.ListOptions
)
