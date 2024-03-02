package socialaccount

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.AccountSocial, error)
	FetchList(ctx context.Context, filter *Filter, order *Order, page *repository.Pagination) ([]*model.AccountSocial, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Disconnect(ctx context.Context, id uint64) (*model.AccountSocial, error)
	FetchSessionList(ctx context.Context, socialAccountID []uint64) ([]*model.AccountSocialSession, error)
}
