// Package user present full API functionality of the specific object
package user

import (
	"context"

	"github.com/geniusrabbit/api-template-base/model"
)

// ListFilter object with filtered values which is not NULL
type ListFilter struct {
	AccountID []uint64
	UserID    []uint64
}

// Repository describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPassword(ctx context.Context, email, password string) (*model.User, error)
	GetByToken(ctx context.Context, token string) (*model.User, *model.Account, error)
	FetchList(ctx context.Context, filter *ListFilter, page, num int) ([]*model.User, error)
	Count(ctx context.Context, filter *ListFilter) (int64, error)
	SetPassword(ctx context.Context, user *model.User, password string) error
	Create(ctx context.Context, user *model.User, password string) (uint64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint64) error
}
