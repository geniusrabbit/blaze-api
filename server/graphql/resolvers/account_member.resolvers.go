package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// InviteAccountMember is the resolver for the inviteAccountMember field.
func (r *mutationResolver) InviteAccountMember(ctx context.Context, accountID uint64, member models.InviteMemberInput) (*models.MemberPayload, error) {
	return r.members.Invite(ctx, accountID, member)
}

// UpdateAccountMember is the resolver for the updateAccountMember field.
func (r *mutationResolver) UpdateAccountMember(ctx context.Context, memberID uint64, member models.MemberInput) (*models.MemberPayload, error) {
	return r.members.Update(ctx, memberID, member)
}

// RemoveAccountMember is the resolver for the removeAccountMember field.
func (r *mutationResolver) RemoveAccountMember(ctx context.Context, memberID uint64) (*models.MemberPayload, error) {
	return r.members.Remove(ctx, memberID)
}

// ApproveAccountMember is the resolver for the approveAccountMember field.
func (r *mutationResolver) ApproveAccountMember(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	return r.members.Approve(ctx, memberID, msg)
}

// RejectAccountMember is the resolver for the rejectAccountMember field.
func (r *mutationResolver) RejectAccountMember(ctx context.Context, memberID uint64, msg string) (*models.MemberPayload, error) {
	return r.members.Reject(ctx, memberID, msg)
}

// ListMembers is the resolver for the listMembers field.
func (r *queryResolver) ListMembers(ctx context.Context, filter *models.MemberListFilter, order *models.MemberListOrder, page *models.Page) (*connectors.CollectionConnection[models.Member, models.MemberEdge], error) {
	return r.members.List(ctx, filter, order, page)
}
