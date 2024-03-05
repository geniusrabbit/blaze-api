package account

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Usecase of the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*model.Account, error)
	GetByTitle(ctx context.Context, title string) (*model.Account, error)
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, pagination *repository.Pagination) ([]*model.Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Store(ctx context.Context, account *model.Account) (uint64, error)
	Register(ctx context.Context, ownerObj *model.User, accountObj *model.Account, password string) (uint64, error)
	Delete(ctx context.Context, id uint64) error
	FetchMembers(ctx context.Context, account *model.Account) ([]*model.AccountMember, error)
	FetchMemberUsers(ctx context.Context, account *model.Account) ([]*model.AccountMember, []*model.User, error)
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
}
