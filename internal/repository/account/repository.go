// Package account present full API functionality of the specific object
package account

import (
	"context"

	"github.com/geniusrabbit/api-template-base/model"
)

// Filter of the objects list
type Filter struct {
	ID       []uint64
	UserID   []uint64
	Page     int
	PageSize int
}

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.Account, error)
	GetByTitle(ctx context.Context, title string) (*model.Account, error)
	LoadPermissions(ctx context.Context, account *model.Account, user *model.User) error
	FetchList(ctx context.Context, filter *Filter) ([]*model.Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, account *model.Account) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.Account) error
	Delete(ctx context.Context, id uint64) error
	FetchMembers(ctx context.Context, account *model.Account) ([]*model.AccountMember, error)
	IsMember(ctx context.Context, user *model.User, account *model.Account) bool
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
}
