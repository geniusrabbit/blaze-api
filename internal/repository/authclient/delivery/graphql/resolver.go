package graphql

import (
	"context"

	"github.com/geniusrabbit/api-template-base/internal/repository/authclient"
	"github.com/geniusrabbit/api-template-base/internal/repository/authclient/repository"
	"github.com/geniusrabbit/api-template-base/internal/repository/authclient/usecase"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/connectors"
	"github.com/geniusrabbit/api-template-base/internal/server/graphql/models"
	"github.com/geniusrabbit/api-template-base/model"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	authClients authclient.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver() *QueryResolver {
	return &QueryResolver{
		authClients: usecase.NewAuthclientUsecase(repository.New()),
	}
}

// AuthClient is the resolver for the authClient field.
func (r *QueryResolver) AuthClient(ctx context.Context, id string) (*models.AuthClientPayload, error) {
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		AuthClientID: client.ID,
		AuthClient:   models.FromAuthClientModel(client),
	}, nil
}

// ListAuthClients is the resolver for the listAuthClients field.
func (r *QueryResolver) ListAuthClients(ctx context.Context,
	filter *models.AuthClientListFilter,
	order []*models.AuthClientListOrder,
	page *models.Page) (*connectors.AuthClientConnection, error) {
	return connectors.NewAuthClientConnection(ctx, r.authClients), nil
}

// CreateAuthClient is the resolver for the createAuthClient field.
func (r *QueryResolver) CreateAuthClient(ctx context.Context, input *models.AuthClientInput) (*models.AuthClientPayload, error) {
	id, err := r.authClients.Create(ctx, &model.AuthClient{
		UserID:             idFromPtr(input.UserID, 0),
		AccountID:          idFromPtr(input.AccountID, 0),
		Title:              input.Title,
		Secret:             input.Secret,
		RedirectURIs:       input.RedirectURIs,
		GrantTypes:         input.GrantTypes,
		ResponseTypes:      input.ResponseTypes,
		Scope:              input.Scope,
		Audience:           input.Audience,
		SubjectType:        input.SubjectType,
		AllowedCORSOrigins: input.AllowedCORSOrigins,
		Public:             input.Public,
		ExpiresAt:          input.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		AuthClientID: client.ID,
		AuthClient:   models.FromAuthClientModel(client),
	}, nil
}

// UpdateAuthClient is the resolver for the updateAuthClient field.
func (r *QueryResolver) UpdateAuthClient(ctx context.Context, id string, input *models.AuthClientInput) (*models.AuthClientPayload, error) {
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update client fields
	client.UserID = idFromPtr(input.UserID, client.UserID)
	client.AccountID = idFromPtr(input.AccountID, client.AccountID)
	client.Title = input.Title
	client.Secret = input.Secret
	client.RedirectURIs = input.RedirectURIs
	client.GrantTypes = input.GrantTypes
	client.ResponseTypes = input.ResponseTypes
	client.Scope = input.Scope
	client.Audience = input.Audience
	client.SubjectType = input.SubjectType
	client.AllowedCORSOrigins = input.AllowedCORSOrigins
	client.Public = input.Public
	client.ExpiresAt = input.ExpiresAt

	if err = r.authClients.Update(ctx, id, client); err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		AuthClientID: client.ID,
		AuthClient:   models.FromAuthClientModel(client),
	}, nil
}

// DeleteAuthClient is the resolver for the deleteAuthClient field.
func (r *QueryResolver) DeleteAuthClient(ctx context.Context, id string, msg *string) (*models.AuthClientPayload, error) {
	if err := r.authClients.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		AuthClientID: id,
	}, nil
}

func idFromPtr(id *int, def uint64) uint64 {
	if id == nil {
		return def
	}
	return uint64(*id)
}
