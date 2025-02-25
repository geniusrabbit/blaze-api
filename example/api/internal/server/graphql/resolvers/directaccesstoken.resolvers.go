package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"
	"fmt"
	"time"

	models1 "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// GenerateDirectAccessToken is the resolver for the generateDirectAccessToken field.
func (r *mutationResolver) GenerateDirectAccessToken(ctx context.Context, userID *uint64, description string, expiresAt *time.Time) (*models.DirectAccessTokenPayload, error) {
	panic(fmt.Errorf("not implemented: GenerateDirectAccessToken - generateDirectAccessToken"))
}

// RevokeDirectAccessToken is the resolver for the revokeDirectAccessToken field.
func (r *mutationResolver) RevokeDirectAccessToken(ctx context.Context, filter models.DirectAccessTokenListFilter) (*models.StatusResponse, error) {
	panic(fmt.Errorf("not implemented: RevokeDirectAccessToken - revokeDirectAccessToken"))
}

// GetDirectAccessToken is the resolver for the getDirectAccessToken field.
func (r *queryResolver) GetDirectAccessToken(ctx context.Context, id uint64) (*models.DirectAccessTokenPayload, error) {
	panic(fmt.Errorf("not implemented: GetDirectAccessToken - getDirectAccessToken"))
}

// ListDirectAccessTokens is the resolver for the listDirectAccessTokens field.
func (r *queryResolver) ListDirectAccessTokens(ctx context.Context, filter *models.DirectAccessTokenListFilter, order *models.DirectAccessTokenListOrder, page *models.Page) (*models1.DirectAccessTokenConnection, error) {
	panic(fmt.Errorf("not implemented: ListDirectAccessTokens - listDirectAccessTokens"))
}
