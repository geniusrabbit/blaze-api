// Package usecase account implementation
package usecase

import (
	"context"
	"database/sql"

	"github.com/geniusrabbit/api-template-base/internal/acl"
	"github.com/geniusrabbit/api-template-base/internal/context/session"
	"github.com/geniusrabbit/api-template-base/internal/repository/account"
	"github.com/geniusrabbit/api-template-base/model"
	"github.com/pkg/errors"
)

// AccountUsecase provides bussiness logic for account access
type AccountUsecase struct {
	accountRepo account.Repository
}

// NewAccountUsecase object controller
func NewAccountUsecase(repo account.Repository) *AccountUsecase {
	return &AccountUsecase{
		accountRepo: repo,
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
func (a *AccountUsecase) FetchList(ctx context.Context, filter *account.Filter) ([]*model.Account, error) {
	if filter == nil {
		filter = &account.Filter{}
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if !acl.HaveAccessList(ctx, &model.Account{}) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	if filter.UserID == nil && len(filter.ID) == 0 {
		filter.UserID = []uint64{session.User(ctx).ID}
		return a.accountRepo.FetchList(ctx, filter)
	}
	list, err := a.accountRepo.FetchList(ctx, filter)
	for _, link := range list {
		if !acl.HaveAccessList(ctx, link) {
			return nil, errors.Wrap(acl.ErrNoPermissions, "list account")
		}
	}
	return list, err
}

// Count of accounts by filter
func (a *AccountUsecase) Count(ctx context.Context, filter *account.Filter) (int64, error) {
	if filter == nil {
		filter = &account.Filter{}
	}
	if !acl.HaveAccessList(ctx, &model.Account{}) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list account")
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
