// Package account present full API functionality of the specific object
package rbac

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/generated"
)

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	generated.RepositoryIface[Role, uint64]
	GetByName(ctx context.Context, name string) (*Role, error)
}
