package user

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Usecase describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error)
	FetchList(ctx context.Context, filter *ListFilter, order *ListOrder, page *repository.Pagination) ([]*model.User, error)
	Count(ctx context.Context, filter *ListFilter) (int64, error)
	Store(ctx context.Context, user *model.User, password string) (uint64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error

	SetPassword(ctx context.Context, user *model.User, password string) error
	ResetPassword(ctx context.Context, email string) (*model.UserPasswordReset, *model.User, error)
	UpdatePassword(ctx context.Context, token, email, password string) error
}
