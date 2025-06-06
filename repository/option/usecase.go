package option

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Usecase of the Option
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, name string, otype model.OptionType, targetID uint64) (*model.Option, error)
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, pagination *repository.Pagination) ([]*model.Option, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Set(ctx context.Context, opt *model.Option) error
	SetOption(ctx context.Context, name string, otype model.OptionType, targetID uint64, value any) error
	Delete(ctx context.Context, name string, otype model.OptionType, targetID uint64) error
}
