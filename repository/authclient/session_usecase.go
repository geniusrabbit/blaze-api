package authclient

import (
	"context"
)

// SessionUsecase defines business logic for OAuth auth_session rows.
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/session_usecase.go
type SessionUsecase interface {
	Get(ctx context.Context, id uint64, opts ...QOption) (*AuthSession, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*AuthSession, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)
	Delete(ctx context.Context, id uint64, opts ...QOption) error
}
