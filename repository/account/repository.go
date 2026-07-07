package account

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/user"
)

// Repository is the core account CRUD repository parameterized by account model type.
// T must be a pointer type implementing Model (e.g. consumer-defined Account struct).
//
//go:generate mockgen -source $GOFILE -package mocks -destination mocks/repository.go
type Repository[T Model] interface {
	// EmptyObject returns a new empty account object of type T.
	EmptyObject() T

	// Get retrieves an account by ID.
	// Returns ErrNotFound if the account does not exist.
	Get(ctx context.Context, id uint64) (T, error)

	// FetchList retrieves a list of accounts based on the provided query options.
	// Returns an empty slice if no accounts match the criteria.
	FetchList(ctx context.Context, opts ...QOption) ([]T, error)

	// Count returns the number of accounts matching the provided query options.
	Count(ctx context.Context, opts ...QOption) (int64, error)

	// Create creates a new account and returns its ID.
	// Returns an error if the account could not be created.
	Create(ctx context.Context, account T) (uint64, error)

	// Update updates an existing account.
	// Returns an error if the account could not be updated.
	Update(ctx context.Context, id uint64, account T) error

	// Delete deletes an account by ID.
	// Returns an error if the account could not be deleted.
	Delete(ctx context.Context, id uint64) error
}

// SessionRepository extends account repository with auth session helpers.
type SessionRepository[TUser user.Model, TAccount Model] interface {
	Repository[TAccount]

	// LoadPermissions loads the permissions for the given user and account.
	LoadPermissions(ctx context.Context, account TAccount, userObj TUser) error

	// GetByToken retrieves the user and account associated with the given session token.
	// Returns ErrInvalidToken if the token is invalid or expired.
	GetByToken(ctx context.Context, token string) (TUser, TAccount, error)
}

// MemberRepository of access to the account members.
// TUser must be a pointer type implementing user.Model (e.g. consumer-defined User struct).
type MemberRepository[TUser user.Model, TAccount Model] interface {
	// EmptyObject returns a new empty member object of type Member[TUser, TAccount].
	EmptyObject() *Member[TUser, TAccount]

	// IsAdmin checks if the user is an admin of the account.
	IsAdmin(ctx context.Context, userID, accountID uint64) bool

	// IsMember checks if the user is a member of the account.
	IsMember(ctx context.Context, userID, accountID uint64) bool

	// FetchListMembers retrieves a list of members for the account.
	FetchListMembers(ctx context.Context, opts ...QOption) ([]*Member[TUser, TAccount], error)

	// CountMembers counts the number of members for the account.
	CountMembers(ctx context.Context, opts ...QOption) (int64, error)

	// Member retrieves a member by user ID and account ID.
	Member(ctx context.Context, userID, accountID uint64) (*Member[TUser, TAccount], error)

	// MemberByID retrieves a member by member ID.
	MemberByID(ctx context.Context, id uint64) (*Member[TUser, TAccount], error)

	// LinkMember links users as members to the account with optional admin role.
	LinkMember(ctx context.Context, account TAccount, isAdmin bool, members ...TUser) error

	// UnlinkMember unlinks users from the account.
	UnlinkMember(ctx context.Context, account TAccount, members ...TUser) error

	// UnlinkAccountMember unlinks a member from the account by member ID.
	SetMemberRoles(ctx context.Context, account TAccount, member TUser, roles ...string) error
}
