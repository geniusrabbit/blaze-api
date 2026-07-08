package account

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Usecase of the account
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/usecase.go
type Usecase[TUser user.Model, TAccount Model] interface {
	// EmptyObject returns a new empty account object of type TAccount.
	EmptyObject() TAccount

	// Get retrieves an account by ID.
	Get(ctx context.Context, id uint64) (TAccount, error)

	// FetchList retrieves a list of accounts based on the provided query options.
	FetchList(ctx context.Context, opts ...QOption) ([]TAccount, error)

	// Count returns the total number of accounts based on the provided query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Update modifies an existing account.
	Update(ctx context.Context, account TAccount) (uint64, error)

	// Register creates a new account with the specified owner and account details.
	Register(ctx context.Context, ownerObj TUser, accountObj TAccount) (uint64, error)

	// Delete removes an account by ID.
	Delete(ctx context.Context, id uint64) error
}

// MemberUsecase of the account members
type MemberUsecase[TUser user.Model, TAccount Model] interface {
	// EmptyObject returns a new empty member object of type Member[TUser, TAccount].
	EmptyObject() *Member[TUser, TAccount]

	// FetchListMembers retrieves a list of members based on the provided query options.
	FetchListMembers(ctx context.Context, opts ...QOption) ([]*Member[TUser, TAccount], error)

	// CountMembers returns the total number of members based on the provided query options.
	CountMembers(ctx context.Context, opts ...QOption) (int64, error)

	// LinkMember associates a user with an account, optionally granting admin privileges.
	LinkMember(ctx context.Context, account TAccount, isAdmin bool, members ...TUser) error

	// UnlinkMember removes the association between a user and an account.
	UnlinkMember(ctx context.Context, account TAccount, members ...TUser) error

	// UnlinkAccountMember removes a member from an account based on the member's ID.
	UnlinkAccountMember(ctx context.Context, memberID uint64) error

	// InviteMember sends an invitation to a user to join an account with specified roles.
	InviteMember(ctx context.Context, accountID, userID uint64, roles ...string) (*Member[TUser, TAccount], error)

	// SetAccountMemeberRoles sets the roles for a user within a specific account.
	SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*Member[TUser, TAccount], error)

	// SetMemberRoles sets the roles for a member based on the member's ID.
	SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*Member[TUser, TAccount], error)
}
