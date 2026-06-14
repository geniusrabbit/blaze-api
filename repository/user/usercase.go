package user

import (
	"context"
)

// Usecase describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPassword(ctx context.Context, email, password string) (*User, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*User, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)

	Create(ctx context.Context, user *User, password string) (uint64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error

	SetPassword(ctx context.Context, user *User, password string) error
	ResetPassword(ctx context.Context, email string) (*UserPasswordReset, *User, error)
	UpdatePassword(ctx context.Context, token, email, password string) error
}
