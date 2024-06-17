package directaccesstoken

import (
	"context"
	"time"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.DirectAccessToken, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, page *repository.Pagination) ([]*model.DirectAccessToken, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Generate(ctx context.Context, userID, accountID uint64, description string, expiresAt time.Time) (*model.DirectAccessToken, error)
	Revoke(ctx context.Context, filter *Filter) error
}
