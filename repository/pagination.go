package repository

import "gorm.io/gorm"

// Pagination of the objects list
type Pagination struct {
	Offset int
	Page   int
	Size   int
}

// PrepareQuery prepare query with pagination
func (p *Pagination) PrepareQuery(q *gorm.DB) *gorm.DB {
	if p == nil {
		return q
	}
	if p.Page > 1 && p.Offset <= 0 {
		p.Offset = (p.Page - 1) * p.Size
	}
	if p.Size <= 0 {
		p.Size = 10
	}
	if p.Offset > 0 {
		q = q.Offset(p.Offset)
	}
	return q.Limit(p.Size)
}
