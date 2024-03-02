package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// DisconnectSocialAccount is the resolver for the disconnectSocialAccount field.
func (r *mutationResolver) DisconnectSocialAccount(ctx context.Context, id uint64) (*models.SocialAccountPayload, error) {
	return r.Resolver.Mutation().DisconnectSocialAccount(ctx, id)
}

// SocialAccount is the resolver for the socialAccount field.
func (r *queryResolver) SocialAccount(ctx context.Context, id uint64) (*models.SocialAccountPayload, error) {
	return r.Resolver.Query().SocialAccount(ctx, id)
}

// CurrentSocialAccounts is the resolver for the currentSocialAccounts field.
func (r *queryResolver) CurrentSocialAccounts(ctx context.Context, filter *models.SocialAccountListFilter, order *models.SocialAccountListOrder) (*connectors.CollectionConnection[models.SocialAccount, models.SocialAccountEdge], error) {
	return r.Resolver.Query().CurrentSocialAccounts(ctx, filter, order)
}

// ListSocialAccounts is the resolver for the listSocialAccounts field.
func (r *queryResolver) ListSocialAccounts(ctx context.Context, filter *models.SocialAccountListFilter, order *models.SocialAccountListOrder, page *models.Page) (*connectors.CollectionConnection[models.SocialAccount, models.SocialAccountEdge], error) {
	return r.Resolver.Query().ListSocialAccounts(ctx, filter, order, page)
}
