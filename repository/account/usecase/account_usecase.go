// Package usecase account implementation
package usecase

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/acl"
	"github.com/geniusrabbit/blaze-api/context/database"
	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// AccountUsecase provides bussiness logic for account access
type AccountUsecase struct {
	userRepo    user.Repository
	accountRepo account.Repository
}

// NewAccountUsecase object controller
func NewAccountUsecase(userRepo user.Repository, accountRepo account.Repository) *AccountUsecase {
	return &AccountUsecase{
		userRepo:    userRepo,
		accountRepo: accountRepo,
	}
}

// Get returns the group by ID if have access
func (a *AccountUsecase) Get(ctx context.Context, id uint64) (*model.Account, error) {
	_, currentAccount := session.UserAccount(ctx)
	if currentAccount.ID == id {
		return currentAccount, nil
	}
	accountObj, err := a.accountRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accountObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	return accountObj, nil
}

// GetByTitle returns the account by title if have access
func (a *AccountUsecase) GetByTitle(ctx context.Context, title string) (*model.Account, error) {
	_, currentAccount := session.UserAccount(ctx)
	if currentAccount.Title == title {
		return currentAccount, nil
	}
	accountObj, err := a.accountRepo.GetByTitle(ctx, title)
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, accountObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	return accountObj, nil
}

// FetchList of accounts by filter
func (a *AccountUsecase) FetchList(ctx context.Context, filter *account.Filter, order *account.ListOrder, pagination *repository.Pagination) ([]*model.Account, error) {
	var err error
	if !acl.HaveAccessList(ctx, session.Account(ctx)) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	// If not `system` access then filter by current user
	if !acl.HaveAccessList(ctx, &model.Account{}) {
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
	if !acl.HaveAccessCount(ctx, &model.Account{}) {
		if filter, err = adjustListFilter(ctx, filter); err != nil {
			return 0, err
		}
	}
	return a.accountRepo.Count(ctx, filter)
}

// Store new object into database
func (a *AccountUsecase) Store(ctx context.Context, accountObj *model.Account) (uint64, error) {
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
	return accountObj.ID, a.accountRepo.Update(ctx, accountObj.ID, accountObj)
}

// Register new account with owner if not exists
func (a *AccountUsecase) Register(ctx context.Context, ownerObj *model.User, accountObj *model.Account, password string) (uint64, error) {
	if !acl.HavePermissions(ctx, "account.register") {
		return 0, errors.Wrap(acl.ErrNoPermissions, "register account")
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
		if err := a.accountRepo.LinkMember(txctx, accountObj, true, ownerObj); err != nil {
			return err
		}
		return nil
	})
	return accountObj.ID, err
}

// Delete delites record by ID
func (a *AccountUsecase) Delete(ctx context.Context, id uint64) error {
	accountObj, err := a.getAccountByID(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "delete account")
	}
	return a.accountRepo.Delete(ctx, id)
}

func (a *AccountUsecase) getAccountByID(ctx context.Context, id uint64) (*model.Account, error) {
	_, currentAccount := session.UserAccount(ctx)
	if currentAccount.ID == id {
		return currentAccount, nil
	}
	return nil, sql.ErrNoRows
}

// FetchMembers returns the list of members from account
func (a *AccountUsecase) FetchMembers(ctx context.Context, accountObj *model.Account) ([]*model.AccountMember, error) {
	if !acl.HaveAccessView(ctx, accountObj) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	if !acl.HaveAccessList(ctx, &model.AccountMember{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list member account")
	}
	return a.accountRepo.FetchMembers(ctx, accountObj)
}

// FetchMemberUsers returns the list of members from account
func (a *AccountUsecase) FetchMemberUsers(ctx context.Context, accountObj *model.Account) ([]*model.AccountMember, []*model.User, error) {
	if !acl.HaveAccessView(ctx, accountObj) {
		return nil, nil, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	if !acl.HaveAccessList(ctx, &model.AccountMember{}) {
		return nil, nil, errors.Wrap(acl.ErrNoPermissions, "list member account")
	}
	return a.accountRepo.FetchMemberUsers(ctx, accountObj)
}

// LinkMember into account
func (a *AccountUsecase) LinkMember(ctx context.Context, accountObj *model.Account, isAdmin bool, members ...*model.User) error {
	if !acl.HaveAccessView(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	if !acl.HaveAccessCreate(ctx, &model.AccountMember{}) {
		return errors.Wrap(acl.ErrNoPermissions, "create member account")
	}
	return a.accountRepo.LinkMember(ctx, accountObj, isAdmin, members...)
}

// UnlinkMember from the account
func (a *AccountUsecase) UnlinkMember(ctx context.Context, accountObj *model.Account, members ...*model.User) error {
	if len(members) == 0 {
		return nil
	}
	if !acl.HaveAccessView(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "view member account")
	}
	for _, member := range members {
		if !acl.HaveAccessDelete(ctx, member) {
			return errors.Wrap(acl.ErrNoPermissions, "delete member account")
		}
	}
	return a.accountRepo.UnlinkMember(ctx, accountObj, members...)
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
