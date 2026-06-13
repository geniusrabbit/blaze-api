// Package usecase account implementation
package usecase

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrOwnerRequired = errors.New("owner is required")

// AccountUsecase provides bussiness logic for account access
type AccountUsecase struct {
	userRepo    user.Repository
	accountRepo account.Repository
	memberRepo  account.MemberRepository
}

// NewAccountUsecase object controller
func NewAccountUsecase(userRepo user.Repository, accountRepo account.Repository, memberRepo account.MemberRepository) *AccountUsecase {
	return &AccountUsecase{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
	}
}

// Get returns the group by ID if have access
func (a *AccountUsecase) Get(ctx context.Context, id uint64) (*account.Account, error) {
	accountObj, err := a.accountRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accountObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	return accountObj, nil
}

// FetchList of accounts by filter
func (a *AccountUsecase) FetchList(ctx context.Context, filter *account.Filter, order *account.ListOrder, pagination *repository.Pagination) ([]*account.Account, error) {
	var err error
	if !acl.HaveAccessList(ctx, session.Account(ctx)) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	// If not `system` access then filter by current user
	if !acl.HaveAccessList(ctx, &account.Account{}) {
		if filter, err = adjustListFilter(ctx, filter); err != nil {
			return nil, err
		}
	}
	return a.accountRepo.FetchList(ctx, filter, order, pagination)
}

// Count of accounts by filter
func (a *AccountUsecase) Count(ctx context.Context, filter *account.Filter) (int64, error) {
	var err error
	if !acl.HaveAccessCount(ctx, session.Account(ctx)) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	// If not `system` access then filter by current user
	if !acl.HaveAccessCount(ctx, &account.Account{}) {
		if filter, err = adjustListFilter(ctx, filter); err != nil {
			return 0, err
		}
	}
	return a.accountRepo.Count(ctx, filter)
}

// Store new object into database
func (a *AccountUsecase) Store(ctx context.Context, accountObj *account.Account) (uint64, error) {
	var err error
	if accountObj.ID == 0 {
		if !acl.HaveAccessCreate(ctx, accountObj) {
			return 0, errors.Wrap(acl.ErrNoPermissions, "create account")
		}
		accountObj.ID, err = a.accountRepo.Create(ctx, accountObj)
		return accountObj.ID, err
	}
	if !acl.HaveAccessUpdate(ctx, accountObj) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "update account")
	}
	return accountObj.ID, a.accountRepo.Update(
		historylog.WithPK(ctx, accountObj.ID),
		accountObj.ID,
		accountObj,
	)
}

// Register new account with owner if not exists
func (a *AccountUsecase) Register(ctx context.Context, ownerObj *user.User, accountObj *account.Account, password string) (uint64, error) {
	if ownerObj == nil || (ownerObj.ID == 0 && ownerObj.Email == "") {
		return 0, errors.Wrap(ErrOwnerRequired, "invalid user data")
	}
	if !acl.HavePermissions(ctx, "account.register") {
		return 0, errors.Wrap(acl.ErrNoPermissions, "register account")
	}
	if ownerObj.ID == 0 {
		if user, _ := a.userRepo.GetByEmail(ctx, ownerObj.Email); user != nil {
			return 0, fmt.Errorf("user with email %s cant be registered", ownerObj.Email)
		}
	}
	// Execute all operations in transaction
	err := database.ContextTransactionExec(ctx, func(txctx context.Context, tx *gorm.DB) error {
		// If user not exists then create it
		if ownerObj.ID == 0 {
			uid, err := a.userRepo.Create(txctx, ownerObj, password)
			if err != nil {
				return err
			}
			ownerObj.ID = uid
		}
		// Create account
		aid, err := a.accountRepo.Create(txctx, accountObj)
		if err != nil {
			return err
		}
		accountObj.ID = aid
		// Link the user to the account as owner (admin)
		if err := a.memberRepo.LinkMember(txctx, accountObj, true, ownerObj); err != nil {
			return err
		}
		return nil
	})
	return accountObj.ID, err
}

// Delete delites record by ID
func (a *AccountUsecase) Delete(ctx context.Context, id uint64) error {
	accountObj, err := a.accountRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete account")
	}
	return a.accountRepo.Delete(historylog.WithPK(ctx, id), id)
}

func adjustListFilter(ctx context.Context, filter *account.Filter) (*account.Filter, error) {
	usr := session.User(ctx)
	if filter == nil {
		return &account.Filter{UserID: []uint64{usr.ID}}, nil
	} else if len(filter.UserID) == 0 {
		filter.UserID = []uint64{usr.ID}
	}
	if len(filter.UserID) != 1 || filter.UserID[0] != usr.ID {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list account (too wide filter)")
	}
	return filter, nil
}
