package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.73

import (
	"context"

	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, login string, password string) (*models.SessionToken, error) {
	return r.Resolver.Mutation().Login(ctx, login, password)
}

// Logout is the resolver for the logout field.
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	return r.Resolver.Mutation().Logout(ctx)
}

// SwitchAccount is the resolver for the switchAccount field.
func (r *mutationResolver) SwitchAccount(ctx context.Context, id uint64) (*models.SessionToken, error) {
	return r.Resolver.Mutation().SwitchAccount(ctx, id)
}

// RegisterAccount is the resolver for the registerAccount field.
func (r *mutationResolver) RegisterAccount(ctx context.Context, input models.AccountCreateInput) (*models.AccountCreatePayload, error) {
	return r.Resolver.Mutation().RegisterAccount(ctx, input)
}

// UpdateAccount is the resolver for the updateAccount field.
func (r *mutationResolver) UpdateAccount(ctx context.Context, id uint64, input models.AccountInput) (*models.AccountPayload, error) {
	return r.Resolver.Mutation().UpdateAccount(ctx, id, input)
}

// ApproveAccount is the resolver for the approveAccount field.
func (r *mutationResolver) ApproveAccount(ctx context.Context, id uint64, msg string) (*models.AccountPayload, error) {
	return r.Resolver.Mutation().ApproveAccount(ctx, id, msg)
}

// RejectAccount is the resolver for the rejectAccount field.
func (r *mutationResolver) RejectAccount(ctx context.Context, id uint64, msg string) (*models.AccountPayload, error) {
	return r.Resolver.Mutation().RejectAccount(ctx, id, msg)
}

// CurrentSession is the resolver for the currentSession field.
func (r *queryResolver) CurrentSession(ctx context.Context) (*models.SessionToken, error) {
	return r.Resolver.Query().CurrentSession(ctx)
}

// CurrentAccount is the resolver for the currentAccount field.
func (r *queryResolver) CurrentAccount(ctx context.Context) (*models.AccountPayload, error) {
	return r.Resolver.Query().CurrentAccount(ctx)
}

// Account is the resolver for the account field.
func (r *queryResolver) Account(ctx context.Context, id uint64) (*models.AccountPayload, error) {
	return r.Resolver.Query().Account(ctx, id)
}

// ListAccounts is the resolver for the listAccounts field.
func (r *queryResolver) ListAccounts(ctx context.Context, filter *models.AccountListFilter, order *models.AccountListOrder, page *models.Page) (*connectors.CollectionConnection[models.Account, models.AccountEdge], error) {
	return r.Resolver.Query().ListAccounts(ctx, filter, order, page)
}

// ListAccountRolesAndPermissions is the resolver for the listAccountRolesAndPermissions field.
func (r *queryResolver) ListAccountRolesAndPermissions(ctx context.Context, accountID uint64, order *models.RBACRoleListOrder) (*connectors.CollectionConnection[models.RBACRole, models.RBACRoleEdge], error) {
	return r.Resolver.Query().ListAccountRolesAndPermissions(ctx, accountID, order)
}
