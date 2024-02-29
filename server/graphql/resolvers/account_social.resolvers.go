package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.44

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// CurrentSocialAccounts is the resolver for the currentSocialAccounts field.
func (r *queryResolver) CurrentSocialAccounts(ctx context.Context, filter *models.SocialAccountListFilter) (*connectors.CollectionConnection[models.SocialAccount, models.SocialAccountEdge], error) {
	return r.socAccounts.CurrentSocialAccounts(ctx, filter)
}