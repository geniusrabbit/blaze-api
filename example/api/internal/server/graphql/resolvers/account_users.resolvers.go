package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input models.UserInput) (*models.UserPayload, error) {
	return r.Resolver.Mutation().CreateUser(ctx, input)
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id uint64, input models.UserInput) (*models.UserPayload, error) {
	return r.Resolver.Mutation().UpdateUser(ctx, id, input)
}

// ApproveUser is the resolver for the approveUser field.
func (r *mutationResolver) ApproveUser(ctx context.Context, id uint64, msg *string) (*models.UserPayload, error) {
	return r.Resolver.Mutation().ApproveUser(ctx, id, msg)
}

// RejectUser is the resolver for the rejectUser field.
func (r *mutationResolver) RejectUser(ctx context.Context, id uint64, msg *string) (*models.UserPayload, error) {
	return r.Resolver.Mutation().RejectUser(ctx, id, msg)
}

// ResetUserPassword is the resolver for the resetUserPassword field.
func (r *mutationResolver) ResetUserPassword(ctx context.Context, email string) (*models.StatusResponse, error) {
	return r.Resolver.Mutation().ResetUserPassword(ctx, email)
}

// UpdateUserPassword is the resolver for the updateUserPassword field.
func (r *mutationResolver) UpdateUserPassword(ctx context.Context, token string, email string, password string) (*models.StatusResponse, error) {
	return r.Resolver.Mutation().UpdateUserPassword(ctx, token, email, password)
}

// CurrentUser is the resolver for the currentUser field.
func (r *queryResolver) CurrentUser(ctx context.Context) (*models.UserPayload, error) {
	return r.Resolver.Query().CurrentUser(ctx)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id uint64, username string) (*models.UserPayload, error) {
	return r.Resolver.Query().User(ctx, id, username)
}

// ListUsers is the resolver for the listUsers field.
func (r *queryResolver) ListUsers(ctx context.Context, filter *models.UserListFilter, order *models.UserListOrder, page *models.Page) (*connectors.CollectionConnection[models.User, models.UserEdge], error) {
	return r.Resolver.Query().ListUsers(ctx, filter, order, page)
}
