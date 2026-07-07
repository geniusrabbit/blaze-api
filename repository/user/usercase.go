package user

import (
	"context"
)

// Usecase is core user business logic parameterized by model type.
// T must be a pointer type (e.g. *MyUser).
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase[T Model] interface {
	EmptyObject() T
	Get(ctx context.Context, id uint64) (T, error)
	FetchList(ctx context.Context, opts ...QOption) ([]T, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)
	Create(ctx context.Context, user T) (uint64, error)
	Update(ctx context.Context, user T) error
	Delete(ctx context.Context, id uint64) error
}

// EmailUsecase provides email lookup operations.
type EmailUsecase[T EmailCapableModel] interface {
	GetByEmail(ctx context.Context, email string) (T, error)
}

// PasswordUsecase provides password management operations (no email lookup).
type PasswordUsecase[T PasswordCapableModel] interface {
	Repo() PasswordRepository[T]
	SetPassword(ctx context.Context, user T, password string) error
	ChangePassword(ctx context.Context, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, userID uint64) (*UserPasswordReset, T, error)
	UpdatePassword(ctx context.Context, userID uint64, token, password string) error
}
