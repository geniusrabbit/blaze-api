// Package account present full API functionality of the specific object
package account

import (
	"context"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
)

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*model.Account, error)
	GetByTitle(ctx context.Context, title string) (*model.Account, error)
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, pagination *repository.Pagination) ([]*model.Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, account *model.Account) (uint64, error)
	Update(ctx context.Context, id uint64, account *model.Account) error
	Delete(ctx context.Context, id uint64) error

	IsAdmin(ctx context.Context, userID, accountID uint64) bool
	IsMember(ctx context.Context, userID, accountID uint64) bool

	FetchListMembers(ctx context.Context, filter *MemberFilter, order *MemberListOrder, pagination *repository.Pagination) ([]*model.AccountMember, error)
	CountMembers(ctx context.Context, filter *MemberFilter) (int64, error)
	Member(ctx context.Context, userID, accountID uint64) (*model.AccountMember, error)
	MemberByID(ctx context.Context, id uint64) (*model.AccountMember, error)
	LinkMember(ctx context.Context, account *model.Account, isAdmin bool, members ...*model.User) error
	UnlinkMember(ctx context.Context, account *model.Account, members ...*model.User) error
	SetMemberRoles(ctx context.Context, account *model.Account, member *model.User, roles ...string) error

	LoadPermissions(ctx context.Context, account *model.Account, user *model.User) error
}
