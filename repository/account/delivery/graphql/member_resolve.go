package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

func ResolveMemberAccount[TUser user.Model, TAccount account.Model](
	ctx context.Context,
	member *account.Member[TUser, TAccount],
	accounts account.Usecase[TUser, TAccount],
) TAccount {
	var zero TAccount
	if member == nil {
		return zero
	}
	if any(member.Account) != any(zero) {
		return member.Account
	}
	if member.AccountID == 0 {
		return zero
	}
	if acc, err := accounts.Get(ctx, member.AccountID); err == nil {
		return acc
	}
	return ModelWithID(accounts.EmptyObject, member.AccountID)
}

func ResolveMemberUser[TUser user.Model, TAccount account.Model](
	ctx context.Context,
	member *account.Member[TUser, TAccount],
	users user.Repository[TUser],
) TUser {
	var zero TUser
	if member == nil {
		return zero
	}
	if any(member.User) != any(zero) {
		return member.User
	}
	if member.UserID == 0 {
		return zero
	}
	if usr, err := users.Get(ctx, member.UserID); err == nil {
		return usr
	}
	return ModelWithID(users.EmptyObject, member.UserID)
}
