package historylog

import (
	"context"

	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
)

// Usecase of the changelog
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Count(ctx context.Context, opts ...QOption) (int64, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*historylogModels.HistoryAction, error)
}
