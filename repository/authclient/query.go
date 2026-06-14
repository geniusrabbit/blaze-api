package authclient

import (
	"github.com/geniusrabbit/blaze-api/repository"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	ID []string
}

// PrepareQuery prepares the GORM query based on the filter fields.
func (f *Filter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if f == nil {
		return q
	}
	if len(f.ID) > 0 {
		q = q.Where(`id IN (?)`, f.ID)
	}
	return q
}

// Type aliases for common repository types.
type (
	Pagination  = repository.Pagination
	QOption     = repository.QOption
	ListOptions = repository.ListOptions
)
