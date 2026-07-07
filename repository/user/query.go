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

// ---------------------------------------------------------------------------
// Filter trait parts
// ---------------------------------------------------------------------------

// FilterBase contains the core ID filter (Model trait only).
// Embed into your custom filter struct; call Apply in PrepareQuery.
type FilterBase struct {
	ID []uint64
}

// Apply applies the ID filter to the query.
func (fl *FilterBase) Apply(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.ID) > 0 {
		q = q.Where(`id IN (?)`, fl.ID)
	}
	return q
}

// FilterEmail contains the email filter (Email trait only).
// Embed into your custom filter struct; call Apply in PrepareQuery.
type FilterEmail struct {
	Emails []string
}

// Apply applies the email filter to the query.
func (fl *FilterEmail) Apply(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.Emails) > 0 {
		q = q.Where(`lower(email) IN (?)`, xtypes.SliceApply(fl.Emails, strings.ToLower))
	}
	return q
}

// ---------------------------------------------------------------------------
// Order trait parts
// ---------------------------------------------------------------------------

// OrderBase contains core order fields (Model trait only).
// Embed into your custom order struct; call Apply in PrepareQuery.
type OrderBase struct {
	ID        Order
	Status    Order
	CreatedAt Order
	UpdatedAt Order
}

// Apply applies the base order fields to the query.
func (ord *OrderBase) Apply(q *gorm.DB) *gorm.DB {
	if ord == nil {
		return q
	}
	q = ord.ID.PrepareQuery(q, "id")
	q = ord.Status.PrepareQuery(q, "approve_status")
	q = ord.CreatedAt.PrepareQuery(q, "created_at")
	q = ord.UpdatedAt.PrepareQuery(q, "updated_at")
	return q
}

// OrderEmail contains the email order field (Email trait only).
type OrderEmail struct {
	Email Order
}

// Apply applies the email order to the query.
func (ord *OrderEmail) Apply(q *gorm.DB) *gorm.DB {
	if ord == nil {
		return q
	}
	q = ord.Email.PrepareQuery(q, "email")
	return q
}

// OrderUsername contains the username order field (Username trait only).
type OrderUsername struct {
	Username Order
}

// Apply applies the username order to the query.
func (ord *OrderUsername) Apply(q *gorm.DB) *gorm.DB {
	if ord == nil {
		return q
	}
	q = ord.Username.PrepareQuery(q, "username")
	return q
}

// ---------------------------------------------------------------------------
// Backward-compatible composite filter / order
// ---------------------------------------------------------------------------

// ListFilter is the backward-compatible composite filter (FilterBase + FilterEmail).
// Consumer projects should compose FilterBase + FilterEmail (+ custom fields) directly.
type ListFilter struct {
	FilterBase
	FilterEmail
}

// PrepareQuery applies all embedded filter parts.
func (fl *ListFilter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	q = fl.FilterBase.Apply(q)
	q = fl.FilterEmail.Apply(q)
	return q
}

// ListOrder is the backward-compatible composite order (OrderBase + OrderEmail).
// Consumer projects should compose OrderBase + OrderEmail (+ custom fields) directly.
type ListOrder struct {
	OrderBase
	OrderEmail
}

// PrepareQuery applies all embedded order parts.
func (ord *ListOrder) PrepareQuery(q *gorm.DB) *gorm.DB {
	if ord == nil {
		return q
	}
	q = ord.OrderBase.Apply(q)
	q = ord.OrderEmail.Apply(q)
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
