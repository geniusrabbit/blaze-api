package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"

	"github.com/geniusrabbit/blaze-api/context/version"
	"github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/generated"
)

// Poke is the resolver for the poke field.
func (r *mutationResolver) Poke(ctx context.Context) (string, error) {
	return "hi", nil
}

// ServiceVersion is the resolver for the serviceVersion field.
func (r *queryResolver) ServiceVersion(ctx context.Context) (string, error) {
	return version.Get(ctx).Public(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
