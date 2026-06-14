package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type MemberQueryResolver struct {
	accounts account.Usecase
	members  account.MemberUsecase
}

func NewMemberQueryResolver(accounts account.Usecase, members account.MemberUsecase) *MemberQueryResolver {
	return &MemberQueryResolver{
		accounts: accounts,
		members:  members,
	}
}

// Invite is the resolver for the inviteAccountMember field.
func (r *MemberQueryResolver) Invite(ctx context.Context, accountID uint64, member models.InviteMemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.members.InviteMember(ctx, accountID, member.Email, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         accountID,
		Member:           FromMemberModel(ctx, accountMember),
	}, nil
}

// Update is the resolver for the updateAccountMember field.
func (r *MemberQueryResolver) Update(ctx context.Context, memberID uint64, member models.MemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.members.SetMemberRoles(ctx, memberID, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
		Member:           FromMemberModel(ctx, accountMember),
	}, nil
}

// Remove is the resolver for the removeAccountMember field.
func (r *MemberQueryResolver) Remove(ctx context.Context, memberID uint64) (*models.MemberPayload, error) {
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
func (r *MemberQueryResolver) Approve(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	panic(fmt.Errorf("not implemented: ApproveAccountMember - approveAccountMember"))
}

// Reject is the resolver for the rejectAccountMember field.
func (r *MemberQueryResolver) Reject(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	panic(fmt.Errorf("not implemented: RejectAccountMember - rejectAccountMember"))
}

// List is the resolver for the listMembers field.
func (r *MemberQueryResolver) List(ctx context.Context, filter *models.MemberListFilter, order *models.MemberListOrder, page *models.Page) (*MemberConnection, error) {
	return NewMemberConnection(ctx, r.members, filter, order, page), nil
}
