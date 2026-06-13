package usecase

import (
	"context"
	"slices"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/pkg/errors"
)

// MemberUsecase provides bussiness logic for account members
type MemberUsecase struct {
	userRepo    user.Repository
	accountRepo account.Repository
	memberRepo  account.MemberRepository
}

// NewMemberUsecase object controller
func NewMemberUsecase(userRepo user.Repository, accountRepo account.Repository, memberRepo account.MemberRepository) *MemberUsecase {
	return &MemberUsecase{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
	}
}

// FetchListMembers returns the list of members from account
func (a *MemberUsecase) FetchListMembers(ctx context.Context, opts ...account.QOption) (_ []*account.AccountMember, err error) {
	if !acl.HaveAccessList(ctx, &accountModels.AccountMember{}) {
		if opts, err = adjustMemberListFilterOpts(ctx, "list", opts); err != nil {
			return nil, err
		}
	}
	return a.memberRepo.FetchListMembers(ctx, opts...)
}

// CountMembers returns the count of members from account
func (a *MemberUsecase) CountMembers(ctx context.Context, opts ...account.QOption) (_ int64, err error) {
	if !acl.HaveAccessCount(ctx, &accountModels.AccountMember{}) {
		if opts, err = adjustMemberListFilterOpts(ctx, "count", opts); err != nil {
			return 0, err
		}
	}
	return a.memberRepo.CountMembers(ctx, opts...)
}

// LinkMember into account
func (a *MemberUsecase) LinkMember(ctx context.Context, accountObj *account.Account, isAdmin bool, members ...*user.User) error {
	if !acl.HaveAccessView(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	if !acl.HaveAccessCreate(ctx, &account.AccountMember{}) {
		return errors.Wrap(acl.ErrNoPermissions, "create member account")
	}
	return a.memberRepo.LinkMember(ctx, accountObj, isAdmin, members...)
}

// UnlinkMember from the account
func (a *MemberUsecase) UnlinkMember(ctx context.Context, accountObj *account.Account, members ...*user.User) error {
	if len(members) == 0 {
		return nil
	}
	for _, member := range members {
		if !acl.HaveAccessDelete(ctx, &account.AccountMember{AccountID: accountObj.ID, UserID: member.ID}) {
			return errors.Wrap(acl.ErrNoPermissions, "delete member account")
		}
	}
	return a.memberRepo.UnlinkMember(ctx, accountObj, members...)
}

// UnlinkAccountMember from the account
func (a *MemberUsecase) UnlinkAccountMember(ctx context.Context, memberID uint64) error {
	member, err := a.memberRepo.MemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	return a.memberRepo.UnlinkMember(ctx, member.Account, member.User)
}

// InviteMember into account by email
func (a *MemberUsecase) InviteMember(ctx context.Context, accountID uint64, email string, roles ...string) (*account.AccountMember, error) {
	// Check permissions for the account object `invite`
	if !acl.HaveObjectPermissions(ctx, &account.AccountMember{AccountID: accountID}, `invite`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "invite member account")
	}
	// Get account by ID
	account, err := a.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, err
	}
	// Get user by email
	usr, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Link the user to the account as member
	if err = a.memberRepo.LinkMember(ctx, account, slices.Contains(roles, "admin"), usr); err != nil {
		return nil, err
	}
	// Set roles for the member
	if err = a.memberRepo.SetMemberRoles(ctx, account, usr, roles...); err != nil {
		return nil, err
	}
	// Return the member object
	member, err := a.memberRepo.Member(ctx, usr.ID, account.ID)
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
func (a *MemberUsecase) SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*account.AccountMember, error) {
	memeber, err := a.memberRepo.Member(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, memeber, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return memeber, a.memberRepo.SetMemberRoles(ctx, memeber.Account, memeber.User, roles...)
}

// SetMemberRoles into account
func (a *MemberUsecase) SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*account.AccountMember, error) {
	memeber, err := a.memberRepo.MemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, memeber, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return memeber, a.memberRepo.SetMemberRoles(ctx, memeber.Account, memeber.User, roles...)
}

func adjustMemberListFilterOpts(ctx context.Context, action string, opts []account.QOption) ([]account.QOption, error) {
	accID := session.Account(ctx).ID
	for _, opt := range opts {
		if f, ok := opt.(*account.MemberFilter); ok {
			if l := len(f.AccountID); l > 1 || (l == 1 && f.AccountID[0] != accID) {
				return nil, errors.Wrap(acl.ErrNoPermissions, action+" member account for that account")
			}
			f.AccountID = []uint64{accID}
			return opts, nil
		}
	}
	return append(opts, &account.MemberFilter{AccountID: []uint64{accID}}), nil
}
