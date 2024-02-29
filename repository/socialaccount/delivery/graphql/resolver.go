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

func (r *QueryResolver) List(ctx context.Context, filter *models.SocialAccountListFilter, order *models.SocialAccountListOrder, page *models.Page) (*connectors.CollectionConnection[models.SocialAccount, models.SocialAccountEdge], error) {
	return connectors.NewSocialAccountConnection(ctx, r.accsounts, filter, order, page), nil
}
