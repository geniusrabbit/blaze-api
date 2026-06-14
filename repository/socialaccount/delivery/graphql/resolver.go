package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver handles GraphQL queries for social accounts.
type QueryResolver struct {
	accounts socialaccount.Usecase
}

// NewQueryResolver creates a new QueryResolver instance.
func NewQueryResolver(uc socialaccount.Usecase) *QueryResolver {
	return &QueryResolver{accounts: uc}
}

// Get retrieves a social account by ID.
func (r *QueryResolver) Get(ctx context.Context, id uint64) (*gqlmodels.SocialAccountPayload, error) {
	obj, err := r.accounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.SocialAccountPayload{
		ClientMutationID: requestid.Get(ctx),
		SocialAccountID:  id,
		SocialAccount:    FromSocialAccountModel(obj),
	}, nil
}

// ListCurrent returns social accounts for the current authenticated user.
func (r *QueryResolver) ListCurrent(
	ctx context.Context,
	filter *gqlmodels.SocialAccountListFilter,
	order *gqlmodels.SocialAccountListOrder,
) (*SocialAccountConnection, error) {
	if filter == nil {
		filter = &gqlmodels.SocialAccountListFilter{}
	}
	if len(filter.UserID) > 1 || (len(filter.UserID) == 1 && filter.UserID[0] != session.User(ctx).ID) {
		return nil, fmt.Errorf("filter by user id is not allowed for current user")
	}
	filter.UserID = append(filter.UserID[:0], session.User(ctx).ID)
	return NewSocialAccountConnection(ctx, r.accounts, filter, order, nil), nil
}

// List returns paginated social accounts with optional filtering and ordering.
func (r *QueryResolver) List(
	ctx context.Context,
	filter *gqlmodels.SocialAccountListFilter,
	order *gqlmodels.SocialAccountListOrder,
	page *gqlmodels.Page,
) (*SocialAccountConnection, error) {
	return NewSocialAccountConnection(ctx, r.accounts, filter, order, page), nil
}

// Disconnect removes a social account association.
func (r *QueryResolver) Disconnect(ctx context.Context, socialAccountID uint64) (*gqlmodels.SocialAccountPayload, error) {
	obj, err := r.accounts.Disconnect(ctx, socialAccountID)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.SocialAccountPayload{
		ClientMutationID: requestid.Get(ctx),
		SocialAccountID:  socialAccountID,
		SocialAccount:    FromSocialAccountModel(obj),
	}, nil
}
