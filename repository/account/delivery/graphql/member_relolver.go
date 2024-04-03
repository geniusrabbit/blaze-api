package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/repository"
	"github.com/geniusrabbit/blaze-api/repository/account/usecase"
	userrepo "github.com/geniusrabbit/blaze-api/repository/user/repository"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type MemberQueryResolver struct {
	accounts account.Usecase
}

func NewMemberQueryResolver() *MemberQueryResolver {
	return &MemberQueryResolver{
		accounts: usecase.NewAccountUsecase(userrepo.New(), repository.New()),
	}
}

// Invite is the resolver for the inviteAccountMember field.
func (r *MemberQueryResolver) Invite(ctx context.Context, accountID uint64, member models.InviteMemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.accounts.InviteMember(ctx, accountID, member.Email, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         accountID,
		Member:           models.FromMemberModel(ctx, accountMember),
	}, nil
}

// Update is the resolver for the updateAccountMember field.
func (r *MemberQueryResolver) Update(ctx context.Context, memberID uint64, member models.MemberInput) (*models.MemberPayload, error) {
	accountMember, err := r.accounts.SetMemberRoles(ctx, memberID, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &models.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
		Member:           models.FromMemberModel(ctx, accountMember),
	}, nil
}

// Remove is the resolver for the removeAccountMember field.
func (r *MemberQueryResolver) Remove(ctx context.Context, memberID uint64) (*models.MemberPayload, error) {
	err := r.accounts.UnlinkAccountMember(ctx, memberID)
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
func (r *MemberQueryResolver) List(ctx context.Context, filter *models.MemberListFilter, order *models.MemberListOrder, page *models.Page) (*connectors.MemberConnection, error) {
	return connectors.NewMemberConnection(ctx, r.accounts, filter, order, page), nil
}
