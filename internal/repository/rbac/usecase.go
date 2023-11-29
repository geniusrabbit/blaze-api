package rbac

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository"
	"github.com/geniusrabbit/api-template-base/model"
)

// Usecase of the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.Role, error)
	GetByName(ctx context.Context, title string) (*model.Role, error)
	FetchList(ctx context.Context, filter *Filter, pagination *repository.Pagination) ([]*model.Role, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, role *model.Role) (uint64, error)
	Update(ctx context.Context, id uint64, role *model.Role) error
	Delete(ctx context.Context, id uint64) error
}
