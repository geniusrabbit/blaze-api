package rbac

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/generated"
)

// Usecase of the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	generated.UsecaseIface[Role, uint64]
	GetByName(ctx context.Context, title string) (*Role, error)
}
