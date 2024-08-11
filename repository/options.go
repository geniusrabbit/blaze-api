package repository

import (
	"reflect"

	"gorm.io/gorm"
)

// QOption prepare query
type QOption interface {
	PrepareQuery(query *gorm.DB) *gorm.DB
}

// ListOptions for query preparation
type ListOptions []QOption

func (opts ListOptions) With(prep QOption) ListOptions {
	updated := false
	for i, opt := range opts {
		if reflect.TypeOf(opt) == reflect.TypeOf(prep) {
			// replace the existing option
			updated = true
			opts[i] = prep
			break
		}
	}
	if !updated {
		opts = append(opts, prep)
	}
	return opts
}

func (opts ListOptions) PrepareQuery(query *gorm.DB) *gorm.DB {
	for _, opt := range opts {
		query = opt.PrepareQuery(query)
	}
	return query
}
