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
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

var ErrOwnerRequired = errors.New("owner is required")

// AccountUsecase provides bussiness logic for account access
type AccountUsecase[TUser user.Model, TAccount account.Model] struct {
	userRepo    user.Repository[TUser]
	accountRepo account.SessionRepository[TUser, TAccount]
	memberRepo  account.MemberRepository[TUser, TAccount]
}

// NewAccountUsecase object controller
func NewAccountUsecase[TUser user.Model, TAccount account.Model](
	userRepo user.Repository[TUser],
	accountRepo account.SessionRepository[TUser, TAccount],
	memberRepo account.MemberRepository[TUser, TAccount],
) *AccountUsecase[TUser, TAccount] {
	return &AccountUsecase[TUser, TAccount]{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
	}
}

// EmptyObject returns a new empty account object of type TAccount.
func (a *AccountUsecase[TUser, TAccount]) EmptyObject() TAccount {
	return a.accountRepo.EmptyObject()
}

// Get returns the group by ID if have access
func (a *AccountUsecase[TUser, TAccount]) Get(ctx context.Context, id uint64) (TAccount, error) {
	var zero TAccount
	accountObj, err := a.accountRepo.Get(ctx, id)
	if err != nil {
		return zero, err
	}
	if !acl.HaveAccessView(ctx, accountObj) {
		return zero, errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	return accountObj, nil
}

// FetchList of accounts by filter
func (a *AccountUsecase[TUser, TAccount]) FetchList(ctx context.Context, opts ...account.QOption) ([]TAccount, error) {
	if !acl.HaveAccessList(ctx, session.Account(ctx)) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	if !acl.HaveAccessList(ctx, a.EmptyObject()) {
		var err error
		if opts, err = account.ListOptions(opts).WithPermissions(ctx, &account.Filter{}); err != nil {
			return nil, errors.Wrap(acl.ErrNoPermissions, err.Error())
		}
	}
	return a.accountRepo.FetchList(ctx, opts...)
}

// Count of accounts by filter
func (a *AccountUsecase[TUser, TAccount]) Count(ctx context.Context, opts ...account.QOption) (int64, error) {
	if !acl.HaveAccessCount(ctx, session.Account(ctx)) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "list account")
	}
	if !acl.HaveAccessCount(ctx, a.EmptyObject()) {
		var err error
		if opts, err = account.ListOptions(opts).WithPermissions(ctx, &account.Filter{}); err != nil {
			return 0, errors.Wrap(acl.ErrNoPermissions, err.Error())
		}
	}
	return a.accountRepo.Count(ctx, opts...)
}

// Update updates the account if have access, or creates new one if ID is 0
func (a *AccountUsecase[TUser, TAccount]) Update(ctx context.Context, accountObj TAccount) (uint64, error) {
	if accountObj.GetID() == 0 {
		if !acl.HaveAccessCreate(ctx, accountObj) {
			return 0, errors.Wrap(acl.ErrNoPermissions, "create account")
		}
		id, err := a.accountRepo.Create(ctx, accountObj)
		return id, err
	}
	if !acl.HaveAccessUpdate(ctx, accountObj) {
		return 0, errors.Wrap(acl.ErrNoPermissions, "update account")
	}
	return accountObj.GetID(), a.accountRepo.Update(
		historylog.WithPK(ctx, accountObj.GetID()),
		accountObj.GetID(),
		accountObj,
	)
}

// Register new account with owner if not exists (requires email + password repositories on userRepo).
func (a *AccountUsecase[TUser, TAccount]) Register(ctx context.Context, ownerObj TUser, accountObj TAccount) (uint64, error) {
	var zeroUser TUser
	type emailLookup interface {
		GetByEmail(context.Context, string) (TUser, error)
	}

	emailRepo, ok := any(a.userRepo).(emailLookup)
	if !ok {
		return 0, errors.New("user repository does not support email lookup")
	}

	ownerEmail := ""
	if em, ok := any(ownerObj).(user.EmailModel); ok {
		ownerEmail = em.GetEmail()
	}

	if any(ownerObj) == any(zeroUser) || (ownerObj.GetID() == 0 && ownerEmail == "") {
		return 0, errors.Wrap(ErrOwnerRequired, "invalid user data")
	}

	// Check permissions for account registration
	if !acl.HavePermissions(ctx, "account.register") {
		return 0, acl.ErrNoPermissions.WithMessage("register account")
	}

	// Check if user already exists by email
	if ownerObj.GetID() == 0 {
		if existing, _ := emailRepo.GetByEmail(ctx, ownerEmail); any(existing) != any(zeroUser) && existing.GetID() != 0 {
			return 0, fmt.Errorf("user with email %s cant be registered", ownerEmail)
		}
	}

	// Execute all operations in transaction
	err := database.ContextTransactionExec(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		// Create user if not exists
		if ownerObj.GetID() == 0 {
			uid, err := a.userRepo.Create(txCtx, ownerObj)
			if err != nil {
				return err
			}
			setUserID(ownerObj, uid)
		}

		// Create account and link owner as member
		aid, err := a.accountRepo.Create(txCtx, accountObj)
		if err != nil {
			return err
		}
		setAccountID(accountObj, aid)

		// Link owner as member with admin role
		return a.memberRepo.LinkMember(txCtx, accountObj, true, ownerObj)
	})
	return accountObj.GetID(), err
}

// Delete delites record by ID
func (a *AccountUsecase[TUser, TAccount]) Delete(ctx context.Context, id uint64) error {
	accountObj, err := a.accountRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !acl.HaveAccessDelete(ctx, accountObj) {
		return acl.ErrNoPermissions.WithMessage("delete account")
	}
	return a.accountRepo.Delete(historylog.WithPK(ctx, id), id)
}
