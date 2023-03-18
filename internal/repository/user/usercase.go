package user

import (
	"context"

	"github.com/geniusrabbit/api-template-base/model"
)

// Usecase describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error)
	FetchList(ctx context.Context, accountID uint64, page, num int) ([]*model.User, error)
	Count(ctx context.Context, accountID uint64) (int64, error)
	SetPassword(ctx context.Context, user *model.User, password string) error
	Store(ctx context.Context, user *model.User, password string) (uint64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error
}
