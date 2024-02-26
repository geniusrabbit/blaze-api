package graphql

import (
	"context"
	"fmt"

	"github.com/geniusrabbit/blaze-api/context/session"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/repository"
	"github.com/geniusrabbit/blaze-api/repository/socialaccount/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type QueryResolver struct {
	accsounts socialaccount.Usecase
}

func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		accsounts: usecase.New(
			repository.New(),
		),
	}
}

func (r *QueryResolver) CurrentSocialAccounts(ctx context.Context, filter *models.SocialAccountListFilter) (*connectors.SocialAccountConnection, error) {
	if filter == nil {
		filter = &models.SocialAccountListFilter{}
	}
	if len(filter.UserID) > 1 || (len(filter.UserID) == 1 && filter.UserID[0] != session.User(ctx).ID) {
		return nil, fmt.Errorf("filter by user id is not allowed for current user")
	}
	filter.UserID = append(filter.UserID[:0], session.User(ctx).ID)
	return connectors.NewSocialAccountConnection(ctx, r.accsounts, filter, nil, nil), nil
}
