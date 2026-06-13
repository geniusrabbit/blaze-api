package auth

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
	userRepository "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

var (
	// errAuthUserIsNotMemberOfAccount is returned when a user is not a member of the target account
	errAuthUserIsNotMemberOfAccount = errors.New("user is not a member of the account")
	// errNoCrossAuthPermission is returned when a user lacks cross-account authentication permissions
	errNoCrossAuthPermission = errors.New("user don't have cross auth permissions")
)

// UserAccountByID retrieves and validates a user and account by their IDs, ensuring proper permissions.
// It checks membership, cross-account permissions, and loads account permissions for the user.
func UserAccountByID(ctx context.Context, uID, accID uint64, preUser *userModels.User, prevAccount *models.Account) (*userModels.User, *models.Account, error) {
	var (
		err      error
		users    = userRepository.NewUserRepository()
		accounts = accountRepository.NewAccountRepository()
		members  = accountRepository.NewMemberRepository()
		account  = prevAccount
		userObj  = preUser
	)

	// Fetch user if ID is provided and doesn't match the pre-loaded user
	if uID > 0 && (preUser == nil || preUser.ID != uID) {
		if userObj, err = users.Get(ctx, uID); err != nil {
			return nil, nil, err
		}
	}

	// Fetch account if ID is provided and doesn't match the pre-loaded account
	if accID > 0 && (prevAccount == nil || prevAccount.ID != accID) {
		if account, err = accounts.Get(ctx, accID); err != nil {
			return nil, nil, err
		}
	}

	// Validate permissions and load account data
	if account != nil {
		// Verify user is a member of the target account
		if userObj != nil && !members.IsMember(ctx, userObj.ID, account.ID) {
			return nil, nil, errAuthUserIsNotMemberOfAccount
		}

		// Verify cross-account access permissions if switching accounts
		if prevAccount != nil && prevAccount.ID != account.ID &&
			!prevAccount.CheckPermissions(ctx, account, session.PermAuthCross) {
			return nil, nil, errNoCrossAuthPermission
		}

		// Load account permissions for the user
		err = accounts.LoadPermissions(ctx, account, userObj)
		if err != nil {
			return nil, nil, err
		}

		// Extend account permissions from previous account if applicable
		if prevAccount != nil {
			account.ExtendPermissions(prevAccount.Permissions)
		}
	}

	return userObj, account, nil
}

// CrossAccountConnect validates and connects to a cross-account context if specified via header.
// Returns the user and account objects with updated permissions.
func CrossAccountConnect(ctx context.Context, crossAccountID string, userObj *userModels.User, accountObj *models.Account) (*userModels.User, *models.Account, error) {
	if crossAccountID != "" {
		userID, accountID := session.ParseCrossAuthHeader(crossAccountID)
		if userID > 0 || accountID > 0 {
			return UserAccountByID(ctx, userID, accountID, userObj, accountObj)
		}
	}
	return userObj, accountObj, nil
}
