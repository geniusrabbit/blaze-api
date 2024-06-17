// Package option present full API functionality of the specific object
package option

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Repository of access to the option
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, name string, otype model.OptionType, targetID uint64) (*model.Option, error)
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, pagination *repository.Pagination) ([]*model.Option, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Set(ctx context.Context, opt *model.Option) error
	Delete(ctx context.Context, name string, otype model.OptionType, targetID uint64) error
}
