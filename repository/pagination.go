package repository

import (
	"bytes"
	"slices"
	"strings"

	"github.com/demdxx/xtypes"
	"gorm.io/gorm"
)

type OrderingColumn struct {
	Name string
	DESC bool
}

// Pagination of the objects list
type Pagination struct {
	After  string
	Offset int
	Page   int
	Size   int
}

// PrepareQuery prepare query with pagination
func (p *Pagination) PrepareQuery(q *gorm.DB) *gorm.DB {
	if p == nil {
		return q
	}
	if p.Size <= 0 {
		p.Size = 10
	}
	if p.Page > 1 && p.Offset <= 0 {
		p.Offset = (p.Page - 1) * p.Size
	}
	if p.Offset > 0 {
		q = q.Offset(p.Offset)
	}
	if p.Size > 0 {
		q = q.Limit(p.Size)
	}
	return q
}

func (p *Pagination) PrepareAfterQuery(q *gorm.DB, idCol string, orderColumns []OrderingColumn) *gorm.DB {
	if p == nil || p.After == "" {
		return q
	}
	containsIDColumn := slices.ContainsFunc(orderColumns, func(c OrderingColumn) bool {
		return c.Name == idCol
	})
	// Prepare columns for order
	columns := strings.Join(
		xtypes.SliceApply(orderColumns, func(c OrderingColumn) string {
			if c.DESC {
				return "-" + c.Name
			}
			return c.Name
		}), ", ")
	// Query example:
	// (name, id) > (SELECT name, id FROM table WHERE id = 'id.value')
	query := bytes.Buffer{}
	_, _ = query.WriteString("(")
	_, _ = query.WriteString(columns)
	if !containsIDColumn {
		if len(orderColumns) > 1 {
			_, _ = query.WriteString(`, `)
		}
		_, _ = query.WriteString(idCol)
	}
	_, _ = query.WriteString(") > (")
	_, _ = query.WriteString(`SELECT `)
	_, _ = query.WriteString(columns)
	if !containsIDColumn {
		if len(orderColumns) > 1 {
			_, _ = query.WriteString(`, `)
		}
		_, _ = query.WriteString(idCol)
	}
	_, _ = query.WriteString(` FROM `)
	_, _ = query.WriteString(queryTable(q))
	_, _ = query.WriteString(` WHERE `)
	_, _ = query.WriteString(idCol)
	_, _ = query.WriteString(` = '`)
	_, _ = query.WriteString(p.After)
	_, _ = query.WriteString(`')`)
	q = q.Where(query.String())
	return q
}

func queryTable(q *gorm.DB) string {
	if q.Statement.Table != "" {
		return q.Statement.Table
	}
	if q.Statement.Model != nil {
		m, _ := q.Statement.Model.(interface{ TableName() string })
		if m != nil {
			return m.TableName()
		}
	}
	return q.Statement.Table
}
