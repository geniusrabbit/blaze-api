package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.43

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// CreateRole is the resolver for the createRole field.
func (r *mutationResolver) CreateRole(ctx context.Context, input models.RBACRoleInput) (*models.RBACRolePayload, error) {
	return r.Resolver.Mutation().CreateRole(ctx, input)
}

// UpdateRole is the resolver for the updateRole field.
func (r *mutationResolver) UpdateRole(ctx context.Context, id uint64, input models.RBACRoleInput) (*models.RBACRolePayload, error) {
	return r.Resolver.Mutation().UpdateRole(ctx, id, input)
}

// DeleteRole is the resolver for the deleteRole field.
func (r *mutationResolver) DeleteRole(ctx context.Context, id uint64, msg *string) (*models.RBACRolePayload, error) {
	return r.Resolver.Mutation().DeleteRole(ctx, id, msg)
}

// Role is the resolver for the role field.
func (r *queryResolver) Role(ctx context.Context, id uint64) (*models.RBACRolePayload, error) {
	return r.Resolver.Query().Role(ctx, id)
}

// CheckPermission is the resolver for the checkPermission field.
func (r *queryResolver) CheckPermission(ctx context.Context, name string, key *string, targetID *string) (*string, error) {
	return r.Resolver.Query().CheckPermission(ctx, name, key, targetID)
}

// ListRoles is the resolver for the listRoles field.
func (r *queryResolver) ListRoles(ctx context.Context, filter *models.RBACRoleListFilter, order *models.RBACRoleListOrder, page *models.Page) (*connectors.CollectionConnection[models.RBACRole, models.RBACRoleEdge], error) {
	return r.Resolver.Query().ListRoles(ctx, filter, order, page)
}
