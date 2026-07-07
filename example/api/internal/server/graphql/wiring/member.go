package wiring

import (
	"context"
	"fmt"

	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountgraphql "github.com/geniusrabbit/blaze-api/repository/account/delivery/graphql"
	"github.com/geniusrabbit/blaze-api/repository/user"
	basemodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// ExampleMemberQueryResolver serves member GraphQL with extended Account on Member.account.
type ExampleMemberQueryResolver[TUser user.Model, TAccount account.Model] struct {
	accounts      account.Usecase[TUser, TAccount]
	members       account.MemberUsecase[TUser, TAccount]
	users         user.Repository[TUser]
	newAccount    func() TAccount
	newUser       func() TUser
	toGraphQL     accountgraphql.AccountGraphQLConverter[TAccount, exmodels.Account]
	toUserGraphQL accountgraphql.UserGraphQLConverter[TUser, accountgraphql.BaseAccountGQLUser]
}

// NewExampleMemberQueryResolver wires member resolver for extended Account schema.
func NewExampleMemberQueryResolver[TUser user.Model, TAccount account.Model](
	accounts account.Usecase[TUser, TAccount],
	members account.MemberUsecase[TUser, TAccount],
	users user.Repository[TUser],
	newAccount func() TAccount,
	newUser func() TUser,
	toGraphQL accountgraphql.AccountGraphQLConverter[TAccount, exmodels.Account],
	toUserGraphQL accountgraphql.UserGraphQLConverter[TUser, accountgraphql.BaseAccountGQLUser],
) *ExampleMemberQueryResolver[TUser, TAccount] {
	return &ExampleMemberQueryResolver[TUser, TAccount]{
		accounts:      accounts,
		members:       members,
		users:         users,
		newAccount:    newAccount,
		newUser:       newUser,
		toGraphQL:     toGraphQL,
		toUserGraphQL: toUserGraphQL,
	}
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) Invite(ctx context.Context, accountID uint64, member *exmodels.InviteMemberInput) (*exmodels.MemberPayload, error) {
	accountMember, err := r.members.InviteMember(ctx, accountID, member.UserID, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &exmodels.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         accountID,
		Member:           r.fromMember(ctx, accountMember),
	}, nil
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) Update(ctx context.Context, memberID uint64, member *exmodels.MemberInput) (*exmodels.MemberPayload, error) {
	accountMember, err := r.members.SetMemberRoles(ctx, memberID, member.AllRoles()...)
	if err != nil {
		return nil, err
	}
	return &exmodels.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
		Member:           r.fromMember(ctx, accountMember),
	}, nil
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) Remove(ctx context.Context, memberID uint64) (*exmodels.MemberPayload, error) {
	if err := r.members.UnlinkAccountMember(ctx, memberID); err != nil {
		return nil, err
	}
	return &exmodels.MemberPayload{
		ClientMutationID: requestid.Get(ctx),
		MemberID:         memberID,
	}, nil
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) Approve(ctx context.Context, memberID uint64, msg string) (*exmodels.MemberPayload, error) {
	return nil, fmt.Errorf("not implemented: approveAccountMember")
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) Reject(ctx context.Context, memberID uint64, msg string) (*exmodels.MemberPayload, error) {
	return nil, fmt.Errorf("not implemented: rejectAccountMember")
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) List(
	ctx context.Context,
	filter *exmodels.MemberListFilter,
	order []*exmodels.MemberListOrder,
	page *basemodels.Page,
) (*accountgraphql.MemberConnection, error) {
	return newExampleMemberConnection(ctx, r.members, filter, order, page, r.toGraphQL, r.accounts, r.users, r.newAccount, r.newUser), nil
}

func (r *ExampleMemberQueryResolver[TUser, TAccount]) fromMember(ctx context.Context, member *account.Member[TUser, TAccount]) *exmodels.Member {
	toAccount := func(acc TAccount) *exmodels.Account {
		gql := r.toGraphQL(acc)
		return &gql
	}
	return accountgraphql.FromMemberModel(ctx, member, toAccount, r.toUserGraphQL, r.accounts, r.users)
}
