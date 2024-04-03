package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver for the social account
type QueryResolver struct {
	accsounts socialaccount.Usecase
}

// NewQueryResolver creates a new instance of the QueryResolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		accsounts: usecase.New(
			repository.New(),
		),
	}
}

// Get Social Account by ID
func (r *QueryResolver) Get(ctx context.Context, id uint64) (*models.SocialAccountPayload, error) {
	obj, err := r.accsounts.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.SocialAccountPayload{
		ClientMutationID: requestid.Get(ctx),
		SocialAccountID:  obj.ID,
		SocialAccount:    models.FromSocialAccountModel(obj),
	}, nil
}

// Current Social Accounts list
func (r *QueryResolver) ListCurrent(ctx context.Context, filter *models.SocialAccountListFilter, order *models.SocialAccountListOrder) (*connectors.SocialAccountConnection, error) {
	if filter == nil {
		filter = &models.SocialAccountListFilter{}
	}
	if len(filter.UserID) > 1 || (len(filter.UserID) == 1 && filter.UserID[0] != session.User(ctx).ID) {
		return nil, fmt.Errorf("filter by user id is not allowed for current user")
	}
	filter.UserID = append(filter.UserID[:0], session.User(ctx).ID)
	return connectors.NewSocialAccountConnection(ctx, r.accsounts, filter, order, nil), nil
}

// List Social Accounts
func (r *QueryResolver) List(ctx context.Context, filter *models.SocialAccountListFilter, order *models.SocialAccountListOrder, page *models.Page) (*connectors.CollectionConnection[models.SocialAccount, models.SocialAccountEdge], error) {
	return connectors.NewSocialAccountConnection(ctx, r.accsounts, filter, order, page), nil
}

// Disconnect Social Account
func (r *QueryResolver) Disconnect(ctx context.Context, socialAccountID uint64) (*models.SocialAccountPayload, error) {
	obj, err := r.accsounts.Disconnect(ctx, socialAccountID)
	if err != nil {
		return nil, err
	}
	return &models.SocialAccountPayload{
		SocialAccount: models.FromSocialAccountModel(obj),
	}, nil
}
