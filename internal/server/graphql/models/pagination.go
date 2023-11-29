package models

import "github.com/geniusrabbit/api-template-base/internal/repository"

func (p *Page) Pagination() *repository.Pagination {
	if p == nil {
		return nil
	}
	return &repository.Pagination{
		Page: valOrDef(p.StartPage, 1),
		Size: valOrDef(p.Size, 0),
	}
}

func valOrDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}
