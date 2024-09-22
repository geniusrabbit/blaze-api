package repository

import (
	"strings"

	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/demdxx/gocast/v2"
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

// PrepareAfterQuery prepare query with pagination
// Requered gorm plugin to support with clause
//
//	 @plugin extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
//		db.Use(extraClausePlugin.New())
func (p *Pagination) PrepareAfterQuery(q *gorm.DB, idCol string, orderColumns []OrderingColumn) *gorm.DB {
	if p == nil || p.After == "" || len(orderColumns) == 0 {
		return q
	}
	order := strings.Join(xtypes.SliceApply(orderColumns,
		func(c OrderingColumn) string { return c.Name + gocast.IfThen(c.DESC, ` DESC`, ``) }), ", ")

	cte := q.Session(&gorm.Session{}).Select(`*, ROW_NUMBER() OVER(ORDER BY ` + order + `) AS rn`).Limit(-1)
	cteAfter := `SELECT rn FROM ctepageall WHERE ` + idCol + ` = '` + p.After + `'`
	q = q.Clauses(
		exclause.NewWith("ctepageall", cte),
		exclause.NewWith("ctepage1", cteAfter),
	)
	q = q.Table("ctepageall")
	q = q.Where("rn > (SELECT rn FROM ctepage1)")
	return q
}
