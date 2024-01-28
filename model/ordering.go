package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order int8

const (
	OrderUndefined Order = 0
	OrderAsc       Order = 1
	OrderDesc      Order = -1
)

// PrepareQuery returns the query with applied order
func OrderFromStr(s string) Order {
	switch s {
	case "ASC":
		return OrderAsc
	case "DESC":
		return OrderDesc
	}
	return OrderUndefined
}

// PrepareQuery returns the query with applied order
func (ord *Order) PrepareQuery(q *gorm.DB, column string) *gorm.DB {
	if ord == nil || *ord == 0 {
		return q
	}
	return q.Order(clause.OrderByColumn{Column: clause.Column{Name: column}, Desc: *ord == OrderDesc})
}

// Set sets the order value from string
func (ord *Order) Set(s string) *Order {
	switch s {
	case "ASC":
		*ord = OrderAsc
	case "DESC":
		*ord = OrderDesc
	default:
		*ord = 0
	}
	return ord
}
