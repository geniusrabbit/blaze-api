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

	FetchListMembers(ctx context.Context, filter *MemberFilter, order *MemberListOrder, pagination *repository.Pagination) ([]*model.AccountMember, error)
	CountMembers(ctx context.Context, filter *MemberFilter) (int64, error)
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
	UnlinkAccountMember(ctx context.Context, memberID uint64) error
	InviteMember(ctx context.Context, accountID uint64, email string, roles ...string) (*model.AccountMember, error)
	SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*model.AccountMember, error)
	SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*model.AccountMember, error)
}
