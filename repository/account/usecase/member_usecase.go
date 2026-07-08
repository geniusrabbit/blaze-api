package usecase

import (
	"context"
	"slices"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/pkg/errors"
)

// MemberUsecase provides bussiness logic for account members
type MemberUsecase[TUser user.Model, TAccount account.Model] struct {
	userRepo    user.Repository[TUser]
	accountRepo account.SessionRepository[TUser, TAccount]
	memberRepo  account.MemberRepository[TUser, TAccount]
}

// NewMemberUsecase object controller
func NewMemberUsecase[TUser user.Model, TAccount account.Model](
	userRepo user.Repository[TUser],
	accountRepo account.SessionRepository[TUser, TAccount],
	memberRepo account.MemberRepository[TUser, TAccount],
) *MemberUsecase[TUser, TAccount] {
	return &MemberUsecase[TUser, TAccount]{
		userRepo:    userRepo,
		accountRepo: accountRepo,
		memberRepo:  memberRepo,
	}
}

func (a *MemberUsecase[TUser, TAccount]) aclMember(accountID, userID uint64) *account.Member[TUser, TAccount] {
	return account.MemberStub[TUser, TAccount](0, accountID, userID)
}

// EmptyObject returns a new empty member object of type Member[TUser, TAccount].
func (a *MemberUsecase[TUser, TAccount]) EmptyObject() *account.Member[TUser, TAccount] {
	return a.memberRepo.EmptyObject()
}

// FetchListMembers returns the list of members from account
func (a *MemberUsecase[TUser, TAccount]) FetchListMembers(ctx context.Context, opts ...account.QOption) (_ []*account.Member[TUser, TAccount], err error) {
	if !acl.HaveAccessList(ctx, a.EmptyObject()) {
		if opts, err = account.ListOptions(opts).WithPermissions(ctx, &account.MemberFilter{}); err != nil {
			return nil, errors.Wrap(acl.ErrNoPermissions, err.Error())
		}
	}
	return a.memberRepo.FetchListMembers(ctx, opts...)
}

// CountMembers returns the count of members from account
func (a *MemberUsecase[TUser, TAccount]) CountMembers(ctx context.Context, opts ...account.QOption) (int64, error) {
	if !acl.HaveAccessCount(ctx, a.EmptyObject()) {
		var err error
		if opts, err = account.ListOptions(opts).WithPermissions(ctx, &account.MemberFilter{}); err != nil {
			return 0, errors.Wrap(acl.ErrNoPermissions, err.Error())
		}
	}
	return a.memberRepo.CountMembers(ctx, opts...)
}

// LinkMember into account
func (a *MemberUsecase[TUser, TAccount]) LinkMember(ctx context.Context, accountObj TAccount, isAdmin bool, members ...TUser) error {
	if !acl.HaveAccessView(ctx, accountObj) {
		return errors.Wrap(acl.ErrNoPermissions, "view account")
	}
	if !acl.HaveAccessCreate(ctx, a.EmptyObject()) {
		return errors.Wrap(acl.ErrNoPermissions, "create member account")
	}
	return a.memberRepo.LinkMember(ctx, accountObj, isAdmin, members...)
}

// UnlinkMember from the account
func (a *MemberUsecase[TUser, TAccount]) UnlinkMember(ctx context.Context, accountObj TAccount, members ...TUser) error {
	if len(members) == 0 {
		return nil
	}
	for _, member := range members {
		if !acl.HaveAccessDelete(ctx, a.aclMember(accountObj.GetID(), member.GetID())) {
			return errors.Wrap(acl.ErrNoPermissions, "delete member account")
		}
	}
	return a.memberRepo.UnlinkMember(ctx, accountObj, members...)
}

// UnlinkAccountMember from the account
func (a *MemberUsecase[TUser, TAccount]) UnlinkAccountMember(ctx context.Context, memberID uint64) error {
	member, err := a.memberRepo.MemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	return a.memberRepo.UnlinkMember(ctx, member.Account, member.User)
}

// InviteMember into account by email (requires email repository on userRepo).
func (a *MemberUsecase[TUser, TAccount]) InviteMember(ctx context.Context, accountID, userID uint64, roles ...string) (*account.Member[TUser, TAccount], error) {
	// Check if user has permission to invite members
	if !acl.HaveObjectPermissions(ctx, a.aclMember(accountID, 0), `invite`) {
		return nil, acl.ErrNoPermissions.WithMessage("invite member account")
	}

	// Fetch account and user by email
	accountObj, err := a.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, err
	}
	usr, err := a.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Link member to account and set roles
	if err = a.memberRepo.LinkMember(ctx, accountObj, slices.Contains(roles, "admin"), usr); err != nil {
		return nil, err
	}
	if err = a.memberRepo.SetMemberRoles(ctx, accountObj, usr, roles...); err != nil {
		return nil, err
	}
	member, err := a.memberRepo.Member(ctx, usr.GetID(), accountObj.GetID())
	if err != nil {
		return nil, err
	}
	if !acl.HaveAccessView(ctx, member) {
		return nil, acl.ErrNoPermissions.WithMessage("view member account")
	}
	return member, nil
}

// SetAccountMemeberRoles into account
func (a *MemberUsecase[TUser, TAccount]) SetAccountMemeberRoles(ctx context.Context, accountID, userID uint64, roles ...string) (*account.Member[TUser, TAccount], error) {
	member, err := a.memberRepo.Member(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, member, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return member, a.memberRepo.SetMemberRoles(ctx, member.Account, member.User, roles...)
}

// SetMemberRoles into account
func (a *MemberUsecase[TUser, TAccount]) SetMemberRoles(ctx context.Context, memberID uint64, roles ...string) (*account.Member[TUser, TAccount], error) {
	member, err := a.memberRepo.MemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if !acl.HaveObjectPermissions(ctx, member, `roles.set.*`) {
		return nil, errors.Wrap(acl.ErrNoPermissions, "update member roles")
	}
	return member, a.memberRepo.SetMemberRoles(ctx, member.Account, member.User, roles...)
}
