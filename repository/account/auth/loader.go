package auth

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var (
	// errAuthUserIsNotMemberOfAccount is returned when a user is not a member of the target account
	errAuthUserIsNotMemberOfAccount = errors.New("user is not a member of the account")
	// errNoCrossAuthPermission is returned when a user lacks cross-account authentication permissions
	errNoCrossAuthPermission = errors.New("user don't have cross auth permissions")
)

// Loader resolves user+account pairs with membership and permission checks.
type Loader[TUser user.Model, TAccount account.Model] struct {
	Users    user.Repository[TUser]
	Accounts account.SessionRepository[TUser, TAccount]
	Members  account.MemberRepository[TUser, TAccount]
}

// NewLoader creates a session loader for auth middleware and authorizers.
func NewLoader[TUser user.Model, TAccount account.Model](
	users user.Repository[TUser],
	accounts account.SessionRepository[TUser, TAccount],
	members account.MemberRepository[TUser, TAccount],
) *Loader[TUser, TAccount] {
	return &Loader[TUser, TAccount]{
		Users:    users,
		Accounts: accounts,
		Members:  members,
	}
}

// UserAccountByID retrieves and validates a user and account by their IDs, ensuring proper permissions.
func (l *Loader[TUser, TAccount]) UserAccountByID(ctx context.Context, uID, accID uint64, preUser TUser, prevAccount TAccount) (TUser, TAccount, error) {
	var (
		err      error
		zeroUser TUser
		zeroAcc  TAccount
		account  = prevAccount
		userObj  = preUser
	)

	if uID > 0 {
		if any(preUser) == any(zeroUser) || preUser.GetID() != uID {
			if userObj, err = l.Users.Get(ctx, uID); err != nil {
				return zeroUser, zeroAcc, err
			}
		}
	}

	if accID > 0 {
		if any(prevAccount) == any(zeroAcc) || prevAccount.GetID() != accID {
			if account, err = l.Accounts.Get(ctx, accID); err != nil {
				return zeroUser, zeroAcc, err
			}
		}
	}

	if any(account) != any(zeroAcc) {
		if any(userObj) != any(zeroUser) && !l.Members.IsMember(ctx, userObj.GetID(), account.GetID()) {
			return zeroUser, zeroAcc, errAuthUserIsNotMemberOfAccount
		}

		if any(prevAccount) != any(zeroAcc) && prevAccount.GetID() != account.GetID() &&
			!prevAccount.CheckPermissions(ctx, account, session.PermAuthCross) {
			return zeroUser, zeroAcc, errNoCrossAuthPermission
		}

		if err = l.Accounts.LoadPermissions(ctx, account, userObj); err != nil {
			return zeroUser, zeroAcc, err
		}

		if any(prevAccount) != any(zeroAcc) {
			account.ExtendPermissions(prevAccount.PermissionsChecker())
		}
	}

	return userObj, account, nil
}

// CrossAccountConnect validates and connects to a cross-account context if specified via header.
func (l *Loader[TUser, TAccount]) CrossAccountConnect(ctx context.Context, crossAccountID string, userObj TUser, accountObj TAccount) (TUser, TAccount, error) {
	if crossAccountID != "" {
		userID, accountID := session.ParseCrossAuthHeader(crossAccountID)
		if userID > 0 || accountID > 0 {
			return l.UserAccountByID(ctx, userID, accountID, userObj, accountObj)
		}
	}
	return userObj, accountObj, nil
}
