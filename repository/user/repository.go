// Package user present full API functionality of the specific object
package user

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Repository describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error)
	FetchList(ctx context.Context, filter *ListFilter, order *ListOrder, page *repository.Pagination) ([]*model.User, error)
	Count(ctx context.Context, filter *ListFilter) (int64, error)
	Create(ctx context.Context, user *model.User, password string) (uint64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error

	SetPassword(ctx context.Context, user *model.User, password string) error
	CreateResetPassword(ctx context.Context, userID uint64) (*model.UserPasswordReset, error)
	GetResetPassword(ctx context.Context, userID uint64, token string) (*model.UserPasswordReset, error)
	EliminateResetPassword(ctx context.Context, userID uint64) error
}
