package graphql

import (
	"context"
	"time"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/repository"
	"github.com/geniusrabbit/blaze-api/repository/authclient/usecase"
	"github.com/geniusrabbit/blaze-api/requestid"
	"github.com/geniusrabbit/blaze-api/server/graphql/connectors"
	"github.com/geniusrabbit/blaze-api/server/graphql/models"
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
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     client.ID,
		AuthClient:       models.FromAuthClientModel(client),
	}, nil
}

// ListAuthClients is the resolver for the listAuthClients field.
func (r *QueryResolver) ListAuthClients(ctx context.Context,
	filter *models.AuthClientListFilter,
	order *models.AuthClientListOrder,
	page *models.Page) (*connectors.AuthClientConnection, error) {
	return connectors.NewAuthClientConnection(ctx, r.authClients, page), nil
}

// CreateAuthClient is the resolver for the createAuthClient field.
func (r *QueryResolver) CreateAuthClient(ctx context.Context, input *models.AuthClientInput) (*models.AuthClientPayload, error) {
	id, err := r.authClients.Create(ctx, &model.AuthClient{
		UserID:             idFromPtr(input.UserID, 0),
		AccountID:          idFromPtr(input.AccountID, 0),
		Title:              valOrDef(input.Title, ""),
		Secret:             valOrDef(input.Secret, ""),
		RedirectURIs:       input.RedirectURIs,
		GrantTypes:         input.GrantTypes,
		ResponseTypes:      input.ResponseTypes,
		Scope:              valOrDef(input.Scope, ""),
		Audience:           input.Audience,
		SubjectType:        input.SubjectType,
		AllowedCORSOrigins: input.AllowedCORSOrigins,
		Public:             valOrDef(input.Public, false),
		ExpiresAt:          valOrDef(input.ExpiresAt, time.Time{}),
	})
	if err != nil {
		return nil, err
	}
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     client.ID,
		AuthClient:       models.FromAuthClientModel(client),
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
	client.Title = valOrDef(input.Title, client.Title)
	client.Secret = valOrDef(input.Secret, client.Secret)
	client.RedirectURIs = input.RedirectURIs
	client.GrantTypes = input.GrantTypes
	client.ResponseTypes = input.ResponseTypes
	client.Scope = valOrDef(input.Scope, client.Scope)
	client.Audience = input.Audience
	client.SubjectType = input.SubjectType
	client.AllowedCORSOrigins = input.AllowedCORSOrigins
	client.Public = valOrDef(input.Public, client.Public)
	client.ExpiresAt = valOrDef(input.ExpiresAt, client.ExpiresAt)

	if err = r.authClients.Update(ctx, id, client); err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     client.ID,
		AuthClient:       models.FromAuthClientModel(client),
	}, nil
}

// DeleteAuthClient is the resolver for the deleteAuthClient field.
func (r *QueryResolver) DeleteAuthClient(ctx context.Context, id string, msg *string) (*models.AuthClientPayload, error) {
	if err := r.authClients.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &models.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     id,
	}, nil
}

func valOrDef[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

func idFromPtr(id *uint64, def uint64) uint64 {
	if id == nil {
		return def
	}
	return uint64(*id)
}
