package historylog

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Usecase of the changelog
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Count(ctx context.Context, filter *Filter) (int64, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, pagination *repository.Pagination) ([]*model.HistoryAction, error)
}
