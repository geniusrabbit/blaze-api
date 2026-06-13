// Package account present full API functionality of the specific object
package historylog

import (
	"context"

	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
)

// Repository of the history actions log
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Count(ctx context.Context, opts ...QOption) (int64, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*historylogModels.HistoryAction, error)
}
