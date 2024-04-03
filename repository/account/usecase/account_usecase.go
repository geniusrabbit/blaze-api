// Package usecase account implementation
package usecase

import (
	"context"
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrOwnerRequired = errors.New("owner is required")

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
		if err := a.accountRepo.LinkMember(txctx, accountObj, true, ownerObj); err != nil {
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
	return a.accountRepo.Delete(ctx, id)
}

// FetchListMembers returns the list of members from account
func (a *AccountUsecase) FetchListMembers(ctx context.Context, filter *account.MemberFilter, order *account.MemberListOrder, pagination *repository.Pagination) (_ []*model.AccountMember, err error) {
	if !acl.HaveAccessList(ctx, &model.AccountMember{}) {
		if filter, err = adjustMemberListFilter(ctx, "list", filter); err != nil {
			return nil, err
		}
	}
	return a.accountRepo.FetchListMembers(ctx, filter, order, pagination)
}

// CountMembers returns the count of members from account
func (a *AccountUsecase) CountMembers(ctx context.Context, filter *account.MemberFilter) (_ int64, err error) {
	if !acl.HaveAccessCount(ctx, &model.AccountMember{}) {
		if filter, err = adjustMemberListFilter(ctx, "count", filter); err != nil {
			return 0, err
		}
	}
	return a.accountRepo.CountMembers(ctx, filter)
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
	for _, member := range members {
		if !acl.HaveAccessDelete(ctx, &model.AccountMember{AccountID: accountObj.ID, UserID: member.ID}) {
			return errors.Wrap(acl.ErrNoPermissions, "delete member account")
		}
	}
	return a.accountRepo.UnlinkMember(ctx, accountObj, members...)
}

// UnlinkAccountMember from the account
func (a *AccountUsecase) UnlinkAccountMember(ctx context.Context, memberID uint64) error {
	member, err := a.accountRepo.MemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	return a.accountRepo.UnlinkMember(ctx, member.Account, member.User)
}

// InviteMember into account by email
func (a *AccountUsecase) InviteMember(ctx context.Context, accountID uint64, email string, roles ...string) (*model.AccountMember, error) {
	// Check permissions for the account object `invite`
	if !acl.HaveObjectPermissions(ctx, &model.AccountMember{AccountID: accountID}, `invite`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "invite member account")
	}
	// Get account by ID
	account, err := a.Get(ctx, accountID)
	if err != nil {
		return nil, err
	}
	// Get user by email
	usr, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Link the user to the account as member
	if err = a.accountRepo.LinkMember(ctx, account, slices.Contains(roles, "admin"), usr); err != nil {
		return nil, err
	}
	// Set roles for the member
	if err = a.accountRepo.SetMemberRoles(ctx, account, usr, roles...); err != nil {
		return nil, err
	}
	// Return the member object
	member, err := a.accountRepo.Member(ctx, usr.ID, account.ID)
	if err != nil {
		return nil, err
	}
	// Check permissions for the member object `view`
	if !acl.HaveAccessView(ctx, member) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "view member account")
	}
	return member, nil
}

// SetAccountMemeberRoles into account
func (a *AccountUsecase) SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*model.AccountMember, error) {
	memeber, err := a.accountRepo.Member(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, memeber, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return memeber, a.accountRepo.SetMemberRoles(ctx, memeber.Account, memeber.User, roles...)
}

// SetMemberRoles into account
func (a *AccountUsecase) SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*model.AccountMember, error) {
	memeber, err := a.accountRepo.MemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, memeber, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return memeber, a.accountRepo.SetMemberRoles(ctx, memeber.Account, memeber.User, roles...)
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

func adjustMemberListFilter(ctx context.Context, action string, filter *account.MemberFilter) (*account.MemberFilter, error) {
	if filter == nil {
		filter = &account.MemberFilter{
			AccountID: []uint64{session.Account(ctx).ID},
		}
	} else {
		if l := len(filter.AccountID); l > 1 || (l == 1 && filter.AccountID[0] != session.Account(ctx).ID) {
			return nil, errors.Wrap(acl.ErrNoPermissions, action+" member account for that account")
		}
		filter.AccountID = append(filter.AccountID[:0], session.Account(ctx).ID)
	}
	return filter, nil
}
