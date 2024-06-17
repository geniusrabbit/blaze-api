package authutils

import (
	"context"
	"errors"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	accountRepository "github.com/geniusrabbit/blaze-api/repository/account/repository"
	userRepository "github.com/geniusrabbit/blaze-api/repository/user/repository"
)

var (
	errAuthUserIsNotMemberOfAccount = errors.New("user is not a member of the account")
	errNoCrossAuthPermission        = errors.New("user don't have cross auth permissions")
)

func UserAccountByID(ctx context.Context, uid, accid uint64, preUser *model.User, prevAccount *model.Account) (*model.User, *model.Account, error) {
	var (
		err      error
		users    = userRepository.New()
		accounts = accountRepository.New()
		account  = prevAccount
		userObj  = preUser
	)
	if uid > 0 && (preUser == nil || preUser.ID != uid) {
		if userObj, err = users.Get(ctx, uid); err != nil {
			return nil, nil, err
		}
	}
	if accid > 0 && (prevAccount == nil || prevAccount.ID != accid) {
		if account, err = accounts.Get(ctx, accid); err != nil {
			return nil, nil, err
		}
	}
	if account != nil {
		if userObj != nil && !accounts.IsMember(ctx, userObj.ID, account.ID) {
			return nil, nil, errAuthUserIsNotMemberOfAccount
		}
		if prevAccount != nil && prevAccount.ID != account.ID &&
			!prevAccount.CheckPermissions(ctx, account, session.PermAuthCross) {
			return nil, nil, errNoCrossAuthPermission
		}
		err = accounts.LoadPermissions(ctx, account, userObj)
		if err != nil {
			return nil, nil, err
		}
		if prevAccount != nil {
			account.ExtendPermissions(prevAccount.Permissions)
		}
	}
	return userObj, account, nil
}

func CrossAccountConnect(ctx context.Context, crossAccountID string, userObj *model.User, accountObj *model.Account) (*model.User, *model.Account, error) {
	if crossAccountID != "" {
		userID, accountID := session.ParseCrossAuthHeader(crossAccountID)
		if userID > 0 || accountID > 0 {
			return UserAccountByID(ctx, userID, accountID, userObj, accountObj)
		}
	}
	return userObj, accountObj, nil
}
