// Package user present full API functionality of the specific object
package user

import (
	"context"
)

// Repository describes basic user methods
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPassword(ctx context.Context, email, password string) (*User, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*User, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)

	Create(ctx context.Context, user *User, password string) (uint64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error

	SetPassword(ctx context.Context, user *User, password string) error

	CreateResetPassword(ctx context.Context, userID uint64) (*UserPasswordReset, error)
	GetResetPassword(ctx context.Context, userID uint64, token string) (*UserPasswordReset, error)
	EliminateResetPassword(ctx context.Context, userID uint64) error
}
