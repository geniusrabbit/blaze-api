package option

import (
	"bytes"

	"github.com/geniusrabbit/blaze-api/model"
	"gorm.io/gorm"
)

// Filter of the objects list
type Filter struct {
	Type        []model.OptionType
	TargetID    []uint64
	Name        []string
	NamePattern []string
}

// PrepareQuery returns the query with applied filters
func (fl *Filter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.Type) > 0 {
		q = q.Where(`type IN (?)`, fl.Type)
	}
	if len(fl.TargetID) > 0 {
		q = q.Where(`target_id IN (?)`, fl.TargetID)
	}
	if len(fl.Name) > 0 {
		q = q.Where(`name IN (?)`, fl.Name)
	}
	if len(fl.NamePattern) > 0 {
		var (
			qbuf   bytes.Buffer
			params []any
		)
		for i, pattern := range fl.NamePattern {
			if i > 0 {
				qbuf.WriteString(` OR `)
			}
			qbuf.WriteString(`name LIKE ?`)
			params = append(params, pattern)
		}
		q = q.Where(qbuf.String(), params...)
	}
	return q
}

// ListOrder object with order values which is not NULL
type ListOrder struct {
	Name      model.Order
	Type      model.Order
	TargetID  model.Order
	CreatedAt model.Order
	UpdatedAt model.Order
}

// PrepareQuery returns the query with applied order
func (ord *ListOrder) PrepareQuery(q *gorm.DB) *gorm.DB {
	if ord != nil {
		q = ord.Name.PrepareQuery(q, `name`)
		q = ord.Type.PrepareQuery(q, `type`)
		q = ord.TargetID.PrepareQuery(q, `target_id`)
		q = ord.CreatedAt.PrepareQuery(q, `created_at`)
		q = ord.UpdatedAt.PrepareQuery(q, `updated_at`)
	}
	return q
}
