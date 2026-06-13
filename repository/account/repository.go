// Package account present full API functionality of the specific object
package account

import (
	"context"
)

// Repository of access to the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository interface {
	Get(ctx context.Context, id uint64) (*Account, error)
	FetchList(ctx context.Context, filter *Filter, order *ListOrder, p *Pagination) ([]*Account, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Create(ctx context.Context, account *Account) (uint64, error)
	Update(ctx context.Context, id uint64, account *Account) error
	Delete(ctx context.Context, id uint64) error

	LoadPermissions(ctx context.Context, account *Account, user *User) error

	// GetByToken returns the user and account objects linked to the token (external session ID)
	GetByToken(ctx context.Context, token string) (*User, *Account, error)
}

// MemberRepository of access to the account members
type MemberRepository interface {
	IsAdmin(ctx context.Context, userID, accountID uint64) bool
	IsMember(ctx context.Context, userID, accountID uint64) bool

	FetchListMembers(ctx context.Context, filter *MemberFilter, order *MemberListOrder, p *Pagination) ([]*AccountMember, error)
	CountMembers(ctx context.Context, filter *MemberFilter) (int64, error)
	Member(ctx context.Context, userID, accountID uint64) (*AccountMember, error)
	MemberByID(ctx context.Context, id uint64) (*AccountMember, error)
	LinkMember(ctx context.Context, account *Account, isAdmin bool, members ...*User) error
	UnlinkMember(ctx context.Context, account *Account, members ...*User) error
	SetMemberRoles(ctx context.Context, account *Account, member *User, roles ...string) error
}
