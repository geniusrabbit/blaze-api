package authclient

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
)

// Usecase of the AuthAclient
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id string) (*model.AuthClient, error)
	FetchList(ctx context.Context, filter *Filter) ([]*model.AuthClient, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, authClient *model.AuthClient) (string, error)
	Update(ctx context.Context, id string, authClient *model.AuthClient) error
	Delete(ctx context.Context, id string) error
}
