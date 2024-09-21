package repository

import (
	"reflect"

	"github.com/demdxx/xtypes"
	"gorm.io/gorm"
)

// QOption prepare query
type QOption interface {
	PrepareQuery(query *gorm.DB) *gorm.DB
}

type OrderingColumnsOption interface {
	OrderingColumns() []OrderingColumn
}

// AfterOption prepare query after
type AfterOption interface {
	PrepareAfterQuery(query *gorm.DB, idCol string, orderColumns []OrderingColumn) *gorm.DB
}

type PreloadOption struct {
	Fields []string
}

func (opt *PreloadOption) PrepareQuery(query *gorm.DB) *gorm.DB {
	if opt == nil {
		return query
	}
	for _, preload := range opt.Fields {
		query = query.Preload(preload)
	}
	return query
}

type GroupOption struct {
	Groups        []string
	SummingFields []string
}

func (opt *GroupOption) PrepareQuery(query *gorm.DB) *gorm.DB {
	if opt == nil || len(opt.Groups) == 0 {
		return query
	}
	for _, group := range opt.Groups {
		query = query.Group(group)
	}
	return query.Select(
		append(opt.Groups,
			xtypes.SliceApply(opt.SummingFields, func(field string) string {
				return "SUM(" + field + ") as " + field
			})...,
		),
	)
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

func (opts ListOptions) PrepareAfterQuery(query *gorm.DB, idCol string, orderColumns ...OrderingColumn) *gorm.DB {
	if len(orderColumns) == 0 {
		for _, opt := range opts {
			if orderingOpt, ok := opt.(OrderingColumnsOption); ok {
				orderColumns = orderingOpt.OrderingColumns()
				break
			}
		}
	}
	for _, opt := range opts {
		if afterOpt, ok := opt.(AfterOption); ok {
			return afterOpt.PrepareAfterQuery(query, idCol, orderColumns)
		}
	}
	return query
}
