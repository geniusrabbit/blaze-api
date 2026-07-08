package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type MemberQueryResolver[TUser user.Model, TAccount account.Model, TGQLAccount any] struct {
	accounts account.Usecase[TUser, TAccount]
	members  account.MemberUsecase[TUser, TAccount]
	users    user.Repository[TUser]
}

type MemberQueryResolverConfig[TUser user.Model, TAccount account.Model, TGQLAccount any] struct {
	Accounts account.Usecase[TUser, TAccount]
	Members  account.MemberUsecase[TUser, TAccount]
	UserRepo user.Repository[TUser]
}

func NewMemberQueryResolver[TUser user.Model, TAccount account.Model, TGQLAccount any](
	cfg MemberQueryResolverConfig[TUser, TAccount, TGQLAccount],
) *MemberQueryResolver[TUser, TAccount, TGQLAccount] {
	return &MemberQueryResolver[TUser, TAccount, TGQLAccount]{
		accounts: cfg.Accounts,
		members:  cfg.Members,
		users:    cfg.UserRepo,
	}
}

// Invite is the resolver for the inviteAccountMember field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) Invite(ctx context.Context, accountID uint64, member *models.InviteMemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.members.InviteMember(ctx, accountID, member.UserID, InviteMemberAllRoles(member)...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         accountID,
		Member:           FromMemberModel(ctx, accountMember, r.accounts, r.users),
	}, nil
}

// Update is the resolver for the updateAccountMember field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) Update(ctx context.Context, memberID uint64, member *models.MemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.members.SetMemberRoles(ctx, memberID, MemberAllRoles(member)...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
		Member:           FromMemberModel(ctx, accountMember, r.accounts, r.users),
	}, nil
}

// Remove is the resolver for the removeAccountMember field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) Remove(ctx context.Context, memberID uint64) (*models.MemberPayload, error) {
	err := r.members.UnlinkAccountMember(ctx, memberID)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
	}, nil
}

// ApproveAccountMember is the resolver for the approveAccountMember field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) Approve(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	panic(fmt.Errorf("not implemented: ApproveAccountMember - approveAccountMember"))
}

// Reject is the resolver for the rejectAccountMember field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) Reject(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	panic(fmt.Errorf("not implemented: RejectAccountMember - rejectAccountMember"))
}

// List is the resolver for the listMembers field.
func (r *MemberQueryResolver[TUser, TAccount, TGQLAccount]) List(ctx context.Context, filter *models.MemberListFilter, order []*models.MemberListOrder, page *models.Page) (*MemberConnection, error) {
	return NewMemberConnection(ctx, r.members, r.accounts, r.users, filter, order, page), nil
}
