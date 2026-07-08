// Package user present full API functionality of the specific object
package user

import (
	"context"
)

// Repository is the core CRUD repository parameterized by user model type.
// T must be a pointer type implementing Model (e.g. *MyUser).
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository[T Model] interface {
	EmptyObject() T
	Get(ctx context.Context, id uint64) (T, error)
	FetchList(ctx context.Context, opts ...QOption) ([]T, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)
	Create(ctx context.Context, user T) (uint64, error)
	Update(ctx context.Context, user T) error
	Delete(ctx context.Context, id uint64) error
}

// EmailRepository provides email lookup for models with UserEmail trait.
type EmailRepository[T EmailCapableModel] interface {
	GetByEmail(ctx context.Context, email string) (T, error)
}

// PasswordRepository provides password auth and reset token operations.
type PasswordRepository[T PasswordCapableModel] interface {
	GetByPassword(ctx context.Context, userID uint64, password string) (T, error)
	CreateWithPassword(ctx context.Context, user T, password string) (uint64, error)
	SetPassword(ctx context.Context, user T, password string) error
	CreateResetPassword(ctx context.Context, userID uint64) (*UserPasswordReset, error)
	GetResetPassword(ctx context.Context, userID uint64, token string) (*UserPasswordReset, error)
	EliminateResetPassword(ctx context.Context, userID uint64) error
}
