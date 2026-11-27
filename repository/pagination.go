package repository

import (
	"errors"
	"regexp"
	"strings"

	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
	"github.com/demdxx/gocast/v2"
	"github.com/demdxx/xtypes"
	"gorm.io/gorm"
)

var (
	// DefaultPageSize is used when the page size is not set or not positive.
	DefaultPageSize = 10
	// MaxPageSize caps the page size requested by clients.
	MaxPageSize = 1000

	errInvalidPaginationColumn = errors.New("pagination: invalid column name")
	columnNameRe               = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)
)

// NormalizePageSize returns the page size limited to 1..MaxPageSize,
// DefaultPageSize for non-positive values.
func NormalizePageSize(size int) int {
	if size <= 0 {
		return DefaultPageSize
	}
	return min(size, MaxPageSize)
}

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
	p.Size = NormalizePageSize(p.Size)
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
	// Column names go into the SQL text, so only plain identifiers are allowed;
	// the cursor value is always a bound parameter.
	if !columnNameRe.MatchString(idCol) {
		_ = q.AddError(errInvalidPaginationColumn)
		return q
	}
	for _, c := range orderColumns {
		if !columnNameRe.MatchString(c.Name) {
			_ = q.AddError(errInvalidPaginationColumn)
			return q
		}
	}
	order := strings.Join(xtypes.SliceApply(orderColumns,
		func(c OrderingColumn) string { return c.Name + gocast.IfThen(c.DESC, ` DESC`, ``) }), ", ")

	cte := q.Session(&gorm.Session{}).Select(`*, ROW_NUMBER() OVER(ORDER BY ` + order + `) AS rn`).Limit(-1)
	q = q.Clauses(
		exclause.NewWith("ctepageall", cte),
		exclause.NewWith("ctepage1", `SELECT rn FROM ctepageall WHERE `+idCol+` = ?`, p.After),
	)
	q = q.Table("ctepageall")
	q = q.Where("rn > (SELECT rn FROM ctepage1)")
	return q
}
