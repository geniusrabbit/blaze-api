package graphql

import (
	"context"

	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/models"
	authclientrepo "github.com/geniusrabbit/blaze-api/repository/authclient/repository"
	authclientusecase "github.com/geniusrabbit/blaze-api/repository/authclient/usecase"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	gqlmodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// QueryResolver implements GQL API methods
type QueryResolver struct {
	authClients authclient.Usecase
}

// NewQueryResolver returns new API resolver
func NewQueryResolver(uc authclient.Usecase) *QueryResolver {
	return &QueryResolver{authClients: uc}
}

// NewDefaultQueryResolver returns new API resolver with default usecase
func NewDefaultQueryResolver() *QueryResolver {
	return &QueryResolver{
		authClients: authclientusecase.NewAuthclientUsecase(
			authclientrepo.NewAuthclientRepository(),
		),
	}
}

// AuthClient is the resolver for the authClient field.
func (r *QueryResolver) AuthClient(ctx context.Context, id string) (*gqlmodels.AuthClientPayload, error) {
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     client.ID,
		AuthClient:       FromAuthClientModel(client),
	}, nil
}

// ListAuthClients is the resolver for the listAuthClients field.
func (r *QueryResolver) ListAuthClients(ctx context.Context,
	filter *gqlmodels.AuthClientListFilter,
	orders []*gqlmodels.AuthClientListOrder,
	page *gqlmodels.Page,
) (*AuthClientConnection, error) {
	return NewAuthClientConnection(ctx, r.authClients, page), nil
}

// CreateAuthClient is the resolver for the createAuthClient field.
func (r *QueryResolver) CreateAuthClient(ctx context.Context, input *gqlmodels.AuthClientCreateInput) (*gqlmodels.AuthClientPayload, error) {
	// Create and fill model from input
	clientObj := CreateFillModel(input, &models.AuthClient{})

	id, err := r.authClients.Create(ctx, clientObj, historylog.Message("GQL create authclient"))
	if err != nil {
		return nil, err
	}
	client, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gqlmodels.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     client.ID,
		AuthClient:       FromAuthClientModel(client),
	}, nil
}

// UpdateAuthClient is the resolver for the updateAuthClient field.
func (r *QueryResolver) UpdateAuthClient(ctx context.Context, id string, input *gqlmodels.AuthClientUpdateInput) (*gqlmodels.AuthClientPayload, error) {
	clientObj, err := r.authClients.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// Update client fields
	UpdateFillModel(input, clientObj)

	if err = r.authClients.Update(ctx, id, clientObj, historylog.Message("GQL update authclient")); err != nil {
		return nil, err
	}
	return &gqlmodels.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     id,
		AuthClient:       FromAuthClientModel(clientObj),
	}, nil
}

// DeleteAuthClient is the resolver for the deleteAuthClient field.
func (r *QueryResolver) DeleteAuthClient(ctx context.Context, id string, msg *string) (*gqlmodels.AuthClientPayload, error) {
	if err := r.authClients.Delete(ctx, id, historylog.Message("GQL delete authclient")); err != nil {
		return nil, err
	}
	return &gqlmodels.AuthClientPayload{
		ClientMutationID: requestid.Get(ctx),
		AuthClientID:     id,
	}, nil
}
