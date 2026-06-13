package account

import (
	"context"
)

// Usecase of the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase interface {
	Get(ctx context.Context, id uint64) (*Account, error)
	FetchList(ctx context.Context, opts ...QOption) ([]*Account, error)
	Count(ctx context.Context, opts ...QOption) (int64, error)
	Store(ctx context.Context, account *Account) (uint64, error)
	Register(ctx context.Context, ownerObj *User, accountObj *Account, password string) (uint64, error)
	Delete(ctx context.Context, id uint64) error
}

// MemberUsecase of the account members
type MemberUsecase interface {
	FetchListMembers(ctx context.Context, opts ...QOption) ([]*AccountMember, error)
	CountMembers(ctx context.Context, opts ...QOption) (int64, error)
	LinkMember(ctx context.Context, account *Account, isAdmin bool, members ...*User) error
	UnlinkMember(ctx context.Context, account *Account, members ...*User) error
	UnlinkAccountMember(ctx context.Context, memberID uint64) error
	InviteMember(ctx context.Context, accountID uint64, email string, roles ...string) (*AccountMember, error)
	SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*AccountMember, error)
	SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*AccountMember, error)
}
