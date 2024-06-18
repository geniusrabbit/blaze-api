package graphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/repository"
	"github.com/geniusrabbit/blaze-api/repository/directaccesstoken/usecase"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

type QueryResolver struct {
	uc *usecase.Usecase
}

// NewQueryResolver creates a new resolver.
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{uc: usecase.New(repository.New())}
}

// Generate is the resolver for the generateDirectAccessToken field.
func (r *QueryResolver) Generate(ctx context.Context, userID *uint64, description string, expiresAt *time.Time) (*models.DirectAccessTokenPayload, error) {
	if expiresAt == nil {
		expiresAt = &time.Time{}
	}
	if expiresAt.IsZero() {
		*expiresAt = time.Now().Add(time.Hour * 24 * 30)
	}
	if expiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("expiresAt should be in future")
	}
	token, err := r.uc.Generate(ctx,
		gocast.IfThenExec(userID != nil, func() uint64 { return *userID }, func() uint64 { return 0 }),
		session.Account(ctx).ID,
		description,
		*expiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &models.DirectAccessTokenPayload{
		ClientMutationID: requestid.Get(ctx),
		Token:            models.FromDirectAccessToken(token),
	}, nil
}

// Revoke is the resolver for the revokeDirectAccessToken field.
func (r *QueryResolver) Revoke(ctx context.Context, filter models.DirectAccessTokenListFilter) (*models.StatusResponse, error) {
	err := r.uc.Revoke(ctx, filter.Filter())
	if err != nil {
		return nil, err
	}
	return &models.StatusResponse{
		ClientMutationID: requestid.Get(ctx),
		Status:           "ok",
		Message:          &[]string{"token(s) revoked"}[0],
	}, nil
}

// Get is the resolver for the getDirectAccessToken field.
func (r *QueryResolver) Get(ctx context.Context, id uint64) (*models.DirectAccessTokenPayload, error) {
	token, err := r.uc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.DirectAccessTokenPayload{
		ClientMutationID: requestid.Get(ctx),
		Token:            models.FromDirectAccessToken(token),
	}, nil
}

// List is the resolver for the listDirectAccessTokens field.
func (r *QueryResolver) List(ctx context.Context, filter *models.DirectAccessTokenListFilter, order *models.DirectAccessTokenListOrder, page *models.Page) (*connectors.CollectionConnection[models.DirectAccessToken, models.DirectAccessTokenEdge], error) {
	return connectors.NewDirectAccessTokenConnection(ctx, r.uc, filter, order, page,
		func(dat *model.DirectAccessToken) *model.DirectAccessToken {
			if dat.CreatedAt.After(time.Now().Add(-time.Minute * 5)) {
				return dat
			}
			m := new(model.DirectAccessToken)
			*m = *dat
			m.Token = strings.Repeat("*", len(m.Token))
			return m
		}), nil
}
