// Package account present full API functionality of the specific object
package authclient

import (
	"context"

	"github.com/geniusrabbit/api-template-base/model"
)

// Filter of the objects list
type Filter struct {
	ID       []string
	Page     int
	PageSize int
}

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id string) (*model.AuthClient, error)
	FetchList(ctx context.Context, filter *Filter) ([]*model.AuthClient, error)
	Create(ctx context.Context, authClient *model.AuthClient) (string, error)
	Update(ctx context.Context, id string, authClient *model.AuthClient) error
	Delete(ctx context.Context, id string) error
}
